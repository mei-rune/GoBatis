package goparser2

import (
	"errors"
	"strconv"
	"strings"

	"github.com/runner-mei/GoBatis/goparser2/astutil"
)

type File struct {
	Ctx        *astutil.Context
	File       *astutil.File
	Source     string
	Package    string
	Imports    []string
	ImportAlas map[string]string // database/sql => sql
	Interfaces []*Interface
}

func Parse(ctx *astutil.Context, filename string) (*File, error) {
	if ctx == nil {
		ctx = astutil.NewContext(nil)
	}
	astFile, err := astutil.ParseFile(ctx, filename)
	if err != nil {
		return nil, err
	}
	file := &File{
		Ctx:        ctx,
		File:       astFile,
		Source:     filename,
		Package:    astFile.Pkg.Name,
		ImportAlas: map[string]string{},
	}

	for _, importSpec := range astFile.Imports {
		pa, err := strconv.Unquote(importSpec.Path.Value)
		if err != nil {
			return nil, err
		}

		file.Imports = append(file.Imports, pa)
		if importSpec.Name != nil {
			if pa == "" {
				file.ImportAlas[importSpec.Path.Value] = importSpec.Name.Name
			} else {
				file.ImportAlas[pa] = importSpec.Name.Name
			}
		}
	}

	for idx := range astFile.Classes {
		if !astFile.Classes[idx].IsInterface {
			continue
		}

		isSkipped := false
		for _, commentText := range astFile.Classes[idx].Comments {
			commentText := strings.TrimSpace(commentText)
			commentText = strings.TrimPrefix(commentText, "//")
			commentText = strings.TrimSpace(commentText)
			if commentText == "@gobatis.ignore" || commentText == "@gobatis.ignore()" {
				isSkipped = true
				break
			}
		}
		if isSkipped {
			continue
		}

		class, err := convertClass(file, &astFile.Classes[idx])
		if err != nil {
			return nil, err
		}

		file.Interfaces = append(file.Interfaces, class)
	}
	return file, nil
}

func convertClass(file *File, class *astutil.Class) (*Interface, error) {
	intf := &Interface{
		File:     file,
		Name:     class.Name,
		Comments: class.Comments,
	}

	for _, embedded := range class.Embedded {
		intf.EmbeddedInterfaces = append(intf.EmbeddedInterfaces, astutil.TypePrint(embedded))
	}

	for idx := range class.Methods {
		method, err := convertMethod(intf, class, &class.Methods[idx])
		if err != nil {
			return nil, err
		}

		intf.Methods = append(intf.Methods, method)
	}
	return intf, nil
}

func convertMethod(intf *Interface, class *astutil.Class,
	methodSpec *astutil.Method) (*Method, error) {
	method, err := NewMethod(intf, methodSpec.Name, methodSpec.Comments)
	if err != nil {
		return nil, errors.New("load document of " + intf.Name + "." + methodSpec.Name + "(...) fail: " + err.Error())
	}

	method.Params = &Params{
		Method: method,
	}

	if methodSpec.Params != nil {
		for idx := range methodSpec.Params.List {
			param, err := convertParam(intf, method, &methodSpec.Params.List[idx])
			if err != nil {
				return nil, err
			}

			method.Params.List = append(method.Params.List, *param)
		}
	}

	method.Results = &Results{
		Method: method,
	}
	if methodSpec.Results != nil {
		for idx := range methodSpec.Results.List {
			result, err := convertReturnResult(intf, method, &methodSpec.Results.List[idx])
			if err != nil {
				return nil, err
			}

			method.Results.List = append(method.Results.List, *result)
		}
	}
	return method, nil
}

func convertParam(intf *Interface, method *Method, paramSpec *astutil.Param) (*Param, error) {
	param := &Param{
		Name:       paramSpec.Name,
		Type:       paramSpec.Typ,
		IsVariadic: paramSpec.IsVariadic,
	}
	return param, nil
}

func convertReturnResult(intf *Interface, method *Method, resultSpec *astutil.Result) (*Result, error) {
	result := &Result{
		Name: resultSpec.Name,
		Type: resultSpec.Typ,
	}
	return result, nil
}

