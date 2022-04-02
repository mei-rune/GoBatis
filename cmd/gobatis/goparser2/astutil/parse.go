package astutil

import (
	"errors"
	"fmt"
	"go/ast"
	"go/format"
	"go/parser"
	"go/token"
	"io"
	"log"
	"os"
	"strings"
)

type (
	File struct {
		Ctx      *Context
		Pkg      *ast.Ident
		Filename string
		Imports  []*ast.ImportSpec
		Classes  []Class
		Types    []*ast.TypeSpec
	}

	parseVisitor struct {
		src *File
	}

	genDeclVisitor struct {
		src      *File
		node     *ast.GenDecl
		Comments []string
	}

	typeSpecVisitor struct {
		src         *File
		node        *ast.TypeSpec
		isInterface bool
		iface       *Class
		name        *ast.Ident
		comments    []string
	}

	Class struct {
		File        *File `json:"-"`
		Node        *ast.TypeSpec
		Name        string
		Comments    []string
		IsInterface bool

		Embedded []ast.Node
		Methods  []Method
		Fields   []Field
	}

	Field struct {
		File  *File  `json:"-"`
		Clazz *Class `json:"-"`
		Node  *ast.Field
		Name  string
		Typ   ast.Expr      // field/method/parameter type
		Tag   *ast.BasicLit // field tag; or nil
	}

	interfaceTypeVisitor struct {
		node     *ast.TypeSpec
		ts       *typeSpecVisitor
		embedded []ast.Node
		methods  []Method
	}

	structVisitor struct {
		node     *ast.TypeSpec
		ts       *typeSpecVisitor
		embedded []ast.Node
		methods  []Method
		fields   []Field
	}

	Method struct {
		File     *File  `json:"-"`
		Clazz    *Class `json:"-"`
		Node     ast.Node
		Name     string
		Comments []string
		Params   *Params
		Results  *Results
	}

	methodVisitor struct {
		depth    int
		node     ast.Node
		list     *[]Method
		name     *ast.Ident
		params   *Params
		results  *Results
		isMethod bool
	}

	Params struct {
		Method *Method `json:"-"`
		List   []Param
	}

	argListVisitor struct {
		list *Params
	}

	Param struct {
		Method     *Method `json:"-"`
		Name       string
		IsVariadic bool
		Typ        ast.Expr
	}

	argVisitor struct {
		node  *ast.TypeSpec
		parts []ast.Expr
		list  *Params
	}

	Results struct {
		Method *Method `json:"-"`
		List   []Result
	}

	resultListVisitor struct {
		list *Results
	}

	Result struct {
		Method *Method `json:"-"`
		Name   string
		Typ    ast.Expr
	}

	resultVisitor struct {
		node  *ast.TypeSpec
		parts []ast.Expr
		list  *Results
	}
)

func (v *parseVisitor) Visit(n ast.Node) ast.Visitor {
	switch rn := n.(type) {
	case *ast.File:
		v.src.Pkg = rn.Name
		return v
	case *ast.ImportSpec:
		v.src.Imports = append(v.src.Imports, rn)
		return nil
	case *ast.FuncDecl:
		if rn.Recv == nil || len(rn.Recv.List) == 0 {
			return nil
		}

		var name string
		if star, ok := rn.Recv.List[0].Type.(*ast.StarExpr); ok {
			name = star.X.(*ast.Ident).Name
		} else if ident, ok := rn.Recv.List[0].Type.(*ast.Ident); ok {
			name = ident.Name
		} else {
			log.Fatalln(fmt.Errorf("func.recv is unknown type - %T", rn.Recv.List[0].Type))
		}
		var class *Class
		for idx := range v.src.Classes {
			if name == v.src.Classes[idx].Name {
				class = &v.src.Classes[idx]
				break
			}
		}

		if class == nil {
			for idx := range v.src.Types {
				if name == v.src.Types[idx].Name.Name {
					return nil
				}
			}

			log.Fatalln(errors.New(v.src.PostionFor(rn.Pos()).String() + ": 请先定义类型，后定义 方法"))
		}

		mv := &methodVisitor{node: &ast.Field{Doc: rn.Doc, Names: []*ast.Ident{rn.Name}, Type: rn.Type}, list: &class.Methods}
		ast.Walk(mv, mv.node)
		return nil
	case *ast.GenDecl:
		if rn.Tok == token.TYPE {
			return &genDeclVisitor{src: v.src, node: rn}
		}
		return v
	default:
		return v
	}
}

