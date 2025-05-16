package gobatis

import (
	"context"
	"database/sql"
	"reflect"

	"github.com/runner-mei/GoBatis/core"
	"github.com/runner-mei/GoBatis/dialects"
)

type Config = core.Config
type DBRunner = core.DBRunner

func WithTx(ctx context.Context, tx DBRunner) context.Context {
	return core.WithTx(ctx, tx)
}

func OpenTxWith(ctx context.Context, conn DBRunner, failIfInTx ...bool) (context.Context, *sql.Tx, error) {
	return core.OpenTxWith(ctx, conn, failIfInTx...)
}

func TxFromContext(ctx context.Context) DBRunner {
	return core.TxFromContext(ctx)
}

func InTx(ctx context.Context, db DBRunner, failIfInTx bool, cb func(ctx context.Context, tx DBRunner) error) error {
	return core.InTx(ctx, db, failIfInTx, cb)
}

func InTxFactory(ctx context.Context, db DbSession, optionalTx DBRunner, failIfInTx bool, cb func(ctx context.Context, tx *Tx) error) error {
	return core.InTxFactory(ctx, db, optionalTx, failIfInTx, cb)
}

func WithDbConnection(ctx context.Context, tx DBRunner) context.Context {
	return core.WithTx(ctx, tx)
}

func DbConnectionFromContext(ctx context.Context) DBRunner {
	return core.TxFromContext(ctx)
}

func RegisterExprFunction(name string, fn func(args ...interface{}) (interface{}, error)) {
	core.RegisterExprFunction(name, fn)
}

func ValidPrintString(value string, inStr bool) error {
	return core.ValidPrintString(value, inStr)
}

type Clob = dialects.Clob

type Tracer = core.Tracer
type TraceWriter = core.TraceWriter
type StdLogger = core.StdLogger

type SqlSession = core.SqlSession

type TableName = core.TableName
type SessionFactory = core.Session
type Session = core.Session
type Tx = core.Tx
type DbSession = core.DbSession
type Reference = core.Reference
type CreateContext = core.CreateContext
type InitContext = core.InitContext
type Dialect = core.Dialect
type Context = core.Context
type SingleRowResult = core.SingleRowResult
type MultRowResult = core.MultRowResult
type MultipleArray = core.MultipleArray
type Multiple = core.Multiple

type StatementType = core.StatementType
type ResultType = core.ResultType
type MappedStatement = core.MappedStatement
type SqlExpression = core.SqlExpression
type Params = core.Params
type Nullable = core.Nullable
type Error = core.Error
type SqlError = core.SqlError
type ErrTableNotExists = core.ErrTableNotExists

var SplitXORM = core.SplitXORM
var SplitDB = core.SplitDB

type TagSplit = core.TagSplit

const (
	OdbcPrefix = "odbc_with_"

	StatementTypeNone   = core.StatementTypeNone
	StatementTypeSelect = core.StatementTypeSelect
	StatementTypeUpdate = core.StatementTypeUpdate
	StatementTypeInsert = core.StatementTypeInsert
	StatementTypeDelete = core.StatementTypeDelete

	ResultUnknown = core.ResultUnknown
	ResultMap     = core.ResultMap
	ResultStruct  = core.ResultStruct
)

var (
	None     = dialects.None
	Postgres = dialects.Postgres
	Kingbase = dialects.Kingbase
	Mysql    = dialects.Mysql
	MSSql    = dialects.MSSql
	Oracle   = dialects.Oracle
	DM       = dialects.DM

	TemplateFuncs = core.TemplateFuncs

	DiscardTracer = core.DiscardTracer
	Constants     = core.Constants

	ErrAlreadyTx = core.ErrAlreadyTx
)

func WithSqlSession(ctx context.Context, sess SqlSession) context.Context {
	return core.WithSqlSession(ctx, sess)
}

func SqlSessionFromContext(ctx context.Context) SqlSession {
	return core.SqlSessionFromContext(ctx)
}

func ClearInit() []func(ctx *InitContext) error {
	return core.ClearInit()
}

func SetInit(callbacks []func(ctx *InitContext) error) []func(ctx *InitContext) error {
	return core.SetInit(callbacks)
}

func Init(cb func(ctx *InitContext) error) {
	core.Init(cb)
}

func NewDialect(driverName string) Dialect {
	return core.NewDialect(driverName)
}

func CompileNamedQuery(txt string) ([]string, Params, error) {
	return core.CompileNamedQuery(txt)
}

type Mapper = core.Mapper
type StructMap = core.StructMap
type FieldInfo = core.FieldInfo

func CreateMapper(prefix string, nameMapper func(string) string, tagMapper func(string, string) []string) *Mapper {
	return core.CreateMapper(prefix, nameMapper, tagMapper)
}

func TagSplitForXORM(s string, fieldName string) []string {
	return core.TagSplitForXORM(s, fieldName)
}

func TagSplitForDb(s string, fieldName string) []string {
	return core.TagSplitForDb(s, fieldName)
}

func ReadTableFields(mapper *Mapper, instance reflect.Type) ([]string, error) {
	return core.ReadTableFields(mapper, instance)
}

func ErrForGenerateStmt(err error, msg string) error {
	return core.ErrForGenerateStmt(err, msg)
}

func NewMapppedStatement(ctx *InitContext, id string, statementType StatementType, resultType ResultType, sqlStr string) (*MappedStatement, error) {
	return core.NewMapppedStatement(&core.StmtContext{InitContext: ctx}, id, statementType, resultType, sqlStr)
}

func NewSqlExpression(ctx *InitContext, sqlstr string) (SqlExpression, error) {
	return core.NewSqlExpression(ctx, sqlstr)
}

func NewMultipleArray() *MultipleArray {
	return core.NewMultipleArray()
}

func NewMultiple() *Multiple {
	return core.NewMultiple()
}

func New(cfg *Config) (*Session, error) {
	return core.New(cfg)
}

func ExecContext(ctx context.Context, conn DBRunner, sqltext string) error {
	return core.ExecContext(ctx, conn, sqltext)
}

func ErrStatementAlreadyExists(id string) error {
	return core.ErrStatementAlreadyExists(id)
}

func IsTableNotExists(dialect dialects.Dialect, e error) bool {
	return core.IsTableNotExists(dialect, e)
}

func AsTableNotExists(dialect dialects.Dialect, e error, ae *ErrTableNotExists) bool {
	return core.AsTableNotExists(dialect, e, ae)
}

func IsTxError(e error, method string, methods ...string) bool {
	return core.IsTxError(e, method, methods...)
}

func IsBeginTx(e error) bool {
	return core.IsTxError(e, "begin")
}

func IsRollbackTx(e error) bool {
	return core.IsTxError(e, "rollback")
}

func IsCommitTx(e error) bool {
	return core.IsTxError(e, "commit")
}
