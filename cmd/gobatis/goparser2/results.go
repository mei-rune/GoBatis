package goparser2

import (
	"go/ast"
	"strings"

	"github.com/runner-mei/GoBatis/cmd/gobatis/goparser2/astutil"
)

type Result struct {
	Name string
	Type ast.Expr
}

func (result Result) Print(ctx *PrintContext) string {
	return result.TypeName()
}

func (result Result) TypeName() string {
	return astutil.TypePrint(result.Type)
}

func (result Result) IsCloser() bool {
	return result.TypeName() == "io.Closer" || result.TypeName() == "Closer"
}

// func (result Result) IsFunc() bool {
// 	_, ok := result.Type.(*types.Signature)
// 	return ok
// }

// func (result Result) IsBatchCallback() bool {
// 	signature, ok := result.Type.(*types.Signature)
// 	if !ok {
// 		return false
// 	}

// 	if signature.Variadic() {
// 		return false
// 	}

// 	if signature.Params().Len() != 1 {
// 		return false
// 	}

// 	typ := signature.Params().At(0).Type()
// 	if _, ok := typ.(*types.Pointer); !ok {
// 		return false
// 	}

// 	if signature.Results().Len() != 2 {
// 		return false
// 	}

// 	typ = signature.Results().At(0).Type()
// 	if typ.String() != "bool" {
// 		return false
// 	}

// 	typ = signature.Results().At(1).Type()
// 	if typ.String() != "error" {
// 		return false
// 	}

// 	return true
// }

// func (result Result) IsCallback() bool {
// 	signature, ok := result.Type.(*types.Signature)
// 	if !ok {
// 		return false
// 	}

// 	if signature.Variadic() {
// 		return false
// 	}

// 	if signature.Params().Len() != 1 {
// 		return false
// 	}

// 	typ := signature.Params().At(0).Type()
// 	if _, ok := typ.(*types.Pointer); !ok {
// 		return false
// 	}

// 	if signature.Results().Len() != 1 {
// 		return false
// 	}

// 	typ = signature.Results().At(0).Type()
// 	if typ.String() != "error" {
// 		return false
// 	}
// 	return true
// }

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
		sb.WriteString(rs.List[idx].TypeName())
	}
}

func (rs *Results) Len() int {
	return len(rs.List)
}

// func ArgFromFunc(typ types.Type) Param {
// 	signature, ok := typ.(*types.Signature)
// 	if !ok {
// 		panic(fmt.Errorf("want *types.Signature got %T", typ))
// 	}

// 	if signature.Params().Len() != 1 {
// 		panic(fmt.Errorf("want params len is 1 got %d", signature.Params().Len()))
// 	}

// 	v := signature.Params().At(0)
// 	return Param{
// 		Name: v.Name(),
// 		Type: v.Type(),
// 	}
// }
