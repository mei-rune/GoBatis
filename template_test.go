package gobatis_test

import (
	"database/sql"
	"net"
	"strings"
	"testing"
	"time"

	gobatis "github.com/runner-mei/GoBatis"
	"github.com/runner-mei/GoBatis/tests"
)

func TestTemplates(t *testing.T) {

	tests.Run(t, func(_ testing.TB, factory *gobatis.SessionFactory) {
		mac, _ := net.ParseMAC("01:02:03:04:A5:A6")
		ip := net.ParseIP("192.168.1.1")
		insertUser1 := tests.User{
			Name:        "张三",
			Nickname:    "haha",
			Password:    "password",
			Description: "地球人",
			Address:     "沪南路1155号",
			HostIP:      ip,
			HostMAC:     mac,
			HostIPPtr:   &ip,
			HostMACPtr:  &mac,
			Sex:         "女",
			ContactInfo: map[string]interface{}{"QQ": "8888888"},
			Birth:       time.Now(),
			CreateTime:  time.Now(),
		}

		insertUser2 := tests.User{
			Name:        "李四",
			Nickname:    "lishi",
			Password:    "password",
			Description: "地球人",
			Address:     "沪南路1155号",
			HostIP:      ip,
			HostMAC:     mac,
			HostIPPtr:   &ip,
			HostMACPtr:  &mac,
			Sex:         "另",
			ContactInfo: map[string]interface{}{"QQ": "8888888"},
			Birth:       time.Now(),
			CreateTime:  time.Now(),
		}

		id1, err := factory.Insert("insertUser", insertUser1)
		if err != nil {
			t.Error(err)
			return
		}

		id2, err := factory.Insert("insertUser", insertUser2)
		if err != nil {
			t.Error(err)
			return
		}

		ref := factory.Reference()
		users := tests.NewUsers(&ref)

		t.Run("id array is nil", func(t *testing.T) {
			list, err := users.Query(nil)
			if err != nil {
				if !strings.Contains(err.Error(), sql.ErrNoRows.Error()) {
					t.Error("excepted is", sql.ErrNoRows)
					t.Error("actual   is", err)
				}
				return
			}

			if len(list) != 2 {
				t.Error("len(list): except 2 got", len(list))
			}
		})

		t.Run("id array is empty", func(t *testing.T) {
			list, err := users.Query([]int64{})
			if err != nil {
				if !strings.Contains(err.Error(), sql.ErrNoRows.Error()) {
					t.Error("excepted is", sql.ErrNoRows)
					t.Error("actual   is", err)
				}
				return
			}

			if len(list) != 2 {
				t.Error("len(list): except 2 got", len(list))
			}
		})

		t.Run("id array is id1", func(t *testing.T) {
			list, err := users.Query([]int64{id1})
			if err != nil {
				t.Error(err)
				return
			}

			if len(list) != 1 {
				t.Error("len(list): except 1 got", len(list))
			}
		})

		t.Run("id array is id1, id2", func(t *testing.T) {
			list, err := users.Query([]int64{id1, id2})
			if err != nil {
				t.Error(err)
				return
			}

			if len(list) != 2 {
				t.Error("len(list): except 2 got", len(list))
			}
		})
	})
}