// func parseTypes(store *File, currentAST *ast.File, files []*ast.File, fset *token.FileSet, importer types.Importer) ([]*Interface, error) {
// 	info := types.Info{Defs: make(map[*ast.Ident]types.Object)}
// 	conf := types.Config{Importer: importer, Error: func(err error) {
// 		logPrint(err)
// 	}}
// 	_, err := conf.Check(store.Package, fset, files, &info)
// 	if err != nil {
// 		logPrint(err)
// 		// return nil, errors.New(err.Error())
// 	}

// 	var ifList []*Interface
// 	for k, obj := range info.Defs {
// 		if k.Obj == nil {
// 			logWarn(k.NamePos, k.Name, "ident object is nil")
// 			continue
// 		}
// 		if k.Obj.Kind != ast.Typ {
// 			logWarn(k.NamePos, k.Name, "ident object kind isnot Type, actual is", k.Obj.Kind.String())
// 			continue
// 		}
// 		if k.Obj.Decl == nil {
// 			logError(k.NamePos, k.Name, "ident object decl is nil")
// 			continue
// 		}
// 		typeSpec, ok := k.Obj.Decl.(*ast.TypeSpec)
// 		if !ok {
// 			logError(k.NamePos, k.Name, "ident object decl isnot TypeSpec, actual is", fmt.Sprintf("%T", k.Obj.Decl))
// 			continue
// 		}
// 		if typeSpec.Type == nil {
// 			logError(k.NamePos, k.Name, "ident object decl type is nil")
// 			continue
// 		}

// 		astInterfaceType, ok := typeSpec.Type.(*ast.InterfaceType)
// 		if !ok {
// 			logWarn(k.NamePos, k.Name, "ident object decl type isnot InterfaceType, actual is", fmt.Sprintf("%T", typeSpec.Type))
// 			continue
// 		}

// 		if obj.Type() == nil {
// 			logError(k.NamePos, k.Name, "object type is nil")
// 			continue
// 		}

// 		// get method name and params/returns
// 		itfType, ok := obj.Type().Underlying().(*types.Interface)
// 		if !ok {
// 			logError(k.NamePos, k.Name, "object type isnot interface{}, actual is",
// 				fmt.Sprintf("%T", obj.Type().Underlying()))
// 			continue
// 		}

// 		if o := currentAST.Scope.Lookup(typeSpec.Name.Name); o == nil {
// 			//fmt.Println(typeSpec.Name.Name, "isnot exists")
// 			continue
// 		}

// 		astObject := currentAST.Scope.Lookup(typeSpec.Name.Name)
// 		if astObject == nil {
// 			//fmt.Println(typeSpec.Name.Name, "isnot exists")
// 			continue
// 		}

// 		visitor := &parseVisitor{find: typeSpec}
// 		ast.Walk(visitor, currentAST)
// 		isSkipped := false
// 		for _, commentText := range visitor.Comments {
// 			commentText := strings.TrimSpace(commentText)
// 			commentText = strings.TrimPrefix(commentText, "//")
// 			commentText = strings.TrimSpace(commentText)
// 			if commentText == "@gobatis.ignore" || commentText == "@gobatis.ignore()" {
// 				isSkipped = true
// 				break
// 			}
// 		}
// 		if isSkipped {
// 			continue
// 		}

// 		itf := &Interface{
// 			File:     store,
// 			Pos:      int(k.Pos()),
// 			Name:     k.Name,
// 			Comments: visitor.Comments,
// 		}
// 		for i := 0; i < itfType.NumEmbeddeds(); i++ {
// 			x := itfType.Embedded(i)
// 			xt := x.Obj()

// 			if xt.Pkg().Name() == store.Package {
// 				itf.EmbeddedInterfaces = append(itf.EmbeddedInterfaces, xt.Name())
// 			} else {
// 				itf.EmbeddedInterfaces = append(itf.EmbeddedInterfaces, x.String())
// 			}
// 		}

