package example

import (
	"fmt"
	"testing"

	gobatis "github.com/runner-mei/GoBatis"
	"github.com/runner-mei/GoBatis/tests"
)

const (
	postgres = `
DROP TABLE IF EXISTS auth_users;

CREATE TABLE IF NOT EXISTS auth_users (
  id bigserial PRIMARY KEY,
  username VARCHAR(32) NOT NULL UNIQUE,
  phone VARCHAR(32),
  address VARCHAR(256),
  status INT,
  birth_day DATE,
  created_at TIMESTAMP default NOW(),
  updated_at TIMESTAMP default NOW()
)`

	mssql = `IF OBJECT_ID('dbo.auth_users', 'U') IS NOT NULL
DROP TABLE auth_users;

CREATE TABLE auth_users (
  id int IDENTITY PRIMARY KEY,
  username VARCHAR(32) NOT NULL UNIQUE,
  phone VARCHAR(32),
  address VARCHAR(256),
  status TINYINT,
  birth_day DATE,
  created_at datetimeoffset default CURRENT_TIMESTAMP,
  updated_at datetimeoffset default CURRENT_TIMESTAMP
)`

	mysql = `
DROP TABLE IF EXISTS auth_users;

CREATE TABLE IF NOT EXISTS auth_users (
  id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
  username VARCHAR(32) NOT NULL UNIQUE,
  phone VARCHAR(32),
  address VARCHAR(256),
  status TINYINT UNSIGNED,
  birth_day DATE,
  created_at TIMESTAMP default CURRENT_TIMESTAMP,
  updated_at TIMESTAMP default CURRENT_TIMESTAMP
) ENGINE=InnoDB DEFAULT CHARSET=utf8`
)

func toString(v interface{}) string {
	if bs, ok := v.([]byte); ok {
		return string(bs)
	}
	return fmt.Sprint(v)
}

func TestConnection(t *testing.T) {
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

		conn := NewConnection(factory)

		t.Run("insertAndGet", func(t *testing.T) {
			_, err := conn.Users().DeleteAll()
			if err != nil {
				t.Error(err)
				return
			}

			id, err := conn.Users().Insert(&insertUser)
			if err != nil {
				t.Error(err)
				return
			}

			u, err := conn.Users().Get(id)
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

			umap, err := conn.Users().GetMap(id)
			if err != nil {
				t.Error(err)
				return
			}

			if toString(umap["username"]) != insertUser.Username {
				t.Error("excepted is", insertUser.Username, ", actual is", umap["username"])
			}

			if toString(umap["phone"]) != insertUser.Phone {
				t.Error("excepted is", insertUser.Phone, ", actual is", umap["phone"])
			}

			if toString(umap["status"]) != fmt.Sprint(insertUser.Status) {
				t.Error("excepted is", insertUser.Status, ", actual is", umap["status"])
			}

			name, err := conn.Users().GetNameByID(id)
			if err != nil {
				t.Error(err)
				return
			}

			if name != insertUser.Username {
				t.Error("excepted is", name, ", actual is", insertUser.Username)
			}

			count, err := conn.Users().Count()
			if err != nil {
				t.Error(err)
				return
			}

			if count != 1 {
				t.Error("excepted is", 1, ", actual is", count)
			}
		})

		t.Run("delete", func(t *testing.T) {
			_, err := conn.Users().DeleteAll()
			if err != nil {
				t.Error(err)
				return
			}

			id, err := conn.Users().Insert(&insertUser)
			if err != nil {
				t.Error(err)
				return
			}

			count, err := conn.Users().Delete(id)
			if err != nil {
				t.Error(err)
				return
			}

			if count != 1 {
				t.Error("excepted is", 1, ", actual is", count)
			}

			count, err = conn.Users().Count()
			if err != nil {
				t.Error(err)
				return
			}

			if count != 0 {
				t.Error("excepted is", 0, ", actual is", count)
			}
		})

		t.Run("update", func(t *testing.T) {
			_, err := conn.Users().DeleteAll()
			if err != nil {
				t.Error(err)
				return
			}

			id, err := conn.Users().Insert(&insertUser)
			if err != nil {
				t.Error(err)
				return
			}

			count, err := conn.Users().UpdateName(id, "newusername")
			if err != nil {
				t.Error(err)
				return
			}

			if count != 1 {
				t.Error("update rows is", count)
			}

			u, err := conn.Users().Get(id)
			if err != nil {
				t.Error(err)
				return
			}

			if u.Username == insertUser.Username {
				t.Error("excepted isnot newusername, actual is", u.Username)
			}

			if u.Username != "newusername" {
				t.Error("excepted is newusername, actual is", u.Username)
			}

			name, err := conn.Users().GetNameByID(id)
			if err != nil {
				t.Error(err)
				return
			}

			if name != "newusername" {
				t.Error("excepted is newusername, actual is", name)
			}

			count, err = conn.Users().Update(id, &AuthUser{
				Username: "tom",
				Phone:    "8734",
				Status:   123,
			})
			if err != nil {
				t.Error(err)
				return
			}

			if count != 1 {
				t.Error("update rows is", count)
			}

			u, err = conn.Users().Get(id)
			if err != nil {
				t.Error(err)
				return
			}

			if u.Status != 123 {
				t.Error("excepted is 123, actual is", u.Status)
			}

			if u.Username != "tom" {
				t.Error("excepted is tom, actual is", u.Username)
			}

			if u.Phone != "8734" {
				t.Error("excepted is 8734, actual is", u.Phone)
			}
		})

		t.Run("list", func(t *testing.T) {
			_, err := conn.Users().DeleteAll()
			if err != nil {
				t.Error(err)
				return
			}

			_, err = conn.Users().Insert(&insertUser)
			if err != nil {
				t.Error(err)
				return
			}

			list, err := conn.Users().List(0, 10)
			if err != nil {
				t.Error(err)
				return
			}
			if len(list) == 0 {
				t.Error("result is empty")
				return
			}
			if len(list) != 1 {
				t.Error("result is ", len(list))
				return
			}
			u := list[0]

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
		t.Run("tx", func(t *testing.T) {
			_, err := conn.Users().DeleteAll()
			if err != nil {
				t.Error(err)
				return
			}

			tx, err := conn.Begin()
			if err != nil {
				t.Error(err)
				return
			}

			userDaoInTx := tx.Users()
			id, err := userDaoInTx.Insert(&insertUser)
			if err != nil {
				t.Error(err)
				return
			}

			if err = tx.Commit(); err != nil {
				t.Error(err)
				return
			}

			userDao := conn.Users()
			_, err = userDao.Delete(id)
			if err != nil {
				t.Error(err)
				return
			}
			tx, err = conn.Begin()
			if err != nil {
				t.Error(err)
				return
			}

			userDaoInTx = tx.Users()
			_, err = userDaoInTx.Insert(&insertUser)
			if err != nil {
				t.Error(err)
				return
			}

			if err = tx.Rollback(); err != nil {
				t.Error(err)
				return
			}

			c, err := userDao.Count()
			if err != nil {
				t.Error(err)
				return
			}
			if c != 0 {
				t.Error("count isnot 0, actual is", c)
			}
		})
	})
}
