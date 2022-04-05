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
		Package  *Package
		Pkg      *ast.Ident
		Filename string
		Imports  []*ast.ImportSpec
		TypeList []*TypeSpec

		Methods map[string][]Method
	}

	TypeSpec struct {
		File      *File `json:"-"`
		Node      *ast.TypeSpec
		Name      string
		Struct    *Struct
		Interface *Interface
	}

	Struct struct {
		Node *ast.StructType

		Embedded []ast.Node
		Fields   []Field
		Methods  []Method
	}

	Interface struct {
		Node *ast.InterfaceType

		Embedded []ast.Node
		Methods  []Method
	}

	Field struct {
		Clazz *TypeSpec `json:"-"`
		Node  *ast.Field
		Name  string
		Typ   ast.Expr      // field/method/parameter type
		Tag   *ast.BasicLit // field tag; or nil
	}

	Function struct {
		Node    *ast.FuncType
		Params  *Params
		Results *Results
	}

	Method struct {
		Clazz    *TypeSpec `json:"-"`
		Node     *ast.Field
		NodeDecl *ast.FuncDecl
		Name     string
		Function
	}

	Params struct {
		Method *Method `json:"-"`
		List   []Param
	}

	Param struct {
		Method     *Method `json:"-"`
		Name       string
		IsVariadic bool
		Typ        ast.Expr
	}

	Results struct {
		Method *Method `json:"-"`
		List   []Result
	}

	Result struct {
		Method *Method `json:"-"`
		Name   string
		Typ    ast.Expr
	}
)

func (sc *File) ImportPath(selectorExpr *ast.SelectorExpr) (string, error) {
	impName := ToString(selectorExpr.X)
	for _, imp := range sc.Imports {
		impPath := strings.Trim(ToString(imp.Path), "\"")
		if imp.Name != nil {
			if ToString(imp.Name) == impName {
				return impPath, nil
			}
		}

		ss := strings.Split(impPath, "/")
		if ss[len(ss)-1] == impName {
			return impPath, nil
		}
	}

	return "", errors.New("'" + impName + "' isnot found")
}

func (sc *File) PostionFor(pos token.Pos) token.Position {
	return sc.Ctx.FileSet.PositionFor(pos, true)
}

func (file *File) GetType(name string) *TypeSpec {
	for _, typ := range file.TypeList {
		if typ.Name == name {
			return typ
		}
	}
	return nil
}

// func (file *File) GetClass(name string) *TypeSpec {
// 	for idx := range file.TypeList {
// 		if file.TypeList[idx].Name == name {
// 			return file.TypeList[idx]
// 		}
// 	}
// 	return nil
// }

// func (method *Method) IsVariadic() bool {
// 	return method.Params.List[len(method.Params.List)-1].IsVariadic
// }

func (fn *Function) IsVariadic() bool {
	return fn.Params.List[len(fn.Params.List)-1].IsVariadic
}

func ToString(typ ast.Node) string {
	fset := token.NewFileSet()
	var buf strings.Builder
	if err := format.Node(&buf, fset, typ); err != nil {
		log.Fatalln(err)
	}
	return buf.String()
}

func (st *Struct) FieldByName(name string) *Field {
	for idx := range st.Fields {
		if st.Fields[idx].Name == name {
			return &st.Fields[idx]
		}
	}
	return nil
}

func (st *Struct) MethodByName(name string) *Method {
	for idx := range st.Methods {
		if st.Methods[idx].Name == name {
			return &st.Methods[idx]
		}
	}
	return nil
}

func (st *Interface) MethodByName(name string) *Method {
	for idx := range st.Methods {
		if st.Methods[idx].Name == name {
			return &st.Methods[idx]
		}
	}
	return nil
}

// var RangeDefineds = map[string]struct {
// 	Start ast.Expr
// 	End   ast.Expr
// }{}

// func AddRangeDefined(typ, start, end string) {
// 	var s ast.Expr = &ast.Ident{Name: strings.TrimPrefix(start, "*")}
// 	var e ast.Expr = &ast.Ident{Name: strings.TrimPrefix(end, "*")}

// 	if strings.HasPrefix(start, "*") {
// 		s = &ast.StarExpr{X: s}
// 	}

// 	if strings.HasPrefix(end, "*") {
// 		e = &ast.StarExpr{X: e}
// 	}

// 	RangeDefineds[typ] = struct {
// 		Start ast.Expr
// 		End   ast.Expr
// 	}{s, e}
// }

// func IsRangeStruct(classes []*TypeSpec, typ ast.Expr) (bool, ast.Expr, ast.Expr) {
// 	name := strings.TrimPrefix(ToString(typ), "*")

// 	if value, ok := RangeDefineds[name]; ok {
// 		return true, value.Start, value.End
// 	}

