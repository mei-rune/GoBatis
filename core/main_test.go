package core_test

import (
	"context"
	"database/sql"
	"net"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/runner-mei/GoBatis/core"
	"github.com/runner-mei/GoBatis/dialects"
	"github.com/runner-mei/GoBatis/tests"
)

func TestNameExists(t *testing.T) {
	tests.Run(t, func(_ testing.TB, factory *core.Session) {
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
			Field1:      2,
			Field2:      2,
			Field3:      2,
			Field4:      2,
			Field5:      "aba",
			Field6:      time.Now(),
		}

		ref := factory.SessionReference()
		users := tests.NewTestUsers(ref)
		//groups := tests.NewTestUserGroups(&ref, users)

		id, err := users.InsertContext(context.Background(), &insertUser)
		if err != nil {
			t.Error(err)
			return
		}
		insertUser.ID = id

		userExists, err := users.NameExist(insertUser.Name)
		if err != nil {
			t.Error(err)
			return
		}

		if !userExists {
			t.Error("user isnot exists")
			return
		}

		userExists, err = users.NameExist("user_not_exists")
		if err != nil {
			t.Error(err)
			return
		}

		if userExists {
			t.Error("user_not_exists is exists")
			return
		}
	})
}

func TestUpdate(t *testing.T) {
	tests.Run(t, func(_ testing.TB, factory *core.Session) {
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
			Field1:      2,
			Field2:      2,
			Field3:      2,
			Field4:      2,
			Field5:      "aba",
			Field6:      time.Now(),
		}

		ref := factory.SessionReference()
		users := tests.NewTestUsers(ref)
		//groups := tests.NewTestUserGroups(&ref, users)

		id, err := users.InsertContext(context.Background(), &insertUser)
		if err != nil {
			t.Error(err)
			return
		}
		insertUser.ID = id

		newUser := insertUser
		newUser.Name = "SetName_abc"
		_, err = users.SetName(id, newUser.Name)
		if err != nil {
			t.Error(err)
			return
		}

		u, err := users.GetByID(id)
		if err != nil {
			t.Error(err)
			return
		}
		tests.AssertUser(t, newUser, u)

		var a tests.User
		err = users.GetByIDWithCallback(id)(&a)
		if err != nil {
			t.Error(err)
			return
		}
		tests.AssertUser(t, newUser, a)

		newUser.Name = "SetName_abc_context"
		_, err = users.SetNameWithContext(context.Background(), id, newUser.Name)
		if err != nil {
			t.Error(err)
			return
		}

		u, err = users.GetByID(id)
		if err != nil {
			t.Error(err)
			return
		}
		tests.AssertUser(t, newUser, u)

		newUser.Birth = time.Now().AddDate(-1, -1, -1)
		_, err = users.SetBirth(id, newUser.Birth)
		if err != nil {
			t.Error(err)
			return
		}

		u, err = users.GetByID(id)
		if err != nil {
			t.Error(err)
			return
		}
		tests.AssertUser(t, newUser, u)

		newUser.Birth = time.Now().AddDate(-3, -3, -3)
		_, err = users.SetBirthWithContext(context.Background(), id, newUser.Birth)
		if err != nil {
			t.Error(err)
			return
		}

		u, err = users.GetByID(id)
		if err != nil {
			t.Error(err)
			return
		}
		tests.AssertUser(t, newUser, u)
	})
}

func TestContextSimple(t *testing.T) {
	tests.Run(t, func(_ testing.TB, factory *core.Session) {
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
			Field1:      2,
			Field2:      2,
			Field3:      2,
			Field4:      2,
			Field5:      "aba",
			Field6:      time.Now(),
		}

		ref := factory.SessionReference()
		users := tests.NewTestUsers(ref)

		id, err := users.InsertContext(context.Background(), &insertUser)
		if err != nil {
			t.Error(err)
			return
		}
		insertUser.ID = id

		u, err := users.GetContext(context.Background(), id)
		if err != nil {
			t.Error(err)
			return
		}

		tests.AssertUser(t, insertUser, *u)

		count, err := users.CountContext(context.Background())
		if count != 1 {
			t.Error("except 1 got ", count)
			return
		}

		all, err := users.GetAllContext(context.Background())
		if err != nil {
			t.Error(err)
			return
		}
		if len(all) != 1 {
			t.Error("except 1 got ", len(all))
			return
		}

		tests.AssertUser(t, insertUser, all[0])
	})
}

