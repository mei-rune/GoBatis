package goparser2

import (
	"errors"
	"strings"
	"unicode"

	gobatis "github.com/runner-mei/GoBatis"
	"github.com/runner-mei/GoBatis/dialects"
)

type SQL struct {
	Filters []gobatis.Filter
	// OrderBy string
}

type Dialect struct {
	DialectNames []string
	SQL          string
}

func (d Dialect) ToGoLiteral() string {
	var s strings.Builder
	for idx, a := range d.DialectNames {
		if idx > 0 {
			s.WriteString(", ")
		}
		s.WriteString(toGoLiteral(a))
	}
	return s.String()
}

func toGoLiteral(dialect string) string {
	switch dialect {
	case "kingbase", "kingbase8":
		return "gobatis.Kingbase"
	case "postgres":
		return "gobatis.Postgres"
	case "opengauss":
		return "gobatis.Opengauss"
	case "gaussdb":
		return "gobatis.GaussDB"
	case "mysql":
		return "gobatis.Mysql"
	case "mssql", "sqlserver":
		return "gobatis.MSSql"
	case "oracle", "ora":
		return "gobatis.Oracle"
	case "dm":
		return "gobatis.DM"
	}
	return "gobatis.NewDialect(\"" + dialect + "\")"
}

type SQLConfig struct {
	Description string
	Reference   *struct {
		Interface string
		Method    string
	}
	StatementType string
	DefaultSQL    string
	Options       map[string]string
	Dialects      []Dialect
	RecordType    string
	SQL           SQL
}

