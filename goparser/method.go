package goparser

import (
	"errors"
	"strings"

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

	// TODO: 增加 filter 配置
	// if len(cfg.Filters) > 0 {
	// 	if cfg.StatementType != "" {
	// 		if cfg.StatementType != "select" && cfg.StatementType != "delete" {
	// 			return nil, errors.New("filter is forbidden while statement type is " + cfg.StatementType)
	// 		}
	// 	} else {
	// 		if !isSelectStatement(name) && !isDeleteStatement(name) {
	// 			return nil, errors.New("filter is forbidden while statement type isnot select or delete")
	// 		}
	// 	}
	// }
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
