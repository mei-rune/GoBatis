package astutil

import (
	"os"
	"path/filepath"
	"strconv"
	"testing"
	"time"
)

func TestSearchDir(t *testing.T) {
	gopath := os.Getenv("GOPATH")
	go111mod := os.Getenv("GO111MODULE")

	reset := func() {
		os.Setenv("GOPATH", gopath)
		os.Setenv("GO111MODULE", go111mod)
	}

	defer reset()

	tmpdir := filepath.Join(os.TempDir(), "test_"+strconv.FormatInt(time.Now().Unix(), 10))

	defer func() {
		os.RemoveAll(tmpdir)
	}()

	mkdir := func(pa string) {
		e := os.MkdirAll(pa, 0777)
		if e == nil || os.IsExist(e) {
			return
		}
		t.Error(e)
		t.FailNow()
	}
	writefile := func(pa, text string) {
		e := os.MkdirAll(filepath.Dir(pa), 0777)
		if e != nil && !os.IsExist(e) {
			t.Error(e)
			t.FailNow()
			return
		}

		e = os.WriteFile(pa, []byte(text), 0666)
		if e != nil {
			t.Error(e)
			t.FailNow()
			return
		}
	}

	mkdir(filepath.Join(tmpdir, "src/a/b/c"))
	mkdir(filepath.Join(tmpdir, "src/a/b/d"))

	mkdir(filepath.Join(tmpdir, "testpkg/d"))
	mkdir(filepath.Join(tmpdir, "testpkg/c"))
	writefile(filepath.Join(tmpdir, "testpkg/go.mod"), "module aaa/bbb/ccc")

	for idx, test := range []struct {
		currentDir string
		pkg        string
		excepted   string
		set        func()
	}{
		{
			currentDir: filepath.Join(tmpdir, "src/a/b/c"),
			pkg:        "a/b/d",
			excepted:   filepath.Join(tmpdir, "src/a/b/d"),
			set: func() {
				os.Setenv("GOPATH", tmpdir)
			},
		},

		{
			currentDir: filepath.Join(tmpdir, "testpkg/d"),
			pkg:        "aaa/bbb/ccc/c",
			excepted:   filepath.Join(tmpdir, "testpkg/c"),
			set: func() {
				os.Setenv("GO111MODULE", "on")
			},
		},
	} {
		reset()

		if test.set != nil {
			test.set()
		}

		ctx := &Context{}
		dir, err := searchDir(ctx, test.currentDir, test.pkg)
		if err != nil {
			t.Error(idx, err)
			return
		}

		if dir != test.excepted {
			t.Error(idx, "want", test.excepted)
			t.Error(idx, "got ", dir)
		}
	}

	for idx, test := range []struct {
		currentDir string
		excepted   string
		set        func()
	}{
		{
			currentDir: filepath.Join(tmpdir, "src/a/b/c"),
			excepted:   "a/b/c",
			set: func() {
				os.Setenv("GOPATH", tmpdir)
				os.Setenv("GO111MODULE", "off")
			},
		},

		{
			currentDir: filepath.Join(tmpdir, "testpkg/d"),
			excepted:   "aaa/bbb/ccc/d",
			set: func() {
				os.Setenv("GO111MODULE", "on")
			},
		},
	} {
		reset()

		if test.set != nil {
			test.set()
		}

		pkgName, err := GetPkgPath(test.currentDir)
		if err != nil {
			t.Error(idx, err)
			return
		}

		if pkgName != test.excepted {
			t.Error(idx, "want", test.excepted)
			t.Error(idx, "got ", pkgName)
		}
	}
}