// 	var cls *Class
// 	for idx := range classes {
// 		if classes[idx].Name == name {
// 			cls = &classes[idx]
// 			break
// 		}
// 	}
// 	if cls == nil {
// 		return false, nil, nil
// 	}

// 	if len(cls.Fields) != 2 {
// 		return false, nil, nil
// 	}

// 	var startType, endType ast.Expr
// 	for _, field := range cls.Fields {
// 		if field.Name == "Start" {
// 			startType = field.Typ
// 		} else if field.Name == "End" {
// 			endType = field.Typ
// 		}
// 	}
// 	if startType == nil || endType == nil {
// 		return false, nil, nil
// 	}

// 	aType := strings.TrimPrefix(ToString(startType), "*")
// 	bType := strings.TrimPrefix(ToString(endType), "*")
// 	if aType != bType {
// 		return false, nil, nil
// 	}
// 	return true, startType, endType
// }

func IsContextType(typ ast.Expr) bool {
	return ToString(typ) == "context.Context"
}

func IsErrorType(typ ast.Expr) bool {
	return ToString(typ) == "error"
}

func IsBooleanType(typ ast.Expr) bool {
	return ToString(typ) == "bool"
}

func IsStringType(typ ast.Expr) bool {
	return isStringType(ToString(typ))
}

func isStringType(s string) bool {
	return s == "string"
}

func isNumericType(name string) bool {
	for _, t := range []string{
		"int8",
		"int16",
		"int32",
		"int64",
		"int",
		"uint8",
		"uint16",
		"uint32",
		"uint64",
		"uint",
		"float32",
		"float64",
		"float",
	} {
		if name == t {
			return true
		}
	}
	return false
}

func IsBasicType(typ ast.Expr) bool {
	if IsPtrType(typ) {
		return false
	}
	return isBasicType(ToString(typ))
}

func isBasicType(name string) bool {
	for _, t := range []string{
		"byte",
		"int8",
		"int16",
		"int32",
		"int64",
		"int",
		"uint8",
		"uint16",
		"uint32",
		"uint64",
		"uint",
		"float32",
		"float64",
		"float",
		"bool",
		"string",
	} {
		if name == t {
			return true
		}
	}
	return false
}
func IsPtrType(typ ast.Expr) bool {
	_, ok := typ.(*ast.StarExpr)
	return ok
}

func IsFuncType(typ ast.Node) bool {
	_, ok := typ.(*ast.FuncType)
	if ok {
		return ok
	}

	_, ok = typ.(*ast.FuncDecl)
	if ok {
		return ok
	}

	_, ok = typ.(*ast.FuncLit)
	if ok {
		return ok
	}

	return false
}

func ToFuncType(typ ast.Node) (*ast.FuncType, bool) {
	fn, ok := typ.(*ast.FuncType)
	if ok {
		return fn, true
	}

	fd, ok := typ.(*ast.FuncDecl)
	if ok {
		return fd.Type, true
	}

	fl, ok := typ.(*ast.FuncLit)
	if ok {
		return fl.Type, true
	}

	return nil, false
}

func ToFunction(fn *ast.FuncType) Function {
	method := Function{
		// File     *File  `json:"-"`
		// Clazz    *Class `json:"-"`
		Node: fn,
		// Name     string
		// Comments []string
		Params: &Params{
			// List: params,
		},
		Results: &Results{
			// List: results,
		},
	}
	// method.Params.Method = method
	// method.Results.Method = method

	if fn.Params != nil {
		for idx := range fn.Params.List {
			params := toParam(fn.Params.List[idx])
			method.Params.List = append(method.Params.List, params...)
		}
	}

	if fn.Results != nil {
		for idx := range fn.Results.List {
			results := toResult(fn.Results.List[idx])
			method.Results.List = append(method.Results.List, results...)
		}
	}
	return method
}

func toParam(fd *ast.Field) []Param {
	var list []Param

	typ := fd.Type

	ellipsis, isVariadic := fd.Type.(*ast.Ellipsis)
	if isVariadic {
		typ = ellipsis.Elt
	}

	if len(fd.Names) == 0 {
		list = append(list, Param{
			IsVariadic: isVariadic,
			Typ:        typ,
		})
		return list
	}

	for _, n := range fd.Names {
		list = append(list, Param{
			Name:       n.Name,
			IsVariadic: isVariadic,
			Typ:        typ,
		})
	}
	return list
}

func toResult(fd *ast.Field) []Result {
	var list []Result
	if len(fd.Names) == 0 {
		list = append(list, Result{
			Typ: fd.Type,
		})
		return list
	}

	for _, n := range fd.Names {
		list = append(list, Result{
			Name: n.Name,
			Typ:  fd.Type,
		})
	}
	return list
}

