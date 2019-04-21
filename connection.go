package gobatis

import (
	"context"
	"database/sql"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"text/template"
)

type Logger interface {
	Println(message string)
	Write(id, sql string, args []interface{})
}

type StdLogger struct {
	Logger *log.Logger
}

func (w StdLogger) Println(message string) {
	w.Logger.Println(message)
}

func (w StdLogger) Write(id, sql string, args []interface{}) {
	w.Logger.Printf(`id:"%s", sql:"%s", params:"%#v"`, id, sql, args)
}

type Config struct {
	Logger            Logger
	ShowSQL           bool
	DumpSQLStatements bool
	Constants         map[string]interface{}

	// DB 和后3个参数任选一个
	DriverName   string
	DB           DBRunner
	DataSource   string
	MaxIdleConns int
	MaxOpenConns int

	XMLPaths      []string
	IsUnsafe      bool
	TagPrefix     string
	TagMapper     func(s string, fieldName string) []string
	TemplateFuncs template.FuncMap
}

type DBRunner interface {
	Prepare(query string) (*sql.Stmt, error)
	ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error)
	QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error)
	QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row
}

type txKeyType struct{ name string }

var txKey = txKeyType{name: "dbtx"}

func WithDbConnection(ctx context.Context, tx DBRunner) context.Context {
	return context.WithValue(ctx, txKey, tx)
}

func DbConnectionFromContext(ctx context.Context) DBRunner {
	v := ctx.Value(txKey)
	if v == nil {
		return nil
	}
	return v.(DBRunner)
}

type Connection struct {
	// logger 用于打印执行的sql
	logger Logger
	// showSQL 显示执行的sql，用于调试，使用logger打印
	showSQL bool

	constants     map[string]interface{}
	dialect       Dialect
	mapper        *Mapper
	db            DBRunner
	sqlStatements map[string]*MappedStatement
	isUnsafe      bool
}

func (conn *Connection) DB() DBRunner {
	return conn.db
}

func (conn *Connection) WithDB(db DBRunner) *Connection {
	newConn := &Connection{}
	*newConn = *conn
	newConn.db = db
	return newConn
}

func (conn *Connection) SetDB(db DBRunner) {
	conn.db = db
}

func (conn *Connection) Dialect() Dialect {
	return conn.dialect
}

func (conn *Connection) Mapper() *Mapper {
	return conn.mapper
}

func (conn *Connection) Insert(ctx context.Context, id string, paramNames []string, paramValues []interface{}, notReturn ...bool) (int64, error) {
	sqlAndParams, _, err := conn.readSQLParams(id, StatementTypeInsert, paramNames, paramValues)
	if err != nil {
		return 0, err
	}

	tx := DbConnectionFromContext(ctx)
	if tx == nil {
		tx = conn.db
	}

	for idx := 0; idx < len(sqlAndParams)-1; idx++ {
		if conn.showSQL {
			conn.logger.Write(id, sqlAndParams[idx].SQL, sqlAndParams[idx].Params)
		}

		_, err := tx.ExecContext(ctx, sqlAndParams[idx].SQL, sqlAndParams[idx].Params...)
		if err != nil {
			return 0, conn.dialect.HandleError(err)
		}
	}

	sqlStr := sqlAndParams[len(sqlAndParams)-1].SQL
	sqlParams := sqlAndParams[len(sqlAndParams)-1].Params

	if conn.showSQL {
		conn.logger.Write(id, sqlStr, sqlParams)
	}

	if len(notReturn) > 0 && notReturn[0] {
		_, err := tx.ExecContext(ctx, sqlStr, sqlParams...)
		return 0, conn.dialect.HandleError(err)
	}

	if conn.dialect.InsertIDSupported() {
		result, err := tx.ExecContext(ctx, sqlStr, sqlParams...)
		if err != nil {
			return 0, conn.dialect.HandleError(err)
		}
		id, err := result.LastInsertId()
		if err != nil {
			err = conn.dialect.HandleError(err)
		}
		return id, err
	}

	var insertID int64
	err = tx.QueryRowContext(ctx, sqlStr, sqlParams...).Scan(&insertID)
	if err != nil {
		return 0, conn.dialect.HandleError(err)
	}
	return insertID, nil
}

func (conn *Connection) Update(ctx context.Context, id string, paramNames []string, paramValues []interface{}) (int64, error) {
	sqlAndParams, _, err := conn.readSQLParams(id, StatementTypeUpdate, paramNames, paramValues)
	if err != nil {
		return 0, err
	}
	return conn.execute(ctx, id, sqlAndParams)
}

