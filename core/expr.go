package core

import (
	"errors"
	"reflect"
	"strconv"
	"strings"
	"time"
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

func isEmptyString(args ...interface{}) (bool, error) {
	isLike := false
	if len(args) != 1 {
		if len(args) != 2 {
			return false, errors.New("args.len() isnot 1 or 2")
		}
		rv := reflect.ValueOf(args[1])
		if rv.Kind() != reflect.Bool {
			return false, errors.New("args[1] isnot bool type")
		}
		isLike = rv.Bool()
	}
	if args[0] == nil {
		return true, nil
	}
	rv := reflect.ValueOf(args[0])
	if rv.Kind() == reflect.String {
		if rv.Len() == 0 {
			return true, nil
		}
		if isLike {
			return rv.String() == "%" || rv.String() == "%%", nil
		}
		return false, nil
	}
	return false, errors.New("value isnot string")
}

func isNil(args ...interface{}) (bool, error) {
	for idx, arg := range args {
		rv := reflect.ValueOf(arg)
		if rv.Kind() != reflect.Ptr &&
			rv.Kind() != reflect.Map &&
			rv.Kind() != reflect.Slice &&
			rv.Kind() != reflect.Interface {
			return false, errors.New("isNil: args(" + strconv.FormatInt(int64(idx), 10) + ") isnot ptr - " + rv.Kind().String())
		}

		if !rv.IsNil() {
			return false, nil
		}
	}

	return true, nil
}

func isNull(args ...interface{}) (interface{}, error) {
	if len(args) == 0 {
		return nil, errors.New("isnull() args is empty")
	}

	b, err := isNil(args...)
	if err != nil {
		return nil, err
	}

	return b, nil
}

func isZero(args ...interface{}) (bool, error) {
	if len(args) == 0 {
		return false, errors.New("isZero() args is empty")
	}

	switch v := args[0].(type) {
	case time.Time:
		return v.IsZero(), nil
	case *time.Time:
		if v == nil {
			return true, nil
		}
		return v.IsZero(), nil
	case int:
		return v == 0, nil
	case int64:
		return v == 0, nil
	case int32:
		return v == 0, nil
	case int16:
		return v == 0, nil
	case int8:
		return v == 0, nil
	case uint:
		return v == 0, nil
	case uint64:
		return v == 0, nil
	case uint32:
		return v == 0, nil
	case uint16:
		return v == 0, nil
	case uint8:
		return v == 0, nil
	default:
		if zero, ok := args[0].(interface{ IsZero() bool}); ok {
			return zero.IsZero(), nil
		}
	}

	return false, nil
}

func isNotNull(args ...interface{}) (interface{}, error) {
	if len(args) == 0 {
		return nil, errors.New("isnotnull() args is empty")
	}

	for _, arg := range args {
		rv := reflect.ValueOf(arg)
		if rv.Kind() != reflect.Ptr &&
			rv.Kind() != reflect.Map &&
			rv.Kind() != reflect.Slice &&
			rv.Kind() != reflect.Interface {
			continue
		}

		if rv.IsNil() {
			return false, nil
		}
	}

	return true, nil
}

func Mapget(args ...interface{}) (interface{}, error) {
	if len(args) != 2 {
		return nil, errors.New("mapget() args isnot 2")
	}

	m, ok := args[0].(map[string]interface{})
	if !ok {
		rv := reflect.ValueOf(args[0])
		if	rv.Kind() != reflect.Map {
			return nil, errors.New("args[0] isnot map")
		}

		key := reflect.ValueOf(args[1])
		value := rv.MapIndex(key)
		return value.Interface(), nil
	}

	key := args[1].(string)
	return m[key], nil
}