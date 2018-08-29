package gobatis

import (
	"log"
	"sync"
)

type SqlSession interface {
	DB() dbRunner
	Dialect() Dialect

	Insert(id string, paramNames []string, paramValues []interface{}, notReturn ...bool) (int64, error)
	Update(id string, paramNames []string, paramValues []interface{}) (int64, error)
	Delete(id string, paramNames []string, paramValues []interface{}) (int64, error)
	SelectOne(id string, paramNames []string, paramValues []interface{}) Result
	Select(id string, paramNames []string, paramValues []interface{}) *Results
}

type Reference struct {
	SqlSession
}

type CreateContext struct {
	Session    *Reference
	Statements map[string]*MappedStatement
}

type InitContext struct {
	Config     *Config
	Logger     *log.Logger
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
