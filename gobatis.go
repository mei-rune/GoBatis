package gobatis

import (
	"database/sql"
	"context"
	"reflect"
	"github.com/runner-mei/GoBatis/core"
	"github.com/runner-mei/GoBatis/dialects"
)

type  Config = core.Config
type  DBRunner = core.DBRunner


func WithTx(ctx context.Context, tx DBRunner) context.Context {
	return core.WithTx(ctx, tx)
}

func OpenTxWith(ctx context.Context, conn DBRunner, failIfInTx ...bool) (context.Context, *sql.Tx, error) {
	return core.OpenTxWith(ctx, conn, failIfInTx...)
}

func TxFromContext(ctx context.Context) DBRunner {
	return core.TxFromContext(ctx)
}

func WithDbConnection(ctx context.Context, tx DBRunner) context.Context {
	return core.WithTx(ctx, tx)
}

func DbConnectionFromContext(ctx context.Context) DBRunner {
	return core.TxFromContext(ctx)
}
type Tracer = core.Tracer
type TraceWriter = core.TraceWriter
type StdLogger = core.StdLogger
type Connection = core.Connection
type SqlSession = core.SqlSession
type SessionFactory = core.SessionFactory
type Tx = core.Tx
type Reference = core.Reference
type CreateContext = core.CreateContext
type InitContext = core.InitContext
type Dialect = core.Dialect
type Context = core.Context
type Result = core.Result
type MultipleArray = core.MultipleArray
type Multiple = core.Multiple

type StatementType = core.StatementType
type ResultType = core.ResultType
type MappedStatement = core.MappedStatement
type Params = core.Params
type Nullable = core.Nullable

const (
	StatementTypeNone    = core.StatementTypeNone
	StatementTypeSelect  = core.StatementTypeSelect
	StatementTypeUpdate  = core.StatementTypeUpdate
	StatementTypeInsert  = core.StatementTypeInsert
	StatementTypeDelete  = core.StatementTypeDelete

	ResultUnknown   = core.ResultUnknown
	ResultMap       = core.ResultMap
	ResultStruct  = core.ResultStruct
)

var (
	None     = dialects.None
	Postgres = dialects.Postgres
	Mysql    = dialects.Mysql
	MSSql    = dialects.MSSql
	Oracle   = dialects.Oracle

	TemplateFuncs = core.TemplateFuncs
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

func ReadTableFields(mapper *Mapper, instance reflect.Type) ([]string, error) {
	return core.ReadTableFields(mapper, instance)
}


func ErrForGenerateStmt(err error, msg string) error {
	return core.ErrForGenerateStmt(err, msg)
}

func NewMapppedStatement(ctx *InitContext, id string, statementType StatementType, resultType ResultType, sqlStr string) (*MappedStatement, error) {
	return core.NewMapppedStatement(ctx, id, statementType, resultType, sqlStr)
}

func NewMultipleArray() *MultipleArray {
	return core.NewMultipleArray()
}

func NewMultiple() *Multiple {
	return core.NewMultiple()
}


func New(cfg *Config) (*SessionFactory, error) {
	return core.New(cfg)
}