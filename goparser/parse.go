package goparser

import (
	"bytes"
	"fmt"
	"go/ast"
	goimporter "go/importer"
	"go/parser"
	"go/token"
	"go/types"
	"log"
	"os/exec"
	"sort"
	"strconv"
	"strings"
)

type Interface struct {
	File *File `json:"-"`
	Pos  int
	Name string

	Methods []*Method
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

type File struct {
	Source     string
	Package    string
	Imports    map[string]string // database/sql => sql
	Interfaces []*Interface
}

func Parse(filename string) (*File, error) {
	err := goBuild(filename)
	if err != nil {
		return nil, err
	}

	return parse(nil, nil, filename, nil)
}

func parse(fset *token.FileSet, importer types.Importer, filename string, src interface{}) (*File, error) {
	if fset == nil {
		fset = token.NewFileSet()
	}
	if importer == nil {
		importer = goimporter.Default()
	}
	f, err := parser.ParseFile(fset, filename, src, parser.ParseComments)
	if err != nil {
		return nil, err
	}

	store := &File{
		Source:  filename,
		Package: f.Name.Name,
		Imports: map[string]string{},
	}
	for _, importSpec := range f.Imports {
		if importSpec.Name != nil {
			pa, err := strconv.Unquote(importSpec.Path.Value)
			if err != nil {
				store.Imports[importSpec.Path.Value] = importSpec.Name.Name
			} else {
				store.Imports[pa] = importSpec.Name.Name
			}
		}
	}
	ifList, err := parseTypes(store, f, fset, importer)
	if err != nil {
		return nil, err
	}

	store.Interfaces = ifList

	return store, nil
}

func goBuild(src string) error {
	cmd := exec.Command("go", "build", "-i", src)
	out, err := cmd.CombinedOutput()
	if bytes.HasSuffix(out, []byte("command-line-arguments\n")) {
		fmt.Printf("%s", out[:len(out)-23])
	} else {
		fmt.Printf("%s", out)
	}
	return err
}

func logWarn(pos token.Pos, name string, args ...interface{}) {
	//log.Println(pos, ":", name, "-", args)
}

func logWarnf(pos token.Pos, name string, fmtStr string, args ...interface{}) {
	//log.Println(pos, ":", name, "-", fmt.Sprintf(fmtStr, args...))
}

func logError(pos token.Pos, name string, args ...interface{}) {
	log.Println(pos, ":", name, "-", args)
}

func logErrorf(pos token.Pos, name string, fmtStr string, args ...interface{}) {
	log.Println(pos, ":", name, "-", fmt.Sprintf(fmtStr, args...))
}

func parseTypes(store *File, f *ast.File, fset *token.FileSet, importer types.Importer) ([]*Interface, error) {
	info := types.Info{Defs: make(map[*ast.Ident]types.Object)}
	conf := types.Config{Importer: importer}
	_, err := conf.Check(f.Name.Name, fset, []*ast.File{f}, &info)
	if err != nil {
		return nil, err
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

		itf := &Interface{
			File: store,
			Pos:  int(k.Pos()),
			Name: k.Name,
		}
		for i := 0; i < itfType.NumMethods(); i++ {
			x := itfType.Method(i)
			astMethod := findMethodByName(astInterfaceType, x.Name())

			doc := readMethodDoc(astMethod)
			m := NewMethod(itf, readMethodPos(astMethod), x.Name(), doc) //readDoc(k.Obj, x.Name()))
			y := x.Type().(*types.Signature)
			m.Params = NewParams(m, y.Params())
			m.Results = NewResults(m, y.Results())
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
	File   *File
	Indent string
}

func printType(ctx *PrintContext, sb *strings.Builder, typ types.Type) {
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
		printType(ctx, sb, t.Elem())
		return
	case *types.Slice:
		sb.WriteString("[]")
		printType(ctx, sb, t.Elem())
		return
	case *types.Map:
		sb.WriteString("map[")
		printType(ctx, sb, t.Key())
		sb.WriteString("]")
		printType(ctx, sb, t.Elem())
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
		sb.WriteString(typ.String())
		return
	}
	if isPointer {
		sb.WriteString("*")
	}

	if named.Obj().Pkg().Name() != ctx.File.Package {
		if a, ok := ctx.File.Imports[named.Obj().Pkg().Path()]; ok {
			sb.WriteString(a)
		} else {
			sb.WriteString(named.Obj().Pkg().Name())
		}
		sb.WriteString(".")
	}
	sb.WriteString(named.Obj().Name())
}
