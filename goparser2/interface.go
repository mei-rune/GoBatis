package goparser2

import (
	"fmt"
	"go/ast"
	"sort"
	"strings"

	gobatis "github.com/runner-mei/GoBatis"
	"github.com/runner-mei/GoBatis/goparser2/astutil"
)

var IgnoreStructs = gobatis.IgnoreStructNames

func IsIgnoreStructTypes(file *File, typ ast.Expr) bool {
	if IsStructType(file, typ) {
		return false
	}
	if file.Ctx.IsPtrType(file.File, typ) {
		return IsIgnoreStructTypes(file, file.Ctx.ElemType(file.File, typ))
	}
	if file.Ctx.IsMapType(file.File, typ) {
		return IsIgnoreStructTypes(file, file.Ctx.MapValueType(file.File, typ))
	}
	if file.Ctx.IsSliceOrArrayType(file.File, typ) {
		return IsIgnoreStructTypes(file, file.Ctx.ElemType(file.File, typ))
	}

	typName := astutil.TypePrint(typ)
	for _, nm := range IgnoreStructs {
		if nm == typName {
			return true
		}
	}

	return false
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

func getElemType(file *File, typ ast.Expr) ast.Expr {
	switch t := typ.(type) {
	case *ast.StructType:
		return t
	case *ast.ArrayType:
		return getElemType(file, t.Elt)
	case *ast.StarExpr:
		return getElemType(file, t.X)
	case *ast.MapType:
		return getElemType(file, t.Value)
	case *ast.Ident:
		return t
	case *ast.SelectorExpr:
		return t
	default:
		return nil
	}
}

type Interface struct {
	File *File `json:"-"`
	Name string

	EmbeddedInterfaces []string
	Comments           []string
	Methods            []*Method
}

func (itf *Interface) DetectRecordType(method *Method) ast.Expr {
	return itf.detectRecordType(method, true)
}

func (itf *Interface) detectRecordType(method *Method, guess bool) ast.Expr {
	if method == nil {
		for _, name := range []string{
			"Get",
			"Insert",
			"List",
			"FindByID",
			"QueryByID",
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

	switch method.StatementType() {
	case gobatis.StatementTypeInsert:
		var list = make([]Param, 0, len(method.Params.List))
		for idx := range method.Params.List {
			if astutil.TypePrint(method.Params.List[idx].Type) == "context.Context" {
				continue
			}
			list = append(list, method.Params.List[idx])
		}

		if len(list) == 1 {
			if method.Name == "InsertRoles1" {
				fmt.Println(fmt.Sprintf("000 %T %s", list[0].Type, astutil.TypePrint(list[0].Type)))
				fmt.Println("IsStructType(itf.File, list[0].Type)", IsStructType(itf.File, list[0].Type))
				fmt.Println("!IsIgnoreStructTypes(itf.File, list[0].Type)", !IsIgnoreStructTypes(itf.File, list[0].Type))

			}
			if IsStructType(itf.File, list[0].Type) &&
				!IsIgnoreStructTypes(itf.File, list[0].Type) {
				return getElemType(itf.File, list[0].Type)
			}

			if method.Name == "InsertRoles1" {
				fmt.Println(fmt.Sprintf("111 %T %s", list[0].Type, astutil.TypePrint(list[0].Type)))
			}
		}

		if guess {
			return itf.detectRecordType(nil, false)
		}
	case gobatis.StatementTypeUpdate:
		if len(method.Params.List) > 0 {
			param := method.Params.List[len(method.Params.List)-1]
			if IsStructType(itf.File, param.Type) && !IsIgnoreStructTypes(itf.File, param.Type) {
				return getElemType(itf.File, param.Type)
			}
		}
		return itf.detectRecordType(nil, false)
	case gobatis.StatementTypeSelect:
		var typ ast.Expr
		if len(method.Results.List) == 1 {
			signature, ok := astutil.ToMethod(method.Results.List[0].Type)
			if !ok {
				typ := method.Results.List[0].Type
				fmt.Println(fmt.Sprintf("%T %#v", typ, typ))
				return nil
			}

			if signature.IsVariadic() {
				return nil
			}

			if len(signature.Params.List) != 1 {
				return nil
			}

			typ = signature.Params.List[0].Typ
		} else if len(method.Results.List) == 2 {
			typ = method.Results.List[0].Type
		} else {
			return nil
		}

		if !IsStructType(itf.File, typ) {
			if guess {
				return itf.detectRecordType(nil, false)
			}
			return nil
		}
		resultType := getElemType(itf.File, typ)
		if !guess {
			if IsStructType(itf.File, resultType) &&
				!IsIgnoreStructTypes(itf.File, resultType) {
				return resultType
			}
			return nil
		}

		if guess {
			fuzzyType := itf.detectRecordType(nil, false)
			if fuzzyType == nil || itf.File.Ctx.IsSameType(itf.File.File, resultType, fuzzyType) {
				return resultType
			}
		}
		return nil
	case gobatis.StatementTypeDelete:
		if guess {
			return itf.detectRecordType(nil, false)
		}
	}
	return nil
}

func (itf *Interface) Print(ctx *PrintContext, sb *strings.Builder) {
	for _, comment := range itf.Comments {
		sb.WriteString(comment)
		sb.WriteString("\r\n")
	}
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

func (itf *Interface) ReferenceInterfaces() []string {
	var names []string
	for _, m := range itf.Methods {
		if m.Config == nil {
			continue
		}
		if m.Config.Reference == nil {
			continue
		}

		names = append(names, m.Config.Reference.Interface)
	}

	if len(names) == 0 {
		return nil
	}
	sort.Strings(names)
	offset := 1
	for i := 1; i < len(names); i++ {
		if names[i] == names[i-1] {
			continue
		}

		if offset != i {
			names[offset] = names[i]
		}
		offset++
	}

	names = names[:offset]
	return names
}
