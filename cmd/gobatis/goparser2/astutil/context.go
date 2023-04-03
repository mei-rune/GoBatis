package astutil

import (
	"errors"
	"fmt"
	"go/ast"
	"go/token"
	"os"
	"path"
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

		basePkg := path.Base(pkg.ImportPath)
		if basePkg == pkgPath {
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

func (ctx *Context) ToTypeSpec(file *File, typ ast.Expr, recursive bool) (*TypeSpec, error) {
	if st, ok := typ.(*ast.StructType); ok {
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
		for idx := range ts.Struct.Embedded {
			ts.Struct.Embedded[idx].Clazz = ts
		}
		return ts, nil
	} else if ident, ok := typ.(*ast.Ident); ok {
		ts := ctx.FindTypeInPackage(file, ident.Name)
		if ts != nil {
			return toTypeSpec(ctx, file, ts, recursive)
		}
	} else if selectorExpr, ok := typ.(*ast.SelectorExpr); ok {
		ts, err := ctx.FindTypeBySelectorExpr(file, selectorExpr)
		if err != nil {
			return nil, err
		}
		return toTypeSpec(ctx, file, ts, recursive)
	}
	return nil, errors.New("'" + ToString(typ) + "' is unknown type")
}

func toTypeSpec(ctx *Context, file *File, ts *TypeSpec, recursive bool) (*TypeSpec, error) {
	if recursive {
		if ts.Node.Assign.IsValid() {
			return ctx.ToTypeSpec(file, ts.Node.Type, recursive)
		}
		_, ok := ts.Node.Type.(*ast.SelectorExpr)
		if ok {
			return ctx.ToTypeSpec(file, ts.Node.Type, recursive)
		}
		_, ok = ts.Node.Type.(*ast.Ident)
		if ok {
			if ts, err := ctx.ToTypeSpec(file, ts.Node.Type, recursive); err == nil {
				return ts, nil
			}
		}
	}
	return ts, nil
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

func (ctx *Context) FindTypeBySelectorExpr(file *File, selectorExpr *ast.SelectorExpr) (*TypeSpec, error) {
	impPath, err := file.ImportPath(selectorExpr)
	if err != nil {
		if !os.IsNotExist(err) {
			return nil, err
		}
		impPath = ToString(selectorExpr.X)
	}

	return ctx.FindType(impPath, selectorExpr.Sel.Name, true)
}

func (ctx *Context) FindType(pkgPath, typeName string, autoLoad bool) (*TypeSpec, error) {
	found := ctx.findPkgByImportPath(pkgPath)

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

func (ctx *Context) GetUnderlyingType(file *File, n ast.Expr) (*File, ast.Expr) {
	if n == nil {
		return nil, nil
	}
	switch node := n.(type) {
	case *ast.Ident:
		if isBasicType(node.Name) {
			return nil, nil
		}

		ts := ctx.FindTypeInPackage(file, node.Name)
		if ts == nil {
			return nil, nil
		}
		if ts.Struct != nil && ts.Interface != nil {
			return nil, nil
		}

		if _, ok := ts.Node.Type.(*ast.SelectorExpr); ok {
			return ctx.GetUnderlyingType(ts.File, ts.Node.Type)
		}

		return ts.File, ts.Node.Type
	case *ast.SelectorExpr:
		pkgType, err := ctx.FindTypeBySelectorExpr(file, node)
		if err != nil {
			panic(err)
		}
		if pkgType.Struct != nil && pkgType.Interface != nil {
			return nil, nil
		}
		return pkgType.File, pkgType.Node.Type
	case *ast.StarExpr:
		return nil, nil
	case *ast.StructType:
		return nil, nil
	case *ast.InterfaceType:
		return nil, nil
	case *ast.MapType:
		return nil, nil
	case *ast.ArrayType:
		return nil, nil
	default:
		panic(fmt.Sprintf("IsBasicType - %T %#v", n, n))
	}
}

func (ctx *Context) GetElemType(file *File, n ast.Expr, recursive bool) (*File, ast.Expr) {
	if n == nil {
		return nil, nil
	}
	if selectorExpr, ok := n.(*ast.SelectorExpr); ok {
		pkgType, err := ctx.FindTypeBySelectorExpr(file, selectorExpr)
		if err != nil {
			panic(err)
		}

		n = pkgType.Node.Type
	}

	elmType := ElemType(n)
	if elmType == nil {
		return file, n
	}
	if recursive {
		for {
			el := ElemType(elmType)
			if el == nil {
				break
			}
			elmType = el
		}
	}
	return file, elmType
}

func (ctx *Context) IsBasicType(file *File, n ast.Expr, checkUnderlying bool) bool {
	if n == nil {
		return false
	}

	switch node := n.(type) {
	case *ast.Ident:
		if isBasicType(node.Name) {
			return true
		}

		ts := ctx.FindTypeInPackage(file, node.Name)
		if ts == nil {
			return false
		}
		if ts.Struct != nil && ts.Interface != nil {
			return false
		}

		if ts.Node.Assign.IsValid() {
			return ctx.IsBasicType(file, ts.Node.Type, checkUnderlying)
		}
		if !checkUnderlying {
			return false
		}
		return ctx.IsBasicType(file, ts.Node.Type, checkUnderlying)
	case *ast.SelectorExpr:
		pkgType, err := ctx.FindTypeBySelectorExpr(file, node)
		if err != nil {
			panic(err)
		}
		return ctx.IsBasicType(file, pkgType.Node.Type, checkUnderlying)
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

func (ctx *Context) IsStringType(file *File, n ast.Expr, checkUnderlying bool) bool {
	if n == nil {
		return false
	}

	switch node := n.(type) {
	case *ast.Ident:
		if isStringType(node.Name) {
			return true
		}
		if isBasicType(node.Name) {
			return false
		}

		ts := ctx.FindTypeInPackage(file, node.Name)
		if ts == nil {
			return false
		}

		if ts.Struct != nil && ts.Interface != nil {
			return false
		}

		if ts.Node.Assign.IsValid() {
			return ctx.IsStringType(file, ts.Node.Type, checkUnderlying)
		}
		if !checkUnderlying {
			return false
		}
		return ctx.IsStringType(file, ts.Node.Type, checkUnderlying)
	case *ast.SelectorExpr:
		pkgType, err := ctx.FindTypeBySelectorExpr(file, node)
		if err != nil {
			panic(err)
		}
		return ctx.IsStringType(file, pkgType.Node.Type, checkUnderlying)
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

func (ctx *Context) IsNumericType(file *File, n ast.Expr, checkUnderlying bool) bool {
	if n == nil {
		return false
	}

	switch node := n.(type) {
	case *ast.Ident:
		if isNumericType(node.Name) {
			return true
		}
		if isBasicType(node.Name) {
			return false
		}

		ts := ctx.FindTypeInPackage(file, node.Name)
		if ts == nil {
			return isNumericType(node.Name)
		}

		if ts.Struct != nil && ts.Interface != nil {
			return false
		}

		if ts.Node.Assign.IsValid() {
			return ctx.IsNumericType(file, ts.Node.Type, checkUnderlying)
		}

		if !checkUnderlying {
			return false
		}
		return ctx.IsNumericType(file, ts.Node.Type, checkUnderlying)
	case *ast.SelectorExpr:
		pkgType, err := ctx.FindTypeBySelectorExpr(file, node)
		if err != nil {
			panic(err)
		}
		return ctx.IsNumericType(file, pkgType.Node.Type, checkUnderlying)
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
	if typ == nil {
		return false
	}
	return IsPtrType(typ)
}

func (ctx *Context) PtrElemType(file *File, typ ast.Expr) (*File, ast.Expr) {
	if typ == nil {
		return nil, nil
	}
	return file, PtrElemType(typ)
}

func (ctx *Context) IsContextType(file *File, expr ast.Expr) bool {
	if expr == nil {
		return false
	}
	return IsContextType(expr)
}

func (ctx *Context) IsInterfaceType(file *File, expr ast.Expr) bool {
	if expr == nil {
		return false
	}
	switch node := expr.(type) {
	case *ast.Ident:
		if isBasicType(node.Name) {
			return false
		}

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
		ts, err := ctx.FindTypeBySelectorExpr(file, node)
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
		panic(fmt.Sprintf("IsInterfaceType - %T %#v", expr, expr))
	}
}

func (ctx *Context) IsStructType(file *File, typ ast.Expr) bool {
	if typ == nil {
		return false
	}

	if IsStructType(typ) {
		return true
	}

	if ident, ok := typ.(*ast.Ident); ok {
		if isBasicType(ident.Name) {
			return false
		}

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
		ts, err := ctx.FindTypeBySelectorExpr(file, selectorExpr)
		if err != nil {
			panic(err)
		}

		if ts == nil {
			panic("import path '" + ToString(selectorExpr) + "' isnot found")
		}
		if ts.Node.Assign.IsValid() {
			return ctx.IsStructType(file, ts.Node.Type)
		}
		return IsStructType(ts.Node.Type)
	}
	return false
}

func IsIgnoreStructTypes(ctx *Context, file *File, typ ast.Expr, ignoreStructs []string) bool {
	if typ == nil {
		return false
	}
	typName := ToString(typ)
	for _, nm := range ignoreStructs {
		if nm == typName {
			return true
		}
	}

	if ctx.IsStructType(file, typ) {
		return false
	}
	if ctx.IsPtrType(file, typ) {
		file, typ = ctx.ElemType(file, typ)
		return IsIgnoreStructTypes(ctx, file, typ, ignoreStructs)
	}
	if ctx.IsMapType(file, typ) {
		file, typ = ctx.MapValueType(file, typ)
		return IsIgnoreStructTypes(ctx, file, typ, ignoreStructs)
	}
	if ctx.IsSliceOrArrayType(file, typ) {
		file, typ = ctx.ElemType(file, typ)
		return IsIgnoreStructTypes(ctx, file, typ, ignoreStructs)
	}
	return false
}

func (ctx *Context) IsArrayOrSliceType(file *File, typ ast.Expr) bool {
	if typ == nil {
		return false
	}
	return IsArrayOrSliceType(typ)
}

func (ctx *Context) IsSliceOrArrayType(file *File, typ ast.Expr) bool {
	if typ == nil {
		return false
	}
	return IsArrayOrSliceType(typ)
}

func (ctx *Context) IsSliceType(file *File, typ ast.Expr) bool {
	if typ == nil {
		return false
	}
	return IsSliceType(typ)
}

func (ctx *Context) IsArrayType(file *File, typ ast.Expr) bool {
	if typ == nil {
		return false
	}
	return IsArrayType(typ)
}

func (ctx *Context) IsEllipsisType(file *File, typ ast.Expr) bool {
	if typ == nil {
		return false
	}
	return IsEllipsisType(typ)
}

func (ctx *Context) IsSameType(file *File, a, b ast.Expr) bool {
	return IsSameType(a, b)
}

func (ctx *Context) IsErrorType(file *File, typ ast.Expr) bool {
	if typ == nil {
		return false
	}
	return IsErrorType(typ)
}

func (ctx *Context) IsMapType(file *File, typ ast.Expr) bool {
	if typ == nil {
		return false
	}
	return IsMapType(typ)
}

func (ctx *Context) MapValueType(file *File, typ ast.Expr) (*File, ast.Expr) {
	return file, MapValueType(typ)
}

func (ctx *Context) MapKeyType(file *File, typ ast.Expr) (*File, ast.Expr) {
	return file, MapKeyType(typ)
}

func (ctx *Context) ElemType(file *File, typ ast.Expr) (*File, ast.Expr) {
	return file, ElemType(typ)
}

func NewContext(fileSet *token.FileSet) *Context {
	if fileSet == nil {
		fileSet = token.NewFileSet()
	}
	return &Context{
		FileSet: fileSet,
	}
}
