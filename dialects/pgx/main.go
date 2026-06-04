package postgres

import (
	"errors"
	"strings"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
	_ "github.com/jackc/pgx/v5/stdlib"

	"github.com/runner-mei/GoBatis/dialects"
)

func init() {
	typeMap := pgtype.NewMap()

	// var a pgtype.Array[int64]
	// err := db.QueryRow("select '{1,2,3}'::bigint[]").Scan(m.SQLScanner(&a))

	dialects.SetHandleError(dialects.DriverPgx.DriverName(), handleError)
	dialects.SetHandleArray(dialects.DriverPgx.DriverName(), makePQArrayValuer(typeMap), makePQArrayScanner(typeMap))
}

func handleError(e error) error {
	if e == nil {
		return nil
	}

	pe, ok := e.(*pgconn.PgError)
	if !ok {
		o := errors.Unwrap(e)
		if o != nil {
			pe, ok = o.(*pgconn.PgError)
		}
	}

	if ok {
		switch pe.Code {
		case "23505":
			detail := strings.TrimPrefix(strings.TrimPrefix(pe.Detail, "Key ("), "键值\"(")
			if pidx := strings.Index(detail, ")"); pidx > 0 {
				return &dialects.Error{Validations: []dialects.ValidationError{
					{Code: "unique_value_already_exists", Message: pe.Detail, Columns: strings.Split(detail[:pidx], ",")},
				}, Err: e}
			}

		case "42P01":
			return dialects.ErrTableNotExists{
				Err:       e,
				Tablename: pe.TableName,
			}

		// case "23503":
		//  return &Error{Validations: []ValidationError{
		//    {Code: "PG.foreign_key_constraint", Message: pe.Message},
		//  }, e: e}
		default:
			return &dialects.Error{Validations: []dialects.ValidationError{
				{Code: "PG." + pe.Code, Message: pe.Message, Columns: []string{pe.ColumnName}},
			}, Err: e}
		}
	}
	return e
}

func makePQArrayValuer(tm *pgtype.Map) func(v interface{}) (interface{}, error) {
	return func(v interface{}) (interface{}, error) {
		// switch a := v.(type) {
		// case []bool:
		// 	var iv = pgtype.FlatArray[bool](a)
		// 	return iv, nil
		// case []float64:
		// 	var iv = pgtype.FlatArray[float64](a)
		// 	return iv, nil
		// case []int64:
		// 	var iv = pgtype.FlatArray[int64](a)
		// 	return iv, nil
		// case []string:
		// 	var iv = pgtype.FlatArray[string](a)
		// 	return iv, nil
		// default:
		// 	return nil, errors.New("must is array, it isnot support - []bool, []float64, []int64 and []string")
		// }

		return v, nil
	}
}

func makePQArrayScanner(tm *pgtype.Map) func(name string, v interface{}) (interface{}, error) {
	return func(name string, v interface{}) (interface{}, error) {
		return tm.SQLScanner(v), nil
	}
}
