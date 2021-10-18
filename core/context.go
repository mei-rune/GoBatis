package core

import (
	"context"
	"sync"
)

type SqlSession interface {
	DB() DBRunner
	Dialect() Dialect

	Insert(ctx context.Context, id string, paramNames []string, paramValues []interface{}, notReturn ...bool) (int64, error)
	InsertQuery(ctx context.Context, id string, paramNames []string, paramValues []interface{}) SingleRowResult
	Update(ctx context.Context, id string, paramNames []string, paramValues []interface{}) (int64, error)
	Delete(ctx context.Context, id string, paramNames []string, paramValues []interface{}) (int64, error)
	SelectOne(ctx context.Context, id string, paramNames []string, paramValues []interface{}) SingleRowResult
	Select(ctx context.Context, id string, paramNames []string, paramValues []interface{}) *MultRowResult
}

type sessionKeyType struct{}

func (*sessionKeyType) String() string {
	return "gobatis-session-key"
}

var sessionKey = &sessionKeyType{}

func WithSqlSession(ctx context.Context, sess SqlSession) context.Context {
	if ctx == nil {
		ctx = context.Background()
	}
	return context.WithValue(ctx, txKey, sess)
}

func SqlSessionFromContext(ctx context.Context) SqlSession {
	if ctx == nil {
		return nil
	}
	v := ctx.Value(txKey)
	if v == nil {
		return nil
	}
	return v.(SqlSession)
}

type Reference struct {
	SqlSession
}

var _ SqlSession = &Reference{}
var _ SqlSession = Reference{}

type CreateContext struct {
	Session    *Reference
	Statements map[string]*MappedStatement
}

type InitContext struct {
	Config     *Config
	Dialect    Dialect
	Mapper     *Mapper
	Statements map[string]*MappedStatement
}

var (
	initLock      sync.Mutex
	initCallbacks []func(ctx *InitContext) error
)

func ClearInit() []func(ctx *InitContext) error {
	initLock.Lock()
	defer initLock.Unlock()
	tmp := initCallbacks
	initCallbacks = nil
	return tmp
}

func SetInit(callbacks []func(ctx *InitContext) error) []func(ctx *InitContext) error {
	initLock.Lock()
	defer initLock.Unlock()
	tmp := initCallbacks
	initCallbacks = callbacks
	return tmp
}

func Init(cb func(ctx *InitContext) error) {
	initLock.Lock()
	defer initLock.Unlock()
	initCallbacks = append(initCallbacks, cb)
}

func runInit(ctx *InitContext) error {
	initLock.Lock()
	defer initLock.Unlock()
	for idx := range initCallbacks {
		if err := initCallbacks[idx](ctx); err != nil {
			return err
		}
	}
	return nil
}
