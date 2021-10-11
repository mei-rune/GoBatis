// This file is copy from https://github.com/Masterminds/squirrel
package gobatis

import (
	"bytes"
	"fmt"
	"strconv"
	"strings"
)

// PlaceholderFormat is the interface that wraps the ReplacePlaceholders method.
//
// ReplacePlaceholders takes a SQL statement and replaces each question mark
// placeholder with a (possibly different) SQL placeholder.
type PlaceholderFormat interface {
	ReplacePlaceholders(sql string) (string, error)
	Format(index int) string

	// Concat(fragments []string, bindParams Params, startIndex int) string
	Get(params SQLProvider) string
}

type SQLProvider interface {
	WithQuestion() string
	WithDollar() string
}

var (
	// Question is a PlaceholderFormat instance that leaves placeholders as
	// question marks.
	Question = questionFormat{}

	// Dollar is a PlaceholderFormat instance that replaces placeholders with
	// dollar-prefixed positional placeholders (e.g. $1, $2, $3).
	Dollar = dollarFormat{}
)

type questionFormat struct{}

func (_ questionFormat) ReplacePlaceholders(sql string) (string, error) {
	return sql, nil
}

func (_ questionFormat)	Format(index int) string {
	return "?"
}

func (_ questionFormat) Get(params SQLProvider) string {
	return params.WithQuestion()
}

// func (_ questionFormat) Concat(fragments []string, names Params, startIndex int) string {
// 	return strings.Join(fragments, "?")
// }

type dollarFormat struct{}

func (_ dollarFormat) ReplacePlaceholders(sql string) (string, error) {
	buf := &bytes.Buffer{}
	i := 0
	for {
		p := strings.Index(sql, "?")
		if p == -1 {
			break
		}

		if len(sql[p:]) > 1 && sql[p:p+2] == "??" { // escape ?? => ?
			buf.WriteString(sql[:p])
			buf.WriteString("?")
			if len(sql[p:]) == 1 {
				break
			}
			sql = sql[p+2:]
		} else {
			i++
			buf.WriteString(sql[:p])
			fmt.Fprintf(buf, "$%d", i)
			sql = sql[p+1:]
		}
	}

	buf.WriteString(sql)
	return buf.String(), nil
}

func (_ dollarFormat)	Format(index int) string {
	switch index {
	case 0:
		return "$1"
	case 1:
		return "$2"
	case 2:
		return "$3"
	case 3:
		return "4"
	case 4:
		return "$5"
	case 5:
		return "$6"
	case 6:
		return "$7"
	case 7:
		return "$8"
	case 8:
		return "$9"
	case 9:
		return "$10"
	case 10:
		return "$11"
	}
	return "$"+strconv.Itoa(index+1)
}

func (_ dollarFormat) Get(params SQLProvider) string {
	return params.WithDollar()
}

// func (_ dollarFormat) Concat(fragments []string, names Params, startIndex int) string {
// 	var sb strings.Builder
// 	sb.WriteString(fragments[0])
// 	for i := 1; i < len(fragments); i++ {
// 		sb.WriteString("$")
// 		sb.WriteString(strconv.Itoa(i + startIndex))
// 		sb.WriteString(fragments[i])
// 	}
// 	return sb.String()
// }

// // Placeholders returns a string with count ? placeholders joined with commas.
// func Placeholders(count int) string {
// 	if count < 1 {
// 		return ""
// 	}

// 	return strings.Repeat(",?", count)[1:]
// }

func Placeholders(format PlaceholderFormat, fragments []string, names Params, startIndex int) string {
	var sb strings.Builder
	sb.WriteString(fragments[0])
	for i := 1; i < len(fragments); i++ {
		sb.WriteString(format.Format(i + startIndex - 1))
		sb.WriteString(fragments[i])
	}
	return sb.String()
}