package gobatis

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
)

type Lazy interface {
	isLazy()
	Read(value interface{}) error

	// For Generic
	// Value() T
	// Get() (T, error)
	// Read(value T) error
}

type LazyValue struct {
	HasValue  bool
	Value     interface{}
	ByteArray []byte
}

func (lv *LazyValue) isLazy() {}

func (lv *LazyValue) ToSQLValue() (interface{}, error) {
	if len(lv.ByteArray) == 0 && lv.HasValue && lv.Value != nil {
		bs, err := json.Marshal(lv.Value)
		if err != nil {
			return nil, err
		}
		lv.ByteArray = bs
	}
	return lv.ByteArray, nil
}

func (lv *LazyValue) Set(value interface{}) {
	lv.HasValue = true
	lv.Value = value
	lv.ByteArray = nil
}

func (lv *LazyValue) Read(value interface{}) error {
	if lv.HasValue {
		if lv.Value == nil {
			return nil
		}

		dst := reflect.ValueOf(value)
		src := reflect.ValueOf(lv.Value)

		if dst.Kind() != reflect.Ptr || dst.IsNil() {
			return errors.New("must pass a pointer")
		}
		if src.Kind() == reflect.Ptr {
			dst.Elem().Set(src.Elem())
		} else {
			dst.Elem().Set(src)
		}
		return nil
	}

	if len(lv.ByteArray) == 0 {
		return nil
	}

	return json.Unmarshal(lv.ByteArray, value)
}

// func (lv *LazyValue) Scan(src interface{}) error {
// 	if src == nil {
// 		lv.HasValue = false
// 		lv.Value = nil
// 		lv.bs = nil
// 		return nil
// 	}

// 	switch v := src.(type) {
// 	case []byte:
// 		lv.HasValue = false
// 		lv.Value = nil
// 		lv.bs = v
// 		return nil
// 	case string:
// 		lv.HasValue = false
// 		lv.Value = nil
// 		lv.bs = []byte(v)
// 		return nil
// 	default:
// 		return fmt.Errorf("except byte array but got '%T'", src)
// 	}
// }

type LazyBytes interface {
	isLazy()

	Value() []byte
	Get() ([]byte, error)
	Read(value interface{}) error
}

type LBytes []byte

func (lb *LBytes) isLazy() {}

func (lb *LBytes) Value() []byte {
	return []byte(*lb)
}

func (lb *LBytes) Get() ([]byte, error) {
	return []byte(*lb), nil
}

func (lb *LBytes) Read(value interface{}) error {
	bp, ok := value.(*[]byte)
	if !ok {
		return fmt.Errorf("converting LBytes to a %T", value)
	}
	*bp = *lb
	return nil
}

func (lb *LBytes) ToSQLValue() (interface{}, error) {
	return []byte(*lb), nil
}

type LazyString interface {
	isLazy()

	Value() string
	Get() (string, error)
	Read(value interface{}) error
}

type LString string

func (ls *LString) isLazy() {}

func (ls *LString) Value() string {
	return string(*ls)
}

func (ls *LString) Get() (string, error) {
	return string(*ls), nil
}

func (ls *LString) Read(value interface{}) error {
	sp, ok := value.(*string)
	if !ok {
		return fmt.Errorf("converting LString to a %T", value)
	}
	*sp = string(*ls)
	return nil
}

func (ls *LString) ToSQLValue() (interface{}, error) {
	return string(*ls), nil
}

type lazyValue struct {
	conn      Connection
	stmt      string
	sqlParams []interface{}
}

func (lv *lazyValue) isLazy() {}

func (lv *lazyValue) Read(value interface{}) error {
	rows, err := lv.conn.db.QueryContext(context.Background(), lv.stmt, lv.sqlParams...)
	if err != nil {
		return err
	}
	defer rows.Close()

	if !rows.Next() {
		if err := rows.Err(); err != nil {
			return err
		}
		return sql.ErrNoRows
	}
	return rows.Scan(value)
}
