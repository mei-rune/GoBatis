package dialects

import (
	"log"
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
	Err         error
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
	} else {
		e = Postgres.HandleError(e)
	}

	if _, ok := e.(ErrTableNotExists); ok {
		return ok
	}
	return false
}

func IsTableNotExistsWithDriverName(driverName string, e error) bool {
	return IsTableNotExists(New(driverName), e)
}

func AsTableNotExists(dialect Dialect, e error, ae *ErrTableNotExists) bool {
	if dialect != nil {
		e = dialect.HandleError(e)
	} else {
		e = Postgres.HandleError(e)
	}
	if te, ok := e.(ErrTableNotExists); ok {
		*ae = te
		return ok
	}
	return false
}

type ErrRecordAlreadyExists struct {
	Err error
}

func (e ErrRecordAlreadyExists) Error() string {
	return e.Err.Error()
}

func IsRecordAlreadyExists(dialect Dialect, e error) bool {
	if dialect != nil {
		e = dialect.HandleError(e)
	} else {
		e = Postgres.HandleError(e)
	}

	if ie, ok := e.(*Error); ok && len(ie.Validations) > 0 {
		for _, v := range ie.Validations {
			if v.Code == "unique_value_already_exists" {
				return true
			}
		}
		return true
	}
	return false
}
