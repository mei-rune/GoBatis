package astutil

import (
	"errors"
	"go/ast"
	"go/token"
	"path/filepath"
	"strings"
)

type Context struct {
	FileSet *token.FileSet

	Files []*File
}

func (ctx *Context) FindTypeInPkg(file *File, pkgName, typeName string) (*File, ast.Expr, error) {
	for _, imp := range file.Imports {
		if imp.Name.Name == pkgName {
			pkgName = strings.Trim(imp.Path.Value, "\"")
		}
	}

	file, err := Load(ctx, filepath.Dir(file.Filename), pkgName)
	if err != nil {
		return nil, nil, err
	}

	typ := file.GetType(typeName)
	if typ == nil {
		return nil, nil, errors.New(pkgName + "." + typeName + "is undefined")
	}
	return file, typ.Type, nil
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
		file, pkgType, err := ctx.FindTypeInPkg(file, TypePrint(selectorExpr.X), selectorExpr.Sel.Name)
		if err != nil {
			panic(err)
		}
		if pkgType == nil {
			panic(TypePrint(selectorExpr) + " isnot found")
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
