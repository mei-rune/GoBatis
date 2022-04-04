package goparser2

import (
	"sort"
	"strings"

	gobatis "github.com/runner-mei/GoBatis"
	"github.com/runner-mei/GoBatis/cmd/gobatis/goparser2/astutil"
)

type Interface struct {
	Ctx  *ParseContext `json:"-"`
	File *File         `json:"-"`
	Name string

	EmbeddedInterfaces []string
	Comments           []string
	Methods            []*Method
}

func (itf *Interface) DetectRecordType(method *Method) *Type {
	return itf.detectRecordType(method, true)
}

func (itf *Interface) detectRecordType(method *Method, guess bool) *Type {
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
			if astutil.ToString(method.Params.List[idx].TypeExpr) == "context.Context" {
				continue
			}
			list = append(list, method.Params.List[idx])
		}

		if len(list) == 1 {
			// if method.Name == "InsertRoles1" {
			// 	fmt.Println(fmt.Sprintf("000 %T %s", list[0].Type, astutil.ToString(list[0].Type)))
			// 	fmt.Println("IsStructType(itf.File, list[0].Type)", IsStructType(itf.File, list[0].Type))
			// 	fmt.Println("!IsIgnoreStructTypes(itf.File, list[0].Type)", !IsIgnoreStructTypes(itf.File, list[0].Type))
			// }

			if list[0].Type().IsStructType() &&
				!list[0].Type().IsIgnoreStructTypes() {
				return list[0].Type().ElemType()
			}

			// if method.Name == "InsertRoles1" {
			// 	fmt.Println(fmt.Sprintf("111 %T %s", list[0].Type, astutil.ToString(list[0].Type)))
			// }
		}

		if guess {
			return itf.detectRecordType(nil, false)
		}
	case gobatis.StatementTypeUpdate:
		if len(method.Params.List) > 0 {
			param := method.Params.List[len(method.Params.List)-1]
			if param.Type().IsStructType() && !param.Type().IsIgnoreStructTypes() {
				return param.Type().ElemType()
			}
		}
		return itf.detectRecordType(nil, false)
	case gobatis.StatementTypeSelect:
		var typ Type
		if len(method.Results.List) == 1 {
			funcType, ok := astutil.ToFuncType(method.Results.List[0].TypeExpr)
			if !ok {
				return nil
			}

			signature := astutil.ToFunction(funcType)

			if signature.IsVariadic() {
				return nil
			}

			if len(signature.Params.List) != 1 {
				return nil
			}

			typ.Ctx = method.Interface.Ctx
			typ.File = method.Interface.File
			typ.TypeExpr = signature.Params.List[0].Typ
		} else if len(method.Results.List) == 2 {
			typ = method.Results.List[0].Type()
		} else {
			return nil
		}

		if !typ.IsStructType() {
			if guess {
				return itf.detectRecordType(nil, false)
			}
			return nil
		}

		resultType := typ.ElemType()
		if !guess {
			if resultType.IsStructType() &&
				!resultType.IsIgnoreStructTypes() {
				return resultType
			}
			return nil
		}

		if guess {
			fuzzyType := itf.detectRecordType(nil, false)
			if fuzzyType == nil || resultType.IsSameType(*fuzzyType) {
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
