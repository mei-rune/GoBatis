package goparser2

import (
	"errors"
	"fmt"
	"go/ast"
	"strings"

	gobatis "github.com/runner-mei/GoBatis"
	"github.com/runner-mei/GoBatis/cmd/gobatis/goparser2/astutil"
)

type Type struct {
	Ctx      *ParseContext `json:"-"`
	File     *File         `json:"-"`
	TypeExpr ast.Expr
}

func (typ Type) String() string {
	if typ.TypeExpr == nil {
		return "**nil**"
	}
	return astutil.ToString(typ.TypeExpr)
}

func (typ Type) ToLiteral() string {
	if typ.TypeExpr == nil {
		return "**nil**"
	}
	return astutil.ToString(typ.TypeExpr)
}

func (typ Type) ToTypeSpec() (*astutil.TypeSpec, error) {
	return typ.Ctx.ToClass(typ.File.File, typ.TypeExpr)
}

func (typ Type) IsSameType(fuzzyType Type) bool {
	if typ.File == fuzzyType.File {
		return astutil.ToString(typ.TypeExpr) == astutil.ToString(fuzzyType.TypeExpr)
	}
	// TODO:
	return false

	// return astutil.ToFullString(typ.File.File, typ.TypeExpr) == astutil.ToFullString(fuzzyType.File.File, fuzzyType.TypeExpr)
}

func (typ Type) IsIgnoreStructTypes() bool {
	return astutil.IsIgnoreStructTypes(typ.Ctx.Context, typ.File.File, typ.TypeExpr, gobatis.IgnoreStructNames)
}

func (typ Type) IsStructType() bool {
	if typ.TypeExpr == nil {
		return false
	}
	return IsStructType(typ.File, typ.TypeExpr)
}

func IsStructType(file *File, typ ast.Expr) bool {
	switch r := typ.(type) {
	case *ast.StructType:
		return true
	case *ast.StarExpr:
		return IsStructType(file, r.X)
	case *ast.MapType:
		return IsStructType(file, r.Value)
	case *ast.ArrayType:
		return IsStructType(file, r.Elt)
	case *ast.Ident:
		return file.File.Ctx.IsStructType(file.File, r)
	case *ast.SelectorExpr:
		return file.File.Ctx.IsStructType(file.File, r)
	}
	return false
}

func (typ Type) IsMapType() bool {
	return typ.Ctx.IsMapType(typ.File.File, typ.TypeExpr)
}

func (typ Type) IsSliceOrArrayType() bool {
	return typ.Ctx.IsSliceOrArrayType(typ.File.File, typ.TypeExpr)
}

func (typ Type) IsContextType() bool {
	return typ.Ctx.IsContextType(typ.File.File, typ.TypeExpr)
}

func (typ Type) ElemType() *Type {
	var elemType ast.Expr
	switch t := typ.TypeExpr.(type) {
	case *ast.StructType:
		elemType = t
	case *ast.ArrayType:
		elemType = t.Elt
	case *ast.StarExpr:
		elemType = t.X
	case *ast.MapType:
		elemType = t.Value
	case *ast.Ident:
		elemType =  t
	case *ast.SelectorExpr:
		elemType =  t
	default:
		elemType =  t
	}

	return &Type{
		Ctx:      typ.Ctx,
		File:     typ.File,
		TypeExpr: elemType,
	}
}

func (typ Type) RecursiveElemType() *Type {
	return &Type{
		Ctx:      typ.Ctx,
		File:     typ.File,
		TypeExpr: getElemType(typ.TypeExpr),
	}
}

func getElemType(typ ast.Expr) ast.Expr {
	switch t := typ.(type) {
	case *ast.StructType:
		return t
	case *ast.ArrayType:
		return getElemType( t.Elt)
	case *ast.StarExpr:
		return getElemType( t.X)
	case *ast.MapType:
		return getElemType( t.Value)
	case *ast.Ident:
		return t
	case *ast.SelectorExpr:
		return t
	default:
		return t
	}
}

func (typ Type) IsPtrType() bool {
	return astutil.IsPtrType(typ.TypeExpr)
}

func (typ Type) IsStringType() bool {
	return astutil.IsStringType(typ.TypeExpr)
}

func (typ Type) IsExceptedType(excepted string, or ...string) bool {
	return isExceptedType(typ.Ctx.Context, typ.File.File, typ.TypeExpr, excepted, or...)
}

func isExceptedType(ctx *astutil.Context, file *astutil.File, typ ast.Expr, excepted string, or ...string) bool {
	if ptr, ok := typ.(*ast.StarExpr); ok {
		if excepted == "ptr" {
			return true
		}
		for _, name := range or {
			if name == "ptr" {
				return true
			}
		}
		return isExceptedType(ctx, file, ptr.X, excepted, or...)
	}
	for _, name := range append([]string{excepted}, or...) {
		switch name {
		case "func":
			return astutil.IsFuncType(typ)
		case "context":
			if astutil.IsContextType(typ) {
				return true
			}
		case "ptr":
		case "error":
			return ctx.IsErrorType(file, typ)
		case "ignoreStructs":
			return astutil.IsIgnoreStructTypes(ctx, file, typ, gobatis.IgnoreStructNames)
		case "underlyingStruct":
			var exp = getElemType(typ)
			if ctx.IsStructType(file, exp) {
				return true
			}
		case "struct":
			if ctx.IsStructType(file, typ) {
				return true
			}
		case "slice":
			return astutil.IsArrayOrSliceType(typ)
		case "numeric":
			return ctx.IsNumericType(file, typ)
		case "bool", "boolean":
			return astutil.IsBooleanType(typ)
		case "string":
			return astutil.IsStringType(typ)
		case "basic":
			return ctx.IsBasicType(file, typ)

		// 	if _, ok := typ.(*types.Basic); ok {
		// 		return true
		// 	}
		// 	typ = typ.Underlying()
		// 	if _, ok := typ.(*types.Basic); ok {
		// 		return true
		// 	}

		case "interface", "interface{}":
			return ctx.IsInterfaceType(file, typ)

		// 	if _, ok := typ.(*types.Interface); ok {
		// 		return true
		// 	}
		// 	typ = typ.Underlying()
		// 	if _, ok := typ.(*types.Interface); ok {
		// 		return true
		// 	}
		default:
			panic(errors.New("'" + fmt.Sprintf("%T %#v", typ, typ) + "' unexcepted type - " + name + "," + strings.Join(or, ",")))
		}
	}

	return false
}

func (typ Type) IsBasicType() bool {
	return typ.Ctx.IsBasicType(typ.File.File, typ.TypeExpr)
}

func (typ Type) IsBasicMap() bool {
	// keyType := getKeyType(recordType)

	returnType := typ.TypeExpr
	for {
		if ptr, ok := returnType.(*ast.StarExpr); !ok {
			break
		} else {
			returnType = ptr.X
		}
	}

	mapType, ok := returnType.(*ast.MapType)
	if !ok {
		return false
	}

	elemType := mapType.Value
	for {
		if ptr, ok := elemType.(*ast.StarExpr); !ok {
			break
		} else {
			elemType = ptr.X
		}
	}

	if typ.Ctx.Context.IsBasicType(typ.File.File, elemType) {
		return true
	}

	switch astutil.ToString(elemType) {
	case "time.Time", "net.IP", "net.HardwareAddr":
		return true
	}
	return false
}
