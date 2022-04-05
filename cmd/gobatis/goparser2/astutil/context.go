package astutil

import (
	"errors"
	"fmt"
	"go/ast"
	"go/token"
	"strings"
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
		ts := ctx.FindTypeInPackage(file, ident.Name)
		if ts != nil {
			return ts, nil
		}
	} else if selectorExpr, ok := typ.(*ast.SelectorExpr); ok {
		impPath, err := file.ImportPath(selectorExpr)
		if err != nil {
			return nil, errors.New("'" + ToString(typ) + "' is unknown type")
		}
		return ctx.FindType(impPath, selectorExpr.Sel.Name, true)
	}
	return nil, errors.New("'" + ToString(typ) + "' is unknown type")
}

func (ctx *Context) FindTypeInPackage(file *File, name string) *TypeSpec {
	ts := file.GetType(name)
	if ts != nil {
		return ts
	}

	for i := 0; i < file.Package.FileCount(); i++ {
		f, err := file.Package.GetFileByIndex(i)
		if err != nil {
			panic(err)
		}

		if f == file {
			continue
		}

		ts = f.GetType(name)
		if ts != nil {
			return ts
		}
	}
	return nil
}

func (ctx *Context) FindType(pkgPath, typeName string, autoLoad bool) (*TypeSpec, error) {
	var found = ctx.findPkgByImportPath(pkgPath)

	if found == nil && autoLoad {
		pkg, err := ctx.LoadPackage(pkgPath)
		if err != nil {
			return nil, err
		}

		found = pkg
	}

	if found != nil {
		for idx := 0; idx < found.FileCount(); idx++ {
			f, err := found.GetFileByIndex(idx)
			if err != nil {
				return nil, errors.New("try load " + pkgPath + "." + typeName + " fail, " + err.Error())
			}

			typ := f.GetType(typeName)
			if typ != nil {
				return typ, nil
			}
		}
		return nil, errors.New(pkgPath + "." + typeName + " is undefined in " + found.OSPath)
	}

	return nil, errors.New(pkgPath + "." + typeName + " is undefined")
}

func (ctx *Context) IsBasicType(file *File, n ast.Expr) bool {
	switch node := n.(type) {
	case *ast.Ident:
		ts := ctx.FindTypeInPackage(file, node.Name)
		if ts == nil {
			return isBasicType(node.Name)
		}
		if ts.Struct != nil && ts.Interface != nil {
			return false
		}

		// if ts.Node.Assign.IsValid() {
		// 	return ctx.IsBasicType(file, ts.Node.Type)
		// }
		return ctx.IsBasicType(file, ts.Node.Type)
		// return false
	case *ast.SelectorExpr:
		impPath, err := file.ImportPath(node)
		if err != nil {
			panic(err)
		}
		pkgType, err := ctx.FindType(impPath, node.Sel.Name, true)
		if err != nil {
			panic(err)
		}
		return ctx.IsBasicType(file, pkgType.Node.Type)
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
		panic(fmt.Sprintf("IsBasicType - %T %#v", n, n))
	}
}

func (ctx *Context) IsStringType(file *File, n ast.Expr) bool {
	if IsStringType(n) {
		return true
	}

	switch node := n.(type) {
	case *ast.Ident:
		ts := ctx.FindTypeInPackage(file, node.Name)
		if ts == nil {
			return isStringType(node.Name)
		}

		if ts.Struct != nil && ts.Interface != nil {
			return false
		}

		// if ts.Node.Assign.IsValid() {
		// 	return ctx.IsNumericType(file, ts.Node.Type)
		// }

		return ctx.IsStringType(file, ts.Node.Type)
	case *ast.SelectorExpr:
		impPath, err := file.ImportPath(node)
		if err != nil {
			panic(err)
		}

		pkgType, err := ctx.FindType(impPath, node.Sel.Name, true)
		if err != nil {
			panic(err)
		}
		return ctx.IsStringType(file, pkgType.Node.Type)
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
		panic(fmt.Sprintf("IsStringType - %T %#v", n, n))
	}
}

func (ctx *Context) IsNumericType(file *File, n ast.Expr) bool {
	switch node := n.(type) {
	case *ast.Ident:
		ts := ctx.FindTypeInPackage(file, node.Name)
		if ts == nil {
			return isNumericType(node.Name)
		}

		if ts.Struct != nil && ts.Interface != nil {
			return false
		}

		// if ts.Node.Assign.IsValid() {
		// 	return ctx.IsNumericType(file, ts.Node.Type)
		// }

		return ctx.IsNumericType(file, ts.Node.Type)
	case *ast.SelectorExpr:
		impPath, err := file.ImportPath(node)
		if err != nil {
			panic(err)
		}

		pkgType, err := ctx.FindType(impPath, node.Sel.Name, true)
		if err != nil {
			panic(err)
		}
		return ctx.IsNumericType(file, pkgType.Node.Type)
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
		panic(fmt.Sprintf("IsNumericType - %T %#v", n, n))
	}
}

func (ctx *Context) IsPtrType(file *File, typ ast.Expr) bool {
	return IsPtrType(typ)
}

func (ctx *Context) PtrElemType(file *File, typ ast.Expr) ast.Expr {
	return PtrElemType(typ)
}

func (ctx *Context) IsContextType(file *File, n ast.Expr) bool {
	return IsContextType(n)
}

func (ctx *Context) IsInterfaceType(file *File, n ast.Expr) bool {
	switch node := n.(type) {
	case *ast.Ident:
		ts := ctx.FindTypeInPackage(file, node.Name)
		if ts == nil {
			return false
		}

		if ts.Struct != nil {
			return false
		}
		if ts.Interface != nil {
			return true
		}

		if ts.Node.Assign.IsValid() {
			return ctx.IsInterfaceType(file, ts.Node.Type)
		}
		return false
	case *ast.SelectorExpr:
		impPath, err := file.ImportPath(node)
		if err != nil {
			panic(err)
		}

		ts, err := ctx.FindType(impPath, node.Sel.Name, true)
		if err != nil {
			panic(err)
		}
		if ts.Struct != nil {
			return false
		}
		if ts.Interface != nil {
			return true
		}

		return ctx.IsInterfaceType(file, ts.Node.Type)
	case *ast.StarExpr:
		return false
	case *ast.StructType:
		return false
	case *ast.InterfaceType:
		return true
	case *ast.MapType:
		return false
	case *ast.ArrayType:
		return false
	default:
		panic(fmt.Sprintf("IsInterfaceType - %T %#v", n, n))
	}
}

func (ctx *Context) IsStructType(file *File, typ ast.Expr) bool {
	if IsStructType(typ) {
		return true
	}

	if ident, ok := typ.(*ast.Ident); ok {
		ts := ctx.FindTypeInPackage(file, ident.Name)
		if ts != nil {
			if ts.Struct != nil {
				return true
			}
			if ts.Interface != nil {
				return false
			}
			if ts.Node.Assign.IsValid() {
				return ctx.IsStructType(file, ts.Node.Type)
			}
			return ctx.IsStructType(file, ts.Node.Type)
		}
	} else if selectorExpr, ok := typ.(*ast.SelectorExpr); ok {
		impPath, err := file.ImportPath(selectorExpr)
		if err != nil {
			panic(err)
		}
		ts, err := ctx.FindType(impPath, selectorExpr.Sel.Name, true)
		if err != nil {
			panic(err)
		}
		if ts == nil {
			panic(ToString(selectorExpr) + " isnot found")
		}
		if ts.Node.Assign.IsValid() {
			return ctx.IsStructType(file, ts.Node.Type)
		}
		return IsStructType(ts.Node.Type)
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