func (v *genDeclVisitor) Visit(n ast.Node) ast.Visitor {
	switch rn := n.(type) {
	case *ast.TypeSpec:
		switch rn.Type.(type) {
		case *ast.InterfaceType:
		default:
			v.src.Types = append(v.src.Types, rn)
		}

		var comments []string
		if v.node.Doc != nil {
			for _, a := range v.node.Doc.List {
				comments = append(comments, a.Text)
			}
		}

		return &typeSpecVisitor{src: v.src, node: rn, comments: comments}
	default:
		return v
	}
}

/*
package foo

type FooService interface {
	Bar(ctx context.Context, i int, s string) (string, error)
}
*/

func (v *typeSpecVisitor) Visit(n ast.Node) ast.Visitor {
	switch rn := n.(type) {
	case *ast.Ident:
		if v.name == nil {
			v.name = rn
		}
		return v
	case *ast.StructType:
		v.isInterface = false
		return &structVisitor{ts: v, methods: []Method{}}
	case *ast.InterfaceType:
		v.isInterface = true
		return &interfaceTypeVisitor{ts: v, methods: []Method{}}
	case nil:
		if v.iface != nil {
			v.iface.IsInterface = v.isInterface
			v.iface.File = v.src
			v.iface.Node = v.node
			v.iface.Name = v.name.Name
			v.iface.Comments = v.comments

			if v.node.Comment != nil {
				for _, a := range v.node.Comment.List {
					v.iface.Comments = append(v.iface.Comments, a.Text)
				}
			}

			if v.node.Doc != nil {
				for _, a := range v.node.Doc.List {
					v.iface.Comments = append(v.iface.Comments, a.Text)
				}
			}

			v.src.Classes = append(v.src.Classes, *v.iface)
		}
		return nil
	default:
		return v
	}
}

func (v *structVisitor) Visit(n ast.Node) ast.Visitor {
	switch rn := n.(type) {
	default:
		return v
	case *ast.FieldList:
		return v
	case *ast.Field:

		if len(rn.Names) == 0 {
			v.embedded = append(v.embedded, rn.Type)
			return nil
		}

		for _, name := range rn.Names {
			v.fields = append(v.fields, Field{Node: rn, Name: name.Name, Typ: rn.Type, Tag: rn.Tag})
		}
		return nil
	case nil:
		v.ts.iface = &Class{Methods: v.methods, Fields: v.fields, Embedded: v.embedded}
		return nil
	}
}

func (v *interfaceTypeVisitor) Visit(n ast.Node) ast.Visitor {
	switch rn := n.(type) {
	case *ast.Field:

		if len(rn.Names) == 0 {
			v.embedded = append(v.embedded, rn.Type)
			return nil
		}

		return &methodVisitor{node: rn, list: &v.methods}
	case nil:
		v.ts.iface = &Class{Methods: v.methods, Embedded: v.embedded}
		for idx := range v.ts.iface.Methods {
			v.ts.iface.Methods[idx].init(v.ts.iface)
		}
		return nil
	default:
		return v
	}
}

func (v *methodVisitor) Visit(n ast.Node) ast.Visitor {
	switch rn := n.(type) {
	default:
		v.depth++
		return v
	case *ast.Ident:
		v.name = rn
		v.depth++
		return v
	case *ast.FuncLit:
		v.depth++
		return v
	case *ast.FuncType:
		v.depth++
		v.isMethod = true
		return v
	case *ast.FieldList:
		if v.params == nil {
			v.params = &Params{}
			return &argListVisitor{list: v.params}
		}
		if v.results == nil {
			v.results = &Results{}
		}
		return &resultListVisitor{list: v.results}
	case *ast.BlockStmt:
		return nil
	case nil:
		v.depth--
		if v.depth == 0 && v.isMethod {
			var comments []string
			if field, ok := v.node.(*ast.Field); ok {
				if field.Comment != nil {
					for _, a := range field.Comment.List {
						comments = append(comments, a.Text)
					}
				}

				if field.Doc != nil {
					for _, a := range field.Doc.List {
						comments = append(comments, a.Text)
					}
				}
			}

			var name string
			if v.name != nil {
				name = v.name.Name
			}

			*v.list = append(*v.list, Method{Node: v.node,
				Name:     name,
				Params:   v.params,
				Results:  v.results,
				Comments: comments})
		}
		return nil
	}
}

func (v *argListVisitor) Visit(n ast.Node) ast.Visitor {
	switch n.(type) {
	default:
		return nil
	case *ast.Field:
		return &argVisitor{list: v.list}
	}
}

