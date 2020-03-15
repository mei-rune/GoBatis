package goparser2

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"strings"
	"testing"

	"github.com/aryann/difflib"
	gobatis "github.com/runner-mei/GoBatis"
	"github.com/runner-mei/GoBatis/goparser2/astutil"
)

const roleText = `
package role

// abc
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
	role "github.com/runner-mei/GoBatis/goparser/tmp/rr"
	g "github.com/runner-mei/GoBatis/goparser/tmp/group"
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

const srcBody = `// aaaaa
type UserDao interface {
	// insert ignore into users(` + "`username`" + `, phone, address, status, birth_day, created, updated)
	// values (?,?,?,?,?,CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
	Insert(u *User) (int64, error)

	// @type insert
	Add(u *User) (int64, error)

	Update(id int, u *User) error

	// @type update
	Edit(id int, u *User) error

	UpdateByID(id int, user map[string]interface{}) error

	// select id, username, phone, address, status, birth_day, created, updated
	// FROM users WHERE id=?
	Get(id uint64) (*User, error)

	Get1(id uint64) func(*User) error

	// select count(1)
	// from users
	Count() (int64, error)

	Ping()

	RemoveAll() (err error)

	List(offset int, size int) ([]*User, error)

	List1(offset int, size int) ([2]*User, error)

	List2(offset int, size int) ([]User, error)

	List3(offset int, size int) (users []User, err error)

	List4(offset int, size int) func(user User) error

	List5() (func(*User) (bool, error), io.Closer)

	ListAll() (map[int]*User, error)

	Roles(id int) ([]role.Role, error)

	InsertRoles1(r *role.Role) (int64, error)

	InsertRoles2(ctx context.Context, r *role.Role) (int64, error)

	QueryRoles3(ctx context.Context) func(t role.Role) error

	QueryRoles4(ctx context.Context) (map[int64]role.Role, error)

	InsertGroup(group *g.Group) (int64, error)

	Groups(id int) ([]g.Group, error)

	Groups1(idList ...int) ([]g.Group, error)

	GroupsWithID(id int) (map[int64]g.Group, error)

	Prefiles(id int) ([]Profile, error)

	// @type insert
	R1() error

	// @type update
	R2() error

	// @type delete
	R3() error

	// @type select
	R4() error

	// @type abc
	R5() error

	// @reference ProfileDao.Insert
	InsertProfile(name string, value string) (int64, error)

	// @reference ProfileDao.Remove
	RemoveProfile(name string) error
}`

const srcProfile = `package user

import "time"

type Profile struct {
	ID        uint64
	Name      string
	Value     string
	CreatedAt time.Time
	UpdatedAt time.Time
}

