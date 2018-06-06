package goparser

import (
	"bufio"
	"go/ast"
	"go/importer"
	"go/parser"
	"go/token"
	"go/types"
	"reflect"
	"strings"
	"testing"

	"github.com/aryann/difflib"
)

const roleText = `
package role

type Role struct {
	ID        uint64     ` + "`json:\"id\"`" + `
	Name  string     ` + "`json:\"name\"`" + `
}
`

const groupText = `
package group

type Group struct {
	ID        uint64     ` + "`json:\"id\"`" + `
	Name  string     ` + "`json:\"name\"`" + `
}
`

const srcHeader = `
package user

import (
	"time"
	role "types/role"
	g "types/group"
)

type Status uint8

type User struct {
	ID        uint64     ` + "`json:\"id\"`" + `
	Username  string     ` + "`json:\"username\"`" + `
	Phone     string     ` + "`json:\"phone\"`" + `
	Address   *string    ` + "`json:\"address\"`" + `
	Status    Status     ` + "`json:\"status\"`" + `
	BirthDay  *time.Time ` + "`json:\"birth_day\"`" + `
	CreatedAt time.Time  ` + "`json:\"created_at\"`" + `
	UpdatedAt time.Time  ` + "`json:\"updated_at\"`" + `
}

`

const srcBody = `type UserDao interface {
	// insert ignore into users(` + "`username`" + `, phone, address, status, birth_day, created, updated)
	// values (?,?,?,?,?,CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
	Insert(u *User) (int64, error)

	// select id, username, phone, address, status, birth_day, created, updated
	// FROM users WHERE id=?
	Get(id uint64) (*User, error)

	// select count(1)
	// from users
	Count() (int64, error)

	List(offset int, size int) ([]*User, error)

	List2(offset int, size int) ([]User, error)

	List3(offset int, size int) (users []User, err error)

	ListAll() (map[int]*User, error)

	UpdateByID(id int, user map[string]interface{}) error

	Roles(id int) ([]role.Role, error)

	Groups(id int) ([]g.Group, error)

	GroupsWithID(id int) (map[int64]g.Group, error)
}`

type testImporter map[string]*types.Package

func (m testImporter) Import(path string) (*types.Package, error) {
	if pkg := m[path]; pkg != nil {
		return pkg, nil
	}
	return importer.Default().Import(path)
}
func TestParse(t *testing.T) {
	fset := token.NewFileSet()
	imports := make(testImporter)
	conf := types.Config{Importer: imports}

	makePkg := func(path, src string) error {
		roleF, err := parser.ParseFile(fset, "types/"+path+".go", src, parser.ParseComments)
		if err != nil {
			return err
		}
		pkg, err := conf.Check("types/"+path, fset, []*ast.File{roleF}, &types.Info{})
		if err != nil {
			return err
		}
		imports["types/"+path] = pkg
		return nil
	}

	for _, pkg := range [][2]string{
		{"role", roleText},
		{"group", groupText},
	} {
		if err := makePkg(pkg[0], pkg[1]); err != nil {
			t.Error(err)
			return
		}
	}

	userF, err := parser.ParseFile(fset, "user.go", srcHeader+srcBody, parser.ParseComments)
	if err != nil {
		t.Error(err)
		return
	}

	f, e := parse(fset, imports, []*ast.File{userF}, "user.go", userF)
	if e != nil {
		t.Error(err)
		return
	}

	ctx := &PrintContext{File: f}
	var sb strings.Builder
	f.Interfaces[0].Print(ctx, &sb)
	genText := sb.String()

	actual := splitLines(genText)
	excepted := splitLines(srcBody)

	if !reflect.DeepEqual(actual, excepted) {
		results := difflib.Diff(excepted, actual)
		for _, result := range results {
			t.Error(result)
		}

		t.Log(f.Imports)
	}

	list3 := f.Interfaces[0].MethodByName("List3")
	signature := list3.MethodSignature(&PrintContext{File: f, Interface: f.Interfaces[0]})
	if excepted := "List3(offset int, size int) (users []User, err error)"; excepted != signature {
		t.Error("actual   is", signature)
		t.Error("excepted is", excepted)
	}

	groups := f.Interfaces[0].MethodByName("Groups")
	signature = groups.MethodSignature(&PrintContext{File: f, Interface: f.Interfaces[0]})
	if excepted := "Groups(id int) ([]g.Group, error)"; excepted != signature {
		t.Error("actual   is", signature)
		t.Error("excepted is", excepted)
	}

	typeName := groups.Results.List[0].TypeName()
	if excepted := "Group"; typeName != excepted {
		t.Error("actual   is", typeName)
		t.Error("excepted is", excepted)
	}

	groups = f.Interfaces[0].MethodByName("GroupsWithID")
	signature = groups.MethodSignature(&PrintContext{File: f, Interface: f.Interfaces[0]})
	if excepted := "GroupsWithID(id int) (map[int64]g.Group, error)"; excepted != signature {
		t.Error("actual   is", signature)
		t.Error("excepted is", excepted)
	}

	typeName = groups.Results.List[0].TypeName()
	if excepted := "map[int64]Group"; typeName != excepted {
		t.Error("actual   is", typeName)
		t.Error("excepted is", excepted)
	}

}

func splitLines(txt string) []string {
	//r := bufio.NewReader(strings.NewReader(s))
	s := bufio.NewScanner(strings.NewReader(txt))
	var ss []string
	for s.Scan() {
		ss = append(ss, s.Text())
	}
	return ss
}
