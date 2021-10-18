package core_test

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/runner-mei/GoBatis/core"
	"github.com/runner-mei/GoBatis/tests"
	"github.com/runner-mei/GoBatis/dialects"
)


func getGoBatis() string {
	for _, pa := range filepath.SplitList(os.Getenv("GOPATH")) {
		dir := filepath.Join(pa, "src/github.com/runner-mei/GoBatis")
		if st, err := os.Stat(dir); err == nil && st.IsDir() {
			return dir
		}
	}

	wd, err := os.Getwd()
	if err == nil {
		for {
			if st, err := os.Stat(filepath.Join(wd, "go.mod")); err == nil && st.IsDir() {
				return wd
			}
			s := filepath.Dir(wd);
			if  len(s) >= len(wd) {
				return wd
			}
			wd = w
		}
	}
	return ""
}

func TestSessionWith(t *testing.T) {
	sess := core.SqlSessionFromContext(nil)
	if sess != nil {
		t.Error("sess != nil")
	}
	ctx := core.WithSqlSession(nil, nil)

	sess = core.SqlSessionFromContext(ctx)
	if sess != nil {
		t.Error("sess == nil")
	}

	sess = core.SqlSessionFromContext(context.Background())
	if sess != nil {
		t.Error("sess != nil")
	}

	ctx = core.WithSqlSession(nil, &core.Reference{})
	sess = core.SqlSessionFromContext(ctx)
	if sess == nil {
		t.Error("sess == nil")
	}

	fmt.Println(ctx)
}

func TestInit(t *testing.T) {
	callbacks := core.ClearInit()
	defer core.SetInit(callbacks)

	exceptederr := errors.New("init error")
	core.Init(func(ctx *core.InitContext) error {
		return exceptederr
	})

	_, err := core.New(&core.Config{DriverName: tests.TestDrv,
		DataSource: tests.GetTestConnURL(),
		XMLPaths: []string{"example/test.xml",
			"../example/test.xml",
			"../../example/test.xml"}})
	if err == nil {
		t.Error("excepted error got ok")
		return
	}

	if !strings.Contains(err.Error(), exceptederr.Error()) {
		t.Error("excepted contains", exceptederr.Error())
		t.Error("got", err.Error())
		return
	}
}

func TestForTestcover(t *testing.T) {
	core.StatementType(0).String()
	core.StatementType(999999).String()

	core.ResultType(0).String()
	core.ResultType(999999).String()

	// bindNamedQuery(nil, nil, nil, nil)
	// bindNamedQuery(Params{{}}, nil, nil, nil)
	// bindNamedQuery(Params{{}}, nil, []interface{}{1, 2, 3}, nil)
}

func TestToDbType(t *testing.T) {
	for _, test := range []struct {
		name   string
		dbType core.Dialect
	}{
		{"postgres", dialects.Postgres},
		{"Postgres", dialects.Postgres},
		{"mysql", dialects.Mysql},
		{"Mysql", dialects.Mysql},
		{"mssql", dialects.MSSql},
		{"Mssql", dialects.MSSql},
		{"sqlserver", dialects.MSSql},
		{"Sqlserver", dialects.MSSql},
		{"oracle", dialects.Oracle},
		{"Oracle", dialects.Oracle},
		{"ora", dialects.Oracle},
		{"aara", dialects.None},
	} {
		if test.dbType != dialects.New(test.name) {
			t.Error(test.name, ", excepted ", test.dbType, "got", dialects.New(test.name))
		}
	}
}

