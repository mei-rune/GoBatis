package gobatis

import (
	"context"
	"database/sql"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
	"text/template"
)

type Config struct {
	Logger  *log.Logger
	ShowSQL bool

	// DB 和后3个参数任选一个
	DriverName   string
	DB           dbRunner
	DataSource   string
	MaxIdleConns int
	MaxOpenConns int

	XMLPaths      []string
	IsUnsafe      bool
	TagPrefix     string
	TagMapper     func(s string) []string
	TemplateFuncs template.FuncMap
}

type dbRunner interface {
	Prepare(query string) (*sql.Stmt, error)
	ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error)
	QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error)
	QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row
}

type Connection struct {
	// logger 用于打印执行的sql
	logger *log.Logger
	// showSQL 显示执行的sql，用于调试，使用logger打印
	showSQL bool

	dialect       Dialect
	mapper        *Mapper
	db            dbRunner
	sqlStatements map[string]*MappedStatement
	isUnsafe      bool
}

func (conn *Connection) DB() dbRunner {
	return conn.db
}

func (conn *Connection) WithDB(db dbRunner) *Connection {
	newConn := &Connection{}
	*newConn = *conn
	newConn.db = db
	return newConn
}

func (conn *Connection) SetDB(db dbRunner) {
	conn.db = db
}

func (conn *Connection) Dialect() Dialect {
	return conn.dialect
}

func (conn *Connection) Insert(ctx context.Context, id string, paramNames []string, paramValues []interface{}, notReturn ...bool) (int64, error) {
	sqlStr, sqlParams, _, err := conn.readSQLParams(id, StatementTypeInsert, paramNames, paramValues)
	if err != nil {
		return 0, err
	}

	if conn.showSQL {
		conn.logger.Printf(`id:"%s", sql:"%s", params:"%+v"`, id, sqlStr, sqlParams)
	}

	if len(notReturn) > 0 && notReturn[0] {
		_, err := conn.db.ExecContext(ctx, sqlStr, sqlParams...)
		return 0, err
	}

	if conn.dialect.InsertIDSupported() {
		result, err := conn.db.ExecContext(ctx, sqlStr, sqlParams...)
		if err != nil {
			return 0, err
		}
		return result.LastInsertId()
	}

	var insertID int64
	err = conn.db.QueryRowContext(ctx, sqlStr, sqlParams...).Scan(&insertID)
	if err != nil {
		return 0, err
	}
	return insertID, nil
}

func (conn *Connection) Update(ctx context.Context, id string, paramNames []string, paramValues []interface{}) (int64, error) {
	sqlStr, sqlParams, _, err := conn.readSQLParams(id, StatementTypeUpdate, paramNames, paramValues)
	if err != nil {
		return 0, err
	}

	if conn.showSQL {
		conn.logger.Printf(`id:"%s", sql:"%s", params:"%+v"`, id, sqlStr, sqlParams)
	}

	result, err := conn.db.ExecContext(ctx, sqlStr, sqlParams...)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected()
}

func (conn *Connection) Delete(ctx context.Context, id string, paramNames []string, paramValues []interface{}) (int64, error) {
	sqlStr, sqlParams, _, err := conn.readSQLParams(id, StatementTypeDelete, paramNames, paramValues)
	if err != nil {
		return 0, err
	}

	if conn.showSQL {
		conn.logger.Printf(`id:"%s", sql:"%s", params:"%+v"`, id, sqlStr, sqlParams)
	}

	result, err := conn.db.ExecContext(ctx, sqlStr, sqlParams...)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected()
}

func (conn *Connection) SelectOne(ctx context.Context, id string, paramNames []string, paramValues []interface{}) Result {
	sql, sqlParams, _, err := conn.readSQLParams(id, StatementTypeSelect, paramNames, paramValues)

	return Result{o: conn,
		ctx:       ctx,
		id:        id,
		sql:       sql,
		sqlParams: sqlParams,
		err:       err}
}

func (conn *Connection) Select(ctx context.Context, id string, paramNames []string, paramValues []interface{}) *Results {
	sql, sqlParams, _, err := conn.readSQLParams(id, StatementTypeSelect, paramNames, paramValues)
	return &Results{o: conn,
		ctx:       ctx,
		id:        id,
		sql:       sql,
		sqlParams: sqlParams,
		err:       err}
}

func (o *Connection) readSQLParams(id string, sqlType StatementType, paramNames []string, paramValues []interface{}) (string, []interface{}, ResultType, error) {
	stmt, ok := o.sqlStatements[id]
	if !ok {
		return "", nil, ResultUnknown, fmt.Errorf("sql '%s' error : statement not found ", id)
	}

	if stmt.sqlType != sqlType {
		return "", nil, ResultUnknown, fmt.Errorf("sql '%s' error : Select type Error, excepted is %s, actual is %s",
			id, sqlType.String(), stmt.sqlType.String())
	}

	ctx, err := NewContext(o.dialect, o.mapper, paramNames, paramValues)
	if err != nil {
		return "", nil, ResultUnknown, fmt.Errorf("sql '%s' error : %s", id, err)
	}

	sql, sqlParams, err := stmt.GenerateSQL(ctx)
	if err != nil {
		return "", nil, ResultUnknown, fmt.Errorf("sql '%s' error : %s", id, err)
	}
	return sql, sqlParams, stmt.result, nil
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
		cfg.Logger = log.New(os.Stdout, "[gobatis] ", log.Flags())
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
		db:            cfg.DB,
		sqlStatements: make(map[string]*MappedStatement),
	}
	var tagPrefix string
	var tagMapper func(string) []string
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
	return base, nil
}
