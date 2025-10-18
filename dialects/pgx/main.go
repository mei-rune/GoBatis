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
	dialects.SetHandleError(dialects.Postgres.Name(), handleError)
	dialects.SetHandleArray(dialects.Postgres.Name(), makePQArrayValuer, makePQArrayScanner)
}

func handleError(e error) error {
	if e == nil {
		return nil
	}

	if pe, ok := e.(*pgconn.PgError); ok {
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
				{Code: "GaussDB." + pe.Code, Message: pe.Message, Columns: []string{pe.ColumnName}},
			}, Err: e}
		}
	}
	return e
}

func makePQArrayValuer(v interface{}) (interface{}, error) {
	switch a := v.(type) {
	case []bool:
		var iv = pgtype.FlatArray[bool](a)
		return iv, nil
	case []float64:
		var iv = pgtype.FlatArray[float64](a)
		return iv, nil
	case []int64:
		var iv = pgtype.FlatArray[int64](a)
		return iv, nil
	case []string:
		var iv = pgtype.FlatArray[string](a)
		return iv, nil
	default:
		return nil, errors.New("must is array, it isnot support - []bool, []float64, []int64 and []string")
	}
}

func makePQArrayScanner(name string, v interface{}) (interface{}, error) {
	switch a := v.(type) {
	case *[]bool:
		var iv = pgtype.FlatArray[bool](*a)
		return &iv, nil
	case *[]float64:
		var iv = pgtype.FlatArray[float64](*a)
		return &iv, nil
	case *[]int64:
		var iv = pgtype.FlatArray[int64](*a)
		return &iv, nil
	case *[]string:
		var iv = pgtype.FlatArray[string](*a)
		return &iv, nil
	default:
		return nil, errors.New("column '" + name + "' is array, it isnot support - []bool, []float64, []int64 and []string")
	}
}