func (conn *Connection) Delete(ctx context.Context, id string, paramNames []string, paramValues []interface{}) (int64, error) {
	sqlAndParams, _, err := conn.readSQLParams(id, StatementTypeDelete, paramNames, paramValues)
	if err != nil {
		return 0, err
	}
	return conn.execute(ctx, id, sqlAndParams)
}

func (conn *Connection) execute(ctx context.Context, id string, sqlAndParams []sqlAndParam) (int64, error) {
	tx := DbConnectionFromContext(ctx)
	if tx == nil {
		tx = conn.db
	}

	rowsAffected := int64(0)
	for idx := range sqlAndParams {
		if conn.showSQL {
			conn.logger.Write(id, sqlAndParams[idx].SQL, sqlAndParams[idx].Params)
		}

		result, err := tx.ExecContext(ctx, sqlAndParams[idx].SQL, sqlAndParams[idx].Params...)
		if err != nil {
			return 0, conn.dialect.HandleError(err)
		}

		affected, err := result.RowsAffected()
		if err != nil {
			return 0, conn.dialect.HandleError(err)
		}
		rowsAffected += affected
	}
	return rowsAffected, nil
}

func (conn *Connection) SelectOne(ctx context.Context, id string, paramNames []string, paramValues []interface{}) Result {
	sqlAndParams, _, err := conn.readSQLParams(id, StatementTypeSelect, paramNames, paramValues)
	if err != nil {
		return Result{o: conn,
			ctx: ctx,
			id:  id,
			err: err,
		}
	}

	if len(sqlAndParams) > 1 {
		return Result{o: conn,
			ctx: ctx,
			id:  id,
			err: ErrMultSQL,
		}
	}
	return Result{o: conn,
		ctx:       ctx,
		id:        id,
		sql:       sqlAndParams[0].SQL,
		sqlParams: sqlAndParams[0].Params,
	}
}

func (conn *Connection) Select(ctx context.Context, id string, paramNames []string, paramValues []interface{}) *Results {
	sqlAndParams, _, err := conn.readSQLParams(id, StatementTypeSelect, paramNames, paramValues)
	if err != nil {
		return &Results{o: conn,
			ctx: ctx,
			id:  id,
			err: err,
		}
	}
	if len(sqlAndParams) > 1 {
		return &Results{o: conn,
			ctx: ctx,
			id:  id,
			err: ErrMultSQL,
		}
	}

	return &Results{o: conn,
		ctx:       ctx,
		id:        id,
		sql:       sqlAndParams[0].SQL,
		sqlParams: sqlAndParams[0].Params,
	}
}

func (o *Connection) readSQLParams(id string, sqlType StatementType, paramNames []string, paramValues []interface{}) ([]sqlAndParam, ResultType, error) {
	stmt, ok := o.sqlStatements[id]
	if !ok {
		return nil, ResultUnknown, fmt.Errorf("sql '%s' error : statement not found ", id)
	}

	if stmt.sqlType != sqlType {
		return nil, ResultUnknown, fmt.Errorf("sql '%s' error : Select type Error, excepted is %s, actual is %s",
			id, sqlType.String(), stmt.sqlType.String())
	}

	ctx, err := NewContext(o.constants, o.dialect, o.mapper, paramNames, paramValues)
	if err != nil {
		return nil, ResultUnknown, fmt.Errorf("sql '%s' error : %s", id, err)
	}

	sqlAndParams, err := stmt.GenerateSQLs(ctx)
	if err != nil {
		return nil, ResultUnknown, fmt.Errorf("sql '%s' error : %s", id, err)
	}
	return sqlAndParams, stmt.result, nil
}

