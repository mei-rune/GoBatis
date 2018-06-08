package example

import (
	"testing"

	gobatis "github.com/runner-mei/GoBatis"
	"github.com/runner-mei/GoBatis/tests"
)

func TestUsers(t *testing.T) {
	insertUser := AuthUser{
		Username: "abc",
		Phone:    "123",
		Status:   1,
	}

	tests.Run(t, func(_ testing.TB, factory *gobatis.SessionFactory) {
		var err error
		switch factory.DbType() {
		case gobatis.DbTypePostgres:
			_, err = factory.DB().Exec(postgres)
		case gobatis.DbTypeMSSql:
			_, err = factory.DB().Exec(mssql)
		default:
			_, err = factory.DB().Exec(mysql)
		}
		if err != nil {
			t.Error(err)
			return
		}

		ref := factory.Reference()
		users := NewUsers(&ref)

		t.Run("insertAndGet", func(t *testing.T) {
			_, err := users.DeleteAll()
			if err != nil {
				t.Error(err)
				return
			}

			id, err := users.Insert(&insertUser)
			if err != nil {
				t.Error(err)
				return
			}

			u, err := users.Get(id)
			if err != nil {
				t.Error(err)
				return
			}

			if u.Username != insertUser.Username {
				t.Error("excepted is", u.Username, ", actual is", insertUser.Username)
			}

			if u.Phone != insertUser.Phone {
				t.Error("excepted is", u.Phone, ", actual is", insertUser.Phone)
			}

			if u.Status != insertUser.Status {
				t.Error("excepted is", u.Status, ", actual is", insertUser.Status)
			}

			count, err := users.Count()
			if err != nil {
				t.Error(err)
				return
			}

			if count != 1 {
				t.Error("excepted is", 1, ", actual is", count)
			}

			count, err = users.Delete(id)
			if err != nil {
				t.Error(err)
				return
			}

			if count != 1 {
				t.Error("excepted is", 1, ", actual is", count)
			}

			count, err = users.Count()
			if err != nil {
				t.Error(err)
				return
			}

			if count != 0 {
				t.Error("excepted is", 0, ", actual is", count)
			}
		})
	})
}
