package gobatis

import (
	"sync"
)

type SqlSession interface {
	DbType() int

	Insert(id string, paramNames []string, paramValues []interface{}) (int64, error)
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
	DbType     int
	Statements map[string]*MappedStatement
}

var (
	initLock      sync.Mutex
	initCallbacks []func(ctx *InitContext) error
)

func ClearInit() {
	initLock.Lock()
	defer initLock.Unlock()
	initCallbacks = nil
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