func toField(fd *ast.Field) []Field {
	var list []Field
	for _, n := range fd.Names {
		doc := fd.Comment
		if doc == nil {
			doc = fd.Doc
		} else if fd.Doc != nil {
			doc.List = append(doc.List, fd.Doc.List...)
		}
		list = append(list, Field{
			// File  *File  `json:"-"`
			// Clazz *Class `json:"-"`
			Node: fd,
			Name: n.Name,
			Typ:  fd.Type,
			Tag:  fd.Tag,
		})
	}
	return list
}

func ToStruct(st *ast.StructType) Struct {
	iface := Struct{
		Node: st,
	}

	for _, fd := range st.Fields.List {
		if len(fd.Names) == 0 {
			iface.Embedded = append(iface.Embedded, fd.Type)
			continue
		}

		fields := toField(fd)
		iface.Fields = append(iface.Fields, fields...)
	}

	return iface
}

func ToInterface(st *ast.InterfaceType) Interface {
	iface := Interface{
		Node: st,
	}
	for _, fd := range st.Methods.List {
		if len(fd.Names) == 0 {
			iface.Embedded = append(iface.Embedded, fd.Type)
			continue
		}

		iface.Methods = append(iface.Methods, toMethod(fd)...)
	}

	initMethods(iface.Methods)
	return iface
}

func initMethods(methods []Method) {
	for idx := range methods {
		methods[idx].Params.Method = &methods[idx]
		for j := range methods[idx].Params.List {
			methods[idx].Params.List[j].Method = &methods[idx]
		}

		methods[idx].Results.Method = &methods[idx]
		for j := range methods[idx].Results.List {
			methods[idx].Results.List[j].Method = &methods[idx]
		}
	}
}

func ToMethodDecl(node *ast.FuncDecl) Method {
	fn := ToFunction(node.Type)
	return Method{
		// Clazz    *TypeSpec `json:"-"`
		NodeDecl: node,
		Name:     node.Name.Name,
		Function: fn,
	}
}

func toMethod(node *ast.Field) []Method {
	fnT, ok := node.Type.(*ast.FuncType)
	if !ok {
		return nil
	}

	fn := ToFunction(fnT)

	var list []Method
	for idx := range node.Names {
		list = append(list, Method{
			// Clazz    *TypeSpec `json:"-"`
			Node:     node,
			Name:     node.Names[idx].Name,
			Function: fn,
		})
	}
	return list
}

func ToTypeSpec(node *ast.TypeSpec) *TypeSpec {
	switch v := node.Type.(type) {
	case *ast.StructType:
		st := ToStruct(v)
		ts := &TypeSpec{
			// File *File `json:"-"`
			Node:   node,
			Name:   node.Name.Name,
			Struct: &st,
		}
		// ts.Struct.Clazz = ts
		for idx := range ts.Struct.Fields {
			ts.Struct.Fields[idx].Clazz = ts
		}
		return ts
	case *ast.InterfaceType:
		itf := ToInterface(v)
		ts := &TypeSpec{
			// File *File `json:"-"`
			Node:      node,
			Name:      node.Name.Name,
			Interface: &itf,
		}
		// ts.Interface.Clazz = ts

		initMethods(ts.Interface.Methods)
		for idx := range ts.Interface.Methods {
			ts.Interface.Methods[idx].Clazz = ts
		}
		return ts
	default:
		return nil
	}
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

type parseVisitor struct {
	src *File
}

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

		funcDecl := ToMethodDecl(rn)

		var class *TypeSpec
		for idx := range v.src.TypeList {
			if name == v.src.TypeList[idx].Name {
				class = v.src.TypeList[idx]
				break
			}
		}

		if class != nil && class.Struct != nil {
			class.Struct.Methods = append(class.Struct.Methods, funcDecl)
			initMethods(class.Struct.Methods)
		} else {
			if v.src.Methods == nil {
				v.src.Methods = map[string][]Method{}
			}
			v.src.Methods[name] = append(v.src.Methods[name], funcDecl)
		}
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

type genDeclVisitor struct {
	src  *File
	node *ast.GenDecl
}

func (v *genDeclVisitor) Visit(n ast.Node) ast.Visitor {
	switch rn := n.(type) {
	case *ast.TypeSpec:
		// FIXME:
		if rn.Doc == nil {
			rn.Doc = v.node.Doc
		}
		ts := ToTypeSpec(rn)
		if ts != nil {
			ts.File = v.src
			v.src.TypeList = append(v.src.TypeList, ts)
			return nil
		}

		v.src.TypeList = append(v.src.TypeList, &TypeSpec{
			File: v.src,
			Node: rn,
			Name: rn.Name.Name,
		})
		return nil
	default:
		return v
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
