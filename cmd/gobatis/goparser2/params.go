package goparser2

import (
	"go/ast"
	"strings"

	"github.com/runner-mei/GoBatis/cmd/gobatis/goparser2/astutil"
)

type Param struct {
	Name       string
	IsVariadic bool
	Type       ast.Expr
}

func (param Param) Print(ctx *PrintContext) string {
	return param.TypeName()
}

func (param Param) TypeName() string {
	return astutil.TypePrint(param.Type)
}

type Params struct {
	Method *Method `json:"-"`
	List   []Param
}

func (ps *Params) Print(ctx *PrintContext, sb *strings.Builder) {
	for idx := range ps.List {
		if idx != 0 {
			sb.WriteString(", ")
		}
		sb.WriteString(ps.List[idx].Name)
		sb.WriteString(" ")

		if ps.List[idx].IsVariadic {
			sb.WriteString("...")
		}
		sb.WriteString(ps.List[idx].TypeName())
	}
}

func (ps *Params) Len() int {
	return len(ps.List)
}
