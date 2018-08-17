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
		users := tests.NewTestUsers(&ref)

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

func TestTemplateFuncs(t *testing.T) {
	isLast := gobatis.TemplateFuncs["isLast"].(func(list interface{}, idx int) bool)
	isNotLast := gobatis.TemplateFuncs["isNotLast"].(func(list interface{}, idx int) bool)
	isFirst := gobatis.TemplateFuncs["isFirst"].(func(list interface{}, idx int) bool)
	isNotFirst := gobatis.TemplateFuncs["isNotFirst"].(func(list interface{}, idx int) bool)
	isEmpty := gobatis.TemplateFuncs["isEmpty"].(func(list interface{}) bool)
	isNotEmpty := gobatis.TemplateFuncs["isNotEmpty"].(func(list interface{}) bool)

	assertTrue := func(r bool, msg string) {
		if !r {
			t.Error(msg)
		}
	}

	assertTrue(!isLast(nil, 0), "isLast(nil)")
	assertTrue(!isLast(0, 0), "isLast(0)")

	assertTrue(!isLast([]int{}, 0), "isLast({},0)")
	assertTrue(isNotLast([]int{}, 0), "isNotLast({},0)")

	assertTrue(!isLast([]int{1, 2}, 0), "isLast({1,2},0)")
	assertTrue(isLast([]int{1, 2}, 1), "isLast({1,2},1)")

	assertTrue(isNotLast([]int{1, 2}, 0), "isNotLast({1,2},0)")
	assertTrue(!isNotLast([]int{1, 2}, 1), "isNotLast({1,2},1)")

	assertTrue(!isFirst(nil, 0), "isFirst(nil)")
	assertTrue(!isFirst(0, 0), "isFirst(0)")

	assertTrue(!isFirst([]int{1, 2}, 1), "isFirst({1,2},1)")
	assertTrue(isFirst([]int{1, 2}, 0), "isFirst({1,2},0)")

	assertTrue(isNotFirst([]int{1, 2}, 1), "isNotFirst({1,2},1)")
	assertTrue(!isNotFirst([]int{1, 2}, 0), "isNotFirst({1,2},0)")

	assertTrue(!isFirst([]int{}, 0), "isFirst({},0)")
	assertTrue(isNotFirst([]int{}, 0), "isNotFirst({},0)")

	assertTrue(isEmpty(nil), "isEmpty(nil)")
	assertTrue(!isNotEmpty(nil), "isNotEmpty(nil)")
	assertTrue(!isEmpty(1), "isEmpty(1)")

	assertTrue(!isEmpty([]int{1, 2}), "isEmpty({1,2})")
	assertTrue(isNotEmpty([]int{1, 2}), "isEmpty({1,2})")
}
