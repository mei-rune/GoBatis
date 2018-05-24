package example

import (
	"database/sql"
	"testing"
	"time"

	// _ "github.com/go-sql-driver/mysql"
	_ "github.com/lib/pq"
	"github.com/runner-mei/GoBatis/tests"
	"github.com/runner-mei/gobatis"
)

func TestSession(t *testing.T) {
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

		user := tests.User{
			Name: "张三",
		}

		t.Run("selectUsers", func(t *testing.T) {
			_, err := factory.Insert("insertUser", insertUser)
			if err != nil {
				t.Error(err)
			}
			_, err = factory.Insert("insertUser", insertUser)
			if err != nil {
				t.Error(err)
			}

			var users []tests.User
			err = factory.Select("selectUsers", user).ScanSlice(&users)
			if err != nil {
				t.Error(err)
				return
			}

			if len(users) != 2 {
				t.Error("excepted size is", 2)
				t.Error("actual size   is", len(users))
			} else {

				insertUser2 := insertUser
				insertUser2.Birth = insertUser2.Birth.UTC()
				insertUser2.CreateTime = insertUser2.CreateTime.UTC()

				for _, u := range users {

					insertUser2.ID = u.ID
					u.Birth = u.Birth.UTC()
					u.CreateTime = u.CreateTime.UTC()

					tests.AssertUser(t, insertUser2, u)
				}
			}
		})

		t.Run("selectUser", func(t *testing.T) {
			_, err := factory.Insert("insertUser", insertUser)
			if err != nil {
				t.Error(err)
			}

			u := tests.User{Name: insertUser.Name}
			err = factory.SelectOne("selectUser", u).Scan(&u)
			if err != nil {
				t.Error(err)
				return
			}

			insertUser.ID = u.ID
			insertUser.Birth = insertUser.Birth.UTC()
			insertUser.CreateTime = insertUser.CreateTime.UTC()
			u.Birth = u.Birth.UTC()
			u.CreateTime = u.CreateTime.UTC()

			tests.AssertUser(t, insertUser, u)
		})

		t.Run("updateUser", func(t *testing.T) {
			_, err := factory.Insert("insertUser", insertUser)
			if err != nil {
				t.Error(err)
			}

			u := tests.User{Name: insertUser.Name}
			err = factory.SelectOne("selectUser", u).Scan(&u)
			if err != nil {
				t.Error(err)
				return
			}

			updateUser := insertUser
			updateUser.ID = u.ID
			updateUser.Nickname = "test@foxmail.com"
			updateUser.Birth = time.Now()
			updateUser.CreateTime = time.Now()
			_, err = factory.Update("updateUser", updateUser)
			if err != nil {
				t.Error(err)
			}

			updateUser.Birth = updateUser.Birth.UTC()
			updateUser.CreateTime = updateUser.CreateTime.UTC()

			err = factory.SelectOne("selectUser", u).Scan(&u)
			if err != nil {
				t.Error(err)
				return
			}
			u.Birth = u.Birth.UTC()
			u.CreateTime = u.CreateTime.UTC()

			tests.AssertUser(t, updateUser, u)
		})

		t.Run("deleteUser", func(t *testing.T) {
			if _, err := factory.DB().Exec(`DELETE FROM gobatis_users`); err != nil {
				t.Error(err)
				return
			}

			_, err := factory.Insert("insertUser", insertUser)
			if err != nil {
				t.Error(err)
			}

			u := tests.User{Name: insertUser.Name}
			err = factory.SelectOne("selectUser", u).Scan(&u)
			if err != nil {
				t.Error(err)
				return
			}

			deleteUser := tests.User{ID: u.ID}
			_, err = factory.Delete("deleteUser", deleteUser)
			if err != nil {
				t.Error(err)
			}

			err = factory.SelectOne("selectUser", u).Scan(&u)
			if err == nil {
				t.Error("DELETE fail")
				return
			}

			if err != sql.ErrNoRows {
				t.Error(err)
			}
		})
	})
}
