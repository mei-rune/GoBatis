package core

import (
	"strings"
	"unicode"
)

const (
	exprTokWS = iota
	exprTokIdent
	exprTokOp
	exprTokStr
)

func readToken(txt []rune) (tok int, end int) {
	c := txt[0]
	if unicode.IsSpace(c) {
		for i := 1; i < len(txt); i++ {
			c = txt[i]
			if !unicode.IsSpace(c) {
				return exprTokWS, i
			}
		}
		return exprTokWS, len(txt)
	}

	if unicode.IsDigit(c) ||
		('a' <= c && c <= 'z') ||
		('A' <= c && c <= 'Z') ||
		c == '.' ||
		c == '_' {

		for i := 1; i < len(txt); i++ {
			c = txt[i]
			if !unicode.IsDigit(c) &&
				!('a' <= c && c <= 'z') &&
				!('A' <= c && c <= 'Z') &&
				c != '.' &&
				c != '_' {

				return exprTokIdent, i
			}
		}
		return exprTokIdent, len(txt)
	}

	if c == '\'' || c == '"' {
		quote := c
		isEscape := false
		for i := 1; i < len(txt); i++ {
			c = txt[i]
			if c == '\\' {
				isEscape = !isEscape
				continue
			}
			if c == quote {
				if isEscape {
					isEscape = false
					continue
				}
				return exprTokStr, i + 1
			}
			isEscape = false
		}
		return exprTokStr, len(txt)
	}

	return exprTokOp, 1
}

func replaceAndOr(s string) string {
	runes := []rune(s)
	var sb strings.Builder

	for len(runes) > 0 {
		tok, next := readToken(runes)

		// fmt.Println(tok, string(runes[:next]))
		if tok != exprTokIdent {
			for idx := 0; idx < next; idx++ {
				sb.WriteRune(runes[idx])
			}
			runes = runes[next:]
			continue
		}

		switch strings.ToLower(string(runes[:next])) {
		case "and":
			sb.WriteString("&&")
		case "or":
			sb.WriteString("||")
		case "gt":
			sb.WriteString(">")
		case "gte":
			sb.WriteString(">=")
		case "lt":
			sb.WriteString("<")
		case "lte":
			sb.WriteString("<=")
		default:
			for idx := 0; idx < next; idx++ {
				sb.WriteRune(runes[idx])
			}
		}
		runes = runes[next:]
	}
	return sb.String()
}
