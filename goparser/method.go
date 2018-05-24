package goparser

import "strings"

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
		return nil, err
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
