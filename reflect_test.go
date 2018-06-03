package gobatis_test

import (
	"testing"
	"time"

	gobatis "github.com/runner-mei/GoBatis"
	"github.com/runner-mei/GoBatis/tests"
)

func TestReflect(t *testing.T) {
	mapper := gobatis.CreateMapper("db", nil)
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

		t.Run("scanMap", func(t *testing.T) {
			id, err := factory.Insert("insertUser", insertUser)
			if err != nil {
				t.Error(err)
			}

			rows, err := factory.DB().Query("select * from gobatis_users where id=$1", id)
			if err != nil {
				t.Error(err)
			}
			defer rows.Close()

			if !rows.Next() {
				t.Error("next")
				return
			}
			var m map[string]interface{}
			err = gobatis.ScanAny(nil, rows, &m, false, true)
			if err != nil {
				t.Error(err)
			}

			var m2 = map[string]interface{}{}
			err = gobatis.ScanAny(nil, rows, &m2, false, true)
			if err != nil {
				t.Error(err)
			}
			if len(m2) == 0 {
				t.Error("m2 is empty")
			}
		})

		t.Run("scanStruct", func(t *testing.T) {
			id, err := factory.Insert("insertUser", insertUser)
			if err != nil {
				t.Error(err)
			}

			rows, err := factory.DB().Query("select * from gobatis_users where id=$1", id)
			if err != nil {
				t.Error(err)
			}
			defer rows.Close()

			if !rows.Next() {
				t.Error("next")
				return
			}
			var m tests.User
			err = gobatis.StructScan(mapper, rows, &m, true)
			if err != nil {
				t.Error(err)
			}
		})

		t.Run("scanMaps", func(t *testing.T) {
			id, err := factory.Insert("insertUser", insertUser)
			if err != nil {
				t.Error(err)
			}

			rows, err := factory.DB().Query("select * from gobatis_users where id=$1", id)
			if err != nil {
				t.Error(err)
			}
			defer rows.Close()

			var m []map[string]interface{}
			err = gobatis.ScanAll(nil, rows, &m, false, true)
			if err != nil {
				t.Error(err)
			}
		})

		// t.Run("scanMapsByID", func(t *testing.T) {
		// 	id, err := factory.Insert("insertUser", insertUser)
		// 	if err != nil {
		// 		t.Error(err)
		// 	}

		// 	rows, err := factory.DB().Query("select * from gobatis_users where id=$1", id)
		// 	if err != nil {
		// 		t.Error(err)
		// 	}
		// 	defer rows.Close()

		// 	var m map[int64]map[string]interface{}
		// 	err = gobatis.ScanAll(mapper, rows, &m, false, true)
		// 	if err != nil {
		// 		t.Error(err)
		// 	}
		// })
		t.Run("scanStructByID", func(t *testing.T) {
			id, err := factory.Insert("insertUser", insertUser)
			if err != nil {
				t.Error(err)
			}

			rows, err := factory.DB().Query("select * from gobatis_users where id=$1", id)
			if err != nil {
				t.Error(err)
			}
			defer rows.Close()

			var m map[int64]tests.User
			err = gobatis.ScanAll(mapper, rows, &m, false, true)
			if err != nil {
				t.Error(err)
			}
		})
		t.Run("scanStructPtrByID", func(t *testing.T) {
			id, err := factory.Insert("insertUser", insertUser)
			if err != nil {
				t.Error(err)
			}

			rows, err := factory.DB().Query("select * from gobatis_users where id=$1", id)
			if err != nil {
				t.Error(err)
			}
			defer rows.Close()

			var m map[int64]*tests.User
			err = gobatis.ScanAll(mapper, rows, &m, false, true)
			if err != nil {
				t.Error(err)
			}
		})
	})
}
