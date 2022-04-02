package astutil

import (
	"errors"
	"go/ast"
	"go/token"
	"path/filepath"
)

type Package struct {
	Context *Context

	ImportPath string
	OSPath     string
	Files      []*File
}

type Context struct {
	FileSet *token.FileSet

	Packages []*Package
}

func (ctx *Context) addFile(pkgPath string, file *File) {
	for _, pkg := range ctx.Packages {
		if pkg.ImportPath == pkgPath {
			pkg.Files = append(pkg.Files, file)
			return
		}
	}

	ctx.Packages = append(ctx.Packages, &Package{
		Context: ctx,

		ImportPath: pkgPath,
		OSPath:     filepath.Dir(file.Filename),
		Files:      []*File{file},
	})
}

func (ctx *Context) findPkg(pkgPath string) *Package {
	for _, pkg := range ctx.Packages {
		if pkg.ImportPath == pkgPath {
			return pkg
		}
	}
	return nil
}

func (ctx *Context) FindType(pkgPath, typeName string, autoLoad bool) (*File, ast.Expr, error) {
	var found = ctx.findPkg(pkgPath)

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
				return file, typ.Type, nil
			}
		}
		return nil, nil, errors.New(pkgPath + "." + typeName + "is undefined in " + found.OSPath)
	}

	return nil, nil, errors.New(pkgPath + "." + typeName + "is undefined")
}

func (ctx *Context) IsPtrType(file *File, typ ast.Expr) bool {
	return IsPtrType(typ)
}

func (ctx *Context) PtrElemType(file *File, typ ast.Expr) ast.Expr {
	return PtrElemType(typ)
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
		file, pkgType, err := ctx.FindType(impPath, selectorExpr.Sel.Name, true)
		if err != nil {
			panic(err)
		}
		if pkgType == nil {
			panic(ToString(selectorExpr) + " isnot found")
		}
		return ctx.IsStructType(file, pkgType)
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

func NewContext(fileSet *token.FileSet) *Context {
	if fileSet == nil {
		fileSet = token.NewFileSet()
	}
	return &Context{
		FileSet: fileSet,
	}
}