// New 创建一个新的Osm，这个过程会打开数据库连接。
//
// cfg 是数据连接的参数，可以是0个1个或2个数字，第一个表示MaxIdleConns，第二个表示MaxOpenConns.
//
// 如：
//  o, err := gobatis.New(&gobatis.Config{DriverName: "mysql",
//         DataSource: "root:root@/51jczj?charset=utf8",
//         XMLPaths: []string{"test.xml"}})
func newConnection(cfg *Config) (*Connection, error) {
	if cfg.Logger == nil {
		cfg.Logger = StdLogger{Logger: log.New(os.Stdout, "[gobatis] ", log.Flags())}
	}
	if cfg.Constants == nil {
		cfg.Constants = map[string]interface{}{}
	}
	if cfg.TemplateFuncs == nil {
		cfg.TemplateFuncs = template.FuncMap{}
	}
	for k, v := range TemplateFuncs {
		cfg.TemplateFuncs[k] = v
	}

	if cfg.DB == nil {
		db, err := sql.Open(cfg.DriverName, cfg.DataSource)
		if err != nil {
			if db != nil {
				db.Close()
			}
			return nil, fmt.Errorf("create gobatis error : %s", err.Error())
		}

		if cfg != nil {
			if cfg.MaxIdleConns > 0 {
				db.SetMaxIdleConns(cfg.MaxIdleConns)
			}
			if cfg.MaxOpenConns > 0 {
				db.SetMaxOpenConns(cfg.MaxOpenConns)
			}
		}
		cfg.DB = db
	}

	base := &Connection{
		logger:        cfg.Logger,
		showSQL:       cfg.ShowSQL,
		constants:     cfg.Constants,
		db:            cfg.DB,
		sqlStatements: make(map[string]*MappedStatement),
	}
	var tagPrefix string
	var tagMapper func(string, string) []string
	if cfg != nil {
		base.isUnsafe = cfg.IsUnsafe
		tagPrefix = cfg.TagPrefix
		tagMapper = cfg.TagMapper
	}
	base.mapper = CreateMapper(tagPrefix, nil, tagMapper)
	base.dialect = ToDbType(cfg.DriverName)
	if base.dialect == DbTypeNone {
		base.dialect = DbTypePostgres
	}

	dbName := strings.ToLower(base.Dialect().Name())
	xmlPaths := []string{}
	for _, xmlPath := range cfg.XMLPaths {
		pathInfo, err := os.Stat(xmlPath)
		if err != nil {
			if os.IsNotExist(err) {
				continue
			}
			return nil, err
		}

		if !pathInfo.IsDir() {
			xmlPaths = append(xmlPaths, xmlPath)
			continue
		}

		fs, err := ioutil.ReadDir(xmlPath)
		if err != nil {
			return nil, err
		}

		for _, fileInfo := range fs {
			if !fileInfo.IsDir() {
				if fileName := fileInfo.Name(); strings.ToLower(filepath.Ext(fileName)) == ".xml" {
					xmlPaths = append(xmlPaths, filepath.Join(xmlPath, fileName))
				}
				continue
			}

			if dbName != strings.ToLower(fileInfo.Name()) {
				continue
			}

			dialectDirs, err := ioutil.ReadDir(filepath.Join(xmlPath, fileInfo.Name()))
			if err != nil {
				return nil, err
			}

			for _, dialectInfo := range dialectDirs {
				if fileName := dialectInfo.Name(); strings.ToLower(filepath.Ext(fileName)) == ".xml" {
					xmlPaths = append(xmlPaths, filepath.Join(xmlPath, fileInfo.Name(), fileName))
				}
			}
		}
	}

	ctx := &InitContext{Config: cfg,
		Logger:     cfg.Logger,
		Dialect:    base.dialect,
		Mapper:     base.mapper,
		Statements: base.sqlStatements}

	for _, xmlPath := range xmlPaths {
		log.Println("load xml -", xmlPath)
		statements, err := readMappedStatements(ctx, xmlPath)
		if err != nil {
			return nil, err
		}

		for _, sm := range statements {
			base.sqlStatements[sm.id] = sm
		}
	}

	if err := runInit(ctx); err != nil {
		return nil, err
	}

	if cfg.DumpSQLStatements || os.Getenv("gobatis_dump_statements") == "true" {
		var sqlStatements = make([][2]string, 0, len(base.sqlStatements))
		keyLen := 0
		for id, stmt := range base.sqlStatements {
			if len(id) > keyLen {
				keyLen = len(id)
			}
			sqlStatements = append(sqlStatements, [2]string{id, stmt.rawSQL})
		}

		sort.Slice(sqlStatements, func(i, j int) bool {
			return sqlStatements[i][0] < sqlStatements[j][0]
		})

		fmt.Println()
		fmt.Println(strings.Repeat("=", 2*keyLen))
		for idx := range sqlStatements {
			id, rawSQL := sqlStatements[idx][0], sqlStatements[idx][1]
			fmt.Println(id+strings.Repeat(" ", keyLen-len(id)), ":", rawSQL)
		}
		fmt.Println()
		fmt.Println()
	}
	return base, nil
}
