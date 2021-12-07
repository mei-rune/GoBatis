package core

import (
	"bytes"
	"database/sql"
	"database/sql/driver"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"reflect"
	"strings"
	"time"

	"github.com/runner-mei/GoBatis/convert"
)

type SQLType interface {
	ToSQLValue() (interface{}, error)
}

func toSQLType(dialect Dialect, param *Param, value interface{}) (interface{}, error) {
	if value == nil {
		if param.NotNull.Valid && param.NotNull.Bool {
			return nil, errors.New("param '" + param.Name + "' is zero value")
		}

		return nil, nil
	}

	switch v := value.(type) {
	case bool:
		return value, nil
	case int8:
		if v == 0 {
			if param.NotNull.Valid && param.NotNull.Bool {
				return nil, errors.New("param '" + param.Name + "' is zero value")
			} else if param.Null.Valid && param.Null.Bool {
				return nil, nil
			}
		}
		return value, nil
	case int16:
		if v == 0 {
			if param.NotNull.Valid && param.NotNull.Bool {
				return nil, errors.New("param '" + param.Name + "' is zero value")
			} else if param.Null.Valid && param.Null.Bool {
				return nil, nil
			}
		}
		return value, nil
	case int32:
		if v == 0 {
			if param.NotNull.Valid && param.NotNull.Bool {
				return nil, errors.New("param '" + param.Name + "' is zero value")
			} else if param.Null.Valid && param.Null.Bool {
				return nil, nil
			}
		}
		return value, nil
	case int:
		if v == 0 {
			if param.NotNull.Valid && param.NotNull.Bool {
				return nil, errors.New("param '" + param.Name + "' is zero value")
			} else if param.Null.Valid && param.Null.Bool {
				return nil, nil
			}
		}
		return value, nil
	case int64:
		if v == 0 {
			if param.NotNull.Valid && param.NotNull.Bool {
				return nil, errors.New("param '" + param.Name + "' is zero value")
			} else if param.Null.Valid && param.Null.Bool {
				return nil, nil
			}
		}
		return value, nil
	case uint8:
		if v == 0 {
			if param.NotNull.Valid && param.NotNull.Bool {
				return nil, errors.New("param '" + param.Name + "' is zero value")
			} else if param.Null.Valid && param.Null.Bool {
				return nil, nil
			}
		}
		return value, nil
	case uint16:
		if v == 0 {
			if param.NotNull.Valid && param.NotNull.Bool {
				return nil, errors.New("param '" + param.Name + "' is zero value")
			} else if param.Null.Valid && param.Null.Bool {
				return nil, nil
			}
		}
		return value, nil
	case uint32:
		if v == 0 {
			if param.NotNull.Valid && param.NotNull.Bool {
				return nil, errors.New("param '" + param.Name + "' is zero value")
			} else if param.Null.Valid && param.Null.Bool {
				return nil, nil
			}
		}
		return value, nil
	case uint:
		if v == 0 {
			if param.NotNull.Valid && param.NotNull.Bool {
				return nil, errors.New("param '" + param.Name + "' is zero value")
			} else if param.Null.Valid && param.Null.Bool {
				return nil, nil
			}
		}
		return value, nil
	case uint64:
		if v == 0 {
			if param.NotNull.Valid && param.NotNull.Bool {
				return nil, errors.New("param '" + param.Name + "' is zero value")
			} else if param.Null.Valid && param.Null.Bool {
				return nil, nil
			}
		}
		return value, nil
	case float32:
		if v == 0 {
			if param.NotNull.Valid && param.NotNull.Bool {
				return nil, errors.New("param '" + param.Name + "' is zero value")
			} else if param.Null.Valid && param.Null.Bool {
				return nil, nil
			}
		}
		return value, nil
	case float64:
		if v == 0 {
			if param.NotNull.Valid && param.NotNull.Bool {
				return nil, errors.New("param '" + param.Name + "' is zero value")
			} else if param.Null.Valid && param.Null.Bool {
				return nil, nil
			}
		}
		return value, nil
	case complex64:
		if v == 0 {
			if param.NotNull.Valid && param.NotNull.Bool {
				return nil, errors.New("param '" + param.Name + "' is zero value")
			} else if param.Null.Valid && param.Null.Bool {
				return nil, nil
			}
		}
		return value, nil
	case complex128:
		if v == 0 {
			if param.NotNull.Valid && param.NotNull.Bool {
				return nil, errors.New("param '" + param.Name + "' is zero value")
			} else if param.Null.Valid && param.Null.Bool {
				return nil, nil
			}
		}
		return value, nil
	case string:
		if v == "" {
			if param.NotNull.Valid && param.NotNull.Bool {
				return nil, errors.New("param '" + param.Name + "' is zero value")
			} else if param.Null.Valid && param.Null.Bool {
				return nil, nil
			}
		}
		return value, nil
	case []byte:
		if len(v) == 0 {
			if param.NotNull.Valid && param.NotNull.Bool {
				return nil, errors.New("param '" + param.Name + "' is zero value")
			} else if param.Null.Valid && param.Null.Bool {
				return nil, nil
			}
		}
		return value, nil
	case *bool:
		if param.NotNull.Valid && param.NotNull.Bool {
			if v == nil {
				return nil, errors.New("param '" + param.Name + "' is nil value")
			}
		}
		return value, nil
	case *float32, *float64:
		if param.NotNull.Valid && param.NotNull.Bool {
			rValue := reflect.ValueOf(value)
			if rValue.IsNil() {
				return nil, errors.New("param '" + param.Name + "' is nil value")
			}
			if rValue.Elem().Float() == 0 {
				return nil, errors.New("param '" + param.Name + "' is zero value")
			}
		} else if param.Null.Valid && param.Null.Bool {
			rValue := reflect.ValueOf(value)
			if !rValue.IsNil() && rValue.Elem().Float() == 0 {
				return nil, nil
			}
		}
		return value, nil

	case *complex64, *complex128:
		if param.NotNull.Valid && param.NotNull.Bool {
			rValue := reflect.ValueOf(value)
			if rValue.IsNil() {
				return nil, errors.New("param '" + param.Name + "' is nil value")
			}
			if rValue.Elem().Complex() == 0 {
				return nil, errors.New("param '" + param.Name + "' is zero value")
			}
		} else if param.Null.Valid && param.Null.Bool {
			rValue := reflect.ValueOf(value)
			if !rValue.IsNil() && rValue.Elem().Complex() == 0 {
				return nil, nil
			}
		}
		return value, nil

	case *int8, *int16, *int32, *int, *int64:
		if param.NotNull.Valid && param.NotNull.Bool {
			rValue := reflect.ValueOf(value)
			if rValue.IsNil() {
				return nil, errors.New("param '" + param.Name + "' is nil value")
			}
			if rValue.Elem().Int() == 0 {
				return nil, errors.New("param '" + param.Name + "' is zero value")
			}
		} else if param.Null.Valid && param.Null.Bool {
			rValue := reflect.ValueOf(value)
			if !rValue.IsNil() && rValue.Elem().Int() == 0 {
				return nil, nil
			}
		}
		return value, nil
	case *uint8, *uint16, *uint32, *uint, *uint64:
		if param.NotNull.Valid && param.NotNull.Bool {
			rValue := reflect.ValueOf(value)
			if rValue.IsNil() {
				return nil, errors.New("param '" + param.Name + "' is nil value")
			}
			if rValue.Elem().Uint() == 0 {
				return nil, errors.New("param '" + param.Name + "' is zero value")
			}
		} else if param.Null.Valid && param.Null.Bool {
			rValue := reflect.ValueOf(value)
			if !rValue.IsNil() && rValue.Elem().Uint() == 0 {
				return nil, nil
			}
		}
		return value, nil
	case *string:
		if param.NotNull.Valid && param.NotNull.Bool {
			if v == nil {
				return nil, errors.New("param '" + param.Name + "' is nil value")
			}
			if *v == "" {
				return nil, errors.New("param '" + param.Name + "' is zero value")
			}
		} else if param.Null.Valid && param.Null.Bool {
			if v != nil && *v == "" {
				return nil, nil
			}
		}
		return value, nil
	case *[]byte:
		if param.NotNull.Valid && param.NotNull.Bool {
			if v == nil {
				return nil, errors.New("param '" + param.Name + "' is nil value")
			}
			if len(*v) == 0 {
				return nil, errors.New("param '" + param.Name + "' is zero value")
			}
		} else if param.Null.Valid && param.Null.Bool {
			if v != nil && len(*v) == 0 {
				return nil, nil
			}
		}
		return value, nil
	case time.Time:
		if v.IsZero() {
			return nil, nil
		}
		return v, nil
	case *time.Time:
		if v == nil || v.IsZero() {
			return nil, nil
		}
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
		if valuer, ok := value.(driver.Valuer); ok {
			return valuer, nil
		}

		//rValue := reflect.ValueOf(value)
		typ := reflect.TypeOf(value)
		kind := typ.Kind()
		if kind == reflect.Ptr {
			kind = typ.Elem().Kind()
		}
		if kind == reflect.Chan || kind == reflect.Func || kind == reflect.UnsafePointer {
			return nil, fmt.Errorf("param '%s' isnot a sql type got %T", param.Name, value)
		}

		if kind == reflect.Slice {
			valuer, err := dialect.MakeArrayValuer(value)
			if err != nil {
				return nil, fmt.Errorf("param '%s' convert to array, %s", param.Name, err)
			}
			return valuer, nil
		}

		bs, err := json.Marshal(v)
		if err != nil {
			return nil, fmt.Errorf("param '%s' convert to json, %s", param.Name, err)
		}
		return string(bs), nil
	}
}

