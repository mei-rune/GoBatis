package kingbase

import (
	"strings"

	pq "gitee.com/runner.mei/gokb" // gokb
	"github.com/runner-mei/GoBatis/dialects"
)

func init() {
	dialects.SetHandleError(dialects.Kingbase.Name(), handleError)
}

func handleError(e error) error {
	if e == nil {
		return nil
	}

	if pe, ok := e.(*pq.Error); ok {
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
				Tablename: pe.Table,
			}

		// case "23503":
		//  return &Error{Validations: []ValidationError{
		//    {Code: "PG.foreign_key_constraint", Message: pe.Message},
		//  }, Err: e}
		default:
			return &dialects.Error{Validations: []dialects.ValidationError{
				{Code: "PG." + pe.Code.Name(), Message: pe.Message, Columns: []string{pe.Column}},
			}, Err: e}
		}
	}
	return e
}
