package gobatis

import (
	"database/sql"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"reflect"
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

type Connection struct {
	// logger 用于打印执行的sql
	logger *log.Logger
	// showSQL 显示执行的sql，用于调试，使用logger打印
	showSQL bool

	dbType        Dialect
	db            dbRunner
	sqlStatements map[string]*MappedStatement
	mapper        *Mapper
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

func (conn *Connection) DbType() Dialect {
	return conn.dbType
}

func (conn *Connection) Insert(id string, paramNames []string, paramValues []interface{}, notReturn ...bool) (int64, error) {
	sqlStr, sqlParams, _, err := conn.readSQLParams(id, StatementTypeInsert, paramNames, paramValues)
	if err != nil {
		return 0, err
	}

	if conn.showSQL {
		conn.logger.Printf(`id:"%s", sql:"%s", params:"%+v"`, id, sqlStr, sqlParams)
	}

	if len(notReturn) > 0 && notReturn[0] {
		_, err := conn.db.Exec(sqlStr, sqlParams...)
		return 0, err
	}

	if conn.dbType != DbTypePostgres && conn.dbType != DbTypeMSSql {
		result, err := conn.db.Exec(sqlStr, sqlParams...)
		if err != nil {
			return 0, err
		}
		return result.LastInsertId()
	}

	var insertID int64
	err = conn.db.QueryRow(sqlStr, sqlParams...).Scan(&insertID)
	if err != nil {
		return 0, err
	}
	return insertID, nil
}

func (conn *Connection) Update(id string, paramNames []string, paramValues []interface{}) (int64, error) {
	sqlStr, sqlParams, _, err := conn.readSQLParams(id, StatementTypeUpdate, paramNames, paramValues)
	if err != nil {
		return 0, err
	}

	if conn.showSQL {
		conn.logger.Printf(`id:"%s", sql:"%s", params:"%+v"`, id, sqlStr, sqlParams)
	}

	result, err := conn.db.Exec(sqlStr, sqlParams...)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected()
}

func (conn *Connection) Delete(id string, paramNames []string, paramValues []interface{}) (int64, error) {
	sqlStr, sqlParams, _, err := conn.readSQLParams(id, StatementTypeDelete, paramNames, paramValues)
	if err != nil {
		return 0, err
	}

	if conn.showSQL {
		conn.logger.Printf(`id:"%s", sql:"%s", params:"%+v"`, id, sqlStr, sqlParams)
	}

	result, err := conn.db.Exec(sqlStr, sqlParams...)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected()
}

func (conn *Connection) SelectOne(id string, paramNames []string, paramValues []interface{}) Result {
	sql, sqlParams, _, err := conn.readSQLParams(id, StatementTypeSelect, paramNames, paramValues)

	return Result{o: conn,
		id:        id,
		sql:       sql,
		sqlParams: sqlParams,
		err:       err}
}

func (conn *Connection) Select(id string, paramNames []string, paramValues []interface{}) *Results {
	sql, sqlParams, _, err := conn.readSQLParams(id, StatementTypeSelect, paramNames, paramValues)
	return &Results{o: conn,
		id:        id,
		sql:       sql,
		sqlParams: sqlParams,
		err:       err}
}

func (o *Connection) readSQLParams(id string, sqlType StatementType, paramNames []string, paramValues []interface{}) (sql string, sqlParams []interface{}, rType ResultType, err error) {
	sqlParams = make([]interface{}, 0)
	stmt, ok := o.sqlStatements[id]
	err = nil

	if !ok {
		err = fmt.Errorf("sql '%s' error : statement not found ", id)
		return
	}
	rType = stmt.result

	if stmt.sqlType != sqlType {
		err = fmt.Errorf("sql '%s' error : Select type Error, excepted is %s, actual is %s",
			id, sqlType.String(), stmt.sqlType.String())
		return
	}

	if stmt.sqlTemplate == nil {
		if stmt.sqlCompiled == nil {
			sql = stmt.sql
			if len(paramNames) != 0 {
				// NOTE: 这里需要调用 toSQLType 么？
				sqlParams = paramValues
			}
			return
		}

		sql = o.dbType.Placeholder().Get(stmt.sqlCompiled)

		sqlParams, err = bindNamedQuery(stmt.sqlCompiled.bindParams, paramNames, paramValues, o.dbType, o.mapper)
		if err != nil {
			err = fmt.Errorf("sql '%s' error : %s", id, err)
		}
		return
	}

	var tplArgs interface{}
	if len(paramNames) == 0 {
		if len(paramValues) == 0 {
			err = fmt.Errorf("sql '%s' error : arguments is missing", id)
			return
		}
		if len(paramValues) > 1 {
			err = fmt.Errorf("sql '%s' error : arguments is exceed 1", id)
			return
		}

		tplArgs = paramValues[0]
	} else if len(paramNames) == 1 {
		tplArgs = paramValues[0]
		if _, ok := tplArgs.(map[string]interface{}); !ok {
			paramType := reflect.TypeOf(tplArgs)
			if paramType.Kind() == reflect.Ptr {
				paramType = paramType.Elem()
			}
			if paramType.Kind() != reflect.Struct {
				tplArgs = map[string]interface{}{paramNames[0]: paramValues[0]}
			}
		}
	} else {
		var args = map[string]interface{}{}
		for idx := range paramNames {
			args[paramNames[idx]] = paramValues[idx]
		}
		tplArgs = args
	}

	var sb strings.Builder
	err = stmt.sqlTemplate.Execute(&sb, tplArgs)
	if err != nil {
		err = fmt.Errorf("1sql '%s' error : %s", id, err)
		return
	}
	sql = sb.String()

	fragments, nameArgs, e := compileNamedQuery(sql)
	if e != nil {
		err = fmt.Errorf("2sql '%s' error : %s", id, e)
		return
	}
	if len(nameArgs) == 0 {
		return
	}

	sql = o.dbType.Placeholder().Concat(fragments, nameArgs)
	sqlParams, err = bindNamedQuery(nameArgs, paramNames, paramValues, o.dbType, o.mapper)
	if err != nil {
		err = fmt.Errorf("3sql '%s' error : %s", id, err)
	}
	return
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
	for k, v := range templateFuncs {
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
		logger:  cfg.Logger,
		showSQL: cfg.ShowSQL,
		dbType:  ToDbType(cfg.DriverName),
	}
	if base.dbType == DbTypeNone {
		base.dbType = DbTypePostgres
	}
	base.db = cfg.DB
	base.sqlStatements = make(map[string]*MappedStatement)
	var tagPrefix string
	var tagMapper func(string) []string
	if cfg != nil {
		base.isUnsafe = cfg.IsUnsafe
		tagPrefix = cfg.TagPrefix
		tagMapper = cfg.TagMapper
	}
	base.mapper = CreateMapper(tagPrefix, nil, tagMapper)

	dbName := strings.ToLower(base.DbType().Name())
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
		DbType:     base.dbType,
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
