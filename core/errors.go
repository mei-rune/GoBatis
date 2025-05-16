package core

import (
	"errors"

	"github.com/runner-mei/GoBatis/dialects"
)

var ErrNotFound = errors.New("not found")
var ErrAlreadyTx = errors.New("open tx fail: already in a tx")

type SqlError struct {
	Err error
	SQL string
}

func (e SqlError) Unwrap() error {
	return e.Err
}

func (e *SqlError) Error() string {
	return e.Err.Error()
}

type errTx struct {
	method string
	inner  error
}

func (e errTx) Unwrap() error {
	return e.inner
}

func (e errTx) Error() string {
	return e.method + " tx fail: " + e.inner.Error()
}

func IsTxError(e error, method string, methods ...string) bool {
	txErr, ok := e.(errTx)
	if !ok {
		return false
	}
	if len(methods) == 0 {
		return txErr.method == method
	}
	if txErr.method == method {
		return true
	}
	for _, m := range methods {
		if m == txErr.method {
			return true
		}
	}
	return false
}

type statementAlreadyExists struct {
	id string
}

func (e statementAlreadyExists) Error() string {
	return "statement '" + e.id + "' already exists"
}

func ErrStatementAlreadyExists(id string) error {
	return statementAlreadyExists{id: id}
}

type ErrTableNotExists = dialects.ErrTableNotExists

func IsTableNotExists(dialect dialects.Dialect, e error) bool {
	return dialects.IsTableNotExists(dialect, e)
}

func AsTableNotExists(dialect dialects.Dialect, e error, ae *ErrTableNotExists) bool {
	return dialects.AsTableNotExists(dialect, e, ae)
}
