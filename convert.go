package gobatis

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"net"
	"reflect"
	"strings"
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
		switch v := value.Interface().(type) {
		case time.Time:
			return v, nil
		case *time.Time:
			return v, nil
		case net.IP:
			if v == nil {
				return nil, nil
			}
			return v.String(), nil
		case *net.IP:
			if v == nil || *v == nil {
				return nil, nil
			}
			return v.String(), nil
		case net.HardwareAddr:
			if v == nil {
				return nil, nil
			}
			return v.String(), nil
		case *net.HardwareAddr:
			if v == nil || *v == nil {
				return nil, nil
			}
			return v.String(), nil
		default:
			bs, err := json.Marshal(v)
			if err != nil {
				return nil, fmt.Errorf("param '%s' convert to json, %s", param.Name, err)
			}
			return string(bs), nil
		}
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
		switch value := field.Interface().(type) {
		case *time.Time:
			return value, nil
		case **time.Time:
			return value, nil
		case *net.IP:
			return &sScanner{name: column, field: field.Elem(), scanFunc: scanIP}, nil
		case *net.HardwareAddr:
			return &sScanner{name: column, field: field.Elem(), scanFunc: scanMAC}, nil
		case **net.IP:
			return &sScanner{name: column, field: field.Elem(), scanFunc: scanIP}, nil
		case **net.HardwareAddr:
			return &sScanner{name: column, field: field.Elem(), scanFunc: scanMAC}, nil
		default:
			if _, ok := value.(sql.Scanner); ok {
				return value, nil
			}
			return &scanner{name: column, value: value}, nil
		}
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

var _ sql.Scanner = &sScanner{}

type sScanner struct {
	name  string
	field reflect.Value

	scanFunc func(s *sScanner, str string) error
}

func scanIP(s *sScanner, str string) error {
	ip := net.ParseIP(str)
	if ip == nil {
		return fmt.Errorf("column %s is invalid ip address - '%s'", s.name, str)
	}

	if s.field.Kind() == reflect.Ptr {
		s.field.Set(reflect.ValueOf(&ip))
	} else {
		s.field.Set(reflect.ValueOf(ip))
	}
	return nil
}

func scanMAC(s *sScanner, str string) error {
	mac, err := net.ParseMAC(str)
	if err != nil {
		return fmt.Errorf("column %s is invalid ip address - '%s'", s.name, str)
	}

	if s.field.Kind() == reflect.Ptr {
		s.field.Set(reflect.ValueOf(&mac))
	} else {
		s.field.Set(reflect.ValueOf(mac))
	}
	return nil
}

func (s *sScanner) Scan(src interface{}) error {
	if src == nil {
		return nil
	}
	str, ok := src.(string)
	if !ok {
		bs, ok := src.([]byte)
		if !ok {
			return fmt.Errorf("column %s should byte array but got '%T', target type '%T'", s.name, src, s.field.Type().String())
		}
		bs = bytes.TrimSpace(bs)
		if len(bs) == 0 {
			return nil
		}
		str = string(bs)
	} else {
		str = strings.TrimSpace(str)
		if len(str) == 0 {
			return nil
		}
	}

	return s.scanFunc(s, str)
}

type nullScanner struct {
	name  string
	value interface{}
}

func (s *nullScanner) Scan(src interface{}) error {
	if src == nil {
		return nil
	}
	return convertAssign(s.value, src)
}
