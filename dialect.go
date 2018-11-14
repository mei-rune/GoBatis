package gobatis

import (
	"encoding/json"
	"errors"
	"strings"

	"github.com/lib/pq"
)

type Dialect interface {
	Name() string
	Placeholder() PlaceholderFormat
	InsertIDSupported() bool
	HandleError(error) error
	MakeArrayValuer(interface{}) (interface{}, error)
	MakeArrayScanner(string, interface{}) (interface{}, error)
}

type dialect struct {
	name            string
	placeholder     PlaceholderFormat
	hasLastInsertID bool
	handleError     func(e error) error

	makeArrayValuer  func(interface{}) (interface{}, error)
	makeArrayScanner func(string, interface{}) (interface{}, error)
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

	DbTypeNone     Dialect = &dialect{name: "unknown", placeholder: Question, hasLastInsertID: true, makeArrayValuer: makeArrayValuer, makeArrayScanner: makeArrayScanner}
	DbTypePostgres Dialect = &dialect{name: "postgres", placeholder: Dollar, hasLastInsertID: false, makeArrayValuer: makePQArrayValuer, makeArrayScanner: makePQArrayScanner, handleError: handlePQError}
	DbTypeMysql    Dialect = &dialect{name: "mysql", placeholder: Question, hasLastInsertID: true, makeArrayValuer: makeArrayValuer, makeArrayScanner: makeArrayScanner}
	DbTypeMSSql    Dialect = &dialect{name: "mssql", placeholder: Question, hasLastInsertID: false, makeArrayValuer: makeArrayValuer, makeArrayScanner: makeArrayScanner}
	DbTypeOracle   Dialect = &dialect{name: "oracle", placeholder: Question, hasLastInsertID: true, makeArrayValuer: makeArrayValuer, makeArrayScanner: makeArrayScanner}
)

func ToDbType(driverName string) Dialect {
	switch strings.ToLower(driverName) {
	case "postgres":
		return DbTypePostgres
	case "mysql":
		return DbTypeMysql
	case "mssql", "sqlserver":
		return DbTypeMSSql
	case "oracle", "ora":
		return DbTypeOracle
	default:
		return DbTypeNone
	}
}
