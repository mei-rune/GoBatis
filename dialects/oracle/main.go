package oracle

import (
	"strings"

	"github.com/runner-mei/GoBatis/dialects"
	_ "github.com/sijms/go-ora/v2"
)

func init() {
	dialects.SetHandleError(dialects.DriverOracle.Name(), handleError)
}

func handleError(e error) error {
	if e == nil {
		return nil
	}

	msg := e.Error()
	if strings.Contains(msg, "ORA-00942") {
		return dialects.ErrTableNotExists{
			Err: e,
		}
	}

	if strings.Contains(msg, "ORA-00001") {
		return &dialects.Error{Validations: []dialects.ValidationError{
			{Code: "unique_value_already_exists", Message: msg},
		}, Err: e}
	}

	return e
}