func TestConnection(t *testing.T) {
	tests.Run(t, func(_ testing.TB, factory *core.SessionFactory) {
		sqlStatements := factory.SqlStatements()
		keyLen := 0
		for _, stmt := range sqlStatements {
			if len(stmt[0]) > keyLen {
				keyLen = len(stmt[0])
			}
		}

		fmt.Println()
		fmt.Println(strings.Repeat("=", 2*keyLen))
		for idx := range sqlStatements {
			id, rawSQL := sqlStatements[idx][0], sqlStatements[idx][1]
			fmt.Println(id+strings.Repeat(" ", keyLen-len(id)), ":", rawSQL)
		}
		fmt.Println()
		fmt.Println()

		insertUser := tests.User{
			Name:        "张三",
			Nickname:    "haha",
			Password:    "password",
			Description: "地球人",
			Address:     "沪南路1155号",
			Sex:         "女",
			ContactInfo: map[string]interface{}{"QQ": "8888888"},
			Birth:       time.Now(),
			CreateTime:  time.Now(),
		}

		// user := tests.User{
		// 	Name: "张三",
		// }

		t.Run("statementNotFound", func(t *testing.T) {
			_, err := factory.Insert(context.Background(), "insertUserNotExists", insertUser)
			if err == nil {
				t.Error("excepted error but got ok")
				return
			}
			if !strings.Contains(err.Error(), "insertUserNotExists") {
				t.Error("excepted is insertUserNotExists")
				t.Error("actual   is", err)
			}
		})

		t.Run("statementTypeError", func(t *testing.T) {
			_, err := factory.Insert(context.Background(), "selectUser", insertUser)
			if err == nil {
				t.Error("excepted error but got ok")
				return
			}
			if !strings.Contains(err.Error(), "selectUser") {
				t.Error("excepted is selectUser")
				t.Error("actual   is", err)
			}

			if !strings.Contains(err.Error(), "type Error") {
				t.Error("excepted is type Error")
				t.Error("actual   is", err)
			}

			_, err = factory.Update(context.Background(), "selectUser", insertUser)
			if err == nil {
				t.Error("excepted error but got ok")
				return
			}
			if !strings.Contains(err.Error(), "selectUser") {
				t.Error("excepted is selectUser")
				t.Error("actual   is", err)
			}

			if !strings.Contains(err.Error(), "type Error") {
				t.Error("excepted is type Error")
				t.Error("actual   is", err)
			}

			_, err = factory.Delete(context.Background(), "selectUser", insertUser)
			if err == nil {
				t.Error("excepted error but got ok")
				return
			}
			if !strings.Contains(err.Error(), "selectUser") {
				t.Error("excepted is selectUser")
				t.Error("actual   is", err)
			}

			if !strings.Contains(err.Error(), "type Error") {
				t.Error("excepted is type Error")
				t.Error("actual   is", err)
			}

			var a int
			err = factory.Select(context.Background(), "deleteUser", insertUser).Scan(&a)
			if err == nil {
				t.Error("excepted error but got ok")
				return
			}
			if !strings.Contains(err.Error(), "deleteUser") {
				t.Error("excepted is deleteUser")
				t.Error("actual   is", err)
			}

			if !strings.Contains(err.Error(), "type Error") {
				t.Error("excepted is type Error")
				t.Error("actual   is", err)
			}

			err = factory.SelectOne(context.Background(), "deleteUser", insertUser).Scan(&a)
			if err == nil {
				t.Error("excepted error but got ok")
				return
			}
			if !strings.Contains(err.Error(), "deleteUser") {
				t.Error("excepted is deleteUser")
				t.Error("actual   is", err)
			}

			if !strings.Contains(err.Error(), "type Error") {
				t.Error("excepted is type Error")
				t.Error("actual   is", err)
			}
		})

		t.Run("argumentError", func(t *testing.T) {
			var u int
			err := factory.SelectOne(context.Background(), "selectUserTplError").Scan(&u)
			if err == nil {
				t.Error("excepted error but got ok")
				return
			}
			if !strings.Contains(err.Error(), "selectUserTplError") {
				t.Error("excepted is selectUserTplError")
				t.Error("actual   is", err)
			}

			if !strings.Contains(err.Error(), "arguments is missing") {
				t.Error("excepted is arguments is missing")
				t.Error("actual   is", err)
			}

			err = factory.SelectOne(context.Background(), "selectUserTplError", 1, 2, 3).Scan(&u)
			if err == nil {
				t.Error("excepted error but got ok")
				return
			}
			if !strings.Contains(err.Error(), "selectUserTplError") {
				t.Error("excepted is selectUserTplError")
				t.Error("actual   is", err)
			}

			if !strings.Contains(err.Error(), "arguments is exceed 1") {
				t.Error("excepted is arguments is exceed 1")
				t.Error("actual   is", err)
			}
		})

		t.Run("bindError", func(t *testing.T) {
			var u int
			err := factory.SelectOne(context.Background(), "selectUser", struct{ s string }{"abc"}).Scan(&u)
			if err == nil {
				t.Error("excepted error but got ok")
				return
			}
			if !strings.Contains(err.Error(), "selectUser") {
				t.Error("excepted is selectUser")
				t.Error("actual   is", err)
			}
		})

		t.Run("compileSQLFailAfterTplOk", func(t *testing.T) {
			var u int
			err := factory.SelectOne(context.Background(), "selectUserTplError", map[string]interface{}{"id": "abc"}).Scan(&u)
			if err == nil {
				t.Error("excepted error but got ok")
				return
			}
			if !strings.Contains(err.Error(), "selectUserTplError") {
				t.Error("excepted is selectUserTplError")
				t.Error("actual   is", err)
			}
		})

		t.Run("bindErrorAfterTplOk", func(t *testing.T) {
			var u int
			err := factory.SelectOne(context.Background(), "selectUserTpl3", map[string]interface{}{"id": "abc"}).Scan(&u)
			if err == nil {
				t.Error("excepted error but got ok")
				return
			}
			if !strings.Contains(err.Error(), "selectUserTpl3") {
				t.Error("excepted is selectUserTpl3")
				t.Error("actual   is", err)
			}
			if !strings.Contains(err.Error(), "'name'") {
				t.Error("excepted is 'name'")
				t.Error("actual   is", err)
			}
		})
	})
}
