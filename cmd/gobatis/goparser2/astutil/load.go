package astutil

import (
	"errors"
	"go/ast"
	"io/ioutil"
	"log"
	"os"
	"path"
	"path/filepath"
	"runtime"
	"strings"

	"golang.org/x/mod/modfile"
)

func (pkg *Package) FileCount() int {
	return len(pkg.Files)
}

func (pkg *Package) GetFileByIndex(index int) (*File, error) {
	if pkg.Files[index] != nil {
		return pkg.Files[index], nil
	}

	filename := filepath.Join(pkg.OSPath, pkg.Filenames[index])
	f, err := ParseFile(pkg.Context, filename)
	if err != nil {
		return nil, err
	}
	f.Package = pkg
	pkg.Files[index] = f
	return f, nil
}

func (pkg *Package) LoadAll() ([]*File, error) {
	for idx := range pkg.Filenames {
		_, err := pkg.GetFileByIndex(idx)
		if err != nil {
			return nil, err
		}
	}
	return pkg.Files, nil
}

func (pkg *Package) getOrLoadFile(filename string) (*File, error) {
	basename := filepath.Base(filename)
	foundIdx := -1
	for idx := range pkg.Filenames {
		if pkg.Filenames[idx] == basename {
			if pkg.Files[idx] != nil {
				return pkg.Files[idx], nil
			}

			foundIdx = idx
			break
		}
	}

	if foundIdx < 0 {
		basename = strings.ToLower(basename)
		for idx := range pkg.Filenames {
			if strings.ToLower(pkg.Filenames[idx]) == basename {
				if pkg.Files[idx] != nil {
					return pkg.Files[idx], nil
				}

				foundIdx = idx
				break
			}
		}

		if foundIdx < 0 {
			return nil, errors.New("'" + filename + "' isnot found in the '" + pkg.OSPath + "'")
		}
	}

	f, err := ParseFile(pkg.Context, filename)
	if err != nil {
		return nil, err
	}

	f.Package = pkg
	pkg.Files[foundIdx] = f
	return f, nil
}

func (ctx *Context) getOrAddPackage(pkgPath, pkgdir string) (*Package, error) {
	pkg := ctx.findPkgByOSPath(pkgdir)
	if pkg != nil {
		return pkg, nil
	}

	fis, err := ioutil.ReadDir(pkgdir)
	if err != nil {
		return nil, err
	}

	var filenames []string
	for _, fi := range fis {
		if ext := filepath.Ext(fi.Name()); strings.ToLower(ext) != ".go" {
			continue
		}
		if strings.HasSuffix(fi.Name(), "_test.go") {
			continue
		}
		if strings.HasSuffix(fi.Name(), ".gobatis.go") {
			continue
		}
		filenames = append(filenames, fi.Name())
	}

	pkgdir = strings.TrimSuffix(pkgdir, "/")
	pkgdir = strings.TrimSuffix(pkgdir, "\\")

	log.Println("load package -", pkgdir)

	pkg = &Package{
		Context:    ctx,
		ImportPath: pkgPath,
		OSPath:     pkgdir,
		Filenames:  filenames,
		Files:      make([]*File, len(filenames)),
	}
	ctx.Packages = append(ctx.Packages, pkg)
	return pkg, nil
}

func (ctx *Context) LoadFile(filename string) (*File, error) {
	absfilename, err := filepath.Abs(filename)
	if err != nil {
		return nil, err
	}

	pkgName, err := GetPkgPath(absfilename)
	if err != nil {
		return nil, err
	}

	pkg, err := ctx.getOrAddPackage(pkgName, filepath.Dir(absfilename))
	if err != nil {
		return nil, err
	}

	return pkg.getOrLoadFile(absfilename)
}

func (ctx *Context) LoadPackage(pkgPath string) (*Package, error) {
	pkgdir := pkgPath
	if !dirExists(pkgdir) {
		currdir, err := os.Getwd()
		if err != nil {
			return nil, err
		}

		dir, err := searchDir(ctx, currdir, pkgPath)
		if err != nil {
			return nil, err
		}
		pkgdir = dir
	}

	var err error
	pkgdir, err = filepath.Abs(pkgdir)
	if err != nil {
		return nil, err
	}

	pkgName, err := GetPkgPath(pkgdir)
	if err != nil {
		return nil, err
	}

	pkg, err := ctx.getOrAddPackage(pkgName, pkgdir)
	if err != nil {
		return nil, err
	}

	_, err = pkg.LoadAll()
	if err != nil {
		return nil, err
	}
	return pkg, nil
}

func Load(ctx *Context, currentDir, pkgName string) (*File, error) {
	dir, err := searchDir(ctx, currentDir, pkgName)
	if err != nil {
		return nil, err
	}
	return ParseFile(ctx, dir)
}

func dirExists(name string) bool {
	st, err := os.Stat(name)
	if err != nil {
		return false
	}
	return st.IsDir()
}

// hasSubdir reports if dir is within root by performing lexical analysis only.
func hasSubdir(root, dir string) (rel string, ok bool) {
	const sep = string(filepath.Separator)
	root = filepath.Clean(root)
	if !strings.HasSuffix(root, sep) {
		root += sep
	}
	dir = filepath.Clean(dir)
	if !strings.HasPrefix(dir, root) {
		return "", false
	}
	return filepath.ToSlash(dir[len(root):]), true
}