var _ sql.Scanner = &scanner{}

type scanner struct {
	name  string
	value interface{}
	Valid bool
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
	if bytes.Equal(bs, []byte("[null]")) {
		return nil
	}
	decoder := json.NewDecoder(bytes.NewReader(bs))
	decoder.UseNumber()
	if err := decoder.Decode(s.value); err != nil {
		return fmt.Errorf("column %s unmarshal error, %s\r\n\t%s", s.name, err, bs)
	}
	s.Valid = true
	return nil
}

func MakJSONScanner(name string, value interface{}) interface{} {
	return &scanner{name: name, value: value}
}

var _ sql.Scanner = &sScanner{}

type sScanner struct {
	name  string
	field reflect.Value

	Valid    bool
	scanFunc func(s *sScanner, str string) error
}

func scanIP(s *sScanner, str string) error {
	ip := net.ParseIP(str)
	if ip == nil {
		if !strings.HasSuffix(str, "/0") {
			return fmt.Errorf("column %s is invalid ip address - '%s'", s.name, str)
		}
		str = strings.TrimSuffix(str, "/0")
		ip = net.ParseIP(str)
		if ip == nil {
			return fmt.Errorf("column %s is invalid ip address - '%s'", s.name, str)
		}
	}

	if s.field.Kind() == reflect.Ptr {
		s.field.Set(reflect.ValueOf(&ip))
	} else {
		s.field.Set(reflect.ValueOf(ip))
	}
	s.Valid = true
	return nil
}

