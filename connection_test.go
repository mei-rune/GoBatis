package gobatis_test

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"

	gobatis "github.com/runner-mei/GoBatis"
	"github.com/runner-mei/GoBatis/tests"
)

func TestSessionWith(t *testing.T) {
	sess := gobatis.SqlSessionFromContext(nil)
	if sess != nil {
		t.Error("sess != nil")
	}
	ctx := gobatis.WithSqlSession(nil, nil)

	sess = gobatis.SqlSessionFromContext(ctx)
	if sess != nil {
		t.Error("sess == nil")
	}

	sess = gobatis.SqlSessionFromContext(context.Background())
	if sess != nil {
		t.Error("sess != nil")
	}

	ctx = gobatis.WithSqlSession(nil, &gobatis.Reference{})
	sess = gobatis.SqlSessionFromContext(ctx)
	if sess == nil {
		t.Error("sess == nil")
	}

	fmt.Println(ctx)
}

func TestInit(t *testing.T) {
	callbacks := gobatis.ClearInit()
	defer gobatis.SetInit(callbacks)

	exceptederr := errors.New("init error")
	gobatis.Init(func(ctx *gobatis.InitContext) error {
		return exceptederr
	})

	_, err := gobatis.New(&gobatis.Config{DriverName: tests.TestDrv,
		DataSource: tests.TestConnURL,
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
	gobatis.StatementType(0).String()
	gobatis.StatementType(999999).String()

	gobatis.ResultType(0).String()
	gobatis.ResultType(999999).String()

	// bindNamedQuery(nil, nil, nil, nil)
	// bindNamedQuery(Params{{}}, nil, nil, nil)
	// bindNamedQuery(Params{{}}, nil, []interface{}{1, 2, 3}, nil)
}

func TestToDbType(t *testing.T) {
	for _, test := range []struct {
		name   string
		dbType gobatis.Dialect
	}{
		{"postgres", gobatis.Postgres},
		{"Postgres", gobatis.Postgres},
		{"mysql", gobatis.Mysql},
		{"Mysql", gobatis.Mysql},
		{"mssql", gobatis.MSSql},
		{"Mssql", gobatis.MSSql},
		{"sqlserver", gobatis.MSSql},
		{"Sqlserver", gobatis.MSSql},
		{"oracle", gobatis.Oracle},
		{"Oracle", gobatis.Oracle},
		{"ora", gobatis.Oracle},
		{"aara", gobatis.None},
	} {
		if test.dbType != gobatis.ToDbType(test.name) {
			t.Error(test.name, ", excepted ", test.dbType, "got", gobatis.ToDbType(test.name))
		}
	}
}

func TestConnection(t *testing.T) {
	tests.Run(t, func(_ testing.TB, factory *gobatis.SessionFactory) {
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