func parseComments(comments []string, prefix string, dbCompatibility bool) (*SQLConfig, error) {
	if len(comments) == 0 {
		return &SQLConfig{}, nil
	}
	sections := splitByEmptyLine(comments)
	if len(sections) == 0 {
		return &SQLConfig{}, nil
	}

	if prefix != "" {
		prefix = strings.ToLower(prefix)
		if !strings.HasPrefix(prefix, "@") {
			prefix = "@" + prefix
		}

		if !strings.HasSuffix(prefix, ".") {
			prefix = prefix + "."
		}
	}

	var sqlCfg = &SQLConfig{}
	sqlCfg.Description = strings.TrimSpace(sections[0])
	for idx := 1; idx < len(sections); idx++ {
		tag, value := splitFirstBySpace(sections[idx])
		if prefix != "" {
			if !strings.HasPrefix(strings.ToLower(tag), prefix) {
				// skip annocations of other app, examples as swagg
				continue
			}
			tag = "@" + tag[len(prefix):]
		}

		value = strings.TrimSpace(value)
		if value == "" {
			if strings.HasPrefix(tag, "@") && strings.Contains(tag, "(") && strings.HasSuffix(tag, ")") {
				// 为了兼容 gogenv1 的格式,
				continue
			}

			if idx == (len(sections) - 1) {
				sqlCfg.DefaultSQL = tag
				break
			}
			return nil, errors.New("'" + sections[idx] + "' is syntex error")
		}
		switch strings.ToLower(tag) {
		case "@reference":
			ss := strings.Split(value, ".")
			if len(ss) != 2 || ss[0] == "" || ss[1] == "" {
				return nil, errors.New("'" + sections[idx] + "' is syntex error - reference must is 'InterfaceName.MethodName'")
			}
			sqlCfg.Reference = &struct {
				Interface string
				Method    string
			}{Interface: ss[0],
				Method: ss[1]}
		case "@type":
			sqlCfg.StatementType = strings.ToLower(strings.TrimSpace(value))
		case "@option":
			optKey, optValue := splitFirstBySpace(value)
			if sqlCfg.Options == nil {
				sqlCfg.Options = map[string]string{strings.TrimSpace(optKey): strings.TrimSpace(optValue)}
			} else {
				sqlCfg.Options[strings.TrimSpace(optKey)] = strings.TrimSpace(optValue)
			}
		case "@default":
			sqlCfg.DefaultSQL = strings.TrimSpace(value)
		case "@record_type":
			sqlCfg.RecordType = strings.TrimSpace(value)
		case "@filter":
			filter, err := splitFilter(strings.TrimSpace(value))
			if err != nil {
				return nil, err
			}
			sqlCfg.SQL.Filters = append(sqlCfg.SQL.Filters, filter)
		case "@orderby":
			return nil, errors.New("@orderby is deleted")
			// sqlCfg.SQL.OrderBy = strings.TrimSpace(value)
		default:
			if s := strings.TrimSpace(sections[idx]); strings.HasPrefix(s, "@") && strings.HasSuffix(s, ")") && strings.Contains(tag, "(") {
				break
			}

			a := Dialect{SQL: strings.TrimSpace(value)}
			tags := strings.Split(strings.TrimPrefix(tag, "@"), ",")
			for _, tagstr := range tags {
				tagstr = strings.TrimSpace(tagstr)
				if tagstr == "" {
					continue
				}
				a.DialectNames = append(a.DialectNames, tagstr)
			}
			sqlCfg.Dialects = append(sqlCfg.Dialects, a)
		}
	}

	if dbCompatibility {
		findDialect := func(list []Dialect, name string) *Dialect {
			for idx := range list {
				for _, dialectName := range list[idx].DialectNames {
					if dialectName == name {
						return &list[idx]
					}
				}
			}
			return nil
		}
		if pg := findDialect(sqlCfg.Dialects, dialects.Postgres.Name()); pg != nil {
			if d := findDialect(sqlCfg.Dialects, dialects.Kingbase.Name()); d == nil {
				pg.DialectNames = append(pg.DialectNames, dialects.Kingbase.Name())
			}
			if d := findDialect(sqlCfg.Dialects, dialects.Opengauss.Name()); d == nil {
				pg.DialectNames = append(pg.DialectNames, dialects.Opengauss.Name())
			}
			if d := findDialect(sqlCfg.Dialects, dialects.GaussDB.Name()); d == nil {
				pg.DialectNames = append(pg.DialectNames, dialects.GaussDB.Name())
			}
		}
		if ora := findDialect(sqlCfg.Dialects, dialects.Oracle.Name()); ora != nil {
			if d := findDialect(sqlCfg.Dialects, dialects.DM.Name()); d == nil {
				ora.DialectNames = append(ora.DialectNames, dialects.DM.Name())
			}
		}
	}

	if sqlCfg.Reference != nil {
		//if strings.ToLower(sqlCfg.DefaultSQL) != "" || len(sqlCfg.Dialects) != 0 {
		if len(sqlCfg.SQL.Filters) != 0 || strings.ToLower(sqlCfg.DefaultSQL) != "" || len(sqlCfg.Dialects) != 0 {
			return nil, errors.New("sql statement or filters is forbidden while reference is exists")
		}
	}

	if len(sqlCfg.SQL.Filters) != 0 {
		if sqlCfg.DefaultSQL != "" || len(sqlCfg.Dialects) != 0 {
			return nil, errors.New("sql statement is forbidden while filters is exists")
		}

		if sqlCfg.StatementType != "" {
			if sqlCfg.StatementType != "select" && sqlCfg.StatementType != "delete" {
				return nil, errors.New("filter is forbidden while statement type is " + sqlCfg.StatementType)
			}
		}
	}
	return sqlCfg, nil
}

// func skipWhitespaces(value string) string {
// 	for idx, c := range value {
// 		if !unicode.IsSpace(c) {
// 			return value[idx:]
// 		}
// 	}
// 	return ""
// }