func TestMaxID(t *testing.T) {
	tests.Run(t, func(_ testing.TB, factory *core.Session) {

		ref := factory.SessionReference()
		users := tests.NewTestUsers(ref)
		groups := tests.NewTestUserGroups(ref, users)

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

func TestInsertUser(t *testing.T) {
	tests.Run(t, func(_ testing.TB, factory *core.Session) {

		ref := factory.SessionReference()
		users := tests.NewTestUsers(ref)

		mac, _ := net.ParseMAC("01:02:03:04:A5:A6")
		ip := net.ParseIP("192.168.1.1")

		now := time.Now()
		boolTrue := true

		id, err := users.InsertByArgs("abc", "an", "aa", "acc", time.Now(), "127.0.0.1",
			ip, mac, &ip, &mac, "male", map[string]interface{}{"abc": "123"},
			1, 2, 3.2, 3.4, "t5", now, &now,
			true, &boolTrue, now)
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

		id, err = users.InsertByArgs("abcad", "anad", "aa", "acc", time.Now(), "127.0.0.1",
			ip, mac, nil, nil, "male", map[string]interface{}{"abc": "123"},
			1, 2, 3.2, 3.4, "t5", now, nil, true, nil, now)
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

	t.Run("测试 Insert 的参数名和字段名同名的情况", func(t *testing.T) {
		tests.Run(t, func(_ testing.TB, factory *core.Session) {

			ref := factory.SessionReference()
			settings := tests.NewTestSettings(ref)

			setting1 := &tests.Setting{
				Name:  "abcaa",
				Value: "abaa",
			}
			id, err := settings.InsertSetting1(setting1)
			if err != nil {
				t.Error(err)
				return
			}
			if id == 0 {
				t.Error("except not 0 got ", id)
				return
			}
			u, err := settings.GetSetting(id)
			if err != nil {
				t.Error(err)
				return
			}

			if u.Name != setting1.Name {
				t.Error("except", setting1.Name, "got", u.Name)
			}

			if u.Value != setting1.Value {
				t.Error("except", setting1.Value, "got", u.Value)
			}

			setting1.Name += "abc"
			id, err = settings.InsertSetting2(setting1)
			if err != nil {
				t.Error(err)
				return
			}
			if id == 0 {
				t.Error("except not 0 got ", id)
				return
			}
			u, err = settings.GetSetting(id)
			if err != nil {
				t.Error(err)
				return
			}

			if u.Name != setting1.Name {
				t.Error("except", setting1.Name, "got", u.Name)
			}

			if u.Value != setting1.Value {
				t.Error("except", setting1.Value, "got", u.Value)
			}

			value1 := tests.ListValue{
				Name: "g1",
			}

			g1, err := settings.InsertListValue1(&value1)
			if err != nil {
				t.Error(err)
				return
			}

			gv1, err := settings.GetListValue(g1)
			if err != nil {
				t.Error(err)
				return
			}
			if gv1.Name != value1.Name {
				t.Error("except", value1.Name, "got", gv1.Name)
			}

			value1.Name += "abc"
			g1, err = settings.InsertListValue1(&value1)
			if err != nil {
				t.Error(err)
				return
			}

			gv1, err = settings.GetListValue(g1)
			if err != nil {
				t.Error(err)
				return
			}
			if gv1.Name != value1.Name {
				t.Error("except", value1.Name, "got", gv1.Name)
			}

		})
	})

	t.Run("测试 Upsert 的参数名和字段名同名的情况", func(t *testing.T) {
		tests.Run(t, func(_ testing.TB, factory *core.Session) {

			ref := factory.SessionReference()
			settings := tests.NewTestSettings(ref)

			setting1 := &tests.Setting{
				Name:  "abcaa",
				Value: "abaa",
			}
			id, err := settings.UpsertSetting1(setting1)
			if err != nil {
				t.Error(err)
				return
			}
			if id == 0 {
				if factory.Dialect() == dialects.DM {
					t.Skip("dm is unsupport upsert")
					// mysql is unsupport
					return
				}

				t.Error("except not 0 got ", id)
				return
			}
			u, err := settings.GetSetting(id)
			if err != nil {
				t.Error(err)
				return
			}

			if u.Name != setting1.Name {
				t.Error("except", setting1.Name, "got", u.Name)
			}

			if u.Value != setting1.Value {
				t.Error("except", setting1.Value, "got", u.Value)
			}

			setting1.Name += "abc"
			id, err = settings.UpsertSetting2(setting1)
			if err != nil {
				t.Error(err)
				return
			}
			if id == 0 {
				t.Error("except not 0 got ", id)
				return
			}
			u, err = settings.GetSetting(id)
			if err != nil {
				t.Error(err)
				return
			}

			if u.Name != setting1.Name {
				t.Error("except", setting1.Name, "got", u.Name)
			}

			if u.Value != setting1.Value {
				t.Error("except", setting1.Value, "got", u.Value)
			}

			if factory.Dialect() == dialects.Mysql {
				// mysql is unsupport
				return
			}

			value1 := tests.ListValue{
				Name: "g1",
			}

			g1, err := settings.UpsertListValue1(&value1)
			if err != nil {
				t.Error(err)
				return
			}

			gv1, err := settings.GetListValue(g1)
			if err != nil {
				t.Error(err)
				return
			}
			if gv1.Name != value1.Name {
				t.Error("except", value1.Name, "got", gv1.Name)
			}

			value1.Name += "abc"
			g1, err = settings.UpsertListValue1(&value1)
			if err != nil {
				t.Error(err)
				return
			}

			gv1, err = settings.GetListValue(g1)
			if err != nil {
				t.Error(err)
				return
			}
			if gv1.Name != value1.Name {
				t.Error("except", value1.Name, "got", gv1.Name)
			}
		})
	})

	t.Run("测试 insert 时返回对象，而不是 ID", func(t *testing.T) {
		tests.Run(t, func(_ testing.TB, factory *core.Session) {

			if factory.Dialect() != dialects.Postgres {
				t.Skip("only support Postgres")
				return
			}

			ref := factory.SessionReference()
			users := tests.NewTestUsers(ref)
			usergroups := tests.NewTestUserGroups(ref, users)

			ug, err := usergroups.InsertByName4("aaaa")
			if err != nil {
				t.Error(err)
				return
			}
			if ug.ID == 0 {
				t.Error("ug.ID == 0")
				return
			}
			if ug.Name != "aaaa" {
				t.Errorf("ug.Name(%q) != \"aaaa\"", ug.Name)
				return
			}
		})
	})
}

func TestInsertOneParam(t *testing.T) {
	tests.Run(t, func(_ testing.TB, factory *core.Session) {

		group0 := tests.UserGroup{
			Name: "g1",
		}

		ref := factory.Reference()
		users := tests.NewTestUsers(&ref)
		groups := tests.NewTestUserGroups(&ref, users)

		g0, err := groups.Insert(&group0)
		if err != nil {
			t.Error(err)
			return
		}

		g2, err := groups.InsertByName("g2")
		if err != nil {
			t.Error(err)
			return
		}
		g3, err := groups.InsertByName2("g3")
		if err != nil {
			t.Error(err)
			return
		}
		g4, err := groups.InsertByName3(context.Background(), "g4")
		if err != nil {
			t.Error(err)
			return
		}

		gv0, err := groups.Get(g0)
		if err != nil {
			t.Error(err)
			return
		}
		if gv0.Name != group0.Name {
			t.Error("except", group0.Name, "got", gv0.Name)
		}

		gv2, err := groups.Get(g2)
		if err != nil {
			t.Error(err)
			return
		}
		if gv2.Name != "g2" {
			t.Error("except 'g2' got", gv2.Name)
		}

		gv3, err := groups.Get(g3)
		if err != nil {
			t.Error(err)
			return
		}
		if gv3.Name != "g3" {
			t.Error("except 'g3' got", gv3.Name)
		}

		gv4, err := groups.Get(g4)
		if err != nil {
			t.Error(err)
			return
		}
		if gv4.Name != "g4" {
			t.Error("except 'g4' got", gv4.Name)
		}

		greset1, err := groups.InsertByName1("greset")
		if err != nil {
			t.Error(err)
			return
		}

		greset2, err := groups.InsertByName1("greset")
		if err != nil {
			t.Error(err)
			return
		}

		if greset1 == greset2 {
			t.Error("except", greset1, "got", greset2)
		}

		gresetv2, err := groups.Get(greset2)
		if err != nil {
			t.Error(err)
			return
		}
		if gresetv2.Name != "greset" {
			t.Error("except 'greset' got", gresetv2.Name)
		}

	})
}

func TestNullXXX(t *testing.T) {
	tests.Run(t, func(_ testing.TB, factory *core.Session) {

		group1 := tests.UserGroup{
			Name: "g1",
		}

		ref := factory.Reference()
		users := tests.NewTestUsers(&ref)
		groups := tests.NewTestUserGroups(&ref, users)

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
		g3, err := groups.InsertByName2("g3")
		if err != nil {
			t.Error(err)
			return
		}
		g4, err := groups.InsertByName3(context.Background(), "g4")
		if err != nil {
			t.Error(err)
			return
		}

		count, err := groups.Count()
		if err != nil {
			t.Error(err)
			return
		}
		if count != 4 {
			t.Error("want 4 got", count)
			return
		}

		for _, id := range []int64{g1, g2, g3, g4} {
			list, err := groups.QueryBy(sql.NullInt64{Valid: true, Int64: id})
			if err != nil {
				t.Error(err)
				return
			}

			if len(list) != 1 {
				t.Error("want 1 got", len(list))
				return
			}
		}

		list, err := groups.QueryBy(sql.NullInt64{Valid: false})
		if err != nil {
			t.Error(err)
			return
		}

		if len(list) != 4 {
			t.Error("want 4 got", len(list))
			return
		}

	})
}

func TestReadOnly(t *testing.T) {
	tests.Run(t, func(_ testing.TB, factory *core.Session) {
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
			Name: "g2",
		}

		group3 := tests.UserGroup{
			Name: "g3",
		}

		ref := factory.Reference()
		users := tests.NewTestUsers(&ref)
		groups := tests.NewTestUserGroups(&ref, users)

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

		err = users.AddToGroup(u1, g1)
		if err != nil {
			t.Error(err)
			return
		}
		err = users.AddToGroup(u1, g2)
		if err != nil {
			t.Error(err)
			return
		}
		err = users.AddToGroup(u2, g2)
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
			t.Error("except 2 got", len(gv2.UserIDs))
			t.Error("except [", u1, ",", u2, "] got", gv2.UserIDs)
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

func TestQueryWithUserQuery(t *testing.T) {
	tests.Run(t, func(_ testing.TB, factory *core.Session) {
		mac, _ := net.ParseMAC("01:02:03:04:A5:A6")
		ip := net.ParseIP("192.168.1.1")
		name := "张三"
		insertUser1 := tests.User{
			Name:        name,
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
			Name:        name + "sss",
			Nickname:    "hahaaa",
			Password:    "password",
			Description: "地球人sss",
			Address:     "沪南路11675号",
			HostIP:      ip,
			HostMAC:     mac,
			HostIPPtr:   &ip,
			HostMACPtr:  &mac,
			Sex:         "女",
			ContactInfo: map[string]interface{}{"QQ": "777777"},
			Birth:       time.Now(),
			CreateTime:  time.Now(),
		}
		users := tests.NewTestUsers(factory.SessionReference())

		_, err := users.Insert(&insertUser1)
		if err != nil {
			t.Error(err)
			return
		}
		_, err = users.Insert(&insertUser2)
		if err != nil {
			t.Error(err)
			return
		}

		assertList := func(t testing.TB, list []tests.User) {
			if len(list) != 1 {
				t.Error("want 1 got", len(list))
				return
			}

			if list[0].Name != name {
				t.Error("want '"+name+"' got", list[0].Name)
				return
			}
		}

		list, err := users.QueryWithUserQuery1(tests.UserQuery{UseUsername: true, Username: name})
		if err != nil {
			t.Error(err)
			return
		}
		assertList(t, list)

		list, err = users.QueryWithUserQuery2(tests.UserQuery{UseUsername: true, Username: name})
		if err != nil {
			t.Error(err)
			return
		}
		assertList(t, list)

		list, err = users.QueryWithUserQuery3(tests.UserQuery{UseUsername: true, Username: name})
		if err != nil {
			t.Error(err)
			return
		}
		assertList(t, list)

		list, err = users.QueryWithUserQuery4(tests.UserQuery{UseUsername: true, Username: name})
		if err != nil {
			t.Error(err)
			return
		}
		assertList(t, list)

		assertList = func(t testing.TB, list []tests.User) {
			if len(list) != 2 {
				t.Error("want 2 got", len(list))
				return
			}

			if list[0].Name != insertUser1.Name {
				t.Error("want '"+insertUser1.Name+"' got", list[0].Name)
				return
			}

			if list[1].Name != insertUser2.Name {
				t.Error("want '"+insertUser2.Name+"' got", list[1].Name)
				return
			}
		}

		list, err = users.QueryWithUserQuery1(tests.UserQuery{UseUsername: false, Username: name})
		if err != nil {
			t.Error(err)
			return
		}
		assertList(t, list)

		list, err = users.QueryWithUserQuery2(tests.UserQuery{UseUsername: false, Username: name})
		if err != nil {
			t.Error(err)
			return
		}
		assertList(t, list)

		list, err = users.QueryWithUserQuery3(tests.UserQuery{UseUsername: false, Username: name})
		if err != nil {
			t.Error(err)
			return
		}
		assertList(t, list)

		list, err = users.QueryWithUserQuery4(tests.UserQuery{UseUsername: false, Username: name})
		if err != nil {
			t.Error(err)
			return
		}
		assertList(t, list)
	})
}

func TestHandleError(t *testing.T) {
	tests.Run(t, func(_ testing.TB, factory *core.Session) {
		if factory.Dialect() != dialects.Postgres {
			t.Skip("db isnot Postgres")
		}
		group1 := tests.UserGroup{
			Name: "g1",
		}

		group2 := tests.UserGroup{
			Name: "g1",
		}

		ref := factory.Reference()
		users := tests.NewTestUsers(&ref)
		groups := tests.NewTestUserGroups(&ref, users)

		_, err := groups.Insert(&group1)
		if err != nil {
			t.Error(err)
			return
		}

		_, err = groups.Insert(&group2)
		if err == nil {
			t.Error("want err is exist got nil")
			return
		}

		e, ok := err.(*core.Error)
		if !ok {
			t.Error("error isnot excepted")
			return
		}

		if len(e.Validations) == 0 {
			t.Error("Validations is empty")
			return
		}
		t.Log(e.Validations)

		if len(e.Validations[0].Columns) == 0 {
			t.Error("Validations.Columns is empty")
			return
		}

		if e.Validations[0].Columns[0] != "name" {
			t.Error("Validations.Columns is empty")
		}

		group2.Name = strings.Repeat("aaa", 50)
		_, err = groups.Insert(&group2)
		if err == nil {
			t.Error("want err is exist got nil")
			return
		}

		e, ok = err.(*core.Error)
		if !ok {
			t.Error("error isnot excepted")
			return
		}

		if len(e.Validations) == 0 {
			t.Error("Validations is empty")
			return
		}
		t.Log(e.Validations)

		// for test cover
		core.ErrForGenerateStmt(e, "aa")
	})
}