// abc
// aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa
type ProfileDao interface {
	Insert(name, value string) (int64, error)

	Remove(name string) error

	Users(name string) ([]User, error)
}
`

func getGoparsers() string {
	for _, pa := range filepath.SplitList(os.Getenv("GOPATH")) {
		dir := filepath.Join(pa, "src/github.com/runner-mei/GoBatis/goparser2")
		if st, err := os.Stat(dir); err == nil && st.IsDir() {
			return dir
		}
	}
	return ""
}

func TestParse(t *testing.T) {

	tmp := filepath.Join(getGoparsers(), "tmp")
	t.Log(tmp)
	// tmp := filepath.Join(getGoparsers(), "tmp")
	// if err := os.RemoveAll(tmp); err != nil && !os.IsNotExist(err) {
	// 	t.Error(err)
	// 	return
	// }
	// if err := os.MkdirAll(tmp, 0666); err != nil && !os.IsExist(err) {
	// 	t.Error(err)
	// 	return
	// }

	fileContents := [][2]string{
		{"rr/rr.go", roleText},
		{"group/group.go", groupText},
		{"user/user.go", srcHeader + srcBody},
		{"user/profile.go", srcProfile},
	}
	for _, pkg := range fileContents {
		pa := filepath.Join(tmp, pkg[0])
		if runtime.GOOS == "windows" {
			if err := os.RemoveAll(filepath.Dir(pa)); err != nil && !os.IsNotExist(err) {
				fmt.Println(err)
				t.Log(err)
			}
		}
		if err := os.MkdirAll(filepath.Dir(pa), 0666); err != nil {
			fmt.Println(err)
			t.Log(err)
		}
		// t.Log("mkdir", filepath.Dir(pa))
	}

	for _, pkg := range fileContents {
		pa := filepath.Join(tmp, pkg[0])
		if err := ioutil.WriteFile(pa, []byte(pkg[1]), 0400); err != nil {
			t.Error(err)
		}
	}

	f, err := Parse(nil, filepath.Join(tmp, "user/user.go"))
	if err != nil {
		t.Error(err)
		return
	}
	if len(f.Interfaces) == 0 {
		t.Error("interfaces is missing")
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

	for _, test := range []struct {
		name       string
		typ        gobatis.StatementType
		typeName   string
		goTypeName string
	}{
		{name: "Insert", typ: gobatis.StatementTypeInsert, typeName: "insert", goTypeName: "gobatis.StatementTypeInsert"},
		{name: "UpdateByID", typ: gobatis.StatementTypeUpdate, typeName: "update", goTypeName: "gobatis.StatementTypeUpdate"},
		{name: "RemoveAll", typ: gobatis.StatementTypeDelete, typeName: "delete", goTypeName: "gobatis.StatementTypeDelete"},
		{name: "Get", typ: gobatis.StatementTypeSelect, typeName: "select", goTypeName: "gobatis.StatementTypeSelect"},
		{name: "Ping", typ: gobatis.StatementTypeNone, typeName: "statementTypeUnknown-Ping", goTypeName: "gobatis.StatementTypeUnknown-Ping"},

		{name: "R1", typ: gobatis.StatementTypeInsert, typeName: "insert", goTypeName: "gobatis.StatementTypeInsert"},
		{name: "R2", typ: gobatis.StatementTypeUpdate, typeName: "update", goTypeName: "gobatis.StatementTypeUpdate"},
		{name: "R3", typ: gobatis.StatementTypeDelete, typeName: "delete", goTypeName: "gobatis.StatementTypeDelete"},
		{name: "R4", typ: gobatis.StatementTypeSelect, typeName: "select", goTypeName: "gobatis.StatementTypeSelect"},
		{name: "R5", typ: gobatis.StatementTypeNone, typeName: "statementTypeUnknown-abc", goTypeName: "gobatis.StatementTypeUnknown-abc"},
	} {
		method := f.Interfaces[0].MethodByName(test.name)
		if test.typ != method.StatementType() {
			t.Error(test.name, ": excepted ", test.typ, "got", method.StatementType())
		}
		if test.typeName != method.StatementTypeName() {
			t.Error(test.name, ": excepted ", test.typeName, "got", method.StatementTypeName())
		}
		if test.goTypeName != method.StatementGoTypeName() {
			t.Error(test.name, ": excepted ", test.goTypeName, "got", method.StatementGoTypeName())
		}
	}

	for _, test := range []struct {
		name     string
		typeName string
	}{
		{name: "Add", typeName: "User"},
		{name: "Insert", typeName: "User"},
		{name: "Get", typeName: "User"},
		{name: "RemoveAll", typeName: "User"},
		{name: "Update", typeName: "User"},
		{name: "Edit", typeName: "User"},
		{name: "Count", typeName: "User"},
		{name: "List1", typeName: "User"},
		{name: "List2", typeName: "User"},
		{name: "List3", typeName: "User"},
		{name: "List4", typeName: "User"},
		{name: "List5", typeName: "User"},

		{name: "ListAll", typeName: "User"},
		{name: "UpdateByID", typeName: "User"},
		{name: "Roles", typeName: ""},
		{name: "R5", typeName: ""},
		{name: "R1", typeName: "User"},

		{name: "InsertRoles1", typeName: "role.Role"},
		{name: "InsertRoles2", typeName: "role.Role"},
		{name: "InsertRoles3", typeName: "role.Role"},
		{name: "InsertRoles4", typeName: "role.Role"},
	} {
		method := f.Interfaces[0].MethodByName(test.name)
		typ := f.Interfaces[0].DetectRecordType(method)

		if typ == nil {
			if test.typeName != "" {
				t.Error(test.name, ": excepted ", test.typeName, "got nil")
			}
			continue
		}

		if test.typeName != astutil.TypePrint(typ) {
			t.Error(test.name, ": excepted ", test.typeName, "got", astutil.TypePrint(typ))
		}
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

	groupsWithID := f.Interfaces[0].MethodByName("GroupsWithID")
	signature = groupsWithID.MethodSignature(&PrintContext{File: f, Interface: f.Interfaces[0]})
	if excepted := "GroupsWithID(id int) (map[int64]g.Group, error)"; excepted != signature {
		t.Error("actual   is", signature)
		t.Error("excepted is", excepted)
	}

	typeName = groupsWithID.Results.List[0].TypeName()
	if excepted := "map[int64]Group"; typeName != excepted {
		t.Error("actual   is", typeName)
		t.Error("excepted is", excepted)
	}

	typeName = groupsWithID.Params.List[0].TypeName()
	if excepted := "int"; typeName != excepted {
		t.Error("actual   is", typeName)
		t.Error("excepted is", excepted)
	}

	updateByID := f.Interfaces[0].MethodByName("UpdateByID")
	typeName = updateByID.Params.List[1].TypeName()
	if excepted := "map[string]interface{}"; typeName != excepted {
		t.Error("actual   is", typeName)
		t.Error("excepted is", excepted)
	}

	a := f.Interfaces[0].ReferenceInterfaces()
	if !reflect.DeepEqual(a, []string{"ProfileDao"}) {
		t.Error(a)
	}

	// for test cover
	f.Interfaces[0].String()
	groupsWithID.String()
	groupsWithID.Params.Len()
	f.Interfaces[0].MethodByName("aaaabc")
	updateByID.Params.List[1].Print(nil)
	updateByID.Results.List[0].Print(nil)
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

func TestParseCommentsFail(t *testing.T) {
	_, err := parseComments([]string{
		"a",
		"@type",
		"@abc",
	})
	if err == nil {
		t.Error("excepted err got ok")
	} else {
		t.Log(err)
	}
}

func TestParseEmbedded(t *testing.T) {

	tmp := filepath.Join(getGoparsers(), "tmp")
	t.Log(tmp)
	// tmp := filepath.Join(getGoparsers(), "tmp")
	// if err := os.RemoveAll(tmp); err != nil && !os.IsNotExist(err) {
	// 	t.Error(err)
	// 	return
	// }
	// if err := os.MkdirAll(tmp, 0666); err != nil && !os.IsExist(err) {
	// 	t.Error(err)
	// 	return
	// }

	srcTest := `package user

