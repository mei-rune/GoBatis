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

	exprTokAnd
	exprTokOr
	exprTokGt
	exprTokGte
	exprTokLt
	exprTokLte
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

		const (
			state_a = iota
			state_an
			state_and
			state_o
			state_or
			state_g
			state_gt
			state_gte
			state_l
			state_lt
			state_lte
			state_unknown
		)
func stateToIdent(state int) int {
	switch state {
	case state_and:
		return exprTokAnd
	case state_or:
		return exprTokOr
	case state_gt:
		return exprTokGt
	case state_gte:
		return exprTokGte
	case state_lt:
		return exprTokLt
	case state_lte:
		return exprTokLte
	}
	return exprTokIdent
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
		i := 1
		state := state_unknown
		switch c {
		case 'A', 'a':
			state = state_a 
		case 'O', 'o':
			state = state_o
		case 'G', 'g':
			state = state_g
		case 'L', 'l':
			state = state_l
		default:
			state = state_unknown
			goto unknownstate
		}
		for ; ; i++ {
			if i >= len(txt) {
				return stateToIdent(state), i
			}

			c = txt[i]

			switch c {
			case 'n', 'N':
				if state == state_a {
					state = state_an
				} else {
					state = state_unknown
					goto unknownstate
				}
			case 'd', 'D':
				if state == state_an {
					state = state_and
				} else {
					state = state_unknown
					goto unknownstate
				}
			case 'r', 'R':
				if state == state_o {
					state = state_or
				} else {
					state = state_unknown
					goto unknownstate
				}

			case 't', 'T':
				if state == state_g {
					state = state_gt
				} else if state == state_l {
					state = state_lt
				} else {
					state = state_unknown
					goto unknownstate
				}

			case 'e', 'E':
				if state == state_gt {
					state = state_gte
				} else if state == state_lt {
					state = state_lte
				} else {
					state = state_unknown
					goto unknownstate
				}
			default:
				if !isExprIdent(c) {
					return stateToIdent(state), i
				}
				state = state_unknown
				goto unknownstate
			}
		}
		
		unknownstate:
		for ; i < len(txt); i++ {
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

		switch tok {
		case exprTokAnd:
			sb.WriteString("&&")
		case exprTokOr:
			sb.WriteString("||")
		case exprTokGt:
			sb.WriteString(">")
		case exprTokGte:
			sb.WriteString(">=")
		case exprTokLt:
			sb.WriteString("<")
		case exprTokLte:
			sb.WriteString("<=")
		default:
			// fmt.Println(tok, string(runes[:next]))
			// if tok != exprTokIdent {
				for idx := 0; idx < next; idx++ {
					sb.WriteRune(runes[idx])
				}
			// } else {
			// 	switch strings.ToLower(string(runes[:next])) {
			// 	case "and":
			// 		sb.WriteString("&&")
			// 	case "or":
			// 		sb.WriteString("||")
			// 	case "gt":
			// 		sb.WriteString(">")
			// 	case "gte":
			// 		sb.WriteString(">=")
			// 	case "lt":
			// 		sb.WriteString("<")
			// 	case "lte":
			// 		sb.WriteString("<=")
			// 	default:
			// 		for idx := 0; idx < next; idx++ {
			// 			sb.WriteRune(runes[idx])
			// 		}
			// 	}
			// }
		}
		runes = runes[next:]
	}
	return sb.String()
}