func searchDir(ctx *Context, currentDir, pkgName string) (string, error) {
	searchVendor := func(root string, isGoroot bool) (bool, string) {
		sub, ok := hasSubdir(root, currentDir)
		if !ok || !strings.HasPrefix(sub, "src/") {
			return false, ""
		}

		for {
			vendor := filepath.Join(root, sub, "vendor")
			if dirExists(vendor) {
				dir := filepath.Join(vendor, pkgName)
				if dirExists(dir) {
					return true, dir
				}
			}
			i := strings.LastIndex(sub, "/")
			if i < 0 {
				break
			}
			sub = sub[:i]
		}
		return false, ""
	}

	goroot := os.Getenv("GOROOT")
	if goroot == "" {
		goroot = runtime.GOROOT()
	}
	if goroot != "" {
		pkgDir := filepath.Join(goroot, "src", pkgName)
		if dirExists(pkgDir) {
			return pkgDir, nil
		}
	}
	gopath := os.Getenv("GOPATH")
	if gopath != "" {
		for _, root := range filepath.SplitList(gopath) {
			if root == "" {
				continue
			}
			if strings.HasPrefix(root, "~") {
				continue
			}

			pkgDir := filepath.Join(root, "src", pkgName)
			if dirExists(pkgDir) {
				return pkgDir, nil
			}

			if ok, pkgDir := searchVendor(root, false); ok {
				return pkgDir, nil
			}
		}
	}

	// If modules are not enabled, then the in-process code works fine and we should keep using it.
	switch os.Getenv("GO111MODULE") {
	case "off":
		return "", errors.New("package '" + pkgName + "' dir not found")
	default: // "", "on", "auto", anything else
		// Maybe use modules.
	}

	// Look to see if there is a go.mod.
	// Since go1.13, it doesn't matter if we're inside GOPATH.
	parent := currentDir
	for {
		info, err := os.Stat(filepath.Join(parent, "go.mod"))
		if err == nil && !info.IsDir() {
			break
		}
		d := filepath.Dir(parent)
		if len(d) >= len(parent) {
			return "", nil
		}
		parent = d
	}

	vendor := filepath.Join(parent, "vendor")
	if dirExists(vendor) {
		dir := filepath.Join(vendor, pkgName)
		if dirExists(dir) {
			return dir, nil
		}
	}

	fileBytes, err := ioutil.ReadFile(filepath.Join(parent, "go.mod"))
	if err != nil {
		return "", err
	}

	packageName := modfile.ModulePath(fileBytes)
	if packageName != "" && strings.HasPrefix(pkgName, packageName) {
		return filepath.Join(parent, strings.TrimPrefix(pkgName, packageName)), nil
	}

	if dir := filepath.Join(os.Getenv("GOMODCACHE"), pkgName); dirExists(dir) {
		return dir, nil
	}

	return "", errors.New("'" + pkgName + "' isnot found")
}

func GetPkgPath(currentDir string) (string, error) {
	currentDir = filepath.Clean(currentDir)

	root, typ := GetSrcRootPath(currentDir)
	if root == "" {
		return "", errors.New("GO ROOT isnot path")
	}

	st, err := os.Stat(currentDir)
	if err != nil {
		return "", err
	}

	switch typ {
	case GOPATH:
		pa := strings.TrimPrefix(currentDir, root)
		if pa == "" {
			return pa, nil
		}

		if !st.IsDir() {
			pa = filepath.Dir(pa)
		}

		pa = filepath.ToSlash(pa)
		return strings.Trim(pa, "/"), nil
	case GOMOD:
		fileBytes, err := ioutil.ReadFile(filepath.Join(root, "go.mod"))
		if err != nil {
			return "", err
		}
		packageName := modfile.ModulePath(fileBytes)

		pa := strings.TrimPrefix(currentDir, root)
		if pa == "" {
			return packageName, nil
		}

		if !st.IsDir() {
			pa = filepath.Dir(pa)
		}

		pa = filepath.ToSlash(pa)
		pa = strings.Trim(pa, "/")
		return path.Join(packageName, pa), nil
	default:
		panic(typ)
	}
}

type RootPathType string

const (
	GOPATH RootPathType = "gopath"
	GOMOD  RootPathType = "gomod"
)

func GetSrcRootPath(currentDir string) (string, RootPathType) {
	// If modules are not enabled, then the in-process code works fine and we should keep using it.
	switch os.Getenv("GO111MODULE") {
	case "off":
		return GetSrcRootPathByGOPATH(currentDir), GOPATH
	default: // "", "on", "auto", anything else
		// Maybe use modules.
	}

	return GetSrcRootPathByGOMOD(currentDir), GOMOD
}

func GetSrcRootPathByGOPATH(currentDir string) string {
	currentDir = strings.Trim(currentDir, "/")
	currentDir = strings.Trim(currentDir, "\\")

	for currentDir != "" {
		name := filepath.Base(currentDir)
		if strings.ToLower(name) == "src" {
			return filepath.Clean(currentDir)
		}
		currentDir = filepath.Dir(currentDir)
	}
	return ""
}

func GetSrcRootPathByGOMOD(currentDir string) string {
	// Look to see if there is a go.mod.
	// Since go1.13, it doesn't matter if we're inside GOPATH.
	parent := currentDir
	for {
		info, err := os.Stat(filepath.Join(parent, "go.mod"))
		if err == nil && !info.IsDir() {
			return parent
		}
		d := filepath.Dir(parent)
		if len(d) >= len(parent) {
			return ""
		}
		parent = d
	}
}

func GetFieldByIndex(fieldList *ast.FieldList, idx int) *ast.Field {
	count := 0
	for _, field := range fieldList.List {
		for _, name := range field.Names {
			if count == idx {
				if len(field.Names) == 1 {
					return field
				}
				newField := &ast.Field{}
				*newField = *field
				newField.Names[0] = name
				newField.Names = newField.Names[:1]
				return newField
			}
			count++
		}
	}

	return fieldList.List[idx]
}