type Test1 interface {
	Test1() error
}

type Test2 interface {
	Test2() error
}

type TestEmbedded interface {
	Test1
	
	Test2
	
	Test3() error
}`

	fileContents := [][2]string{
		{"test.go", srcTest},
	}
	for _, pkg := range fileContents {
		pa := filepath.Join(tmp, pkg[0])
		if runtime.GOOS == "windows" {
			if err := os.RemoveAll(filepath.Dir(pa)); err != nil && !os.IsNotExist(err) {
				fmt.Println(err)
				t.Log(err)
			}
		}
		if err := os.MkdirAll(filepath.Dir(pa), 0666); err != nil {
			fmt.Println(err)
			t.Log(err)
		}
		t.Log(pa)
	}

	for _, pkg := range fileContents {
		pa := filepath.Join(tmp, pkg[0])
		if err := ioutil.WriteFile(pa, []byte(pkg[1]), 0400); err != nil {
			t.Error(err)
		}
	}

	f, err := Parse(nil, filepath.Join(tmp, "test.go"))
	if err != nil {
		t.Error(err)
		return
	}
	if len(f.Interfaces) != 3 {
		t.Error("interfaces is missing -", len(f.Interfaces))
		return
	}

	embedded := f.Interfaces[2]
	if len(embedded.Methods) != 1 {
		t.Error("want 1 got", len(embedded.Methods))
		return
	}

	test3 := embedded.MethodByName("Test3")
	signature := test3.MethodSignature(&PrintContext{File: f, Interface: f.Interfaces[0]})
	if excepted := "Test3() error"; excepted != signature {
		t.Error("actual   is", signature)
		t.Error("excepted is", excepted)
	}

	if len(embedded.EmbeddedInterfaces) != 2 {
		t.Error("want 2 got", len(embedded.EmbeddedInterfaces))
		return
	}

	if excepted, actual := "Test1", embedded.EmbeddedInterfaces[0]; excepted != actual {
		t.Error("actual   is", actual)
		t.Error("excepted is", excepted)
	}

	if excepted, actual := "Test2", embedded.EmbeddedInterfaces[1]; excepted != actual {
		t.Error("actual   is", actual)
		t.Error("excepted is", excepted)
	}
}
