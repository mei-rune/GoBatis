package goparser2

import (
	"errors"
	"fmt"
	"go/ast"
	"go/token"
	"log"
	"reflect"
	"strconv"
	"strings"
	"unicode"

	"github.com/runner-mei/GoBatis/cmd/gobatis/goparser2/astutil"
)

type ParseContext struct {
	*astutil.Context
	Mapper TypeMapper

	AnnotationPrefix string
	DbCompatibility bool
}

type TypeMapper struct {
	TagName  string
	TagSplit func(string, string) []string
}

func (mapper *TypeMapper) Fields(st *Type, cb func(string) bool) (string, bool) {
	ts, err := st.ToTypeSpec()
	if err != nil {
		panic(err)
	}

	var queue []*ast.StructType
	queue = append(queue, ts.Struct.Node)
	for len(queue) != 0 {
		cur := queue[0]
		queue = queue[1:]

		for _, v := range cur.Fields.List {
			if len(v.Names) > 0 {
				if cb(v.Names[0].Name) {
					return v.Names[0].Name, true
				}

				if v.Tag != nil {
					if tagValue, ok := reflect.StructTag(v.Tag.Value).Lookup(mapper.TagName); !ok && tagValue != "" {
						parts := mapper.TagSplit(tagValue, v.Names[0].Name)
						fieldName := parts[0]
						if fieldName != "" && fieldName != "-" {
							if cb(fieldName) {
								return v.Names[0].Name, true
							}
						}
					}
				}
				continue
			}

			// 嵌入字段

			typ := v.Type

			if p, ok := typ.(*ast.StarExpr); ok {
				typ = p.X
			}

			if ident, ok := typ.(*ast.Ident); ok {
				ts := st.File.Ctx.FindTypeInPackage(st.File, ident.Name)
				if ts == nil {
					panic(errors.New("'" + ident.Name + "' not found"))
				}

				if ts.Interface != nil {
					continue
				}
				if ts.Struct != nil {
					queue = append(queue, ts.Struct.Node)
					continue
				}
				typ = ts.Node.Type
			}

			if selectorExpr, ok := typ.(*ast.SelectorExpr); ok {
				ts, err := st.File.Ctx.FindTypeBySelectorExpr(st.File, selectorExpr)
				if err != nil {
					panic(err)
				}
				if ts == nil {
					panic(errors.New("'" + astutil.ToString(selectorExpr) + "' not found"))
				}

				if ts.Interface != nil {
					continue
				}
				if ts.Struct != nil {
					queue = append(queue, ts.Struct.Node)
					continue
				}
				typ = ts.Node.Type
			}

			if t, ok := typ.(*ast.StructType); ok {
				queue = append(queue, t)
			}
		}
	}
	return "", false
}

type File struct {
	Ctx *ParseContext
	// Ctx        *astutil.Context
	File        *astutil.File
	Source      string
	Package     string
	Imports     []string
	ImportAlias map[string]string // database/sql => sql
	Interfaces  []*Interface
}

func Parse(ctx *ParseContext, filename string) (*File, error) {
	if ctx == nil || ctx.Context == nil {
		panic("ParseContext is null")
		// ctx = astutil.NewContext(nil)
	}
	astFile, err := ctx.Context.LoadFile(filename)
	if err != nil {
		return nil, err
	}
	file := &File{
		Ctx:         ctx,
		File:        astFile,
		Source:      filename,
		Package:     astFile.Pkg.Name,
		ImportAlias: map[string]string{},
	}

	for _, importSpec := range astFile.Imports {
		pa, err := strconv.Unquote(importSpec.Path.Value)
		if err != nil {
			return nil, err
		}

		file.Imports = append(file.Imports, pa)
		if importSpec.Name != nil {
			if pa == "" {
				file.ImportAlias[importSpec.Path.Value] = importSpec.Name.Name
			} else {
				file.ImportAlias[pa] = importSpec.Name.Name
			}
		}
	}

	for idx := range astFile.TypeList {
		if astFile.TypeList[idx].Interface == nil {
			continue
		}

		isSkipped := false
		useNamespace := false
		customNamespace := ""
		sqlFragments := map[string][]Dialect{}

		commentTexts := splitByEmptyLine(joinComments(astFile.TypeList[idx].Node.Doc, astFile.TypeList[idx].Node.Comment))

			for _, commentText := range commentTexts {
				commentText = strings.TrimSpace(commentText)
				commentText = strings.TrimPrefix(commentText, "//")
				commentText = strings.TrimSpace(commentText)
				if commentText == "@gobatis.ignore" || commentText == "@gobatis.ignore()" {
					isSkipped = true
					continue
				}
				if commentText == "@gobatis.namespace" || commentText == "@gobatis.namespace()" {
					useNamespace = true
					continue
				}
				if strings.HasPrefix(commentText, "@gobatis.namespace(") {
					useNamespace = true
					customNamespace = strings.TrimPrefix(commentText, "@gobatis.namespace(")
					customNamespace = strings.TrimSuffix(customNamespace, ")")
					customNamespace = strings.TrimSpace(customNamespace)
					if strings.HasPrefix(customNamespace, "value=") {
						customNamespace = strings.TrimPrefix(customNamespace, "value=")
					} else if strings.HasPrefix(customNamespace, "value =") {
						customNamespace = strings.TrimPrefix(customNamespace, "value =")
					} else {
						return nil, errors.New("load document of " + astFile.TypeList[idx].Name + " fail: namespace invalid syntex")
					}
					customNamespace = strings.TrimSpace(customNamespace)
					continue
				}

				if strings.HasPrefix(commentText, "@gobatis.namespace ") {
					useNamespace = true
					customNamespace = strings.TrimPrefix(commentText, "@gobatis.namespace ")
					customNamespace = strings.TrimSpace(customNamespace)
					continue
				}

				if strings.HasPrefix(commentText, "@gobatis.namespace\t") {
					useNamespace = true
					customNamespace = strings.TrimPrefix(commentText, "@gobatis.namespace\t")
					customNamespace = strings.TrimSpace(customNamespace)
					continue
				}

				if strings.HasPrefix(commentText, "@gobatis.sql ") || strings.HasPrefix(commentText, "@gobatis.sql\t") {
					id, dialect, err := convertSqlFragment(ctx, file, strings.TrimPrefix(commentText, "@gobatis.sql"))
					if err != nil {
						return nil, errors.New("load document of " + astFile.TypeList[idx].Name + " fail: " + err.Error())
					}
					sqlFragments[id] = append(sqlFragments[id], dialect)
					continue
				}
			}
			
		if isSkipped {
			continue
		}

		class, err := convertClass(ctx, file, astFile.TypeList[idx])
		if err != nil {
			return nil, err
		}
		class.SqlFragments = sqlFragments
		class.UseNamespace = useNamespace
		if customNamespace == "" {
			class.CustomNamespace = class.Namespace
		} else {
			class.CustomNamespace = customNamespace
		}
		file.Interfaces = append(file.Interfaces, class)
	}
	return file, nil
}

