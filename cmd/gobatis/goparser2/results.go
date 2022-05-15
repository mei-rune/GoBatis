package goparser2

import (
	"fmt"
	"go/ast"
	"strings"

	"github.com/runner-mei/GoBatis/cmd/gobatis/goparser2/astutil"
)

type Result struct {
	Results  *Results `json:"-"`
	Name     string
	TypeExpr ast.Expr
}

func (result Result) Print(ctx *PrintContext) string {
	return result.ToTypeLiteral()
}

func (result Result) ToTypeLiteral() string {
	return astutil.ToString(result.TypeExpr)
}

func (result Result) Type() Type {
	return Type{
		Type: astutil.Type{
			// Ctx:      result.Results.Method.Interface.Ctx,
			File: result.Results.Method.Interface.File.File,
			Expr: result.TypeExpr,
		},
	}
}

func (result Result) IsCloser() bool {
	return result.ToTypeLiteral() == "io.Closer" || result.ToTypeLiteral() == "Closer"
}

func (result Result) IsFuncType() bool {
	return astutil.IsFuncType(result.TypeExpr)
}

func (result Result) IsBatchCallback() bool {
	funcType, ok := astutil.ToFuncType(result.TypeExpr)
	if !ok {
		return false
	}
	signature := astutil.ToFunction(funcType)

	if signature.IsVariadic() {
		return false
	}

	if len(signature.Params.List) != 1 {
		return false
	}

	typ := signature.Params.List[0].Expr
	if !astutil.IsPtrType(typ) {
		return false
	}

	if len(signature.Results.List) != 2 {
		return false
	}

	typ = signature.Results.List[0].Expr
	if !astutil.IsBooleanType(typ) {
		return false
	}

	typ = signature.Results.List[1].Expr
	if !astutil.IsErrorType(typ) {
		return false
	}
	return true
}

func (result Result) IsCallback() bool {
	funcType, ok := astutil.ToFuncType(result.TypeExpr)
	if !ok {
		return false
	}
	signature := astutil.ToFunction(funcType)

	if signature.IsVariadic() {
		return false
	}

	if len(signature.Params.List) != 1 {
		return false
	}

	typ := signature.Params.List[0].Expr
	if !astutil.IsPtrType(typ) {
		return false
	}

	if len(signature.Results.List) != 1 {
		return false
	}

	typ = signature.Results.List[0].Expr
	if !astutil.IsErrorType(typ) {
		return false
	}
	return true
}

type Results struct {
	Method *Method `json:"-"`
	// Tuple  *types.Tuple `json:"-"`
	List []Result
}

// func NewResults(method *Method, tuple *types.Tuple) *Results {
// 	rs := &Results{
// 		Method: method,
// 		Tuple:  tuple,
// 		List:   make([]Result, tuple.Len()),
// 	}

// 	for i := 0; i < tuple.Len(); i++ {
// 		v := tuple.At(i)
// 		rs.List[i] = Result{
// 			Name: v.Name(),
// 			Type: v.Type(),
// 		}
// 	}
// 	return rs
// }

func (rs *Results) Print(ctx *PrintContext, sb *strings.Builder) {
	for idx := range rs.List {
		if idx != 0 {
			sb.WriteString(", ")
		}
		if rs.List[idx].Name != "" {
			sb.WriteString(rs.List[idx].Name)
			sb.WriteString(" ")
		}
		sb.WriteString(rs.List[idx].ToTypeLiteral())
	}
}

func (rs *Results) Len() int {
	return len(rs.List)
}

func ArgFromFunc(typ Type) Param {
	funcType, ok := astutil.ToFuncType(typ.Expr)
	if !ok {
		panic(fmt.Errorf("want *ast.FuncType got %T", typ.Expr))
	}
	signature := astutil.ToFunction(funcType)

	// if signature.IsVariadic() {
	// 	return false
	// }

	// if len(signature.Params.List) != 1 {
	// 	return false
	// }

	for idx := range signature.Params.List {
		if astutil.IsContextType(signature.Params.List[idx].Expr) {
			continue
		}

		return Param{
			Params: &Params{
				Method: &Method{
					Interface: &Interface{
						Ctx: &ParseContext{
							Context: typ.File.Ctx,
						},
						File: &File{
							File: typ.File,
						},
					},
				},
			},
			Name:     signature.Params.List[idx].Name,
			TypeExpr: signature.Params.List[idx].Expr,
		}
	}
	panic(fmt.Errorf("want *ast.FuncType got %T", typ.Expr))

	// signature, ok := typ.(*types.Signature)
	// if !ok {
	// 	panic(fmt.Errorf("want *types.Signature got %T", typ))
	// }

	// if signature.Params().Len() != 1 {
	// 	panic(fmt.Errorf("want params len is 1 got %d", signature.Params().Len()))
	// }

	// v := signature.Params().At(0)
	// return Param{
	// 	Name: v.Name(),
	// 	Type: v.Type(),
	// }
}
