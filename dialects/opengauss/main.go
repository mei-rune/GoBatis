package opengauss

import (
	"errors"
	"strings"

	pq "gitee.com/opengauss/openGauss-connector-go-pq" // openGauss
	"github.com/runner-mei/GoBatis/dialects"
)

func init() {
	dialects.SetHandleError(dialects.Opengauss.Name(), handleError)
	dialects.SetHandleArray(dialects.Opengauss.Name(), makePQArrayValuer, makePQArrayScanner)
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
		//  }, e: e}
		default:
			return &dialects.Error{Validations: []dialects.ValidationError{
				{Code: "PG." + pe.Code.Name(), Message: pe.Message, Columns: []string{pe.Column}},
			}, Err: e}
		}
	}
	return e
}

func makePQArrayValuer(v interface{}) (interface{}, error) {
	value := pq.Array(v)
	return value, nil
}
func makePQArrayScanner(name string, v interface{}) (interface{}, error) {
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
