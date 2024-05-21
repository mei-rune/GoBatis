package example

import (
	"context"
	"fmt"
	"testing"

	gobatis "github.com/runner-mei/GoBatis"
	"github.com/runner-mei/GoBatis/tests"
)

func TestUsers(t *testing.T) {
	insertUser := User{
		Username: "abc",
		Phone:    "123",
		Status:   1,
	}

	tests.Run(t, func(_ testing.TB, factory *gobatis.Session) {
		sqltext := GetTestSQL(factory.Dialect().Name())
		err := gobatis.ExecContext(context.Background(), factory.DB(), sqltext)
		if err != nil {
			t.Error(factory.Dialect().Name())
			if e, ok := err.(*gobatis.SqlError); ok {
				t.Error(e.SQL)
			}
			t.Error(err)
			return
		}

		ref := factory.SessionReference()
		users := NewUsers(ref)

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
		})

		t.Run("insertAndDelete", func(t *testing.T) {
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

		t.Run("insertAndDeleteAll", func(t *testing.T) {
			_, err := users.DeleteAll()
			if err != nil {
				t.Error(err)
				return
			}

			_, err = users.Insert(&insertUser)
			if err != nil {
				t.Error(err)
				return
			}
			count, err := users.Count()
			if err != nil {
				t.Error(err)
				return
			}

			if count != 1 {
				t.Error("excepted is", 1, ", actual is", count)
			}

			count, err = users.DeleteAll()
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

		t.Run("update", func(t *testing.T) {
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

			updateUser := insertUser
			updateUser.Username = insertUser.Username + "_udpated"
			updateUser.Phone = insertUser.Phone + "_udpated"
			updateUser.Status = insertUser.Status + 123

			_, err = users.Update(id, &updateUser)
			if err != nil {
				t.Error(err)
				return
			}

			u, err := users.Get(id)
			if err != nil {
				t.Error(err)
				return
			}
			if u.Username != updateUser.Username {
				t.Error("excepted is", updateUser.Username, ", actual is", u.Username)
			}

			if u.Phone != updateUser.Phone {
				t.Error("excepted is", updateUser.Phone, ", actual is", u.Phone)
			}

			if u.Status != updateUser.Status {
				t.Error("excepted is", updateUser.Status, ", actual is", u.Status)
			}
		})

		t.Run("selectMap", func(t *testing.T) {
			_, err := users.DeleteAll()
			if err != nil {
				t.Error(err)
				return
			}

			_, err = users.Insert(&insertUser)
			if err != nil {
				t.Error(err)
				return
			}

			listmap, err := users.ListForMapArray()
			if err != nil {
				t.Error(err)
				return
			}
			if len(listmap) != 1 {
				t.Error("excepted len is", 1, ", actual is", len(listmap))
				return
			}
			if fmt.Sprint(listmap[0]["username"]) != insertUser.Username {
				t.Error("excepted is", insertUser.Username, ", actual is", fmt.Sprint(listmap[0]["username"]))
			}

			if fmt.Sprint(listmap[0]["phone"]) != insertUser.Phone {
				t.Error("excepted is", insertUser.Phone, ", actual is", fmt.Sprint(listmap[0]["phone"]))
			}

			if fmt.Sprint(listmap[0]["status"]) != fmt.Sprint(insertUser.Status) {
				t.Error("excepted is", insertUser.Status, ", actual is", fmt.Sprint(listmap[0]["status"]))
			}
		})
	})
}
