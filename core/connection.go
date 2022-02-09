package core

import (
	"context"
	"database/sql"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"text/template"

	"github.com/runner-mei/GoBatis/dialects"
)

type errTx struct {
	method string
	inner error
}

func (e errTx) Unwrap() error {
	return e.inner
}

func (e errTx) Error() string {
	return e.method + " tx fail: " + e.inner.Error()
}

func IsTxError(e error, method string, methods ...string) bool {
	txErr, ok := e.(errTx)
	if !ok {
		return false
	}
	if len(methods) == 0 {
		return txErr.method == method	
	}
	if txErr.method == method {
		return true
	}
	for _, m := range methods {
		if m == txErr.method {
			return true
		}
	}
	return false
}

var ErrAlreadyTx = errors.New("open tx fail: already in a tx")

type statementAlreadyExists struct {
	id string
}

func (e statementAlreadyExists) Error() string {
	return "statement '" + e.id + "' already exists"
}

func ErrStatementAlreadyExists(id string) error {
	return statementAlreadyExists{id: id}
}

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
		fmt.Fprintf(w.Output, "id:\"%s\", sql:\"%s\", params:\"%#v\", err:%q\r\n", id, sql, args, err)
	} else {
		fmt.Fprintf(w.Output, "id:\"%s\", sql:\"%s\", params:\"%#v\", err: null\r\n", id, sql, args)
	}
}

var (
	DiscardTracer = NullTracer{}
	Constants     = map[string]interface{}{}
)

