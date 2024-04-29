package core

import (
	"bufio"
	"io"
	"log"
	"strings"
)

func endsWithSemicolon(line string) bool {
	prev := ""
	scanner := bufio.NewScanner(strings.NewReader(line))
	scanner.Split(bufio.ScanWords)

	for scanner.Scan() {
		word := scanner.Text()
		if strings.HasPrefix(word, "--") {
			break
		}
		prev = word
	}

	return strings.HasSuffix(prev, ";")
}

// Split the given sql script into individual statements.
//
// The base case is to simply split on semicolons, as these
// naturally terminate a statement.
//
// However, more complex cases like pl/pgsql can have semicolons
// within a statement. For these cases, we provide the explicit annotations
// 'StatementBegin' and 'StatementEnd' to allow the script to
// tell us to ignore semicolons.
func splitSQLStatements(r io.Reader) (stmts []string) {
	return SplitSQLStatements(r, "gobatis")
}

func SplitSQLStatements(r io.Reader, prefix string) []string {
	var buf strings.Builder
	scanner := bufio.NewScanner(r)

	isFirst := true
	statementEnded := false
	ignoreSemicolons := false
	var stmts []string

	for scanner.Scan() {
		text := scanner.Text()

		if line := strings.TrimSpace(text); strings.HasPrefix(line, "--") {
			ss := strings.Fields(line)
			var cmd string
			if prefix == "" {
				if len(ss) == 2 {
					// -- +StatementBegin
					cmd = strings.TrimPrefix(ss[1], "+")
				}
			} else {
				if len(ss) == 3 && (ss[1] == prefix || ss[1] == "+"+prefix) {
					// -- +gobatis StatementBegin
					cmd = ss[2]
				} else if len(ss) == 2 {
					// -- +StatementBegin
					cmd = strings.TrimPrefix(ss[1], "+")
				}
			}

			// handle any gobatis-specific commands
			switch cmd {
			case "StatementBegin", "statementBegin":
				ignoreSemicolons = true
			case "StatementEnd", "statementEnd":
				statementEnded = (ignoreSemicolons == true)
				ignoreSemicolons = false
			}
		} else {
			if isFirst {
				isFirst = false
			} else {
				buf.WriteString("\n")
			}
			if _, err := buf.WriteString(text); err != nil {
				log.Fatalf("io err: %v", err)
			}
		}

		if !ignoreSemicolons && (statementEnded || endsWithSemicolon(text)) {
			statementEnded = false
			stmts = append(stmts, buf.String())
			buf.Reset()
		}
	}

	if err := scanner.Err(); err != nil {
		log.Fatalf("scanning migration: %v", err)
	}

	if buf.Len() > 0 {
		stmt := strings.TrimSpace(buf.String())
		if stmt != "" {
			stmts = append(stmts, buf.String())
		}
	}

	// diagnose likely migration script errors
	if ignoreSemicolons {
		log.Println("WARNING: saw '-- +gobatis StatementBegin' with no matching '-- +gobatis StatementEnd'")
	}

	return stmts
}

var SplitXORM = &TagSplit{
	Prefix: "xorm",
	Split: TagSplitForXORM,
}

var SplitDB = &TagSplit{
	Prefix: "db",
	Split: TagSplitForDb,
}

type TagSplit struct {
	Prefix string
	Split  func(s string, fieldName string) []string
}