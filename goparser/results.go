package goparser

import (
	"go/types"
	"strings"
)

type Result struct {
	Name string
	Type types.Type
}

type Results struct {
	Method *Method      `json:"-"`
	Tuple  *types.Tuple `json:"-"`
	List   []Result
}

func NewResults(method *Method, tuple *types.Tuple) *Results {
	rs := &Results{
		Method: method,
		Tuple:  tuple,
		List:   make([]Result, tuple.Len()),
	}

	for i := 0; i < tuple.Len(); i++ {
		v := tuple.At(i)
		rs.List[i] = Result{
			Name: v.Name(),
			Type: v.Type(),
		}
	}
	return rs
}

func (rs *Results) Print(ctx *PrintContext, sb *strings.Builder) {
	for idx := range rs.List {
		if idx != 0 {
			sb.WriteString(", ")
		}
		if rs.List[idx].Name != "" {
			sb.WriteString(rs.List[idx].Name)
			sb.WriteString(" ")
		}
		printType(ctx, sb, rs.List[idx].Type)
	}
}

func (rs *Results) Len() int {
	return len(rs.List)
}

func (rs *Results) ByName(name string) *Result {
	name = strings.Trim(name, "`")
	if name == "" {
		panic("name must not blank")
	}

	for idx := range rs.List {
		if rs.List[idx].Name == name {
			return &rs.List[idx]
		}
	}
	return nil
}