func joinComments(doc ...*ast.CommentGroup) []string {
	var results []string
	for _, c := range doc {
		if c == nil {
			continue
		}

		for _, comment := range c.List {
			results = append(results, comment.Text)
		}
	}
	return results
}

func convertSqlFragment(ctx *ParseContext, file *File, sqlstr string) (string, Dialect, error) {
	sqlstr = strings.TrimSpace(sqlstr)
	idx := strings.IndexFunc(sqlstr, unicode.IsSpace)
	if idx < 0 {
		return "", Dialect{}, errors.New("id of sql fragment is missing")
	}

	id := sqlstr[:idx]
	sqlstr = sqlstr[idx:]
	sqlstr = strings.TrimSpace(sqlstr)

	idx = strings.IndexFunc(sqlstr, unicode.IsSpace)
	if idx < 0 {
		return "", Dialect{}, errors.New("dialect of sql fragment is missing")
	}

	dialect := sqlstr[:idx]
	sqlstr = sqlstr[idx:]
	sqlstr = strings.TrimSpace(sqlstr)

	return id, Dialect{
		DialectNames: []string{dialect},
		SQL:     sqlstr,
	}, nil
}

func convertClass(ctx *ParseContext, file *File, class *astutil.TypeSpec) (*Interface, error) {
	intf := &Interface{
		Ctx:      ctx,
		File:     file,
		Name:     class.Name,
		Comments: joinComments(class.Node.Doc, class.Node.Comment),
	}
	if class.File != nil && class.File.Pkg != nil {
		intf.Namespace = class.File.Pkg.Name
	}

	for _, embedded := range class.Interface.Embedded {
		intf.EmbeddedInterfaces = append(intf.EmbeddedInterfaces, astutil.ToString(embedded))
	}

	for idx := range class.Interface.Methods {
		if class.Interface.Methods[idx].Name == "WithDB" {
			continue
		}
		method, err := convertMethod(intf, class, &class.Interface.Methods[idx], ctx.AnnotationPrefix, ctx.DbCompatibility)
		if err != nil {
			return nil, err
		}

		intf.Methods = append(intf.Methods, method)
	}
	return intf, nil
}

func convertMethod(intf *Interface, class *astutil.TypeSpec,
	methodSpec *astutil.Method, annotationPrefix string, dbCompatibility bool) (*Method, error) {
	method, err := NewMethod(intf, methodSpec.Name,
		joinComments(methodSpec.Node.Doc, methodSpec.Node.Comment), annotationPrefix, dbCompatibility)
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

			param.Params = method.Params
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

			result.Results = method.Results
			method.Results.List = append(method.Results.List, *result)
		}
	}
	return method, nil
}

func convertParam(intf *Interface, method *Method, paramSpec *astutil.Param) (*Param, error) {
	param := &Param{
		Name:       paramSpec.Name,
		TypeExpr:   paramSpec.Expr,
		IsVariadic: paramSpec.IsVariadic,
	}
	return param, nil
}

func convertReturnResult(intf *Interface, method *Method, resultSpec *astutil.Result) (*Result, error) {
	result := &Result{
		Name:     resultSpec.Name,
		TypeExpr: resultSpec.Expr,
	}
	return result, nil
}

func logPrint(err error) {
	log.Println("3", err)
}

func logWarn(pos token.Pos, name string, args ...interface{}) {
	//log.Println(pos, ":", name, "-", args)
}

func logWarnf(pos token.Pos, name string, fmtStr string, args ...interface{}) {
	//log.Println(pos, ":", name, "-", fmt.Sprintf(fmtStr, args...))
}

func logError(pos token.Pos, name string, args ...interface{}) {
	log.Println("1", pos, ":", name, "-", args)
}

func logErrorf(pos token.Pos, name string, fmtStr string, args ...interface{}) {
	log.Println("2", pos, ":", name, "-", fmt.Sprintf(fmtStr, args...))
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
// 						itf.EmbeddedInterfaces = append(itf.EmbeddedInterfaces, ToString(ename))
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
// 			if a, ok := ctx.File.ImportAlias[other.Path()]; ok {
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
// 		if a, ok := ctx.File.ImportAlias[named.Obj().Pkg().Path()]; ok {
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
