package core_test

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/runner-mei/GoBatis/core"
	"github.com/runner-mei/GoBatis/tests"
)

func TestReflect(t *testing.T) {
	var placeholder core.PlaceholderFormat = core.Question
	mapper := core.CreateMapper("db", nil, nil)
	tests.Run(t, func(_ testing.TB, factory *core.Session) {
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

		placeholder = factory.Dialect().Placeholder()

		replacePlaceholders := func(s string) string {
			s, _ = placeholder.ReplacePlaceholders(s)
			return s
		}

		t.Run("scanMap", func(t *testing.T) {
			id, err := factory.Insert(context.Background(), "insertUser", insertUser)
			if err != nil {
				t.Error(err)
				return
			}

			rows, err := factory.DB().QueryContext(context.Background(), replacePlaceholders("select * from gobatis_users where id=?"), id)
			if err != nil {
				t.Error(err)
				return
			}
			defer rows.Close()

			if !rows.Next() {
				t.Error("next")
				return
			}
			var m map[string]interface{}
			err = core.ScanAny(factory.Dialect(), nil, rows, &m, false, true)
			if err != nil {
				t.Error(err)
			}

			var m2 = map[string]interface{}{}
			err = core.ScanAny(factory.Dialect(), nil, rows, &m2, false, true)
			if err != nil {
				t.Error(err)
			}
			if len(m2) == 0 {
				t.Error("m2 is empty")
			}

			m2 = map[string]interface{}{}
			err = core.ScanAny(factory.Dialect(), nil, rows, m2, false, true)
			if err != nil {
				t.Error(err)
			}
			if len(m2) == 0 {
				t.Error("m2 is empty")
			}
		})

		t.Run("scanError", func(t *testing.T) {
			id, err := factory.Insert(context.Background(), "insertUser", insertUser)
			if err != nil {
				t.Error(err)
				return
			}

			rows, err := factory.DB().QueryContext(context.Background(), replacePlaceholders("select * from gobatis_users where id=?"), id)
			if err != nil {
				t.Error(err)
				return
			}
			defer rows.Close()

			if !rows.Next() {
				t.Error("next")
				return
			}

			var notpointer string
			err = core.ScanAny(factory.Dialect(), mapper, rows, notpointer, false, true)
			if err == nil {
				t.Error("excepted is error got ok")
			} else if !strings.Contains(err.Error(), "must pass a pointer") {
				t.Error("excepted is must pass a pointer")
				t.Error("actual   is", err)
			}

			var nilpointer *string
			err = core.ScanAny(factory.Dialect(), mapper, rows, nilpointer, false, true)
			if err == nil {
				t.Error("excepted is error got ok")
			} else if !strings.Contains(err.Error(), "nil pointer passed") {
				t.Error("excepted is nil pointer passed")
				t.Error("actual   is", err)
			}

			var notstruct struct{}
			err = core.ScanAny(factory.Dialect(), mapper, rows, &notstruct, true, true)
			if err == nil {
				t.Error("excepted is error got ok")
			} else if !strings.Contains(err.Error(), "struct") {
				t.Error("excepted is struct")
				t.Error("actual   is", err)
			}

			var errColumns string
			err = core.ScanAny(factory.Dialect(), mapper, rows, &errColumns, false, true)
			if err == nil {
				t.Error("excepted is error got ok")
			} else if !strings.Contains(err.Error(), "scannable dest type string with >1 columns") {
				t.Error("excepted is scannable dest type string with >1 columns")
				t.Error("actual   is", err)
			}

			err = core.ScanAll(factory.Dialect(), mapper, rows, notpointer, false, true)
			if err == nil {
				t.Error("excepted is error got ok")
			} else if !strings.Contains(err.Error(), "must pass a pointer") {
				t.Error("excepted is must pass a pointer")
				t.Error("actual   is", err)
			}

			err = core.ScanAll(factory.Dialect(), mapper, rows, nilpointer, false, true)
			if err == nil {
				t.Error("excepted is error got ok")
			} else if !strings.Contains(err.Error(), "nil pointer passed") {
				t.Error("excepted is nil pointer passed")
				t.Error("actual   is", err)
			}

			err = core.ScanAll(factory.Dialect(), mapper, rows, &notstruct, true, true)
			if err == nil {
				t.Error("excepted is error got ok")
			} else if !strings.Contains(err.Error(), "struct") {
				t.Error("excepted is struct")
				t.Error("actual   is", err)
			}

			var errArrayColumns []string
			err = core.ScanAll(factory.Dialect(), mapper, rows, &errArrayColumns, false, true)
			if err == nil {
				t.Error("excepted is error got ok")
			} else if !strings.Contains(err.Error(), "dest type string with >1 columns") {
				t.Error("excepted is dest type string with >1 columns")
				t.Error("actual   is", err)
			}
		})

		t.Run("scanStruct", func(t *testing.T) {
			id, err := factory.Insert(context.Background(), "insertUser", insertUser)
			if err != nil {
				t.Error(err)
				return
			}

			rows, err := factory.DB().QueryContext(context.Background(), replacePlaceholders("select * from gobatis_users where id=?"), id)
			if err != nil {
				t.Error(err)
				return
			}
			defer rows.Close()

			if !rows.Next() {
				t.Error("next")
				return
			}
			var m tests.User
			err = core.StructScan(factory.Dialect(), mapper, rows, &m, true)
			if err != nil {
				t.Error(err)
			}
		})

		t.Run("scanMaps", func(t *testing.T) {
			id, err := factory.Insert(context.Background(), "insertUser", insertUser)
			if err != nil {
				t.Error(err)
				return
			}

			rows, err := factory.DB().QueryContext(context.Background(), replacePlaceholders("select * from gobatis_users where id=?"), id)
			if err != nil {
				t.Error(err)
				return
			}
			defer rows.Close()

			var m []map[string]interface{}
			err = core.ScanAll(factory.Dialect(), nil, rows, &m, false, true)
			if err != nil {
				t.Error(err)
			}
		})

		// t.Run("scanMapsByID", func(t *testing.T) {
		// 	id, err := factory.Insert("insertUser", insertUser)
		// 	if err != nil {
		// 		t.Error(err)
		// 	}

		// 	rows, err := factory.DB().QueryContext(context.Background(),"select * from gobatis_users where id=$1", id)
		// 	if err != nil {
		// 		t.Error(err)
		// 	}
		// 	defer rows.Close()

		// 	var m map[int64]map[string]interface{}
		// 	err = core.ScanAll(mapper, rows, &m, false, true)
		// 	if err != nil {
		// 		t.Error(err)
		// 	}
		// })
		t.Run("scanStructByID", func(t *testing.T) {
			id, err := factory.Insert(context.Background(), "insertUser", insertUser)
			if err != nil {
				t.Error(err)
				return
			}

			rows, err := factory.DB().QueryContext(context.Background(), replacePlaceholders("select * from gobatis_users where id=?"), id)
			if err != nil {
				t.Error(err)
				return
			}
			defer rows.Close()

			var m map[int64]tests.User
			err = core.ScanAll(factory.Dialect(), mapper, rows, &m, false, true)
			if err != nil {
				t.Error(err)
			}
		})
		t.Run("scanStructPtrByID", func(t *testing.T) {
			id, err := factory.Insert(context.Background(), "insertUser", insertUser)
			if err != nil {
				t.Error(err)
				return
			}

			rows, err := factory.DB().QueryContext(context.Background(), replacePlaceholders("select * from gobatis_users where id=?"), id)
			if err != nil {
				t.Error(err)
				return
			}
			defer rows.Close()

			var m map[int64]*tests.User
			err = core.ScanAll(factory.Dialect(), mapper, rows, &m, false, true)
			if err != nil {
				t.Error(err)
			}
		})
	})
}
