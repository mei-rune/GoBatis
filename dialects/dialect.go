package dialects

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"strings"
	"time"
)

const OdbcPrefix = "odbc_with_"

var newDialects []func(driverName string) Dialect

func RegisterDialectFactory(create func(driverName string) Dialect) {
	if create == nil {
		return
	}
	newDialects = append(newDialects, create)
}

func SetToDate(driverName string, toDate func(time.Time) interface{}) {
	d := New(driverName)
	o, ok := d.(*dialect)
	if ok {
		o.toDate = toDate
	} else {
		log.Println("set toDate fail, dialect isnot *dialect type")
	}
}

func New(driverName string) Dialect {
	driverName = strings.ToLower(driverName)
retrySwitch:
	switch driverName {
	case "kingbase", "kingbase8":
		return Kingbase
	case "postgres":
		return Postgres
	case "pgx", "pgx/v5":
		return Pgx
	case "opengauss":
		return Opengauss
	case "gaussdb":
		return GaussDB
	case "mysql":
		return Mysql
	case "mariadb":
		return Mariadb
	case "mssql", "sqlserver":
		return MSSql
	case "oracle", "ora":
		return Oracle
	case "dm":
		return DM
	case "sqlite":
		return Sqlite
	default:
		if strings.HasPrefix(driverName, OdbcPrefix) {
			driverName = strings.TrimPrefix(driverName, OdbcPrefix)
			goto retrySwitch
		}
		for _, newDialect := range newDialects {
			d := newDialect(driverName)
			if d != nil {
				return d
			}
		}
		return None
	}
	// panic("Unsupported database type: " + driverName)
}

type KeyMethodType int

const (
	KeyMethodLastInsertID KeyMethodType = iota
	KeyMethodReturning
	KeyMethodReturnInto
	KeyMethodOutput
)

type Dialect interface {
	Name() string
	Compatibility() string
	Quote(string) string
	BooleanStr(bool) string
	Placeholder() PlaceholderFormat
	KeyMethod() KeyMethodType
	HasAS() bool

	HandleError(error) error
	Limit(int64, int64) string

	ToDate(time.Time) interface{}
	ClobSupported() bool
	NewClob(*string) Clob
	BlobSupported() bool
	NewBlob(*[]byte) Blob
	MakeArrayValuer(interface{}) (interface{}, error)
	MakeArrayScanner(string, interface{}) (interface{}, error)
}

type dialect struct {
	name          string
	compatibility string
	placeholder   PlaceholderFormat
	keyMethod     KeyMethodType
	hasAS         bool
	quoteFunc     func(string) string
	trueStr       string
	falseStr      string
	handleError   func(error) error
	limitFunc     func(offset, limit int64) string

	toDate           func(time.Time) interface{}
	clobSupported    bool
	newClob          func(*string) Clob
	blobSupported    bool
	newBlob          func(*[]byte) Blob
	makeArrayValuer  func(interface{}) (interface{}, error)
	makeArrayScanner func(string, interface{}) (interface{}, error)
}

func (d *dialect) Compatibility() string {
	if d.compatibility == "" {
		return d.name
	}
	return d.compatibility
}

func (d *dialect) Quote(name string) string {
	if d.quoteFunc == nil {
		return name
	}
	return d.quoteFunc(name)
}

func (d *dialect) BooleanStr(b bool) string {
	if b {
		return d.trueStr
	}
	return d.falseStr
}

func (d *dialect) ToDate(t time.Time) interface{} {
	if d.toDate != nil {
		return d.toDate(t)
	}
	return t
}

func (d *dialect) Limit(offset, limit int64) string {
	if d.limitFunc != nil {
		return d.limitFunc(offset, limit)
	}
	return limitByOffsetLimit(offset, limit)
}

func limitByLimitMN(offset, limit int64) string {
	if offset > 0 {
		if limit > 0 {
			return fmt.Sprintf(" LIMIT %d, %d ", offset, limit)
		}
		return fmt.Sprintf(" OFFSET %d ", offset)
	}
	if limit > 0 {
		return fmt.Sprintf(" LIMIT %d ", limit)
	}
	return ""
}

