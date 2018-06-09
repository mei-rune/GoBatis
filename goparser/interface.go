package goparser

import (
	"go/types"
	"strings"

	"github.com/runner-mei/GoBatis"
)

func IsStructType(typ types.Type) bool {
	if _, ok := typ.(*types.Struct); ok {
		return true
	}

	if ptr, ok := typ.(*types.Pointer); ok {
		return IsStructType(ptr.Elem())
	}

	if m, ok := typ.(*types.Map); ok {
		return IsStructType(m.Elem())
	}

	if ar, ok := typ.(*types.Array); ok {
		return IsStructType(ar.Elem())
	}

	if slice, ok := typ.(*types.Slice); ok {
		return IsStructType(slice.Elem())
	}

	if named, ok := typ.(*types.Named); ok {
		return IsStructType(named.Underlying())
	}

	return false
}

func GetElemType(typ types.Type) types.Type {
	switch t := typ.(type) {
	case *types.Struct:
		return t
	case *types.Array:
		return GetElemType(t.Elem())
	case *types.Slice:
		return GetElemType(t.Elem())
	case *types.Pointer:
		return GetElemType(t.Elem())
	case *types.Map:
		return GetElemType(t.Elem())
	case *types.Named:
		return t // underlyingType(t.Underlying())
	default:
		return nil
	}
}

type Interface struct {
	File *File `json:"-"`
	Pos  int
	Name string

	Methods []*Method
}

func (itf *Interface) DetectRecordType(method *Method) types.Type {
	return itf.detectRecordType(method, true)
}

func (itf *Interface) detectRecordType(method *Method, fuzzy bool) types.Type {
	if method == nil {
		for _, name := range []string{
			"Get",
			"Insert",
			"List",
		} {
			method = itf.MethodByName(name)
			if method != nil {
				typ := itf.detectRecordType(method, false)
				if typ != nil {
					return typ
				}
			}
		}
		return nil
	}
	if method.Name == "Insert" {
		if len(method.Params.List) == 1 {
			if IsStructType(method.Params.List[0].Type) {
				return GetElemType(method.Params.List[0].Type)
			}
		}
		return nil
	}

	if method.Name == "Get" || method.Name == "List" || method.Name == "Query" {
		if len(method.Results.List) == 2 {
			if IsStructType(method.Results.List[0].Type) {
				return GetElemType(method.Results.List[0].Type)
			}
		}
	}
	if method.Name == "Update" {
		if len(method.Params.List) > 0 {
			param := method.Params.List[len(method.Params.List)-1]
			if IsStructType(param.Type) {
				return GetElemType(param.Type)
			}
		}
		return nil
	}

	switch method.StatementType() {
	case gobatis.StatementTypeInsert:
		if len(method.Params.List) == 1 {
			if IsStructType(method.Params.List[0].Type) {
				return GetElemType(method.Params.List[0].Type)
			}
		}
		return nil
	case gobatis.StatementTypeUpdate:
		if len(method.Params.List) > 0 {
			param := method.Params.List[len(method.Params.List)-1]
			if IsStructType(param.Type) {
				return GetElemType(param.Type)
			}
		}
		return nil
	case gobatis.StatementTypeSelect:
		if len(method.Results.List) == 2 {
			if !IsStructType(method.Results.List[0].Type) {
				if fuzzy {
					return itf.detectRecordType(nil, false)
				}
				return nil
			}
			resultType := GetElemType(method.Results.List[0].Type)
			fuzzyType := itf.detectRecordType(nil, false)

			if types.Identical(resultType, fuzzyType) {
				return resultType
			}
		}
		return nil
	case gobatis.StatementTypeDelete:
		if fuzzy {
			return itf.detectRecordType(nil, false)
		}
	}
	return nil
}

func (itf *Interface) Print(ctx *PrintContext, sb *strings.Builder) {
	sb.WriteString("type ")
	sb.WriteString(itf.Name)
	sb.WriteString(" interface {")
	var oldIndent string
	if ctx != nil {
		oldIndent = ctx.Indent
		ctx.Indent = ctx.Indent + "	"
	}
	for idx, m := range itf.Methods {
		if idx > 0 {
			sb.WriteString("\r\n")
		}
		sb.WriteString("\r\n")
		m.Print(ctx, true, sb)
	}

	if ctx != nil {
		ctx.Indent = oldIndent
	}
	sb.WriteString("\r\n")
	sb.WriteString("}")
}

func (itf *Interface) String() string {
	var sb strings.Builder
	itf.Print(&PrintContext{}, &sb)
	return sb.String()
}

func (itf *Interface) MethodByName(name string) *Method {
	for idx := range itf.Methods {
		if itf.Methods[idx].Name == name {
			return itf.Methods[idx]
		}
	}
	return nil
}
