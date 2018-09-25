package gobatis

import "strings"

type Dialect interface {
	Name() string
	Placeholder() PlaceholderFormat
	InsertIDSupported() bool
	HandleError(error) error
}

type dialect struct {
	name            string
	placeholder     PlaceholderFormat
	hasLastInsertID bool
	handleError     func(e error) error
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

var (
	DbTypeNone     Dialect = &dialect{name: "unknown", placeholder: Question, hasLastInsertID: true}
	DbTypePostgres Dialect = &dialect{name: "postgres", placeholder: Dollar, hasLastInsertID: false, handleError: handlePQError}
	DbTypeMysql    Dialect = &dialect{name: "mysql", placeholder: Question, hasLastInsertID: true}
	DbTypeMSSql    Dialect = &dialect{name: "mssql", placeholder: Question, hasLastInsertID: false}
	DbTypeOracle   Dialect = &dialect{name: "oracle", placeholder: Question, hasLastInsertID: true}
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
