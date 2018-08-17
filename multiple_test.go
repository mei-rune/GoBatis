package gobatis_test

import (
	"net"
	"strings"
	"testing"
	"time"

	gobatis "github.com/runner-mei/GoBatis"
	"github.com/runner-mei/GoBatis/tests"
)

func TestMultiple(t *testing.T) {
	tests.Run(t, func(_ testing.TB, factory *gobatis.SessionFactory) {
		mac, _ := net.ParseMAC("01:02:03:04:A5:A6")
		ip := net.ParseIP("192.168.1.1")
		insertUser := tests.User{
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
		ref := factory.Reference()
		users := tests.NewTestUsers(&ref)

		_, err := users.Insert(&insertUser)
		if err != nil {
			t.Error(err)
			return
		}

		t.Run("QueryFieldNotExist", func(t *testing.T) {
			_, _, err := users.QueryFieldNotExist1()
			if err == nil {
				t.Error("except error got ok")
				return
			}

			if !strings.Contains(err.Error(), "isnot exists in the names - u,name") {
				t.Error("excepted is", "isnot exists in the names - u,name")
				t.Error("actual   is", err)
			}

			_, _, err = users.QueryFieldNotExist2()
			if err == nil {
				t.Error("except error got ok")
				return
			}

			if !strings.Contains(err.Error(), "isnot exists in the names - u,name") {
				t.Error("excepted is", "isnot exists in the names - u,name")
				t.Error("actual   is", err)
			}
		})

		t.Run("QueryReturnDupError", func(t *testing.T) {
			_, _, err := users.QueryReturnDupError1()
			if err == nil {
				t.Error("except error got ok")
				return
			}

			if !strings.Contains(err.Error(), "column 'name' is duplicated with") {
				t.Error("excepted is", "column 'name' is duplicated with")
				t.Error("actual   is", err)
			}

			_, _, err = users.QueryReturnDupError2()
			if err == nil {
				t.Error("except error got ok")
				return
			}

			if !strings.Contains(err.Error(), "column 'name_name' is duplicated with") {
				t.Error("excepted is", "column 'name_name' is duplicated with")
				t.Error("actual   is", err)
			}
		})
	})
}
