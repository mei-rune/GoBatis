package gobatis

import "strings"

type Dialect interface {
	Name() string
}

type dialect struct {
	name string
}

func (d *dialect) Name() string {
	return d.name
}

var (
	DbTypeNone     Dialect = &dialect{name: "unknown"}
	DbTypePostgres Dialect = &dialect{name: "postgres"}
	DbTypeMysql    Dialect = &dialect{name: "mysql"}
	DbTypeMSSql    Dialect = &dialect{name: "mssql"}
	DbTypeOracle   Dialect = &dialect{name: "oracle"}
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
