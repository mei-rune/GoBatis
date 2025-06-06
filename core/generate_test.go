package core_test

import (
	"bufio"
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/aryann/difflib"
	"github.com/runner-mei/GoBatis/cmd/gobatis/generator"
)

func TestGenerate(t *testing.T) {
	wd := getGoBatis()

	// for _, cmd := range []string{
	// 	"github.com/runner-mei/GoBatis/cmd/gobatis",
	// 	"github.com/runner-mei/GoBatis/gentest",
	// 	"github.com/runner-mei/GoBatis/gentest/fail",
	// } {
	// 	var gen = generator.Generator{}
	// 	if err := gen.Run([]string{cmd}); err != nil {
	// 		t.Error(err)
	// 	}
	// }

	for _, name := range []string{
		"user",
		"role",
		"users",
		"interface",
		"upsert",
		"embedded",
		"external",
	} {
		t.Log("=====================", name)
		os.Remove(filepath.Join(wd, "gentest", name+".gobatis.go"))
		// fmt.Println(filepath.Join(wd, "gentest", name+".gobatis.go"))

		var gen = generator.Generator{}
		gen.SetDbCompatibility(false)
		if err := gen.Run([]string{filepath.Join(wd, "gentest", name+".go")}); err != nil {
			fmt.Println(err)
			t.Error(err)
			continue
		}

		actual := readFile(filepath.Join(wd, "gentest", name+".gobatis.go"), false)
		excepted := readFile(filepath.Join(wd, "gentest", name+".gobatis.txt"), false)
		if !reflect.DeepEqual(actual, excepted) {
			results := difflib.Diff(excepted, actual)
			for _, result := range results {
				if result.Delta == difflib.Common {
					continue
				}

				t.Error(result)
			}
		}
	}

	for _, name := range []string{"interface"} {
		t.Log("===================== fail/", name)
		os.Remove(filepath.Join(wd, "gentest", "fail", name+".gobatis.go"))
		// fmt.Println(filepath.Join(wd, "gentest", "fail", name+".gobatis.go"))

		var gen = generator.Generator{}
		gen.SetDbCompatibility(false)
		if err := gen.Run([]string{filepath.Join(wd, "gentest", "fail", name+".go")}); err != nil {
			fmt.Println(err)
			t.Error(err)
			continue
		}

		actual := readFile(filepath.Join(wd, "gentest", "fail", name+".gobatis.go"), true)
		excepted := readFile(filepath.Join(wd, "gentest", "fail", name+".gobatis.txt"), true)
		if !reflect.DeepEqual(actual, excepted) {
			results := difflib.Diff(excepted, actual)
			for _, result := range results {
				if result.Delta == difflib.Common {
					continue
				}
				t.Error(result)
			}
		}
		// os.Remove(filepath.Join(wd, "gentest", "fail", name+".gobatis.go"))
	}
}

func readFile(pa string, trimSpace bool) []string {
	bs, err := ioutil.ReadFile(pa)
	if err != nil {
		panic(err)
	}

	return splitLines(bs, trimSpace)
}

func splitLines(txt []byte, trimSpace bool) []string {
	//r := bufio.NewReader(strings.NewReader(s))
	s := bufio.NewScanner(bytes.NewReader(txt))
	var ss []string
	for s.Scan() {
		if trimSpace {
			ss = append(ss, strings.TrimSpace(s.Text()))
		} else {
			ss = append(ss, s.Text())
		}
	}
	return ss
}
