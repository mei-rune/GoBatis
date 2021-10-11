package gobatis

import (
	"github.com/runner-mei/GoBatis/dialects"
)

type Dialect = dialects.Dialect

func ToDbType(driverName string) Dialect {
	return dialects.New(driverName)
}

var (
	None  = dialects.None
	Postgres  = dialects.Postgres
	Mysql  = dialects.Mysql
	MSSql  = dialects.MSSql
	Oracle  = dialects.Oracle
)