// 		if len(itf.EmbeddedInterfaces) == 0 {
// 			// go1.13 版本不知是为什么 itfType.NumEmbeddeds() 为 0
// 			for _, method := range astInterfaceType.Methods.List {
// 				if len(method.Names) == 0 {
// 					if ename, ok := method.Type.(*ast.Ident); ok {
// 						itf.EmbeddedInterfaces = append(itf.EmbeddedInterfaces, ename.Name)
// 					} else if ename, ok := method.Type.(*ast.SelectorExpr); ok {
// 						itf.EmbeddedInterfaces = append(itf.EmbeddedInterfaces, typePrint(ename))
// 					}
// 				}
// 			}
// 		}

// 		for i := 0; i < itfType.NumMethods(); i++ {
// 			x := itfType.Method(i)
// 			astMethod := findMethodByName(astInterfaceType, x.Name())
// 			if astMethod == nil {
// 				// this is method of embedded interface
// 				continue
// 			}

// 			doc := readMethodDoc(astMethod)
// 			pos := readMethodPos(astMethod)
// 			m, err := NewMethod(itf, pos, x.Name(), doc)
// 			if err != nil {
// 				return nil, errors.New("load document of method(" + x.Name() + ") fail at the file:" + strconv.Itoa(pos) + ": " + err.Error())
// 			}
// 			y := x.Type().(*types.Signature)
// 			m.Params = NewParams(m, y.Params(), y.Variadic())
// 			m.Results = NewResults(m, y.Results())
// 			itf.Methods = append(itf.Methods, m)
// 		}
// 		ifList = append(ifList, itf)
// 	}

// 	// 本函数对功能没有任何作用，只让接口和方法按文件中的顺序排序
// 	// 本函数主要是为了 TestParse()
// 	sort.Slice(ifList, func(i, j int) bool {
// 		return ifList[i].Pos <= ifList[j].Pos
// 	})
// 	for idx := range ifList {
// 		sort.Slice(ifList[idx].Methods, func(i, j int) bool {
// 			return ifList[idx].Methods[i].Pos <= ifList[idx].Methods[j].Pos
// 		})
// 	}
// 	return ifList, nil
// }

// func findMethodByName(ift *ast.InterfaceType, name string) *ast.Field {
// 	if ift == nil {
// 		return nil
// 	}

// 	for _, field := range ift.Methods.List {
// 		if len(field.Names) == 0 {
// 			// type a interface {}
// 			// type b  interface { a }
// 			continue
// 		}
// 		if field.Names[0].Name == name {
// 			return field
// 		}
// 	}
// 	return nil
// }

// func readMethodDoc(field *ast.Field) []string {
// 	if field == nil {
// 		return nil
// 	}

// 	if field.Doc == nil {
// 		return nil
// 	}
// 	ss := make([]string, len(field.Doc.List))
// 	for idx := range field.Doc.List {
// 		ss = append(ss, field.Doc.List[idx].Text)
// 	}
// 	return ss
// }

// func readMethodPos(field *ast.Field) int {
// 	if field == nil {
// 		return 0
// 	}

// 	return int(field.Pos())
// }

type PrintContext struct {
	File      *File
	Interface *Interface
	Indent    string
}

// func printTypename(sb *strings.Builder, typ types.Type, isVariadic bool) {
// 	var named *types.Named
// 	switch t := typ.(type) {
// 	case *types.Array:
// 		printTypename(sb, t.Elem(), isVariadic)
// 		return
// 	case *types.Slice:
// 		printTypename(sb, t.Elem(), isVariadic)
// 		return
// 	case *types.Map:
// 		sb.WriteString("map[")
// 		printTypename(sb, t.Key(), isVariadic)
// 		sb.WriteString("]")
// 		printTypename(sb, t.Elem(), isVariadic)
// 		return
// 	case *types.Pointer:
// 		if base := t.Elem(); base != nil {
// 			named, _ = base.(*types.Named)
// 		}
// 	case *types.Named:
// 		named = t
// 	}
// 	if named == nil || named.Obj() == nil || named.Obj().Pkg() == nil {
// 		sb.WriteString(typ.String())
// 		return
// 	}
// 	sb.WriteString(named.Obj().Name())
// }

