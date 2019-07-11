package goparser

import (
	"errors"
	"go/types"
	"strings"
	"unicode"

	gobatis "github.com/runner-mei/GoBatis"
)

type Method struct {
	Itf      *Interface `json:"-"`
	Pos      int
	Name     string
	Comments []string
	Config   *SQLConfig
	Params   *Params
	Results  *Results
}

func NewMethod(itf *Interface, pos int, name string, comments []string) (*Method, error) {
	m := &Method{Itf: itf, Pos: pos, Name: name, Comments: comments}
	cfg, err := parseComments(comments)
	if err != nil {
		return nil, errors.New("method '" + m.Name + "' error : " + err.Error())
	}

	if len(cfg.SQL.Filters) > 0 {
		if cfg.StatementType != "" {
			if cfg.StatementType != "select" && cfg.StatementType != "delete" {
				return nil, errors.New("filter is forbidden while statement type is " + cfg.StatementType)
			}
		} else {
			if !isSelectStatement(name) && !isDeleteStatement(name) {
				return nil, errors.New("filter is forbidden while statement type isnot select or delete")
			}
		}
	}
	m.Config = cfg
	return m, nil
}

func (m *Method) MethodSignature(ctx *PrintContext) string {
	var sb strings.Builder
	m.Print(ctx, false, &sb)
	return sb.String()
}

func (m *Method) String() string {
	var sb strings.Builder
	m.Print(nil, true, &sb)
	return sb.String()
}

func (m *Method) Print(ctx *PrintContext, comment bool, sb *strings.Builder) {
	if comment {
		for idx := range m.Comments {
			if strings.TrimSpace(m.Comments[idx]) == "" {
				continue
			}
			if ctx != nil {
				sb.WriteString(ctx.Indent)
			}
			sb.WriteString(m.Comments[idx])
			sb.WriteString("\r\n")
		}
	}
	if ctx != nil {
		sb.WriteString(ctx.Indent)
	}
	sb.WriteString(m.Name)
	sb.WriteString("(")
	m.Params.Print(ctx, sb)
	sb.WriteString(")")
	switch m.Results.Len() {
	case 0:
	case 1:
		sb.WriteString(" ")
		if m.Results.List[0].Name != "" {
			sb.WriteString("(")
			m.Results.Print(ctx, sb)
			sb.WriteString(")")
		} else {
			m.Results.Print(ctx, sb)
		}
	default:
		sb.WriteString(" (")
		m.Results.Print(ctx, sb)
		sb.WriteString(")")
	}
}

func (m *Method) StatementGoTypeName() string {
	switch m.StatementType() {
	case gobatis.StatementTypeSelect:
		return "gobatis.StatementTypeSelect"
	case gobatis.StatementTypeUpdate:
		return "gobatis.StatementTypeUpdate"
	case gobatis.StatementTypeInsert:
		return "gobatis.StatementTypeInsert"
	case gobatis.StatementTypeDelete:
		return "gobatis.StatementTypeDelete"
	default:
		if m.Config != nil && m.Config.StatementType != "" {
			return "gobatis.StatementTypeUnknown-" + m.Config.StatementType
		}
		return "gobatis.StatementTypeUnknown-" + m.Name
	}
}

func (m *Method) StatementTypeName() string {
	switch m.StatementType() {
	case gobatis.StatementTypeSelect:
		return "select"
	case gobatis.StatementTypeUpdate:
		return "update"
	case gobatis.StatementTypeInsert:
		return "insert"
	case gobatis.StatementTypeDelete:
		return "delete"
	default:
		if m.Config != nil && m.Config.StatementType != "" {
			return "statementTypeUnknown-" + m.Config.StatementType
		}
		return "statementTypeUnknown-" + m.Name
	}
}

