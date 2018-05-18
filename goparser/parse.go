package goparser

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/importer"
	"go/parser"
	"go/token"
	"go/types"
	"log"
	"os/exec"
)

type Interface struct {
	File *File  `json:"-"`
	Name string // IUser

	Methods []*Method
}

type File struct {
	Source     string
	Package    string
	Imports    map[string]string // database/sql => sql
	Interfaces []*Interface
}

func Parse(src string) (*File, error) {
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, src, nil, parser.ParseComments)
	if err != nil {
		return nil, err
	}

	store := &File{
		Source:  src,
		Package: f.Name.Name,
		Imports: map[string]string{},
	}

	err = goBuild(src)
	if err != nil {
		return nil, err
	}

	// err = extractInterfaceTypes(store, f)
	// if err != nil {
	// 	return nil, err
	// }

	ifList, err := parseTypes(store, f, fset)
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

func parseTypes(store *File, f *ast.File, fset *token.FileSet) ([]*Interface, error) {
	info := types.Info{Defs: make(map[*ast.Ident]types.Object)}
	conf := types.Config{Importer: importer.Default()}
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
			Name: k.Name,
		}
		for i := 0; i < itfType.NumMethods(); i++ {
			x := itfType.Method(i)

			doc := readMethodDocByName(astInterfaceType, x.Name())
			m := NewMethod(itf, x.Name(), doc) //readDoc(k.Obj, x.Name()))
			y := x.Type().(*types.Signature)
			m.Params = NewParams(m, y.Params())
			m.Results = NewResults(m, y.Results())
			itf.Methods = append(itf.Methods, m)
		}
		ifList = append(ifList, itf)
	}
	return ifList, nil
}

func readMethodDocByName(ift *ast.InterfaceType, name string) []string {
	if ift == nil {
		return nil
	}

	for _, field := range ift.Methods.List {
		if field.Names[0].Name == name {
			if field.Doc == nil {
				return nil
			}
			ss := make([]string, len(field.Doc.List))
			for idx := range field.Doc.List {
				ss = append(ss, field.Doc.List[idx].Text)
			}
			return ss
		}
	}
	return nil
}
