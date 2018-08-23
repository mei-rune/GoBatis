package gobatis_test

import (
	"database/sql"
	"net"
	"reflect"
	"testing"
	"time"

	gobatis "github.com/runner-mei/GoBatis"
	"github.com/runner-mei/GoBatis/tests"
)

func TestMaxID(t *testing.T) {
	tests.Run(t, func(_ testing.TB, factory *gobatis.SessionFactory) {

		ref := factory.Reference()
		groups := tests.NewTestUserGroups(&ref)

		_, err := groups.MaxID()
		if err == nil {
			t.Error("except error got ok")
			return
		}

		if err != sql.ErrNoRows {
			t.Error("except ErrNoRows got ", err)
		}

		count, err := groups.Count()
		if err != nil {
			t.Error("except ok got ", err)
			return
		}

		if count != 0 {
			t.Error("except 0 got ", count)
		}
	})
}

func TestInsetOneParam(t *testing.T) {
	tests.Run(t, func(_ testing.TB, factory *gobatis.SessionFactory) {

		group1 := tests.UserGroup{
			Name: "g1",
		}

		ref := factory.Reference()
		groups := tests.NewTestUserGroups(&ref)

		g1, err := groups.Insert(&group1)
		if err != nil {
			t.Error(err)
			return
		}
		g2, err := groups.InsertByName("g2")
		if err != nil {
			t.Error(err)
			return
		}

		gv1, err := groups.Get(g1)
		if err != nil {
			t.Error(err)
			return
		}
		if gv1.Name != group1.Name {
			t.Error("except", group1.Name, "got", gv1.Name)
		}

		gv2, err := groups.Get(g2)
		if err != nil {
			t.Error(err)
			return
		}
		if gv2.Name != "g2" {
			t.Error("except 'g2' got", gv2.Name)
		}
	})
}

func TestReadOnly(t *testing.T) {
	tests.Run(t, func(_ testing.TB, factory *gobatis.SessionFactory) {
		mac, _ := net.ParseMAC("01:02:03:04:A5:A6")
		ip := net.ParseIP("192.168.1.1")
		user1 := tests.User{
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

		user2 := tests.User{
			Name:        "李四",
			Nickname:    "haha",
			Password:    "password",
			Description: "地球人",
			Address:     "ss",
			HostIP:      ip,
			HostMAC:     mac,
			HostIPPtr:   &ip,
			HostMACPtr:  &mac,
			Sex:         "女",
			ContactInfo: map[string]interface{}{"QQ": "abc"},
			Birth:       time.Now(),
			CreateTime:  time.Now(),
		}

		group1 := tests.UserGroup{
			Name: "g1",
		}

		group2 := tests.UserGroup{
			Name: "g1",
		}

		group3 := tests.UserGroup{
			Name: "g3",
		}

		ref := factory.Reference()
		users := tests.NewTestUsers(&ref)
		groups := tests.NewTestUserGroups(&ref)

		u1, err := users.Insert(&user1)
		if err != nil {
			t.Error(err)
			return
		}

		u2, err := users.Insert(&user2)
		if err != nil {
			t.Error(err)
			return
		}

		g1, err := groups.Insert(&group1)
		if err != nil {
			t.Error(err)
			return
		}

		g2, err := groups.Insert(&group2)
		if err != nil {
			t.Error(err)
			return
		}

		g3, err := groups.Insert(&group3)
		if err != nil {
			t.Error(err)
			return
		}

		users.AddToGroup(u1, g1)
		users.AddToGroup(u1, g2)
		users.AddToGroup(u2, g2)

		gv1, err := groups.Get(g1)
		if err != nil {
			t.Error(err)
			return
		}

		if gv1.Name != group1.Name {
			t.Error("except", group1.Name, "got", gv1.Name)
		}

		if len(gv1.UserIDs) != 1 {
			t.Error("except 1 got", len(gv1.UserIDs))
		} else if gv1.UserIDs[0] != u1 {
			t.Error("except [1] got", gv1.UserIDs)
		}

		gv2, err := groups.Get(g2)
		if err != nil {
			t.Error(err)
			return
		}

		if gv2.Name != group2.Name {
			t.Error("except", group2.Name, "got", gv2.Name)
		}

		if len(gv2.UserIDs) != 2 {
			t.Error("except 1 got", len(gv2.UserIDs))
		} else if !reflect.DeepEqual(gv2.UserIDs, []int64{u1, u2}) &&
			!reflect.DeepEqual(gv2.UserIDs, []int64{u2, u1}) {
			t.Error("except [", u1, ",", u2, "] got", gv2.UserIDs)
		}

		gv3, err := groups.Get(g3)
		if err != nil {
			t.Error(err)
			return
		}

		if gv3.Name != group3.Name {
			t.Error("except", group3.Name, "got", gv3.Name)
		}

		if len(gv3.UserIDs) != 0 {
			t.Error("except 0 got", len(gv3.UserIDs))
			t.Error(gv3.UserIDs)
		}

	})
}
