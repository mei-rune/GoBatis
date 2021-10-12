package gobatis

import (
	"errors"
	"strings"

	"github.com/runner-mei/GoBatis/dialects"
)

type Dialect = dialects.Dialect

func ToDbType(driverName string) Dialect {
	return dialects.New(driverName)
}

var (
	None     = dialects.None
	Postgres = dialects.Postgres
	Mysql    = dialects.Mysql
	MSSql    = dialects.MSSql
	Oracle   = dialects.Oracle
)

var ErrMultSQL = errors.New("mult sql is unsupported")

// ValidationError store the Message & Key of a validation error
type ValidationError = dialects.ValidationError

// Error store a error with validation errors
type Error = dialects.Error

func ErrForGenerateStmt(err error, msg string) error {
	return errors.New(msg + ": " + err.Error())
}

// PlaceholderFormat is the interface that wraps the ReplacePlaceholders method.
//
// ReplacePlaceholders takes a SQL statement and replaces each question mark
// placeholder with a (possibly different) SQL placeholder.
type PlaceholderFormat = dialects.PlaceholderFormat

type SQLPrintable = dialects.SQLPrintable

var (
	// Question is a PlaceholderFormat instance that leaves placeholders as
	// question marks.
	Question = dialects.Question

	// Dollar is a PlaceholderFormat instance that replaces placeholders with
	// dollar-prefixed positional placeholders (e.g. $1, $2, $3).
	Dollar = dialects.Dollar
)

func Placeholders(format PlaceholderFormat, fragments []string, names Params, startIndex int) string {
	var sb strings.Builder
	sb.WriteString(fragments[0])
	for i := 1; i < len(fragments); i++ {
		sb.WriteString(format.Format(i + startIndex - 1))
		sb.WriteString(fragments[i])
	}
	return sb.String()
}
