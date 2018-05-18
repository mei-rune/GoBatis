package goparser

import (
	"go/types"
	"strings"
)

type Param struct {
	Name string
	Type types.Type
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

func (ps *Params) ByName(name string) *Param {
	name = strings.Trim(name, "`")
	if name == "" {
		panic("name must not blank")
	}

	for idx := range ps.List {
		if ps.List[idx].Name == name {
			return &ps.List[idx]
		}
	}
	return nil
}
