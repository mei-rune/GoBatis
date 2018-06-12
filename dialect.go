package gobatis

import "strings"

type Dialect interface {
	Name() string
	Placeholder() PlaceholderFormat
}

type dialect struct {
	name        string
	placeholder PlaceholderFormat
}

func (d *dialect) Name() string {
	return d.name
}

func (d *dialect) Placeholder() PlaceholderFormat {
	return d.placeholder
}

var (
	DbTypeNone     Dialect = &dialect{name: "unknown", placeholder: Question}
	DbTypePostgres Dialect = &dialect{name: "postgres", placeholder: Dollar}
	DbTypeMysql    Dialect = &dialect{name: "mysql", placeholder: Question}
	DbTypeMSSql    Dialect = &dialect{name: "mssql", placeholder: Question}
	DbTypeOracle   Dialect = &dialect{name: "oracle", placeholder: Question}
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