func (v *argVisitor) Visit(n ast.Node) ast.Visitor {
	switch t := n.(type) {
	case *ast.CommentGroup, *ast.BasicLit:
		return nil
	case *ast.Ident: //Expr -> everything, but clarity
		if t.Name != "_" {
			v.parts = append(v.parts, t)
		}
	case ast.Expr:
		v.parts = append(v.parts, t)
	case nil:
		names := v.parts[:len(v.parts)-1]
		tp := v.parts[len(v.parts)-1]
		if len(names) == 0 {
			v.list.List = append(v.list.List, Param{Typ: tp})
			return nil
		}
		for _, n := range names {
			v.list.List = append(v.list.List, Param{
				Name: n.(*ast.Ident).Name,
				Typ:  tp,
			})
		}
	}
	return nil
}

func (v *resultListVisitor) Visit(n ast.Node) ast.Visitor {
	switch n.(type) {
	default:
		return nil
	case *ast.Field:
		return &resultVisitor{list: v.list}
	}
}

func (v *resultVisitor) Visit(n ast.Node) ast.Visitor {
	switch t := n.(type) {
	case *ast.CommentGroup:
		return nil
	case *ast.BasicLit:
		return nil
	case *ast.Ident: //Expr -> everything, but clarity
		if t.Name != "_" {
			v.parts = append(v.parts, t)
		}
	case ast.Expr:
		v.parts = append(v.parts, t)
	case nil:
		names := v.parts[:len(v.parts)-1]
		tp := v.parts[len(v.parts)-1]
		if len(names) == 0 {
			v.list.List = append(v.list.List, Result{Typ: tp})
			return nil
		}
		for _, n := range names {
			v.list.List = append(v.list.List, Result{
				Name: n.(*ast.Ident).Name,
				Typ:  tp,
			})
		}
	}
	return nil
}

func (sc *File) ImportPath(selectorExpr *ast.SelectorExpr) string {
	impName := ToString(selectorExpr.X)
	for _, imp := range sc.Imports {
		impPath := ToString(imp.Path)
		if imp.Name != nil {
			if ToString(imp.Name) == impName {
				return impPath
			}
		}

		ss := strings.Split(impPath, "/")
		if ss[len(ss)-1] == impName {
			return impPath
		}
	}

	return ""
}

func (sc *File) PostionFor(pos token.Pos) token.Position {
	return sc.Ctx.FileSet.PositionFor(pos, true)
}

func (file *File) GetType(name string) *ast.TypeSpec {
	for _, typ := range file.Types {
		if typ.Name.Name == name {
			return typ
		}
	}
	return nil
}

func (file *File) GetClass(name string) *Class {
	for idx := range file.Classes {
		if file.Classes[idx].Name == name {
			return &file.Classes[idx]
		}
	}
	return nil
}

func (sc *File) validate() error {
	//	for _, i := range sc.Classes {
	//		for _, m := range i.Methods {
	//			if m.Results == nil || len(m.Results.List) < 1 {
	//				return fmt.Errorf("method %q of interface %q has no result types", m.Name, i.Name)
	//			}
	//		}
	//	}
	return nil
}

func (method *Method) init(iface *Class) {
	method.Clazz = iface
	if method.Params != nil {
		method.Params.Method = method
		for j := range method.Params.List {
			method.Params.List[j].Method = method
		}
	}

	if method.Results != nil {
		method.Results.Method = method
		for j := range method.Results.List {
			method.Results.List[j].Method = method
		}
	}
}

func Parse(ctx *Context, filename string, source io.Reader) (*File, error) {
	f, err := parser.ParseFile(ctx.FileSet, filename, source, parser.DeclarationErrors|parser.ParseComments)
	if err != nil {
		return nil, errors.New("parsing input file '" + filename + "': " + err.Error())
	}

	file := &File{
		Filename: filename,
		Ctx:      ctx,
	}
	visitor := &parseVisitor{src: file}
	ast.Walk(visitor, f)

	for classIdx := range file.Classes {
		file.Classes[classIdx].File = file

		for idx := range file.Classes[classIdx].Methods {
			method := &file.Classes[classIdx].Methods[idx]
			method.File = file
			method.init(&file.Classes[classIdx])
		}

		for idx := range file.Classes[classIdx].Fields {
			file.Classes[classIdx].Fields[idx].File = file
		}
	}
	if err := file.validate(); err != nil {
		return nil, errors.New("examining input file '" + filename + "': " + err.Error())
	}

	return file, nil
}

func ParseFile(ctx *Context, filename string) (*File, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, errors.New("error while opening '" + filename + "': " + err.Error())
	}
	defer file.Close()
	return Parse(ctx, filename, file)
}

