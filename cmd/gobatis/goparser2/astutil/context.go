package astutil

import (
	"errors"
	"go/ast"
	"go/token"
	"strings"
	"fmt"
)

type Package struct {
	Context *Context

	ImportPath string
	OSPath     string
	Filenames  []string
	Files      []*File
}

type Context struct {
	FileSet *token.FileSet

	Packages []*Package
}

func (ctx *Context) findPkgByImportPath(pkgPath string) *Package {
	for _, pkg := range ctx.Packages {
		if pkg.ImportPath == pkgPath {
			return pkg
		}
	}
	return nil
}

func (ctx *Context) findPkgByOSPath(osPath string) *Package {
	osPath = strings.TrimSuffix(osPath, "/")
	osPath = strings.TrimSuffix(osPath, "\\")
	osPath = strings.ToLower(osPath)
	for _, pkg := range ctx.Packages {
		if strings.ToLower(pkg.OSPath) == osPath {
			return pkg
		}
	}
	return nil
}

func (ctx *Context) ToClass(file *File, typ ast.Expr) (*TypeSpec, error) {
	st, ok := typ.(*ast.StructType)
	if ok {
		t := ToStruct(st)
		ts := &TypeSpec{
			// File *File `json:"-"`
			// Node: node,
			Name:   "*** class ***",
			Struct: &t,
		}
		// ts.Struct.Clazz = ts
		for idx := range ts.Struct.Fields {
			ts.Struct.Fields[idx].Clazz = ts
		}
		return ts, nil
	}

	if ident, ok := typ.(*ast.Ident); ok {
		class := file.GetClass(ident.Name)
		if class != nil {
			return class, nil
		}
	} else if selectorExpr, ok := typ.(*ast.SelectorExpr); ok {
		impPath := file.ImportPath(selectorExpr)
		return ctx.FindClass(impPath, selectorExpr.Sel.Name, true)
	}
	return nil, errors.New("'" + ToString(typ) + "' is unknown type")
}

func (ctx *Context) FindClass(pkgPath, typeName string, autoLoad bool) (*TypeSpec, error) {
	var found = ctx.findPkgByImportPath(pkgPath)

	if found == nil && autoLoad {
		pkg, err := ctx.LoadPackage(pkgPath)
		if err != nil {
			return nil, err
		}

		found = pkg
	}

	if found != nil {
		for _, file := range found.Files {
			typ := file.GetClass(typeName)
			if typ != nil {
				return typ, nil
			}
		}

		for _, file := range found.Files {
			typ := file.GetType(typeName)
			if typ != nil {
				if typ.Assign.IsValid() {
					return ctx.ToClass(file, typ.Type)
				}
				return ctx.ToClass(file, typ.Type)
			}
		}
		return nil, errors.New(pkgPath + "." + typeName + "is undefined in " + found.OSPath)
	}

	return nil, errors.New(pkgPath + "." + typeName + "is undefined")
}

func (ctx *Context) findType(pkgPath, typeName string, autoLoad bool) (*File, *ast.TypeSpec, error) {
	var found = ctx.findPkgByImportPath(pkgPath)

	if found == nil && autoLoad {
		pkg, err := ctx.LoadPackage(pkgPath)
		if err != nil {
			return nil, nil, err
		}

		found = pkg
	}

	if found != nil {
		for _, file := range found.Files {
			typ := file.GetType(typeName)
			if typ != nil {
				return file, typ, nil
			}
		}
		return nil, nil, errors.New(pkgPath + "." + typeName + "is undefined in " + found.OSPath)
	}

	return nil, nil, errors.New(pkgPath + "." + typeName + "is undefined")
}

func (ctx *Context) IsBasicType(file *File, n ast.Expr) bool {
	switch node := n.(type) {
	case *ast.Ident:
		ts := file.GetType(node.Name)
		if ts == nil {
			return isBasicType(node.Name)
		}
		if cls := file.GetClass(node.Name); cls != nil {
			return false
		}

		if ts.Assign.IsValid() {
			return ctx.IsBasicType(file, ts.Type)
		}
		return false
	case *ast.SelectorExpr:
		impPath := file.ImportPath(node)
		file, pkgType, err := ctx.findType(impPath, node.Sel.Name, true)
		if err != nil {
			panic(err)
		}
		return ctx.IsBasicType(file, pkgType.Type)
	case *ast.StarExpr:
		return false
	case *ast.StructType:
		return false
	case *ast.InterfaceType:
		return false
	case *ast.MapType:
		return false
	case *ast.ArrayType:
		return false
	default:
		panic(fmt.Sprintf("%T %#v", n, n))
	}
}