// func readString(value string) (string, string, error) {
// 	value = skipWhitespaces(value)
// 	var sb strings.Builder
// 	isQuote := false
// 	hasQuote := false
// 	isEscape := false
// 	for idx, c := range value {
// 		if idx == 0 {
// 			if c == '"' {
// 				isQuote = true
// 				hasQuote = true
// 				continue
// 			}
// 		}
//
// 		if isQuote {
// 			if isEscape {
// 				isEscape = false
//
// 				switch c {
// 				case '\\':
// 					sb.WriteRune(c)
// 				case '"':
// 					sb.WriteRune(c)
// 				case 'n':
// 					sb.WriteRune('\n')
// 				case 'r':
// 					sb.WriteRune('\r')
// 				case 't':
// 					sb.WriteRune('\t')
// 				default:
// 					sb.WriteRune('\\')
// 					sb.WriteRune(c)
// 				}
// 				continue
// 			}
//
// 			if c == '\\' {
// 				isEscape = true
// 				continue
// 			}
// 			if c == '"' {
// 				isQuote = false
// 				continue
// 			}
// 		}
//
// 		if unicode.IsSpace(c) {
// 			if !isQuote {
// 				return sb.String(), value[idx:], nil
// 			}
// 		} else if hasQuote && !isQuote {
// 			return "", "", errors.New("invalid syntex")
// 		}
//
// 		sb.WriteRune(c)
// 	}
// 	return sb.String(), "", nil
// }

func splitFilter(value string) (gobatis.Filter, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return gobatis.Filter{}, errors.New("expression is empty")
	}
	if !strings.HasPrefix(value, "-") {
		return gobatis.Filter{Expression: value}, nil
	}
	idx := strings.IndexFunc(value, unicode.IsSpace)
	if idx < 0 {
		return gobatis.Filter{}, errors.New("expression is empty")
	}

	key := value[:idx]
	expression := strings.TrimSpace(value[idx:])
	return gobatis.Filter{Expression: expression, Dialect: strings.TrimPrefix(key, "-")}, nil

	// name, nameNext, err := readString(value)
	// if err != nil {
	// 	return Filter{}, errors.New("name is invalid syntex")
	// }
	// if name == "" {
	// 	return Filter{}, errors.New("name is missing")
	// }
	// op, opNext, err := readString(nameNext)
	// if err != nil {
	// 	return Filter{}, errors.New("op is invalid syntex")
	// }
	// if op == "" {
	// 	return Filter{}, errors.New("op is missing")
	// }

	// value, valueNext, err := readString(opNext)
	// if err != nil {
	// 	return Filter{}, errors.New("value is invalid syntex")
	// }
	// if value == "" {
	// 	return Filter{}, errors.New("value is missing")
	// }
	// var values = []string{value}
	// if strings.ToLower(op) == "between" {
	// 	value, valueNext, err = readString(valueNext)
	// 	if err != nil {
	// 		return Filter{}, errors.New("value2 is invalid syntex")
	// 	}
	// 	if value == "" {
	// 		return Filter{}, errors.New("value2 is missing")
	// 	}
	// 	values = append(values, value)
	// }
	// dialect, dialectNext, err := readString(valueNext)
	// if err != nil {
	// 	return Filter{}, errors.New("dialect is invalid syntex")
	// }
	// if strings.TrimSpace(dialectNext) != "" {
	// 	return Filter{}, errors.New("invalid syntex")
	// }
	// return Filter{Name: name, Op: op, Values: values, Dialect: dialect}, nil
}

func splitFirstBySpace(comment string) (string, string) {
	comment = strings.TrimSpace(comment)
	idx := strings.IndexFunc(comment, unicode.IsSpace)
	if idx >= 0 {
		return strings.TrimSpace(comment[:idx]), comment[idx:]
	}
	return comment, ""
}

func splitByEmptyLine(comments []string) []string {
	var ss []string
	var sb strings.Builder
	for _, comment := range comments {
		comment = strings.TrimSpace(comment)
		comment = strings.TrimPrefix(comment, "//")

		if s := strings.TrimSpace(comment); strings.HasPrefix(s, "@") {
			ss = append(ss, sb.String())
			sb.Reset()
			comment = s
		}

		if sb.Len() != 0 {
			sb.WriteString("\r\n")
		}
		sb.WriteString(comment)
	}
	if sb.Len() != 0 {
		ss = append(ss, sb.String())
	}
	return ss
}
