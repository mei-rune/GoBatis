package goparser

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
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

	Ping()

	RemoveAll() (err error)

	List(offset int, size int) ([]*User, error)

	List2(offset int, size int) ([]User, error)

	List3(offset int, size int) (users []User, err error)

	ListAll() (map[int]*User, error)

	UpdateByID(id int, user map[string]interface{}) error

	Roles(id int) ([]role.Role, error)

	Groups(id int) ([]g.Group, error)

	GroupsWithID(id int) (map[int64]g.Group, error)

	Prefiles(id int) ([]Profile, error)
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

type ProfileDao interface {
	Insert(name, value string) (int64, error)

	Remove(name string) error

	Users(name string) ([]User, error)
}
`

func getGoparsers() string {
	for _, pa := range filepath.SplitList(os.Getenv("GOPATH")) {
		dir := filepath.Join(pa, "src/github.com/runner-mei/GoBatis/goparser")
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
		if err := os.RemoveAll(filepath.Dir(pa)); err != nil && !os.IsNotExist(err) {
			fmt.Println(err)
			t.Log(err)
		}
		if err := os.MkdirAll(filepath.Dir(pa), 0666); err != nil {
			fmt.Println(err)
			t.Log(err)
		}
		t.Log("mkdir", filepath.Dir(pa))
	}

	for _, pkg := range fileContents {
		pa := filepath.Join(tmp, pkg[0])
		// f, err := os.Create(pa)
		// if err != nil {
		// 	t.Error(err)
		// 	return
		// }
		// _, err = f.WriteString(pkg[1])
		// if err != nil {
		// 	t.Error(err)
		// 	return
		// }
		// if err = f.Close(); err != nil {
		// 	t.Error(err)
		// }

		if err := ioutil.WriteFile(pa, []byte(pkg[1]), 0400); err != nil {
			t.Error(err)
		}
	}

	f, err := Parse(filepath.Join(tmp, "user/user.go"))
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

	f.Interfaces[0].String()
	groupsWithID.String()
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
