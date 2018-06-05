package gobatis_test

import (
	"errors"
	"strings"
	"testing"
	"time"

	gobatis "github.com/runner-mei/GoBatis"
	"github.com/runner-mei/GoBatis/tests"
)

func TestInit(t *testing.T) {
	defer gobatis.ClearInit()
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

func TestToDbType(t *testing.T) {
	for _, test := range []struct {
		name   string
		dbType int
	}{
		{"postgres", gobatis.DbTypePostgres},
		{"Postgres", gobatis.DbTypePostgres},
		{"mysql", gobatis.DbTypeMysql},
		{"Mysql", gobatis.DbTypeMysql},
		{"mssql", gobatis.DbTypeMSSql},
		{"Mssql", gobatis.DbTypeMSSql},
		{"sqlserver", gobatis.DbTypeMSSql},
		{"Sqlserver", gobatis.DbTypeMSSql},
		{"oracle", gobatis.DbTypeOracle},
		{"Oracle", gobatis.DbTypeOracle},
		{"ora", gobatis.DbTypeOracle},
		{"aara", gobatis.DbTypeNone},
	} {
		if test.dbType != gobatis.ToDbType(test.name) {
			t.Error(test.name, ", excepted ", test.dbType, "got", gobatis.ToDbType(test.name))
		}
	}

	if "unknown" != gobatis.ToDbName(10000) {
		t.Error("a")
	}
}

func TestConnection(t *testing.T) {
	tests.Run(t, func(_ testing.TB, factory *gobatis.SessionFactory) {
		insertUser := tests.User{
			Name:        "张三",
			Nickname:    "haha",
			Password:    "password",
			Description: "地球人",
			Address:     "沪南路1155号",
			Sex:         "女",
			ContactInfo: `{"QQ":"8888888"}`,
			Birth:       time.Now(),
			CreateTime:  time.Now(),
		}

		// user := tests.User{
		// 	Name: "张三",
		// }

		t.Run("statementNotFound", func(t *testing.T) {
			_, err := factory.Insert("insertUserNotExists", insertUser)
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
			_, err := factory.Insert("selectUser", insertUser)
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

			_, err = factory.Update("selectUser", insertUser)
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

			_, err = factory.Delete("selectUser", insertUser)
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
			err = factory.Select("deleteUser", insertUser).Scan(&a)
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

			err = factory.SelectOne("deleteUser", insertUser).Scan(&a)
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
			err := factory.SelectOne("selectUserTplError").Scan(&u)
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

			err = factory.SelectOne("selectUserTplError", 1, 2, 3).Scan(&u)
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
			err := factory.SelectOne("selectUser", struct{ s string }{"abc"}).Scan(&u)
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
			err := factory.SelectOne("selectUserTplError", map[string]interface{}{"id": "abc"}).Scan(&u)
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
			err := factory.SelectOne("selectUserTpl3", map[string]interface{}{"id": "abc"}).Scan(&u)
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
