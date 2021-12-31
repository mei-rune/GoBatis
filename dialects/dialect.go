package dialects

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/lib/pq"
)

func New(driverName string) Dialect {
	switch strings.ToLower(driverName) {
	case "postgres":
		return Postgres
	case "mysql":
		return Mysql
	case "mssql", "sqlserver":
		return MSSql
	case "oracle", "ora":
		return Oracle
	case "dm":
		return DM
	default:
		return None
	}
	// panic("Unsupported database type: " + driverName)
}

type Dialect interface {
	Name() string
	Quote(string) string
	BooleanStr(bool) string
	Placeholder() PlaceholderFormat
	InsertIDSupported() bool
	HandleError(error) error
	Limit(int64, int64) string

	MakeArrayValuer(interface{}) (interface{}, error)
	MakeArrayScanner(string, interface{}) (interface{}, error)
}

type dialect struct {
	name            string
	placeholder     PlaceholderFormat
	hasLastInsertID bool
	quoteChars      string
	trueStr         string
	falseStr        string
	handleError     func(e error) error

	makeArrayValuer  func(interface{}) (interface{}, error)
	makeArrayScanner func(string, interface{}) (interface{}, error)
}

func (d *dialect) Quote(name string) string {
	return d.quoteChars + name + d.quoteChars
}

func (d *dialect) BooleanStr(b bool) string {
	if b {
		return d.trueStr
	}
	return d.falseStr
}

func (d *dialect) Limit(offset, limit int64) string {
	if offset > 0 {
		if limit > 0 {
			return fmt.Sprintf(" OFFSET %d LIMIT %d ", offset, limit)
		}
		return fmt.Sprintf(" OFFSET %d ", offset)
	}
	if limit > 0 {
		return fmt.Sprintf(" LIMIT %d ", limit)
	}
	return ""
}

func (d *dialect) Name() string {
	return d.name
}

func (d *dialect) Placeholder() PlaceholderFormat {
	return d.placeholder
}

func (d *dialect) InsertIDSupported() bool {
	return d.hasLastInsertID
}

func (d *dialect) HandleError(e error) error {
	if d.handleError == nil {
		return e
	}
	return d.handleError(e)
}

func (d *dialect) MakeArrayValuer(v interface{}) (interface{}, error) {
	return d.makeArrayValuer(v)
}

func (d *dialect) MakeArrayScanner(name string, v interface{}) (interface{}, error) {
	return d.makeArrayScanner(name, v)
}

var (
	makeArrayValuer = func(v interface{}) (interface{}, error) {
		bs, err := json.Marshal(v)
		return bs, err
	}
	makeArrayStringValuer = func(v interface{}) (interface{}, error) {
		bs, err := json.Marshal(v)
		if err != nil {
			return nil, err
		}
		return string(bs), nil
	}

	makeArrayScanner = func(name string, v interface{}) (interface{}, error) {
		return &scanner{name: name, value: v}, nil
	}

	makePQArrayValuer = func(v interface{}) (interface{}, error) {
		value := pq.Array(v)
		return value, nil
	}
	makePQArrayScanner = func(name string, v interface{}) (interface{}, error) {
		switch v.(type) {
		case *[]bool:
		case *[]float64:
		case *[]int64:
		case *[]string:
		default:
			return nil, errors.New("column '" + name + "' is array, it isnot support - []bool, []float64, []int64 and []string")
		}

		value := pq.Array(v)
		return value, nil
	}

	None Dialect = &dialect{
		name: "unknown", placeholder: Question,
		hasLastInsertID:  true,
		trueStr:          "true",
		falseStr:         "false",
		quoteChars:       "\"",
		makeArrayValuer:  makeArrayValuer,
		makeArrayScanner: makeArrayScanner,
	}
	Postgres Dialect = &dialect{
		name: "postgres", placeholder: Dollar,
		hasLastInsertID:  false,
		trueStr:          "true",
		falseStr:         "false",
		quoteChars:       "`",
		makeArrayValuer:  makePQArrayValuer,
		makeArrayScanner: makePQArrayScanner,
		handleError:      handlePQError,
	}
	Mysql Dialect = &dialect{
		name:             "mysql",
		placeholder:      Question,
		hasLastInsertID:  true,
		trueStr:          "1",
		falseStr:         "0",
		quoteChars:       "\"",
		makeArrayValuer:  makeArrayValuer,
		makeArrayScanner: makeArrayScanner,
	}
	MSSql Dialect = &dialect{
		name:             "mssql",
		placeholder:      Question,
		hasLastInsertID:  false,
		trueStr:          "true",
		falseStr:         "false",
		quoteChars:       "\"",
		makeArrayValuer:  makeArrayValuer,
		makeArrayScanner: makeArrayScanner,
	}
	Oracle Dialect = &dialect{
		name:             "oracle",
		placeholder:      Question,
		hasLastInsertID:  true,
		trueStr:          "true",
		falseStr:         "false",
		quoteChars:       "\"",
		makeArrayValuer:  makeArrayValuer,
		makeArrayScanner: makeArrayScanner,
	}
	DM Dialect = &dialect{
		name:             "dm",
		placeholder:      Question,
		hasLastInsertID:  true,
		trueStr:          "true",
		falseStr:         "false",
		quoteChars:       "\"",
		makeArrayValuer:  makeArrayStringValuer,
		makeArrayScanner: makeArrayScanner,
	}
)

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
