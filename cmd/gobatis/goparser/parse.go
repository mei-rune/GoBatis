package goparser

import (
	"errors"
	"fmt"
	"go/ast"
	"go/format"
	goimporter "go/importer"
	"go/parser"
	"go/token"
	"go/types"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"runtime"
	"sort"
	"strconv"
	"strings"
)

type ParseContext struct {
	Mapper TypeMapper
}

type TypeMapper struct {
	TagName  string
	TagSplit func(string, string) []string
}

func (mapper *TypeMapper) Fields(st *types.Struct, cb func(string) bool) (string, bool) {
	var queue []*types.Struct
	queue = append(queue, st)
	for len(queue) != 0 {
		cur := queue[0]
		queue = queue[1:]

		for idx := 0; idx < cur.NumFields(); idx++ {
			v := cur.Field(idx)

			if v.Anonymous() {
				typ := v.Type()
				if p, ok := typ.(*types.Pointer); ok {
					typ = p.Elem()
				}
				if named, ok := typ.(*types.Named); ok {
					typ = named.Underlying()
				}
				if t, ok := typ.(*types.Struct); ok {
					queue = append(queue, t)
				}
			}

			if cb(v.Name()) {
				return v.Name(), true
			}

			tag := cur.Tag(idx)
			if tagValue, ok := reflect.StructTag(tag).Lookup(mapper.TagName); !ok && tagValue != "" {
				parts := mapper.TagSplit(tagValue, v.Name())
				fieldName := parts[0]
				if fieldName != "" && fieldName != "-" {
					if cb(fieldName) {
						return v.Name(), true
					}
				}
			}
		}
	}
	return "", false
}

type File struct {
	*ParseContext

	Source      string
	Package     string
	Imports     []string
	ImportAlias map[string]string // database/sql => sql
	Interfaces  []*Interface
}


func atoi(s string) int {
	i, _ := strconv.Atoi(s)
	return i
}
func goVersion() (int, int, int) { 
	version := strings.TrimPrefix(runtime.Version(), "go")
	ss := strings.Split(version, ".")
	switch len(ss) {
	case 0:
		return 0, 0, 0
	case 1:
		return atoi(ss[0]), 0, 0
	case 2:
		return atoi(ss[0]), atoi(ss[1]), 0
	case 3:
		return atoi(ss[0]), atoi(ss[1]), atoi(ss[2])
	default:
		return atoi(ss[0]), atoi(ss[1]), atoi(ss[2])
	}
}

func Parse(filename string, ctx *ParseContext) (*File, error) {
	goBuild(filename)

	dir := filepath.Dir(filename)
	if dir == "" {
		dir = "."
	}

	fset := token.NewFileSet()

	importer := goimporter.Default()
	if modEnable := os.Getenv("GO111MODULE"); modEnable == "on" {
		importer = goimporter.ForCompiler(fset, "source", nil)
	} else if modEnable == "" {
		if _, v, _ := goVersion(); v >= 17 {
			importer = goimporter.ForCompiler(fset, "source", nil)
		}
	}
	filenames, err := filepath.Glob(filepath.Join(dir, "*.go"))
	if err != nil {
		return nil, err
	}

	var files []*ast.File
	var current *ast.File
	for _, fname := range filenames {
		if strings.HasSuffix(fname, "_test.go") {
			continue
		}

		f, err := parser.ParseFile(fset, fname, nil, parser.ParseComments)
		if err != nil {
			if strings.HasSuffix(fname, "gobatis.go") {
				continue
			}
			return nil, err
		}
		files = append(files, f)

		if strings.HasSuffix(strings.ToLower(filepath.ToSlash(fname)),
			strings.ToLower(filepath.ToSlash(filename))) {
			current = f
		}
	}

	if current == nil {
		return nil, errors.New("`" + filename + "` isnot found")
	}

	return parse(ctx, fset, importer, files, filename, current)
}

