package goparser

import (
	"errors"
	"strings"
	"unicode"
)

type SQLConfig struct {
	Description string
	Reference   *struct {
		Interface string
		Method    string
	}
	StatementType  string
	DefaultSQL     string
	Options        map[string]string
	Dialects       map[string]string
	FieldDelimiter string
}

func parseComments(comments []string) (*SQLConfig, error) {
	if len(comments) == 0 {
		return &SQLConfig{}, nil
	}
	sections := splitByEmptyLine(comments)
	if len(sections) == 0 {
		return &SQLConfig{}, nil
	}
	var sqlCfg = &SQLConfig{}
	sqlCfg.Description = strings.TrimSpace(sections[0])
	for idx := 1; idx < len(sections); idx++ {
		tag, value := splitFirstBySpace(sections[idx])
		value = strings.TrimSpace(value)
		if value == "" {
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
			sqlCfg.StatementType = strings.TrimSpace(value)
		case "@option":
			optKey, optValue := splitFirstBySpace(value)
			if sqlCfg.Options == nil {
				sqlCfg.Options = map[string]string{strings.TrimSpace(optKey): strings.TrimSpace(optValue)}
			} else {
				sqlCfg.Options[strings.TrimSpace(optKey)] = strings.TrimSpace(optValue)
			}
		case "@default":
			sqlCfg.DefaultSQL = strings.TrimSpace(value)
		case "@field_delimiter":
			sqlCfg.FieldDelimiter = strings.TrimSpace(value)
		default:
			if sqlCfg.Dialects == nil {
				sqlCfg.Dialects = map[string]string{strings.TrimPrefix(tag, "@"): strings.TrimSpace(value)}
			} else {
				sqlCfg.Dialects[strings.TrimPrefix(tag, "@")] = strings.TrimSpace(value)
			}
		}
	}

	if sqlCfg.Reference != nil {
		if strings.ToLower(sqlCfg.DefaultSQL) != "" || len(sqlCfg.Dialects) != 0 {
			return nil, errors.New("sql statement is unnecessary while reference is exists")
		}
	}
	return sqlCfg, nil
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
