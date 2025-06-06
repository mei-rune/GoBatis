package dialects

import (
	"log"
	"strings"

	"github.com/lib/pq"
)

func SetHandleError(driverName string, handleError func(error) error) {
	d := New(driverName)
	o, ok := d.(*dialect)
	if ok {
		o.handleError = handleError
	} else {
		log.Println("set handleError fail, dialect isnot *dialect type")
	}
}

// ValidationError store the Message & Key of a validation error
type ValidationError struct {
	Code, Message string
	Columns       []string
}

// Error store a error with validation errors
type Error struct {
	Validations []ValidationError
	Err           error
}

func (err *Error) Error() string {
	return err.Err.Error()
}

type ErrTableNotExists struct {
	Tablename string
	Err       error
}

func (e ErrTableNotExists) Error() string {
	return e.Err.Error()
}

func IsTableNotExists(dialect Dialect, e error) bool {
	if dialect != nil {
		e = dialect.HandleError(e)
		if _, ok := e.(ErrTableNotExists); ok {
			return ok
		}
	}
	if err, ok := e.(*pq.Error); ok && "42P01" == err.Code {
		return true
	}
	return false
}

func AsTableNotExists(dialect Dialect, e error, ae *ErrTableNotExists) bool {
	if dialect != nil {
		e = dialect.HandleError(e)
		if te, ok := e.(ErrTableNotExists); ok {
			*ae = te
			return ok
		}
	}
	if err, ok := e.(*pq.Error); ok && "42P01" == err.Code {
		ae.Tablename = err.Table
		ae.Err = err
		return true
	}
	return false
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
				}, Err: e}
			}

		case "42P01":
			return ErrTableNotExists{
				Err:       e,
				Tablename: pe.Table,
			}

		// case "23503":
		//  return &Error{Validations: []ValidationError{
		//    {Code: "PG.foreign_key_constraint", Message: pe.Message},
		//  }, e: e}
		default:
			return &Error{Validations: []ValidationError{
				{Code: "PG." + pe.Code.Name(), Message: pe.Message, Columns: []string{pe.Column}},
			}, Err: e}
		}
	}
	return e
}
