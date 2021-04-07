package goparser

import (
	"fmt"
	"go/ast"
	"go/types"
	"strings"
)

type Param struct {
	Name       string
	IsVariadic bool
	Type       types.Type
}

func (param Param) Print(ctx *PrintContext) string {
	return PrintType(ctx, param.Type, param.IsVariadic)
}

func (param Param) TypeName() string {
	var sb strings.Builder
	printTypename(&sb, param.Type, param.IsVariadic)
	return sb.String()
}

type Params struct {
	Method *Method      `json:"-"`
	Tuple  *types.Tuple `json:"-"`
	List   []Param
}

func NewParams(method *Method, fieldList *ast.FieldList, tuple *types.Tuple, isVariadic bool) *Params {
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
		if "updatedAt" == v.Name() {
			fmt.Println("=========", v.Name(), fmt.Sprintf("%T %+v", v.Type(), v.Type()))
			fmt.Println("=========", v.Name(), fmt.Sprintf("%T %+v", fieldList.List[i].Type, fieldList.List[i].Type))
		}
	}

	if tuple.Len() > 0 {
		ps.List[tuple.Len()-1].IsVariadic = isVariadic
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
		printType(ctx, sb, ps.List[idx].Type, ps.List[idx].IsVariadic)
	}
}

func (ps *Params) Len() int {
	return len(ps.List)
}
