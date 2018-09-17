package goparser

import (
	"go/types"
	"strings"
)

type Result struct {
	Name string
	Type types.Type
}

func (result Result) Print(ctx *PrintContext) string {
	return PrintType(ctx, result.Type, false)
}

func (result Result) TypeName() string {
	var sb strings.Builder
	printTypename(&sb, result.Type, false)
	return sb.String()
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
		printType(ctx, sb, rs.List[idx].Type, false)
	}
}

func (rs *Results) Len() int {
	return len(rs.List)
}
