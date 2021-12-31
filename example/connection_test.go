package example

import (
	"context"
	"fmt"
	"testing"

	gobatis "github.com/runner-mei/GoBatis"
	"github.com/runner-mei/GoBatis/tests"
)

const (
	postgres = `
DROP TABLE IF EXISTS auth_users_and_roles;

DROP TABLE IF EXISTS user_profiles;

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
);

CREATE TABLE IF NOT EXISTS user_profiles (
  id bigserial PRIMARY KEY,
  user_id int NOT NULL,
  name varchar(45) DEFAULT NULL,
  value varchar(255) DEFAULT NULL,
  created_at TIMESTAMP default NOW(),
  updated_at TIMESTAMP default NOW()
);

DROP TABLE IF EXISTS auth_roles;

CREATE TABLE IF NOT EXISTS auth_roles (
  id bigserial PRIMARY KEY,
  name VARCHAR(32) NOT NULL UNIQUE,
  created_at TIMESTAMP default NOW(),
  updated_at TIMESTAMP default NOW()
);

CREATE TABLE IF NOT EXISTS auth_users_and_roles (
  user_id bigint,
  role_id bigint,

  PRIMARY KEY(user_id, role_id)
);`

	mssql = `IF OBJECT_ID('dbo.user_profiles', 'U') IS NOT NULL 
		DROP TABLE user_profiles;

IF OBJECT_ID('dbo.auth_users_and_roles', 'U') IS NOT NULL
DROP TABLE auth_users_and_roles;

	IF OBJECT_ID('dbo.auth_users', 'U') IS NOT NULL
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
);

CREATE TABLE user_profiles (
  id int IDENTITY NOT NULL PRIMARY KEY,
  user_id int NOT NULL,
  name varchar(45) DEFAULT NULL,
  value varchar(255) DEFAULT NULL,
  created_at datetimeoffset default CURRENT_TIMESTAMP,
  updated_at datetimeoffset default CURRENT_TIMESTAMP
);

IF OBJECT_ID('dbo.auth_roles', 'U') IS NOT NULL
DROP TABLE auth_roles;

CREATE TABLE auth_roles (
  id int IDENTITY PRIMARY KEY,
  name VARCHAR(32) NOT NULL UNIQUE,
  created_at TIMESTAMP default NOW(),
  updated_at TIMESTAMP default NOW()
);

CREATE TABLE auth_users_and_roles (
  user_id int,
  role_id int,
  
  PRIMARY KEY(user_id, role_id)
);`

	mysql = `DROP TABLE IF EXISTS auth_users_and_roles;
DROP TABLE IF EXISTS user_profiles;
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
) ENGINE=InnoDB DEFAULT CHARSET=utf8;

CREATE TABLE IF NOT EXISTS user_profiles (
  id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
  user_id int NOT NULL,
  name varchar(45) DEFAULT NULL,
  value varchar(255) DEFAULT NULL,
  created_at TIMESTAMP default CURRENT_TIMESTAMP,
  updated_at TIMESTAMP default CURRENT_TIMESTAMP
) ENGINE=InnoDB DEFAULT CHARSET=utf8;


DROP TABLE IF EXISTS auth_roles;

CREATE TABLE IF NOT EXISTS auth_roles (
  id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
  name VARCHAR(32) NOT NULL UNIQUE,
  created_at TIMESTAMP default NOW(),
  updated_at TIMESTAMP default NOW()
) ENGINE=InnoDB DEFAULT CHARSET=utf8;

CREATE TABLE IF NOT EXISTS auth_users_and_roles (
  user_id BIGINT,
  role_id BIGINT,
  
  PRIMARY KEY(user_id, role_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;`

	dmsql = `DROP TABLE IF EXISTS user_profiles;
DROP TABLE IF EXISTS auth_users_and_roles;
DROP TABLE IF EXISTS auth_roles;
DROP TABLE IF EXISTS auth_users;

CREATE TABLE auth_users (
  id int IDENTITY PRIMARY KEY,
  username VARCHAR(32) NOT NULL UNIQUE,
  phone VARCHAR(32),
  address VARCHAR(256),
  status TINYINT,
  birth_day DATE,
  created_at TIMESTAMP WITH TIME ZONE default CURRENT_TIMESTAMP,
  updated_at TIMESTAMP WITH TIME ZONE default CURRENT_TIMESTAMP
);

CREATE TABLE user_profiles (
  id int IDENTITY NOT NULL PRIMARY KEY,
  user_id int NOT NULL,
  name varchar(45) DEFAULT NULL,
  value varchar(255) DEFAULT NULL,
  created_at TIMESTAMP WITH TIME ZONE default CURRENT_TIMESTAMP,
  updated_at TIMESTAMP WITH TIME ZONE default CURRENT_TIMESTAMP
);


CREATE TABLE auth_roles (
  id int IDENTITY NOT NULL PRIMARY KEY,
  name VARCHAR(32) NOT NULL UNIQUE,
  created_at TIMESTAMP WITH TIME ZONE default NOW(),
  updated_at TIMESTAMP WITH TIME ZONE default NOW()
);

CREATE TABLE auth_users_and_roles (
  user_id int,
  role_id int,
  
  PRIMARY KEY(user_id, role_id)
);`
)

