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
	exprTokFloat
	exprTokInt
)

func isExprIdent(c rune) bool {
	return unicode.IsDigit(c) ||
		('a' <= c && c <= 'z') ||
		('A' <= c && c <= 'Z') ||
		c == '.' || c == '_'
}

// func isExprIdentFirst(c rune) bool {
// 	return unicode.IsDigit(c) ||
// 		('a' <= c && c <= 'z') ||
// 		('A' <= c && c <= 'Z') ||
// 		c == '_'
// }

func readExprString(txt []rune) int {
	quote := txt[0]
	isEscape := false
	for i := 1; i < len(txt); i++ {
		c := txt[i]
		if c == '\\' {
			isEscape = !isEscape
			continue
		}
		if c == quote {
			if isEscape {
				isEscape = false
				continue
			}
			return i + 1
		}
		isEscape = false
	}
	return len(txt)
}

func readExprToken(txt []rune) (tok int, end int) {
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

	/////// begin
	// 我们无需区分 number 和 ident, 所以下面删除了

	// if unicode.IsDigit(c) {
	// 	i := 1
	// 	for ; ; i++ {
	// 		if i >= len(txt) {
	// 			return exprTokInt, len(txt)
	// 		}
	// 		if !unicode.IsDigit(txt[i]) {
	// 			break
	// 		}
	// 	}
	// 	if txt[i] == '.' {
	// 		for ; ; i++ {
	// 			if i >= len(txt) {
	// 				return exprTokFloat, len(txt)
	// 			}
	// 			if !unicode.IsDigit(txt[i]) {
	// 				break
	// 			}
	// 		}
	// 		return exprTokFloat, len(txt)
	// 	}
	// 	return exprTokInt, len(txt)
	// }
	// if c == '.' {
	// 	for i := 1; i < len(txt); i++ {
	// 		if !unicode.IsDigit(txt[i]) {
	// 			return exprTokIdent, i
	// 		}
	// 	}
	// 	return exprTokFloat, len(txt)
	// }

	// if isExprIdentFirst(c) {
	/////// end

	if isExprIdent(c) {
		for i := 1; i < len(txt); i++ {
			c = txt[i]
			if !isExprIdent(c) {
				return exprTokIdent, i
			}
		}
		return exprTokIdent, len(txt)
	}

	if c == '\'' || c == '"' {
		end := readExprString(txt)
		return exprTokStr, end
	}

	return exprTokOp, 1
}

func replaceAndOr(s string) string {
	runes := []rune(s)
	var sb strings.Builder

	for len(runes) > 0 {
		tok, next := readExprToken(runes)

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
