package gobatis

import "strings"

type Dialect interface {
	Name() string
	Placeholder() PlaceholderFormat
	InsertIDSupported() bool
}

type dialect struct {
	name            string
	placeholder     PlaceholderFormat
	hasLastInsertID bool
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

var (
	DbTypeNone     Dialect = &dialect{name: "unknown", placeholder: Question, hasLastInsertID: true}
	DbTypePostgres Dialect = &dialect{name: "postgres", placeholder: Dollar, hasLastInsertID: false}
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
