package core

import (
	"bufio"
	"io"
	"log"
	"strings"
)

const sqlCmdPrefix = "-- +gobatis "

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
	var buf strings.Builder
	scanner := bufio.NewScanner(r)

	statementEnded := false
	ignoreSemicolons := false

	isFirst := true
	for scanner.Scan() {
		text := scanner.Text()

		if line := strings.TrimSpace(text); strings.HasPrefix(line, "--") {
			// handle any gobatis-specific commands
			if strings.HasPrefix(line, sqlCmdPrefix) {
				cmd := strings.TrimSpace(line[len(sqlCmdPrefix):])
				switch cmd {
				case "StatementBegin":
					ignoreSemicolons = true
					break

				case "StatementEnd":
					statementEnded = (ignoreSemicolons == true)
					ignoreSemicolons = false
					break
				}
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

	return
}
