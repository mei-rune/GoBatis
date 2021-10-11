// This file is copy from https://github.com/Masterminds/squirrel
package gobatis

import (
	"strings"

	"github.com/runner-mei/GoBatis/dialects"
)

// PlaceholderFormat is the interface that wraps the ReplacePlaceholders method.
//
// ReplacePlaceholders takes a SQL statement and replaces each question mark
// placeholder with a (possibly different) SQL placeholder.
type PlaceholderFormat = dialects.PlaceholderFormat

type SQLProvider = dialects.SQLProvider

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
