package gobatis

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"reflect"
	"time"
)

func toSQLType(param *Param, value interface{}) (interface{}, error) {
	if value == nil {
		return nil, nil
	}
	return toSQLTypeWith(param, reflect.ValueOf(value))
}

func toSQLTypeWith(param *Param, value reflect.Value) (interface{}, error) {
	kind := value.Type().Kind()
	if kind == reflect.Ptr {
		kind = value.Type().Elem().Kind()
	}
	switch kind {
	case reflect.Bool,
		reflect.Int,
		reflect.Int8,
		reflect.Int16,
		reflect.Int32,
		reflect.Int64,
		reflect.Uint,
		reflect.Uint8,
		reflect.Uint16,
		reflect.Uint32,
		reflect.Uint64,
		reflect.Uintptr,
		reflect.Float32,
		reflect.Float64,
		reflect.Complex64,
		reflect.Complex128,
		reflect.String:
		return value.Interface(), nil
	case reflect.Chan, reflect.Func, reflect.UnsafePointer:
		return nil, fmt.Errorf("param '%s' isnot a sql type got %T", param.Name, value.Interface())
	default:
		v := value.Interface()
		if _, ok := v.(*time.Time); ok {
			return v, nil
		}

		bs, err := json.Marshal(v)
		if err != nil {
			return nil, fmt.Errorf("param '%s' convert to json, %s", param.Name, err)
		}
		return string(bs), nil
	}
}

func toGOTypeWith(column string, instance, field reflect.Value) (interface{}, error) {
	kind := field.Type().Kind()
	if kind == reflect.Ptr {
		kind = field.Type().Elem().Kind()
		if kind == reflect.Ptr {
			kind = field.Type().Elem().Kind()
		}
	}
	switch kind {
	case reflect.Bool,
		reflect.Int,
		reflect.Int8,
		reflect.Int16,
		reflect.Int32,
		reflect.Int64,
		reflect.Uint,
		reflect.Uint8,
		reflect.Uint16,
		reflect.Uint32,
		reflect.Uint64,
		reflect.Uintptr,
		reflect.Float32,
		reflect.Float64,
		reflect.Complex64,
		reflect.Complex128,
		reflect.String:
		return field.Interface(), nil
	case reflect.Chan, reflect.Func, reflect.UnsafePointer:
		return nil, fmt.Errorf("cannot convert column '%s' into '%T' fail, error type %T", column, instance.Interface(), field.Interface())
	default:
		var value = field.Interface()
		if _, ok := value.(*time.Time); ok {
			return value, nil
		}
		if _, ok := value.(**time.Time); ok {
			return value, nil
		}
		if _, ok := value.(sql.Scanner); ok {
			return value, nil
		}
		return &scanner{name: column, value: value}, nil
	}
}

var _ sql.Scanner = &scanner{}

type scanner struct {
	name  string
	value interface{}
}

func (s *scanner) Scan(src interface{}) error {
	if src == nil {
		return nil
	}
	bs, ok := src.([]byte)
	if !ok {
		str, ok := src.(string)
		if !ok {
			return fmt.Errorf("column %s should byte array but got '%T', target type '%T'", s.name, src, s.value)
		}
		bs = []byte(str)
	}
	bs = bytes.TrimSpace(bs)
	if len(bs) == 0 {
		return nil
	}
	if err := json.Unmarshal(bs, s.value); err != nil {
		return fmt.Errorf("column %s unmarshal error, %s\r\n\t%s", s.name, err, bs)
	}
	return nil
}
