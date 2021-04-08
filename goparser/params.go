package goparser

import (
	"go/ast"
	"go/types"
	"strings"

	"github.com/runner-mei/GoBatis/goparser2/astutil"
)

type Param struct {
	Name       string
	IsVariadic bool
	Type       types.Type
	Expr       ast.Expr
}

func (param Param) Print(ctx *PrintContext, sb *strings.Builder) {
	sb.WriteString(param.Name)
	sb.WriteString(" ")
	sb.WriteString(param.TypeName())
}

func (param Param) TypeName() string {
	return astutil.TypePrint(param.Expr)
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
			Expr: astutil.GetFieldByIndex(fieldList, i).Type,
		}

		// if "updatedAt" == v.Name() {
		// 	fmt.Println("=========", v.Name(), fmt.Sprintf("%T %+v", v.Type(), v.Type()))
		// 	fmt.Println("=========", v.Name(), fmt.Sprintf("%T %+v", fieldList.List[i].Type, fieldList.List[i].Type))
		// }
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
		ps.List[idx].Print(ctx, sb)
	}
}

func (ps *Params) Len() int {
	return len(ps.List)
}