func GetTestSQL(name string) string {
	switch name {
	case gobatis.Postgres.Name():
		return postgres
	case gobatis.MSSql.Name():
		return mssql
	case gobatis.DM.Name():
		return dmsql
	default:
		return mysql
	}
}

func toString(v interface{}) string {
	if bs, ok := v.([]byte); ok {
		return string(bs)
	}
	return fmt.Sprint(v)
}

func TestConnection(t *testing.T) {
	insertUser := User{
		Username: "abc",
		Phone:    "123",
		Status:   1,
	}

	tests.Run(t, func(_ testing.TB, factory *gobatis.SessionFactory) {
		sqltext := GetTestSQL(factory.Dialect().Name())
		err := gobatis.ExecContext(context.Background(), factory.DB(), sqltext)
		if err != nil {
			t.Error(factory.Dialect().Name())
			t.Error(sqltext)
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
				if toString(umap["USERNAME"]) != insertUser.Username {
					t.Error("excepted is", insertUser.Username, ", actual is", umap["username"])
				}
			}

			if toString(umap["phone"]) != insertUser.Phone {
				if toString(umap["PHONE"]) != insertUser.Phone {
					t.Error("excepted is", insertUser.Phone, ", actual is", umap["phone"])
				}
			}

			if toString(umap["status"]) != fmt.Sprint(insertUser.Status) {
				if toString(umap["STATUS"]) != fmt.Sprint(insertUser.Status) {
					t.Error("excepted is", insertUser.Status, ", actual is", umap["status"])
				}
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

			count, err = conn.Users().Update(id, &User{
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

			c, err := userDao.Count()
			if err != nil {
				t.Error(err)
				return
			}
			if c != 0 {
				t.Error("count isnot 0, actual is", c)
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

			c, err = userDao.Count()
			if err != nil {
				t.Error(err)
				return
			}
			if c != 0 {
				t.Error("count isnot 0, actual is", c)
			}
		})

		t.Run("roles", func(t *testing.T) {
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

			rolename1 := "role1"
			roleid1, err := conn.Roles().Insert(rolename1)
			if err != nil {
				t.Error(err)
				return
			}

			err = conn.Roles().AddUser(insertUser.Username, rolename1)
			if err != nil {
				t.Error(err)
				return
			}

			var count int
			err = conn.DB().QueryRowContext(context.Background(), "select count(*) from auth_users_and_roles").Scan(&count)
			if err != nil {
				t.Error(err)
				return
			}

			if count != 1 {
				t.Error("excepted is", 1)
				t.Error("actual   is", count)
			}

			users, err := conn.Roles().Users(roleid1)
			if err != nil {
				t.Error(err)
				return
			}

			if len(users) != 1 {
				t.Error("excepted is", 1)
				t.Error("actual   is", len(users))
			}

			err = conn.Roles().RemoveUser(insertUser.Username, rolename1)
			if err != nil {
				t.Error(err)
				return
			}

			err = conn.DB().QueryRowContext(context.Background(), "select count(*) from auth_users_and_roles").Scan(&count)
			if err != nil {
				t.Error(err)
				return
			}

			if count != 0 {
				t.Error("excepted is", 0)
				t.Error("actual   is", count)
			}

		})

	})
}
