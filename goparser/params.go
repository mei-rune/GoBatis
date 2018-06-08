package goparser

import (
	"go/types"
	"strings"
)

type Param struct {
	Name string
	Type types.Type
}

func (param Param) Print(ctx *PrintContext) string {
	var sb strings.Builder
	printType(ctx, &sb, param.Type)
	return sb.String()
}

func (param Param) TypeName() string {
	var sb strings.Builder
	printTypename(&sb, param.Type)
	return sb.String()
}

type Params struct {
	Method *Method      `json:"-"`
	Tuple  *types.Tuple `json:"-"`
	List   []Param
}

func NewParams(method *Method, tuple *types.Tuple) *Params {
	ps := &Params{
		Method: method,
		Tuple:  tuple,
		List:   make([]Param, tuple.Len()),
	}

	for i := 0; i < tuple.Len(); i++ {
		v := tuple.At(i)
		ps.List[i] = Param{
			Name: v.Name(),
			Type: v.Type(),
		}
	}

	return ps
}

func (ps *Params) Print(ctx *PrintContext, sb *strings.Builder) {
	for idx := range ps.List {
		if idx != 0 {
			sb.WriteString(", ")
		}
		sb.WriteString(ps.List[idx].Name)
		sb.WriteString(" ")
		printType(ctx, sb, ps.List[idx].Type)
	}
}

func (ps *Params) Len() int {
	return len(ps.List)
}
