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

func TestInsert(t *testing.T) {
	tests.Run(t, func(_ testing.TB, factory *gobatis.SessionFactory) {

		ref := factory.Reference()
		users := tests.NewTestUsers(&ref)

		mac, _ := net.ParseMAC("01:02:03:04:A5:A6")
		ip := net.ParseIP("192.168.1.1")

		now := time.Now()
		id, err := users.InsertByArgs("abc", "an", "aa", "acc", time.Now(), "127.0.0.1",
			ip, mac, &ip, &mac, "male", map[string]interface{}{"abc": "123"},
			1, 2, 3.2, 3.4, "t5", now, &now, now)
		if err != nil {
			t.Error(err)
			return
		}

		if id == 0 {
			t.Error("except not 0 got ", id)
			return
		}

		u, err := users.Get(id)
		if err != nil {
			t.Error(err)
			return
		}

		if u.Name != "abc" {
			t.Error(1)
		}
		if u.Nickname != "an" {
			t.Error(1)
		}
		if u.Password != "aa" {
			t.Error(1)
		}
		if u.Description != "acc" {
			t.Error(1)
		}
		if u.Address != "127.0.0.1" {
			t.Error(1)
		}

		if !u.HostIP.Equal(ip) {
			t.Error("except", ip, "got", u.HostIP)
		}

		if mac.String() != u.HostMAC.String() {
			t.Error("except", mac, "got", u.HostMAC)
		}

		if !u.HostIPPtr.Equal(ip) {
			t.Error("except", ip, "got", u.HostIP)
		}

		if mac.String() != u.HostMACPtr.String() {
			t.Error("except", mac, "got", u.HostMACPtr)
		}

		if u.Sex != "male" {
			t.Error("except male got", u.Sex)
		}

		if !reflect.DeepEqual(u.ContactInfo, map[string]interface{}{"abc": "123"}) {
			t.Error("except map[string]interface{}{\"abc\": \"123\"} got", u.ContactInfo)
		}

		if u.Field1 != 1 {
			t.Error("except 1 got", u.Field1)
		}

		if u.Field2 != 2 {
			t.Error("except 2 got", u.Field2)
		}

		if u.Field3 != 3.2 {
			t.Error("except 3.2 got", u.Field3)
		}

		if u.Field4 != 3.4 {
			t.Error("except 3.4 got", u.Field4)
		}

		if u.Field5 != "t5" {
			t.Error("except t5 got", u.Field5)
		}

		timeEQU := func(a, b time.Time) bool {
			interval := a.Sub(b)
			if interval < 0 {
				interval = -interval
			}
			return interval > 2*time.Second
		}
		if timeEQU(now, u.Field6) {
			t.Error("except ", now, " got", u.Field6)
		}

		if u.Field7 == nil || timeEQU(now, *u.Field7) {
			t.Error("except ", now, " got", u.Field7)
		}

		if timeEQU(now, u.CreateTime) {
			t.Error("except ", now, " got", u.CreateTime)
		}

		id, err = users.InsertByArgs("abc", "an", "aa", "acc", time.Now(), "127.0.0.1",
			ip, mac, nil, nil, "male", map[string]interface{}{"abc": "123"},
			1, 2, 3.2, 3.4, "t5", now, nil, now)
		if err != nil {
			t.Error(err)
			return
		}

		if id == 0 {
			t.Error("except not 0 got ", id)
			return
		}

		u, err = users.Get(id)
		if err != nil {
			t.Error(err)
			return
		}

		if u.HostIPPtr != nil && len(*u.HostIPPtr) > 0 {
			t.Error("except", nil, "got", u.HostIPPtr)
		}

		if u.HostMACPtr != nil && len(*u.HostMACPtr) > 0 {
			t.Error("except", nil, "got", u.HostMACPtr)
		}

		if u.Field7 != nil {
			t.Error("except", nil, "got", u.Field7)
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