// func PrintType(ctx *PrintContext, typ types.Type, isVariadic bool) string {
// 	var sb strings.Builder
// 	printType(ctx, &sb, typ, isVariadic)
// 	return sb.String()
// }

// func printType(ctx *PrintContext, sb *strings.Builder, typ types.Type, isVariadic bool) {
// 	if ctx == nil || ctx.File == nil {
// 		sb.WriteString(typ.String())
// 		return
// 	}

// 	var isPointer bool
// 	var named *types.Named
// 	switch t := typ.(type) {
// 	case *types.Array:
// 		sb.WriteString("[")
// 		sb.WriteString(strconv.FormatInt(t.Len(), 10))
// 		sb.WriteString("]")
// 		printType(ctx, sb, t.Elem(), false)
// 		return
// 	case *types.Slice:
// 		if isVariadic {
// 			sb.WriteString("...")
// 		} else {
// 			sb.WriteString("[]")
// 		}
// 		printType(ctx, sb, t.Elem(), false)
// 		return
// 	case *types.Map:
// 		sb.WriteString("map[")
// 		printType(ctx, sb, t.Key(), false)
// 		sb.WriteString("]")
// 		printType(ctx, sb, t.Elem(), false)
// 		return
// 	case *types.Pointer:
// 		if base := t.Elem(); base != nil {
// 			var ok bool
// 			if named, ok = base.(*types.Named); ok {
// 				isPointer = true
// 			}
// 		}
// 	case *types.Named:
// 		named = t
// 	}
// 	if named == nil || named.Obj() == nil || named.Obj().Pkg() == nil {
// 		sb.WriteString(types.TypeString(typ, types.Qualifier(func(other *types.Package) string {
// 			if a, ok := ctx.File.ImportAlas[other.Path()]; ok {
// 				return a
// 			}
// 			if ctx.File.Package == other.Path() {
// 				return "" // same package; unqualified
// 			}
// 			return other.Path()
// 		})))
// 		return
// 	}
// 	if isPointer {
// 		sb.WriteString("*")
// 	}

// 	if named.Obj().Pkg().Name() != ctx.File.Package {
// 		if a, ok := ctx.File.ImportAlas[named.Obj().Pkg().Path()]; ok {
// 			sb.WriteString(a)
// 		} else {
// 			sb.WriteString(named.Obj().Pkg().Name())
// 		}
// 		sb.WriteString(".")
// 	}
// 	sb.WriteString(named.Obj().Name())
// }

// type (
// 	parseVisitor struct {
// 		Comments []string
// 		find     *ast.TypeSpec
// 	}

// 	genDeclVisitor struct {
// 		node     *ast.GenDecl
// 		find     *ast.TypeSpec
// 		Comments *[]string
// 	}
// )

// func (v *parseVisitor) Visit(n ast.Node) ast.Visitor {
// 	switch rn := n.(type) {
// 	case *ast.File:
// 		return v
// 	case *ast.ImportSpec:
// 		return nil
// 	case *ast.FuncDecl:
// 		return nil
// 	case *ast.GenDecl:
// 		if rn.Tok == token.TYPE {
// 			return &genDeclVisitor{node: rn, find: v.find, Comments: &v.Comments}
// 		}
// 		return nil
// 	default:
// 		return v
// 	}
// }

// func (v *genDeclVisitor) Visit(n ast.Node) ast.Visitor {
// 	switch rn := n.(type) {
// 	case *ast.TypeSpec:
// 		if v.find == rn {
// 			if v.node.Doc != nil {
// 				for _, a := range v.node.Doc.List {
// 					*v.Comments = append(*v.Comments, a.Text)
// 				}
// 			}
// 		}
// 		return nil
// 	default:
// 		return v
// 	}
// }

// func typePrint(typ ast.Node) string {
// 	fset := token.NewFileSet()
// 	var buf strings.Builder
// 	if err := format.Node(&buf, fset, typ); err != nil {
// 		log.Fatalln(err)
// 	}
// 	return buf.String()
// }
