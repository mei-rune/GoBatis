package gobatis

import (
	"context"
	"database/sql"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"text/template"
)

type Tracer interface {
	Write(ctx context.Context, id, sql string, args []interface{}, err error)
}

type NullTracer struct{}

func (w NullTracer) Write(ctx context.Context, id, sql string, args []interface{}, err error) {}

type StdLogger struct {
	Logger *log.Logger
}

func (w StdLogger) Write(ctx context.Context, id, sql string, args []interface{}, err error) {
	if err != nil {
		w.Logger.Printf(`id:"%s", sql:"%s", params:"%#v", err:%q`, id, sql, args, err)
	} else {
		w.Logger.Printf(`id:"%s", sql:"%s", params:"%#v", err: null`, id, sql, args)
	}
}

type TraceWriter struct {
	Output io.Writer
}

func (w TraceWriter) Write(ctx context.Context, id, sql string, args []interface{}, err error) {
	if err != nil {
		fmt.Fprintf(w.Output, `id:"%s", sql:"%s", params:"%#v", err:%q`, id, sql, args, err)
	} else {
		fmt.Fprintf(w.Output, `id:"%s", sql:"%s", params:"%#v", err: null`, id, sql, args)
	}
}

type Config struct {
	Tracer          Tracer
	EnabledSQLCheck bool
	Constants       map[string]interface{}

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
	PrepareContext(ctx context.Context, query string) (*sql.Stmt, error)
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
	if ctx == nil {
		return nil
	}
	v := ctx.Value(txKey)
	if v == nil {
		return nil
	}
	return v.(DBRunner)
}

type Connection struct {
	// logger 用于打印执行的sql
	tracer Tracer

	constants     map[string]interface{}
	dialect       Dialect
	mapper        *Mapper
	db            DBRunner
	sqlStatements map[string]*MappedStatement
	isUnsafe      bool
}

func (conn *Connection) SqlStatements() [][2]string {
	var sqlStatements = make([][2]string, 0, len(conn.sqlStatements))
	for id, stmt := range conn.sqlStatements {
		sqlStatements = append(sqlStatements, [2]string{id, stmt.rawSQL})
	}

	sort.Slice(sqlStatements, func(i, j int) bool {
		return sqlStatements[i][0] < sqlStatements[j][0]
	})
	return sqlStatements
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
	sqlAndParams, _, err := conn.readSQLParams(ctx, id, StatementTypeInsert, paramNames, paramValues)
	if err != nil {
		return 0, err
	}

	tx := DbConnectionFromContext(ctx)
	if tx == nil {
		tx = conn.db
	}

	for idx := 0; idx < len(sqlAndParams)-1; idx++ {
		_, err := tx.ExecContext(ctx, sqlAndParams[idx].SQL, sqlAndParams[idx].Params...)
		conn.tracer.Write(ctx, id, sqlAndParams[idx].SQL, sqlAndParams[idx].Params, err)
		if err != nil {
			return 0, conn.dialect.HandleError(err)
		}
	}

	sqlStr := sqlAndParams[len(sqlAndParams)-1].SQL
	sqlParams := sqlAndParams[len(sqlAndParams)-1].Params

	if len(notReturn) > 0 && notReturn[0] {
		_, err := tx.ExecContext(ctx, sqlStr, sqlParams...)
		conn.tracer.Write(ctx, id, sqlStr, sqlParams, err)
		return 0, conn.dialect.HandleError(err)
	}

	if conn.dialect.InsertIDSupported() {
		result, err := tx.ExecContext(ctx, sqlStr, sqlParams...)
		if err != nil {
			conn.tracer.Write(ctx, id, sqlStr, sqlParams, err)
			return 0, conn.dialect.HandleError(err)
		}
		insertID, err := result.LastInsertId()
		conn.tracer.Write(ctx, id, sqlStr, sqlParams, err)
		if err != nil {
			err = conn.dialect.HandleError(err)
		}
		return insertID, err
	}

	var insertID int64
	err = tx.QueryRowContext(ctx, sqlStr, sqlParams...).Scan(&insertID)
	conn.tracer.Write(ctx, id, sqlStr, sqlParams, err)
	if err != nil {
		return 0, conn.dialect.HandleError(err)
	}
	return insertID, nil
}

func (conn *Connection) Update(ctx context.Context, id string, paramNames []string, paramValues []interface{}) (int64, error) {
	sqlAndParams, _, err := conn.readSQLParams(ctx, id, StatementTypeUpdate, paramNames, paramValues)
	if err != nil {
		return 0, err
	}
	return conn.execute(ctx, id, sqlAndParams)
}

func (conn *Connection) Delete(ctx context.Context, id string, paramNames []string, paramValues []interface{}) (int64, error) {
	sqlAndParams, _, err := conn.readSQLParams(ctx, id, StatementTypeDelete, paramNames, paramValues)
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

		result, err := tx.ExecContext(ctx, sqlAndParams[idx].SQL, sqlAndParams[idx].Params...)
		if err != nil {
			conn.tracer.Write(ctx, id, sqlAndParams[idx].SQL, sqlAndParams[idx].Params, err)
			return 0, conn.dialect.HandleError(err)
		}

		affected, err := result.RowsAffected()
		conn.tracer.Write(ctx, id, sqlAndParams[idx].SQL, sqlAndParams[idx].Params, err)
		if err != nil {
			return 0, conn.dialect.HandleError(err)
		}
		rowsAffected += affected
	}
	return rowsAffected, nil
}

func (conn *Connection) SelectOne(ctx context.Context, id string, paramNames []string, paramValues []interface{}) Result {
	sqlAndParams, _, err := conn.readSQLParams(ctx, id, StatementTypeSelect, paramNames, paramValues)
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
	sqlAndParams, _, err := conn.readSQLParams(ctx, id, StatementTypeSelect, paramNames, paramValues)
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

func (o *Connection) readSQLParams(ctx context.Context, id string, sqlType StatementType, paramNames []string, paramValues []interface{}) ([]sqlAndParam, ResultType, error) {
	stmt, ok := o.sqlStatements[id]
	if !ok {
		return nil, ResultUnknown, fmt.Errorf("sql '%s' error : statement not found ", id)
	}

	if stmt.sqlType != sqlType {
		return nil, ResultUnknown, fmt.Errorf("sql '%s' error : Select type Error, excepted is %s, actual is %s",
			id, sqlType.String(), stmt.sqlType.String())
	}

	genCtx, err := NewContext(o.constants, o.dialect, o.mapper, paramNames, paramValues)
	if err != nil {
		return nil, ResultUnknown, fmt.Errorf("sql '%s' error : %s", id, err)
	}

	sqlAndParams, err := stmt.GenerateSQLs(genCtx)
	if err != nil {
		o.tracer.Write(ctx, id, stmt.rawSQL, nil, err)
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
	if cfg.Tracer == nil {
		cfg.Tracer = NullTracer{} // StdLogger{Logger: log.New(os.Stdout, "[gobatis] ", log.Flags())}
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
		tracer:        cfg.Tracer,
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

	return base, nil
}