func scanIPNet(s *sScanner, str string) error {
	_, ipnet, err := net.ParseCIDR(str)
	if err != nil {
		return fmt.Errorf("column %s is invalid cidr - '%s'", s.name, str)
	}

	if s.field.Kind() == reflect.Ptr {
		s.field.Set(reflect.ValueOf(ipnet))
	} else {
		s.field.Set(reflect.ValueOf(*ipnet))
	}
	s.Valid = true
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
	s.Valid = true
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

type emptyScanner struct{}

func (s *emptyScanner) Scan(src interface{}) error {
	return nil
}

var emptyScan sql.Scanner = &emptyScanner{}

type Nullable struct {
	Name  string
	Value interface{}

	Valid bool
}

func (s *Nullable) Scan(src interface{}) error {
	if src == nil {
		return nil
	}
	s.Valid = true
	return convert.ConvertAssign(s.Value, src)
}

func MakeIPScanner(fieldName string, value *net.IP) sql.Scanner {
	return &sScanner{name: fieldName, field: reflect.ValueOf(value).Elem(), scanFunc: scanIP}
}

func MakeIPNetScanner(fieldName string, value *net.IPNet) sql.Scanner {
	return &sScanner{name: fieldName, field: reflect.ValueOf(value).Elem(), scanFunc: scanIPNet}
}