func limitByOffsetLimit(offset, limit int64) string {
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

func limitByFetchNext(offset, limit int64) string {
	if offset > 0 {
		if limit > 0 {
			return fmt.Sprintf(" OFFSET %d ROWS FETCH NEXT %d ROWS ONLY ", offset, limit)
		}
		return fmt.Sprintf(" OFFSET %d ROWS ", offset)
	}
	if limit > 0 {
		return fmt.Sprintf(" OFFSET 0 ROWS FETCH NEXT %d ROWS ONLY ", limit)
	}
	return ""
}

func (d *dialect) Name() string {
	return d.name
}

func (d *dialect) Placeholder() PlaceholderFormat {
	return d.placeholder
}

func (d *dialect) KeyMethod() KeyMethodType {
	return d.keyMethod
}

func (d *dialect) HasAS() bool {
	return d.hasAS
}

func (d *dialect) HandleError(e error) error {
	if d.handleError == nil {
		return e
	}
	return d.handleError(e)
}

func (d *dialect) ClobSupported() bool {
	return d.clobSupported
}

func (d *dialect) NewClob(addr *string) Clob {
	return d.newClob(addr)
}

func (d *dialect) BlobSupported() bool {
	return d.blobSupported
}

func (d *dialect) NewBlob(addr *[]byte) Blob {
	return d.newBlob(addr)
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

	None Dialect = &dialect{
		name: "unknown", placeholder: Question,
		keyMethod: KeyMethodLastInsertID,
		hasAS:     true,
		trueStr:   "true",
		falseStr:  "false",
		quoteFunc: defaultQuote,

		newClob:          newClob,
		newBlob:          newBlob,
		makeArrayValuer:  makeArrayValuer,
		makeArrayScanner: makeArrayScanner,
	}
	Kingbase Dialect = &dialect{
		name:          "kingbase",
		compatibility: "postgres",
		placeholder:   Dollar,
		keyMethod:     KeyMethodReturning,
		hasAS:         true,
		trueStr:       "true",
		falseStr:      "false",
		quoteFunc:     defaultQuote,
		newClob:       newClob,
		newBlob:       newBlob,

		makeArrayValuer:  makeArrayValuerForUnsupport("please import \"github.com/runner-mei/GoBatis/dialects/kingbase\""),
		makeArrayScanner: makeArrayScanForUnsupport("please import \"github.com/runner-mei/GoBatis/dialects/kingbase\""),
		handleError:      makeHandleErrorForUnsupport("please import \"github.com/runner-mei/GoBatis/dialects/kingbase\""),
	}
	Postgres Dialect = &dialect{
		name:             "postgres",
		placeholder:      Dollar,
		keyMethod:        KeyMethodReturning,
		hasAS:            true,
		trueStr:          "true",
		falseStr:         "false",
		quoteFunc:        defaultQuote,
		newClob:          newClob,
		newBlob:          newBlob,
		makeArrayValuer:  makeArrayValuerForUnsupport("please import \"github.com/runner-mei/GoBatis/dialects/pq\" or \"github.com/runner-mei/GoBatis/dialects/pgx\""),
		makeArrayScanner: makeArrayScanForUnsupport("please import \"github.com/runner-mei/GoBatis/dialects/pq\" or \"github.com/runner-mei/GoBatis/dialects/pgx\""),
		handleError:      makeHandleErrorForUnsupport("please import \"github.com/runner-mei/GoBatis/dialects/pq\" or \"github.com/runner-mei/GoBatis/dialects/pgx\""),
	}

	Pgx Dialect = &dialect{
		name:             "pgx",
		compatibility:    "postgres",
		placeholder:      Dollar,
		keyMethod:        KeyMethodReturning,
		hasAS:            true,
		trueStr:          "true",
		falseStr:         "false",
		quoteFunc:        defaultQuote,
		newClob:          newClob,
		newBlob:          newBlob,
		makeArrayValuer:  makeArrayValuerForUnsupport("please import \"github.com/runner-mei/GoBatis/dialects/pq\" or \"github.com/runner-mei/GoBatis/dialects/pgx\""),
		makeArrayScanner: makeArrayScanForUnsupport("please import \"github.com/runner-mei/GoBatis/dialects/pq\" or \"github.com/runner-mei/GoBatis/dialects/pgx\""),
		handleError:      makeHandleErrorForUnsupport("please import \"github.com/runner-mei/GoBatis/dialects/pq\" or \"github.com/runner-mei/GoBatis/dialects/pgx\""),
	}
	Opengauss Dialect = &dialect{
		name:             "opengauss",
		compatibility:    "postgres",
		placeholder:      Dollar,
		keyMethod:        KeyMethodReturning,
		hasAS:            true,
		trueStr:          "true",
		falseStr:         "false",
		quoteFunc:        defaultQuote,
		newClob:          newClob,
		newBlob:          newBlob,
		makeArrayValuer:  makeArrayValuerForUnsupport("please import \"github.com/runner-mei/GoBatis/dialects/opengauss\""),
		makeArrayScanner: makeArrayScanForUnsupport("please import \"github.com/runner-mei/GoBatis/dialects/opengauss\""),
		handleError:      makeHandleErrorForUnsupport("please import \"github.com/runner-mei/GoBatis/dialects/opengauss\""),
	}
	GaussDB Dialect = &dialect{
		name:             "gaussdb",
		compatibility:    "postgres",
		placeholder:      Dollar,
		keyMethod:        KeyMethodReturning,
		hasAS:            true,
		trueStr:          "true",
		falseStr:         "false",
		quoteFunc:        defaultQuote,
		newClob:          newClob,
		newBlob:          newBlob,
		makeArrayValuer:  makeArrayValuerForUnsupport("please import \"github.com/runner-mei/GoBatis/dialects/gaussdb\""),
		makeArrayScanner: makeArrayScanForUnsupport("please import \"github.com/runner-mei/GoBatis/dialects/gaussdb\""),
		handleError:      makeHandleErrorForUnsupport("please import \"github.com/runner-mei/GoBatis/dialects/gaussdb\""),
	}

	Mysql Dialect = &dialect{
		name:             "mysql",
		placeholder:      Question,
		keyMethod:        KeyMethodLastInsertID,
		hasAS:            false,
		trueStr:          "1",
		falseStr:         "0",
		quoteFunc:        defaultMysqlQuote,
		newClob:          newClob,
		newBlob:          newBlob,
		makeArrayValuer:  makeArrayValuer,
		makeArrayScanner: makeArrayScanner,
		limitFunc:        limitByLimitMN,
	}
	Mariadb Dialect = &dialect{
		name:             "mariadb",
		compatibility:    "mysql",
		placeholder:      Question,
		keyMethod:        KeyMethodReturning,
		hasAS:            false,
		trueStr:          "1",
		falseStr:         "0",
		quoteFunc:        defaultMysqlQuote,
		newClob:          newClob,
		newBlob:          newBlob,
		makeArrayValuer:  makeArrayValuer,
		makeArrayScanner: makeArrayScanner,
		limitFunc:        limitByLimitMN,
	}
	MSSql Dialect = &dialect{
		name:             "mssql",
		placeholder:      Question,
		keyMethod:        KeyMethodOutput,
		hasAS:            true,
		trueStr:          "true",
		falseStr:         "false",
		quoteFunc:        defaultQuote,
		newClob:          newClob,
		newBlob:          newBlob,
		makeArrayValuer:  makeArrayValuer,
		makeArrayScanner: makeArrayScanner,
		limitFunc:        limitByFetchNext,
	}
	Oracle Dialect = &dialect{
		name:             "oracle",
		placeholder:      Question,
		keyMethod:        KeyMethodReturnInto, // 它是支持 output 子句的，有空支持一下
		hasAS:            true,
		trueStr:          "1",
		falseStr:         "0",
		quoteFunc:        defaultOracleQuote,
		newClob:          newClob,
		newBlob:          newBlob,
		makeArrayValuer:  makeArrayValuer,
		makeArrayScanner: makeArrayScanner,
		limitFunc:        limitByOffsetLimit,
	}
	Sqlite Dialect = &dialect{
		name:             "sqlite",
		placeholder:      Question,
		keyMethod:        KeyMethodReturning,
		hasAS:            true,
		trueStr:          "1",
		falseStr:         "0",
		quoteFunc:        defaultQuote,
		newClob:          newClob,
		newBlob:          newBlob,
		makeArrayValuer:  makeArrayValuer,
		makeArrayScanner: makeArrayScanner,
		limitFunc:        limitByOffsetLimit,
	}
	DM Dialect = &dialect{
		name:             "dm",
		compatibility:    "oracle",
		placeholder:      Question,
		keyMethod:        KeyMethodReturnInto,
		hasAS:            true,
		trueStr:          "1",
		falseStr:         "0",
		quoteFunc:        defaultDMQuote,
		clobSupported:    true,
		newClob:          newDMClob,
		blobSupported:    true,
		newBlob:          newDMBlob,
		makeArrayValuer:  makeArrayStringValuer,
		makeArrayScanner: makeArrayScanner,
		limitFunc:        limitByOffsetLimit,
	}
)

func defaultQuote(name string) string {
	if name == "select" {
		return "\"select\""
	}
	if name == "from" {
		return "\"from\""
	}
	if name == "order" {
		return "\"order\""
	}
	if name == "group" {
		return "\"group\""
	}
	return name
}

func defaultDMQuote(name string) string {
	// if name == "type" {
	// 	return "\"type\""
	// }
	if name == "interval" {
		return "\"interval\""
	}
	if name == "match" {
		return "\"match\""
	}
	if name == "model" {
		return "\"model\""
	}

	return name
}

func defaultMysqlQuote(name string) string {
	if name == "interval" {
		return "`interval`"
	}
	if name == "match" {
		return "`match`"
	}
	return name
}

func defaultOracleQuote(name string) string {
	if name == "interval" {
		return "\"interval\""
	}
	if name == "match" {
		return "\"match\""
	}
	if name == "model" {
		return "\"model\""
	}
	return name
}

var createDmClob func(*string) Clob
var createDmBlob func(*[]byte) Blob

func SetNewDMClob(create func(*string) Clob) {
	createDmClob = create
}

func SetNewDMBlob(create func(*[]byte) Blob) {
	createDmBlob = create
}

func newDMClob(addr *string) Clob {
	if createDmClob != nil {
		return createDmClob(addr)
	}
	return newClob(addr)
}

func newDMBlob(addr *[]byte) Blob {
	if createDmBlob != nil {
		return createDmBlob(addr)
	}
	return newBlob(addr)
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

func SetHandleArray(driverName string, makeArrayValuer func(interface{}) (interface{}, error), makeArrayScanner func(string, interface{}) (interface{}, error)) {
	d := New(driverName)
	o, ok := d.(*dialect)
	if ok {
		o.makeArrayValuer = makeArrayValuer
		o.makeArrayScanner = makeArrayScanner
	} else {
		log.Println("set handleError fail, dialect isnot *dialect type")
	}
}

func makeArrayValuerForUnsupport(message string) func(v interface{}) (interface{}, error) {
	err := errors.New(message)
	return func(v interface{}) (interface{}, error) {
		return nil, err
	}
}

func makeArrayScanForUnsupport(message string) func(name string, v interface{}) (interface{}, error) {
	err := errors.New(message)
	return func(name string, v interface{}) (interface{}, error) {
		return nil, err
	}
}

func makeHandleErrorForUnsupport(message string) func(error) error {
	msgErr := errors.New(message)
	return func(err error) error {
		panic(msgErr)
		return err
	}
}