func (ctx *Context) IsNumericType(file *File, n ast.Expr) bool {
	switch node := n.(type) {
	case *ast.Ident:
		ts := file.GetType(node.Name)
		if ts == nil {
			return isNumericType(node.Name)
		}
		if cls := file.GetClass(node.Name); cls != nil {
			return false
		}

		if ts.Assign.IsValid() {
			return ctx.IsNumericType(file, ts.Type)
		}
		return false
	case *ast.SelectorExpr:
		impPath := file.ImportPath(node)
		file, pkgType, err := ctx.findType(impPath, node.Sel.Name, true)
		if err != nil {
			panic(err)
		}
		return ctx.IsNumericType(file, pkgType.Type)
	case *ast.StarExpr:
		return false
	case *ast.StructType:
		return false
	case *ast.InterfaceType:
		return false
	case *ast.MapType:
		return false
	case *ast.ArrayType:
		return false
	default:
		panic(fmt.Sprintf("%T %#v", n, n))
	}
}

func (ctx *Context) IsPtrType(file *File, typ ast.Expr) bool {
	return IsPtrType(typ)
}

func (ctx *Context) PtrElemType(file *File, typ ast.Expr) ast.Expr {
	return PtrElemType(typ)
}

func (ctx *Context) IsInterfaceType(file *File, n ast.Expr) bool {
	switch node := n.(type) {
	case *ast.Ident:
		ts := file.GetType(node.Name)
		if ts == nil {
			return false
		}
		if cls := file.GetClass(node.Name); cls != nil {
			return cls.Interface != nil
		}

		if ts.Assign.IsValid() {
			return ctx.IsInterfaceType(file, ts.Type)
		}
		return false
	case *ast.SelectorExpr:
		impPath := file.ImportPath(node)
		file, pkgType, err := ctx.findType(impPath, node.Sel.Name, true)
		if err != nil {
			panic(err)
		}
		return ctx.IsInterfaceType(file, pkgType.Type)
	case *ast.StarExpr:
		return false
	default:
		panic(fmt.Sprintf("%T %#v", n, n))
	}
}

func (ctx *Context) IsStructType(file *File, typ ast.Expr) bool {
	if IsStructType(typ) {
		return true
	}

	if ident, ok := typ.(*ast.Ident); ok {
		class := file.GetClass(ident.Name)
		if class != nil {
			return true
		}
	} else if selectorExpr, ok := typ.(*ast.SelectorExpr); ok {
		impPath := file.ImportPath(selectorExpr)
		file, pkgType, err := ctx.findType(impPath, selectorExpr.Sel.Name, true)
		if err != nil {
			panic(err)
		}
		if pkgType == nil {
			panic(ToString(selectorExpr) + " isnot found")
		}
		if pkgType.Assign.IsValid() {
			return ctx.IsStructType(file, pkgType.Type)
		}
		return IsStructType(pkgType.Type)
	}
	return false
}

func IsIgnoreStructTypes(ctx *Context, file *File, typ ast.Expr, ignoreStructs []string) bool {
	if !ctx.IsStructType(file, typ) {
		return false
	}
	if ctx.IsPtrType(file, typ) {
		return IsIgnoreStructTypes(ctx, file, ctx.ElemType(file, typ), ignoreStructs)
	}
	if ctx.IsMapType(file, typ) {
		return IsIgnoreStructTypes(ctx, file, ctx.MapValueType(file, typ), ignoreStructs)
	}
	if ctx.IsSliceOrArrayType(file, typ) {
		return IsIgnoreStructTypes(ctx, file, ctx.ElemType(file, typ), ignoreStructs)
	}

	typName := ToString(typ)
	for _, nm := range ignoreStructs {
		if nm == typName {
			return true
		}
	}

	return false
}

func (ctx *Context) IsArrayOrSliceType(file *File, typ ast.Expr) bool {
	return IsArrayOrSliceType(typ)
}

func (ctx *Context) IsSliceOrArrayType(file *File, typ ast.Expr) bool {
	return IsArrayOrSliceType(typ)
}

func (ctx *Context) IsSliceType(file *File, typ ast.Expr) bool {
	return IsSliceType(typ)
}

func (ctx *Context) IsArrayType(file *File, typ ast.Expr) bool {
	return IsArrayType(typ)
}

func (ctx *Context) IsEllipsisType(file *File, typ ast.Expr) bool {
	return IsEllipsisType(typ)
}

func (ctx *Context) IsMapType(file *File, typ ast.Expr) bool {
	return IsMapType(typ)
}

func (ctx *Context) IsSameType(file *File, a, b ast.Expr) bool {
	return IsSameType(a, b)
}

func (ctx *Context) MapValueType(file *File, typ ast.Expr) ast.Expr {
	return MapValueType(typ)
}

func (ctx *Context) MapKeyType(file *File, typ ast.Expr) ast.Expr {
	return MapKeyType(typ)
}

func (ctx *Context) ElemType(file *File, typ ast.Expr) ast.Expr {
	return ElemType(typ)
}

func (ctx *Context) IsErrorType(file *File, typ ast.Expr) bool {
	return IsErrorType(typ)
}



func NewContext(fileSet *token.FileSet) *Context {
	if fileSet == nil {
		fileSet = token.NewFileSet()
	}
	return &Context{
		FileSet: fileSet,
	}
}