func (method *Method) IsVariadic() bool {
	return method.Params.List[len(method.Params.List)-1].IsVariadic
}

func ToMethod(node ast.Node) (*Method, bool) {
	field, ok := node.(*ast.Field)
	if !ok {
		funcType, ok := node.(*ast.FuncType)
		if !ok {
			return nil, false
		}

		list := make([]Method, 0, 1)
		visitor := &methodVisitor{node: node, list: &list}
		ast.Walk(visitor, funcType)
		return &list[0], true
	}

	list := make([]Method, 0, 1)
	visitor := &methodVisitor{node: field, list: &list}
	ast.Walk(visitor, field)
	return &list[0], true
}

func ToString(typ ast.Node) string {
	fset := token.NewFileSet()
	var buf strings.Builder
	if err := format.Node(&buf, fset, typ); err != nil {
		log.Fatalln(err)
	}
	return buf.String()
}

var RangeDefineds = map[string]struct {
	Start ast.Expr
	End   ast.Expr
}{}

func AddRangeDefined(typ, start, end string) {
	var s ast.Expr = &ast.Ident{Name: strings.TrimPrefix(start, "*")}
	var e ast.Expr = &ast.Ident{Name: strings.TrimPrefix(end, "*")}

	if strings.HasPrefix(start, "*") {
		s = &ast.StarExpr{X: s}
	}

	if strings.HasPrefix(end, "*") {
		e = &ast.StarExpr{X: e}
	}

	RangeDefineds[typ] = struct {
		Start ast.Expr
		End   ast.Expr
	}{s, e}
}

func IsRangeStruct(classes []Class, typ ast.Expr) (bool, ast.Expr, ast.Expr) {
	name := strings.TrimPrefix(ToString(typ), "*")

	if value, ok := RangeDefineds[name]; ok {
		return true, value.Start, value.End
	}

	var cls *Class
	for idx := range classes {
		if classes[idx].Name == name {
			cls = &classes[idx]
			break
		}
	}
	if cls == nil {
		return false, nil, nil
	}

	if len(cls.Fields) != 2 {
		return false, nil, nil
	}

	var startType, endType ast.Expr
	for _, field := range cls.Fields {
		if field.Name == "Start" {
			startType = field.Typ
		} else if field.Name == "End" {
			endType = field.Typ
		}
	}
	if startType == nil || endType == nil {
		return false, nil, nil
	}

	aType := strings.TrimPrefix(ToString(startType), "*")
	bType := strings.TrimPrefix(ToString(endType), "*")
	if aType != bType {
		return false, nil, nil
	}
	return true, startType, endType
}

func IsPtrType(typ ast.Expr) bool {
	_, ok := typ.(*ast.StarExpr)
	return ok
}

func PtrElemType(typ ast.Expr) ast.Expr {
	star, ok := typ.(*ast.StarExpr)
	if ok {
		return star.X
	}
	return nil
}

func IsStructType(typ ast.Expr) bool {
	_, ok := typ.(*ast.StructType)
	return ok
}

func IsArrayOrSliceType(typ ast.Expr) bool {
	_, ok := typ.(*ast.ArrayType)
	return ok
}

func IsSliceType(typ ast.Expr) bool {
	aType, ok := typ.(*ast.ArrayType)
	if !ok {
		return false
	}
	return aType.Len == nil
}

func IsArrayType(typ ast.Expr) bool {
	aType, ok := typ.(*ast.ArrayType)
	if !ok {
		return false
	}
	return aType.Len != nil
}

func IsEllipsisType(typ ast.Expr) bool {
	_, ok := typ.(*ast.Ellipsis)
	return ok
}

func IsMapType(typ ast.Expr) bool {
	_, ok := typ.(*ast.MapType)
	return ok
}

func IsSameType(a, b ast.Expr) bool {
	as := ToString(a)
	bs := ToString(b)
	return as == bs
}

func KeyType(typ ast.Expr) ast.Expr {
	m, ok := typ.(*ast.MapType)
	if ok {
		return m.Key
	}
	return nil
}

func MapValueType(typ ast.Expr) ast.Expr {
	m, ok := typ.(*ast.MapType)
	if ok {
		return m.Value
	}
	return nil
}

func MapKeyType(typ ast.Expr) ast.Expr {
	m, ok := typ.(*ast.MapType)
	if ok {
		return m.Key
	}
	return nil
}

func ElemType(typ ast.Expr) ast.Expr {
	switch t := typ.(type) {
	case *ast.StarExpr:
		return t.X
	case *ast.ArrayType:
		return t.Elt
	case *ast.Ellipsis:
		return t.Elt
	}
	return nil
}
