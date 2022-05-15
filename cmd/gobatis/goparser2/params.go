package goparser2

import (
	"go/ast"
	"strings"

	"github.com/runner-mei/GoBatis/cmd/gobatis/goparser2/astutil"
)

type Param struct {
	Params     *Params `json:"-"`
	Name       string
	IsVariadic bool
	TypeExpr   ast.Expr
}

func (param Param) Print(ctx *PrintContext) string {
	return param.ToTypeLiteral()
}

func (param Param) ToTypeLiteral() string {
	return astutil.ToString(param.TypeExpr)
}

func (param Param) Type() Type {
	return Type{
		Type: astutil.Type{
			// Ctx:      param.Params.Method.Interface.Ctx,
			File: param.Params.Method.Interface.File.File,
			Expr: param.TypeExpr,
		},
	}
}

func (param Param) IsEllipsis() bool {
	return param.IsVariadic
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
		sb.WriteString(ps.List[idx].ToTypeLiteral())
	}
}

func (ps *Params) Len() int {
	return len(ps.List)
}