func (m *Method) StatementType() gobatis.StatementType {
	if m.Config != nil && m.Config.StatementType != "" {
		switch strings.ToLower(m.Config.StatementType) {
		case "insert":
			return gobatis.StatementTypeInsert
		case "update":
			return gobatis.StatementTypeUpdate
		case "delete":
			return gobatis.StatementTypeDelete
		case "select":
			return gobatis.StatementTypeSelect
		}
		return gobatis.StatementTypeNone
	}
	if isInsertStatement(m.Name) {
		return gobatis.StatementTypeInsert
	}
	if isUpdateStatement(m.Name) {
		return gobatis.StatementTypeUpdate
	}
	if isDeleteStatement(m.Name) {
		return gobatis.StatementTypeDelete
	}
	if isSelectStatement(m.Name) {
		return gobatis.StatementTypeSelect
	}
	return gobatis.StatementTypeNone
}

func (m *Method) UpsertKeys() []string {
	if m.Params.Len() == 0 {
		return []string{}
	}

	params := make([]string, 0, m.Params.Len()-1)
	lastParamType := m.Params.List[m.Params.Len()-1].Type
	if IsStructType(lastParamType) && !IsIgnoreStructTypes(lastParamType) {

		for _, param := range m.Params.List[:m.Params.Len()-1] {
			if param.Type.String() == "context.Context" {
				continue
			}
			params = append(params, param.Name)
		}
	}
	return params
}

func (m *Method) QueryKeys() []string {
	pos := strings.Index(m.Name, "By")
	if pos < 0 {
		return nil
	}
	keyStr := m.Name[pos+len("By"):]
	if keyStr == "" {
		return nil
	}

	ss := strings.Split(keyStr, "By")
	if len(ss) == 0 {
		return nil
	}

	params := make([]string, len(ss))
	for idx, nm := range ss {
		param, ok := m.findParam(nm)
		if !ok {
			panic(errors.New("param '" + nm + "' isnot found"))
		}
		params[idx] = param
	}
	return params
}

func (m *Method) IsOneParam() bool {
	count := 0
	for idx := range m.Params.List {
		if m.Params.List[idx].Type.String() == "context.Context" {
			continue
		}
		count++
	}
	return count == 1
}

func (m *Method) findParam(name string) (string, bool) {
	lowerName := strings.ToLower(name)

	if m.IsOneParam() {
		var param *Param
		for idx := range m.Params.List {
			if m.Params.List[idx].Type.String() == "context.Context" {
				continue
			}
			param = &m.Params.List[idx]
		}

		typ := param.Type

		if p, ok := typ.(*types.Pointer); ok {
			typ = p.Elem()
		}
		if named, ok := typ.(*types.Named); ok {
			typ = named.Underlying()
		}

		if st, ok := typ.(*types.Struct); ok {
			return filter(lowerName, func(cb func(string) bool) (string, bool) {
				for idx := 0; idx < st.NumFields(); idx++ {
					v := st.Field(idx)
					if cb(v.Name()) {
						return v.Name(), true
					}
				}
				return "", false
			})
		}
	}

	return filter(lowerName, func(cb func(string) bool) (string, bool) {
		for idx := range m.Params.List {
			if cb(m.Params.List[idx].Name) {
				return m.Params.List[idx].Name, true
			}
		}
		return "", false
	})
}

func filter(lowerName string, search func(func(string) bool) (string, bool)) (string, bool) {
	if nm, ok := search(func(paramName string) bool {
		return strings.ToLower(paramName) == lowerName
	}); ok {
		return nm, ok
	}
	if nm, ok := search(func(paramName string) bool {
		return strings.ToLower(paramName) == lowerName+"id"
	}); ok {
		return nm, ok
	}
	if nm, ok := search(func(paramName string) bool {
		return strings.ToLower(toAbbreviation(paramName)) == lowerName
	}); ok {
		return nm, ok
	}
	if nm, ok := search(func(paramName string) bool {
		return strings.ToLower(toAbbreviation(paramName)) == lowerName+"id"
	}); ok {
		return nm, ok
	}
	return "", false
}

func toAbbreviation(name string) string {
	var abbreviation strings.Builder
	for idx, c := range name {
		if idx == 0 {
			abbreviation.WriteRune(c)
			continue
		}

		if unicode.IsUpper(c) {
			abbreviation.WriteRune(c)
		}
	}
	return abbreviation.String()
}