func parse(ctx *ParseContext, fset *token.FileSet, importer types.Importer, files []*ast.File, filename string, f *ast.File) (*File, error) {
	store := &File{
		ParseContext: ctx,

		Source:      filename,
		Package:     f.Name.Name,
		ImportAlias: map[string]string{},
	}
	for _, importSpec := range f.Imports {
		pa, err := strconv.Unquote(importSpec.Path.Value)
		if err != nil {
			panic(err)
		}

		store.Imports = append(store.Imports, pa)
		if importSpec.Name != nil {
			if pa == "" {
				store.ImportAlias[importSpec.Path.Value] = importSpec.Name.Name
			} else {
				store.ImportAlias[pa] = importSpec.Name.Name
			}
		}
	}

	ifList, err := parseTypes(ctx, store, f, files, fset, importer)
	if err != nil {
		return nil, err
	}

	store.Interfaces = ifList
	return store, nil
}

func goBuild(src string) error {
	cmd := exec.Command("go", "build", "-i")
	cmd.Dir = filepath.Dir(src)
	out, err := cmd.CombinedOutput()
	if len(out) > 0 {
		fmt.Println("go build -i")
		fmt.Println(string(out))
	}
	if err != nil {
		fmt.Println(err)
	} else if len(out) == 0 {
		fmt.Println("run `go build -i` on `" + cmd.Dir + "` ok")
	}
	return err
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

func parseTypes(ctx *ParseContext, store *File, currentAST *ast.File, files []*ast.File, fset *token.FileSet, importer types.Importer) ([]*Interface, error) {
	info := types.Info{Defs: make(map[*ast.Ident]types.Object)}
	conf := types.Config{Importer: importer, Error: func(err error) {
		logPrint(err)
	}}
	_, err := conf.Check(store.Package, fset, files, &info)
	if err != nil {
		logPrint(err)
		// return nil, errors.New(err.Error())
	}

	var ifList []*Interface
	for k, obj := range info.Defs {
		if k.Obj == nil {
			logWarn(k.NamePos, k.Name, "ident object is nil")
			continue
		}
		if k.Obj.Kind != ast.Typ {
			logWarn(k.NamePos, k.Name, "ident object kind isnot Type, actual is", k.Obj.Kind.String())
			continue
		}
		if k.Obj.Decl == nil {
			logError(k.NamePos, k.Name, "ident object decl is nil")
			continue
		}
		typeSpec, ok := k.Obj.Decl.(*ast.TypeSpec)
		if !ok {
			logError(k.NamePos, k.Name, "ident object decl isnot TypeSpec, actual is", fmt.Sprintf("%T", k.Obj.Decl))
			continue
		}
		if typeSpec.Type == nil {
			logError(k.NamePos, k.Name, "ident object decl type is nil")
			continue
		}

		astInterfaceType, ok := typeSpec.Type.(*ast.InterfaceType)
		if !ok {
			logWarn(k.NamePos, k.Name, "ident object decl type isnot InterfaceType, actual is", fmt.Sprintf("%T", typeSpec.Type))
			continue
		}

		if obj.Type() == nil {
			logError(k.NamePos, k.Name, "object type is nil")
			continue
		}

		// get method name and params/returns
		itfType, ok := obj.Type().Underlying().(*types.Interface)
		if !ok {
			logError(k.NamePos, k.Name, "object type isnot interface{}, actual is",
				fmt.Sprintf("%T", obj.Type().Underlying()))
			continue
		}

		if o := currentAST.Scope.Lookup(typeSpec.Name.Name); o == nil {
			//fmt.Println(typeSpec.Name.Name, "isnot exists")
			continue
		}

		astObject := currentAST.Scope.Lookup(typeSpec.Name.Name)
		if astObject == nil {
			//fmt.Println(typeSpec.Name.Name, "isnot exists")
			continue
		}

		visitor := &parseVisitor{find: typeSpec}
		ast.Walk(visitor, currentAST)
		isSkipped := false
		for _, commentText := range visitor.Comments {
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

		itf := &Interface{
			ParseContext: ctx,
			File:         store,
			Pos:          int(k.Pos()),
			Name:         k.Name,
			Comments:     visitor.Comments,
		}
		for i := 0; i < itfType.NumEmbeddeds(); i++ {
			x := itfType.Embedded(i)
			xt := x.Obj()

			if xt.Pkg().Name() == store.Package {
				itf.EmbeddedInterfaces = append(itf.EmbeddedInterfaces, xt.Name())
			} else {
				itf.EmbeddedInterfaces = append(itf.EmbeddedInterfaces, x.String())
			}
		}

		if len(itf.EmbeddedInterfaces) == 0 {
			// go1.13 版本不知是为什么 itfType.NumEmbeddeds() 为 0
			for _, method := range astInterfaceType.Methods.List {
				if len(method.Names) == 0 {
					if ename, ok := method.Type.(*ast.Ident); ok {
						itf.EmbeddedInterfaces = append(itf.EmbeddedInterfaces, ename.Name)
					} else if ename, ok := method.Type.(*ast.SelectorExpr); ok {
						itf.EmbeddedInterfaces = append(itf.EmbeddedInterfaces, typePrint(ename))
					}
				}
			}
		}

		for i := 0; i < itfType.NumMethods(); i++ {
			x := itfType.Method(i)
			astMethod := findMethodByName(astInterfaceType, x.Name())
			if astMethod == nil {
				// this is method of embedded interface
				continue
			}

			doc := readMethodDoc(astMethod)
			pos := readMethodPos(astMethod)
			m, err := NewMethod(itf, pos, x.Name(), doc)
			if err != nil {
				return nil, errors.New("load document of method(" + x.Name() + ") fail at the file:" + strconv.Itoa(pos) + ": " + err.Error())
			}
			y := x.Type().(*types.Signature)
			method := astMethod.Type.(*ast.FuncType)
			m.Params = NewParams(m, method.Params, y.Params(), y.Variadic())
			m.Results = NewResults(m, method.Results, y.Results())
			itf.Methods = append(itf.Methods, m)
		}
		ifList = append(ifList, itf)
	}

	// 本函数对功能没有任何作用，只让接口和方法按文件中的顺序排序
	// 本函数主要是为了 TestParse()
	sort.Slice(ifList, func(i, j int) bool {
		return ifList[i].Pos <= ifList[j].Pos
	})
	for idx := range ifList {
		sort.Slice(ifList[idx].Methods, func(i, j int) bool {
			return ifList[idx].Methods[i].Pos <= ifList[idx].Methods[j].Pos
		})
	}
	return ifList, nil
}

func findMethodByName(ift *ast.InterfaceType, name string) *ast.Field {
	if ift == nil {
		return nil
	}

	for _, field := range ift.Methods.List {
		if len(field.Names) == 0 {
			// type a interface {}
			// type b  interface { a }
			continue
		}
		if field.Names[0].Name == name {
			return field
		}
	}
	return nil
}

func readMethodDoc(field *ast.Field) []string {
	if field == nil {
		return nil
	}

	if field.Doc == nil {
		return nil
	}
	ss := make([]string, len(field.Doc.List))
	for idx := range field.Doc.List {
		ss = append(ss, field.Doc.List[idx].Text)
	}
	return ss
}

func readMethodPos(field *ast.Field) int {
	if field == nil {
		return 0
	}

	return int(field.Pos())
}

type PrintContext struct {
	File      *File
	Interface *Interface
	Indent    string
}

func printTypename(sb *strings.Builder, typ types.Type, isVariadic bool) {
	var named *types.Named
	switch t := typ.(type) {
	case *types.Array:
		printTypename(sb, t.Elem(), isVariadic)
		return
	case *types.Slice:
		printTypename(sb, t.Elem(), isVariadic)
		return
	case *types.Map:
		sb.WriteString("map[")
		printTypename(sb, t.Key(), isVariadic)
		sb.WriteString("]")
		printTypename(sb, t.Elem(), isVariadic)
		return
	case *types.Pointer:
		if base := t.Elem(); base != nil {
			named, _ = base.(*types.Named)
		}
	case *types.Named:
		named = t
	}
	if named == nil || named.Obj() == nil || named.Obj().Pkg() == nil {
		sb.WriteString(typ.String())
		return
	}
	sb.WriteString(named.Obj().Name())
}

func PrintType(ctx *PrintContext, typ types.Type, isVariadic bool) string {
	var sb strings.Builder
	printType(ctx, &sb, typ, isVariadic)
	return sb.String()
}

func printType(ctx *PrintContext, sb *strings.Builder, typ types.Type, isVariadic bool) {
	if ctx == nil || ctx.File == nil {
		sb.WriteString(typ.String())
		return
	}

	var isPointer bool
	var named *types.Named
	switch t := typ.(type) {
	case *types.Array:
		sb.WriteString("[")
		sb.WriteString(strconv.FormatInt(t.Len(), 10))
		sb.WriteString("]")
		printType(ctx, sb, t.Elem(), false)
		return
	case *types.Slice:
		if isVariadic {
			sb.WriteString("...")
		} else {
			sb.WriteString("[]")
		}
		printType(ctx, sb, t.Elem(), false)
		return
	case *types.Map:
		sb.WriteString("map[")
		printType(ctx, sb, t.Key(), false)
		sb.WriteString("]")
		printType(ctx, sb, t.Elem(), false)
		return
	case *types.Pointer:
		if base := t.Elem(); base != nil {
			var ok bool
			if named, ok = base.(*types.Named); ok {
				isPointer = true
			}
		}
	case *types.Named:
		named = t
	}
	if named == nil || named.Obj() == nil || named.Obj().Pkg() == nil {
		sb.WriteString(types.TypeString(typ, types.Qualifier(func(other *types.Package) string {
			if a, ok := ctx.File.ImportAlias[other.Path()]; ok {
				return a
			}
			if ctx.File.Package == other.Path() {
				return "" // same package; unqualified
			}
			return other.Path()
		})))
		return
	}
	if isPointer {
		sb.WriteString("*")
	}

	if named.Obj().Pkg().Name() != ctx.File.Package {
		if a, ok := ctx.File.ImportAlias[named.Obj().Pkg().Path()]; ok {
			sb.WriteString(a)
		} else {
			sb.WriteString(named.Obj().Pkg().Name())
		}
		sb.WriteString(".")
	}
	sb.WriteString(named.Obj().Name())
}

type (
	parseVisitor struct {
		Comments []string
		find     *ast.TypeSpec
	}

	genDeclVisitor struct {
		node     *ast.GenDecl
		find     *ast.TypeSpec
		Comments *[]string
	}
)

func (v *parseVisitor) Visit(n ast.Node) ast.Visitor {
	switch rn := n.(type) {
	case *ast.File:
		return v
	case *ast.ImportSpec:
		return nil
	case *ast.FuncDecl:
		return nil
	case *ast.GenDecl:
		if rn.Tok == token.TYPE {
			return &genDeclVisitor{node: rn, find: v.find, Comments: &v.Comments}
		}
		return nil
	default:
		return v
	}
}

func (v *genDeclVisitor) Visit(n ast.Node) ast.Visitor {
	switch rn := n.(type) {
	case *ast.TypeSpec:
		if v.find == rn {
			if v.node.Doc != nil {
				for _, a := range v.node.Doc.List {
					*v.Comments = append(*v.Comments, a.Text)
				}
			}
		}
		return nil
	default:
		return v
	}
}

func typePrint(typ ast.Node) string {
	fset := token.NewFileSet()
	var buf strings.Builder
	if err := format.Node(&buf, fset, typ); err != nil {
		log.Fatalln(err)
	}
	return buf.String()
}
