package gobatis

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"time"
)

func toSQLType(param *Param, value interface{}) (interface{}, error) {
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
		bs, err := json.Marshal(value.Interface())
		if err != nil {
			return nil, fmt.Errorf("param '%s' convert to json, %s", param.Name, err)
		}
		return string(bs), nil
	}
}

func toGOTypeWith(instance, field reflect.Value) (interface{}, error) {
	kind := field.Type().Kind()
	if kind == reflect.Ptr {
		kind = field.Type().Elem().Kind()
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
		return nil, fmt.Errorf("target '%T' of '%T' is invalid", instance.Interface(), field.Interface())
	default:
		var value = field.Interface()
		if _, ok := value.(*time.Time); ok {
			return value, nil
		}
		return &valueScanner{value: value}, nil
	}
}

var _ sql.Scanner = &valueScanner{}

type valueScanner struct {
	value interface{}
}

func (s *valueScanner) Scan(src interface{}) error {
	return errors.New("not implemented")
}
