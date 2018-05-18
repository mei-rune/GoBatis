package goparser

import "strings"

type Method struct {
	Itf     *Interface `json:"-"`
	Pos     int
	Name    string
	Doc     []string
	Params  *Params
	Results *Results
}

func NewMethod(itf *Interface, pos int, name string, doc []string) *Method {
	return &Method{Itf: itf, Pos: pos, Name: name, Doc: doc}
}

func (m *Method) MethodSignature() string {
	var sb strings.Builder
	m.Print(nil, false, &sb)
	return sb.String()
}

func (m *Method) String() string {
	var sb strings.Builder
	m.Print(nil, true, &sb)
	return sb.String()
}

func (m *Method) Print(ctx *PrintContext, comment bool, sb *strings.Builder) {
	if comment {
		for idx := range m.Doc {
			if strings.TrimSpace(m.Doc[idx]) == "" {
				continue
			}
			if ctx != nil {
				sb.WriteString(ctx.Indent)
			}
			sb.WriteString(m.Doc[idx])
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
