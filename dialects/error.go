package dialects

import (
	"strings"

	"github.com/lib/pq"
)

// ValidationError store the Message & Key of a validation error
type ValidationError struct {
	Code, Message string
	Columns       []string
}

// Error store a error with validation errors
type Error struct {
	Validations []ValidationError
	e           error
}

func (err *Error) Error() string {
	return err.e.Error()
}

func handlePQError(e error) error {
	if e == nil {
		return nil
	}

	if pe, ok := e.(*pq.Error); ok {
		switch pe.Code {
		case "23505":
			detail := strings.TrimPrefix(strings.TrimPrefix(pe.Detail, "Key ("), "键值\"(")
			if pidx := strings.Index(detail, ")"); pidx > 0 {
				return &Error{Validations: []ValidationError{
					{Code: "unique_value_already_exists", Message: pe.Detail, Columns: strings.Split(detail[:pidx], ",")},
				}, e: e}
			}
		// case "23503":
		//  return &Error{Validations: []ValidationError{
		//    {Code: "PG.foreign_key_constraint", Message: pe.Message},
		//  }, e: e}
		default:
			return &Error{Validations: []ValidationError{
				{Code: "PG." + pe.Code.Name(), Message: pe.Message, Columns: []string{pe.Column}},
			}, e: e}
		}
	}
	return e
}