type Config struct {
	Tracer          Tracer
	EnabledSQLCheck bool
	Constants       map[string]interface{}

	// DB 和后3个参数任选一个
	DriverName           string
	TransformDbOwnership bool
	DB                   DBRunner
	DataSource           string
	MaxIdleConns         int
	MaxOpenConns         int

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

type txKeyType struct{}

func (*txKeyType) String() string {
	return "gobatis-tx-key"
}

var txKey = &txKeyType{}

func WithTx(ctx context.Context, tx DBRunner) context.Context {
	if ctx == nil {
		ctx = context.Background()
	}
	return context.WithValue(ctx, txKey, tx)
}

func OpenTxWith(ctx context.Context, conn DBRunner, failIfInTx ...bool) (context.Context, *sql.Tx, error) {
	shouldFail := false
	if len(failIfInTx) > 0 {
		shouldFail = failIfInTx[0]
	}

	switch db := conn.(type) {
	case *sql.DB:
		tx, err := db.Begin()
		if err != nil {
			return ctx, nil, errTx{method: "begin", inner: err}
		}
		return WithTx(ctx, tx), tx, nil
	case *sql.Tx:
		if shouldFail {
			return WithTx(ctx, db), db, ErrAlreadyTx
		}
		return WithTx(ctx, db), db, nil
	default:
		return ctx, nil, fmt.Errorf("bad conn arguments: unknown type '%T'", conn)
	}
}

func TxFromContext(ctx context.Context) DBRunner {
	if ctx == nil {
		return nil
	}
	v := ctx.Value(txKey)
	if v == nil {
		return nil
	}
	return v.(DBRunner)
}

func InTx(ctx context.Context, db DBRunner, failIfInTx bool, cb func(ctx context.Context, tx DBRunner) error) (rerr error) {
	if tx := TxFromContext(ctx); tx != nil {
		return cb(ctx, tx)
	}

	hasTx := true
	ctxWithTx, tx, err := OpenTxWith(ctx, db, failIfInTx)
	if err != nil {
		if err != ErrAlreadyTx {
			return ErrAlreadyTx
		}
		hasTx = false
	} else {
		defer func() {
			err := tx.Rollback()
			if err != nil && err != sql.ErrTxDone {
				rerr = errTx{method: "rollback", inner: err}
			}
		}()
	}

	err = cb(ctxWithTx, tx)
	if err != nil {
		return err
	}

	if hasTx {
		if err = tx.Commit(); err != nil {
			return errTx{method: "commit", inner: err}
		}
	}

	return nil
}

type connection struct {
	// logger 用于打印执行的sql
	tracer Tracer

	constants     map[string]interface{}
	dialect       Dialect
	mapper        *Mapper
	dbOwner       bool
	db            DBRunner
	sqlStatements map[string]*MappedStatement
	isUnsafe      bool
}

func (conn *connection) Close() (err error) {
	if !conn.dbOwner {
		return
	}

	if conn.db == nil {
		err = fmt.Errorf("db no opened")
	} else {
		sqlDb, ok := conn.db.(*sql.DB)
		if ok {
			err = sqlDb.Close()
		} else {
			err = fmt.Errorf("db unknown")
		}
	}
	return
}

func (conn *connection) SqlStatements() [][2]string {
	var sqlStatements = make([][2]string, 0, len(conn.sqlStatements))
	for id, stmt := range conn.sqlStatements {
		sqlStatements = append(sqlStatements, [2]string{id, stmt.rawSQL})
	}

	sort.Slice(sqlStatements, func(i, j int) bool {
		return sqlStatements[i][0] < sqlStatements[j][0]
	})
	return sqlStatements
}

func (conn *connection) ToXML() (map[string]*xmlConfig, error) {
	var sqlStatements = map[string]*xmlConfig{}
	for id, stmt := range conn.sqlStatements {
		pos := strings.IndexByte(id, '.')
		filename := id

		if pos > 0 {
			filename = id[:pos]
		}
		cfg := sqlStatements[filename]
		if cfg == nil {
			cfg = &xmlConfig{}
			sqlStatements[filename] = cfg
		}

		xmlStmt := &stmtXML{
			ID:  id,
			SQL: stmt.rawSQL,
		}
		switch stmt.sqlType {
		case StatementTypeSelect:
			cfg.Selects = append(cfg.Selects, *xmlStmt)
		case StatementTypeUpdate:
			cfg.Updates = append(cfg.Updates, *xmlStmt)
		case StatementTypeInsert:
			cfg.Inserts = append(cfg.Inserts, *xmlStmt)
		case StatementTypeDelete:
			cfg.Deletes = append(cfg.Deletes, *xmlStmt)
		default:
			return nil, errors.New("statement '" + id + "' type is unknown")
		}
	}

	return sqlStatements, nil
}

func (conn *connection) ToXMLFiles(dir string) error {
	if err := os.MkdirAll(dir, 0777); err != nil && !os.IsExist(err) {
		return err
	}

	files, err := conn.ToXML()
	if err != nil {
		return err
	}
	for id, file := range files {
		err = func(filename string, cfg *xmlConfig) error {
			w, err := os.Create(filepath.Join(dir, filename+".xml"))
			if err != nil {
				return err
			}
			defer w.Close()

			io.WriteString(w, "<?xml version=\"1.0\" encoding=\"utf-8\"?>\r\n")
			encoder := xml.NewEncoder(w)
			encoder.Indent("", "  ")
			err = encoder.Encode(cfg)
			if err != nil {
				return err
			}
			err = encoder.Flush()
			if err != nil {
				return err
			}
			return nil
		}(id, file)
		if err != nil {
			return err
		}
	}
	return nil
}

func (conn *connection) DB() DBRunner {
	return conn.db
}

func (conn *connection) WithDB(db DBRunner) *connection {
	newConn := &connection{}
	*newConn = *conn
	newConn.db = db
	return newConn
}

func (conn *connection) SetDB(db DBRunner) {
	conn.db = db
}

func (conn *connection) DriverName() string {
	return conn.dialect.Name()
}

func (conn *connection) Dialect() Dialect {
	return conn.dialect
}

func (conn *connection) Mapper() *Mapper {
	return conn.mapper
}

func (conn *connection) QueryRow(ctx context.Context, sqlstr string, params []interface{}) SingleRowResult {
	return SingleRowResult{o: conn,
		ctx:       ctx,
		id:        "<empty>",
		sql:       sqlstr,
		sqlParams: params,
	}
}

func (conn *connection) Query(ctx context.Context, sqlstr string, params []interface{}) *MultRowResult {
	return &MultRowResult{o: conn,
		ctx:       ctx,
		id:        "<empty>",
		sql:       sqlstr,
		sqlParams: params,
	}
}

func (conn *connection) Insert(ctx context.Context, id string, paramNames []string, paramValues []interface{}, notReturn ...bool) (int64, error) {
	sqlAndParams, _, err := conn.readSQLParams(ctx, id, StatementTypeInsert, paramNames, paramValues)
	if err != nil {
		return 0, err
	}

	tx := TxFromContext(ctx)
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

func (conn *connection) Update(ctx context.Context, id string, paramNames []string, paramValues []interface{}) (int64, error) {
	sqlAndParams, _, err := conn.readSQLParams(ctx, id, StatementTypeUpdate, paramNames, paramValues)
	if err != nil {
		return 0, err
	}
	return conn.execute(ctx, id, sqlAndParams)
}

func (conn *connection) Delete(ctx context.Context, id string, paramNames []string, paramValues []interface{}) (int64, error) {
	sqlAndParams, _, err := conn.readSQLParams(ctx, id, StatementTypeDelete, paramNames, paramValues)
	if err != nil {
		return 0, err
	}
	return conn.execute(ctx, id, sqlAndParams)
}

func (conn *connection) execute(ctx context.Context, id string, sqlAndParams []sqlAndParam) (int64, error) {
	tx := TxFromContext(ctx)
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

func (conn *connection) SelectOne(ctx context.Context, id string, paramNames []string, paramValues []interface{}) SingleRowResult {
	return conn.selectOneOrInsert(ctx, id, StatementTypeSelect, paramNames, paramValues)
}

func (conn *connection) InsertQuery(ctx context.Context, id string, paramNames []string, paramValues []interface{}) SingleRowResult {
	return conn.selectOneOrInsert(ctx, id, StatementTypeInsert, paramNames, paramValues)
}

func (conn *connection) selectOneOrInsert(ctx context.Context, id string, sqlType StatementType, paramNames []string, paramValues []interface{}) SingleRowResult {
	sqlAndParams, _, err := conn.readSQLParams(ctx, id, sqlType, paramNames, paramValues)
	if err != nil {
		return SingleRowResult{o: conn,
			ctx: ctx,
			id:  id,
			err: err,
		}
	}

	if len(sqlAndParams) > 1 {
		return SingleRowResult{o: conn,
			ctx: ctx,
			id:  id,
			err: ErrMultSQL,
		}
	}
	return SingleRowResult{o: conn,
		ctx:       ctx,
		id:        id,
		sql:       sqlAndParams[0].SQL,
		sqlParams: sqlAndParams[0].Params,
	}
}

func (conn *connection) Select(ctx context.Context, id string, paramNames []string, paramValues []interface{}) *MultRowResult {
	sqlAndParams, _, err := conn.readSQLParams(ctx, id, StatementTypeSelect, paramNames, paramValues)
	if err != nil {
		return &MultRowResult{o: conn,
			ctx: ctx,
			id:  id,
			err: err,
		}
	}
	if len(sqlAndParams) > 1 {
		return &MultRowResult{o: conn,
			ctx: ctx,
			id:  id,
			err: ErrMultSQL,
		}
	}

	return &MultRowResult{o: conn,
		ctx:       ctx,
		id:        id,
		sql:       sqlAndParams[0].SQL,
		sqlParams: sqlAndParams[0].Params,
	}
}

func (o *connection) readSQLParams(ctx context.Context, id string, sqlType StatementType, paramNames []string, paramValues []interface{}) ([]sqlAndParam, ResultType, error) {
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
//  o, err := core.New(&core.Config{DriverName: "mysql",
//         DataSource: "root:root@/51jczj?charset=utf8",
//         XMLPaths: []string{"test.xml"}})
func newConnection(cfg *Config) (*connection, error) {
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

	dbOwner := cfg.TransformDbOwnership
	if cfg.DB == nil {
		driverName := cfg.DriverName
		if strings.HasPrefix(cfg.DriverName, dialects.OdbcPrefix) {
			driverName = "odbc"
		}
		db, err := sql.Open(driverName, cfg.DataSource)
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
		dbOwner = true
	}

	base := &connection{
		tracer:        cfg.Tracer,
		constants:     cfg.Constants,
		dbOwner:       dbOwner,
		db:            cfg.DB,
		sqlStatements: make(map[string]*MappedStatement),
	}

	for key, value := range Constants {
		_, ok := base.constants[key]
		if !ok {
			base.constants[key] = value
		}
	}

	var tagPrefix string
	var tagMapper func(string, string) []string
	if cfg != nil {
		base.isUnsafe = cfg.IsUnsafe
		tagPrefix = cfg.TagPrefix
		tagMapper = cfg.TagMapper
	}
	base.mapper = CreateMapper(tagPrefix, nil, tagMapper)
	base.dialect = NewDialect(cfg.DriverName)
	if base.dialect == dialects.None {
		base.dialect = dialects.Postgres
	}

	ctx := &InitContext{Config: cfg,
		Dialect:    base.dialect,
		Mapper:     base.mapper,
		Statements: base.sqlStatements}

	xmlFiles, err := loadXmlFiles(base, cfg)
	if err != nil {
		return nil, err
	}
	for _, xmlFile := range xmlFiles {
		log.Println("load xml -", xmlFile)
		statements, err := readMappedStatementsFromXMLFile(ctx, xmlFile)
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

	if dir := os.Getenv("gobatis_generate_xml_dir"); dir != "" {
		err := base.ToXMLFiles(dir)
		if err != nil {
			return nil, err
		}
	}

	return base, nil
}

func loadXmlFiles(base *connection, cfg *Config) ([]string, error) {
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

	return xmlPaths, nil
}

func ExecContext(ctx context.Context, conn DBRunner, sqltext string) (rerr error) {
	texts := splitSQLStatements(strings.NewReader(sqltext))

	txctx, tx, err := OpenTxWith(ctx, conn)
	if err != nil {
		return err
	}
	defer func() {
		err := tx.Rollback()
		if err != nil && err != sql.ErrTxDone {
			rerr = errTx{method: "rollback", inner: err}
		}
	}()


	for _, text := range texts {
		_, err = conn.ExecContext(txctx, text)
		if err != nil {
			return err
		}
	}

	err = tx.Commit()
	if err != nil {
		return errTx{method: "commit", inner: err}
	}
	return nil
}
