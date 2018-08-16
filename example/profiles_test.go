package example

import (
	"testing"

	gobatis "github.com/runner-mei/GoBatis"
	"github.com/runner-mei/GoBatis/tests"
)

func TestUserProfiles(t *testing.T) {
	insertUser1 := AuthUser{
		Username: "abc",
		Phone:    "123",
		Status:   1,
	}

	insertUser2 := AuthUser{
		Username: "xyz",
		Phone:    "123",
		Status:   1,
	}

	tests.Run(t, func(_ testing.TB, factory *gobatis.SessionFactory) {
		var err error
		switch factory.DbType() {
		case gobatis.DbTypePostgres:
			_, err = factory.DB().Exec(postgres)
		case gobatis.DbTypeMSSql:
			_, err = factory.DB().Exec(mssql)
		default:
			_, err = factory.DB().Exec(mysql)
		}
		if err != nil {
			t.Error(err)
			return
		}

		ref := factory.Reference()
		users := NewUsers(&ref)
		profiles := NewUserProfiles(&ref)
		_, err = users.DeleteAll()
		if err != nil {
			t.Error(err)
			return
		}

		_, err = profiles.DeleteAll()
		if err != nil {
			t.Error(err)
			return
		}

		uid1, err := users.Insert(&insertUser1)
		if err != nil {
			t.Error(err)
			return
		}

		uid2, err := users.Insert(&insertUser2)
		if err != nil {
			t.Error(err)
			return
		}

		pid1, err := profiles.Insert(&UserProfile{
			UserID: uid1,
			Key:    "k1",
			Value:  "v1",
		})
		if err != nil {
			t.Error(err)
			return
		}

		pid2, err := profiles.Insert(&UserProfile{
			UserID: uid2,
			Key:    "k2",
			Value:  "v2",
		})
		if err != nil {
			t.Error(err)
			return
		}

		t.Run("FindByID1", func(t *testing.T) {

			p, u, err := profiles.FindByID1(pid1)
			if err != nil {
				t.Error(err)
				return
			}

			if p.ID != pid1 || p.UserID != uid1 || p.Key != "k1" || p.Value != "v1" {
				t.Error("id: except", pid1, "got", p.ID)
				t.Error("userid: except", uid1, "got", p.UserID)
				t.Error("key: except", "k1", "got", p.Key)
				t.Error("value: except", "v1", "got", p.Value)
				return
			}

			if u.ID != uid1 || u.Username != insertUser1.Username {
				t.Error("id: except", uid1, "got", u.ID)
				t.Error("username: except", insertUser1.Username, "got", u.Username)
				return
			}
		})

		t.Run("FindByID2", func(t *testing.T) {
			p2, u2, err := profiles.FindByID2(pid2)
			if err != nil {
				t.Error(err)
				return
			}
			p := &p2
			u := &u2

			if p.ID != pid2 || p.UserID != uid2 || p.Key != "k2" || p.Value != "v2" {
				t.Error("id: except", pid2, "got", p.ID)
				t.Error("userid: except", uid2, "got", p.UserID)
				t.Error("key: except", "k2", "got", p.Key)
				t.Error("value: except", "v2", "got", p.Value)
				return
			}

			if u.ID != uid2 || u.Username != insertUser2.Username {
				t.Error("id: except", uid2, "got", u.ID)
				t.Error("username: except", insertUser2.Username, "got", u.Username)
				return
			}

		})

		t.Run("FindByID3", func(t *testing.T) {
			p3, userid, username, err := profiles.FindByID3(pid2)
			if err != nil {
				t.Error(err)
				return
			}
			p := &p3
			u := &AuthUser{ID: userid, Username: username}

			if p.ID != pid2 || p.UserID != uid2 || p.Key != "k2" || p.Value != "v2" {
				t.Error("id: except", pid2, "got", p.ID)
				t.Error("userid: except", uid2, "got", p.UserID)
				t.Error("key: except", "k2", "got", p.Key)
				t.Error("value: except", "v2", "got", p.Value)
				return
			}

			if u.ID != uid2 || u.Username != insertUser2.Username {
				t.Error("id: except", uid2, "got", u.ID)
				t.Error("username: except", insertUser2.Username, "got", u.Username)
				return
			}
		})

		t.Run("FindByID4", func(t *testing.T) {
			p4, userid2, username2, err := profiles.FindByID4(pid2)
			if err != nil {
				t.Error(err)
				return
			}
			p := p4
			u := &AuthUser{ID: *userid2, Username: *username2}

			if p.ID != pid2 || p.UserID != uid2 || p.Key != "k2" || p.Value != "v2" {
				t.Error("id: except", pid2, "got", p.ID)
				t.Error("userid: except", uid2, "got", p.UserID)
				t.Error("key: except", "k2", "got", p.Key)
				t.Error("value: except", "v2", "got", p.Value)
				return
			}

			if u.ID != uid2 || u.Username != insertUser2.Username {
				t.Error("id: except", uid2, "got", u.ID)
				t.Error("username: except", insertUser2.Username, "got", u.Username)
				return
			}

		})

		t.Run("ListByUserID1", func(t *testing.T) {
			plist, ulist, err := profiles.ListByUserID1(uid1)
			if err != nil {
				t.Error(err)
				return
			}

			if len(plist) != 1 {
				t.Error("plist: except", 1, "got", len(plist))
				return
			}

			if len(ulist) != 1 {
				t.Error("ulist: except", 1, "got", len(ulist))
				return
			}

			p := plist[0]
			u := ulist[0]

			if p.ID != pid1 || p.UserID != uid1 || p.Key != "k1" || p.Value != "v1" {
				t.Error("id: except", pid1, "got", p.ID)
				t.Error("userid: except", uid1, "got", p.UserID)
				t.Error("key: except", "k1", "got", p.Key)
				t.Error("value: except", "v1", "got", p.Value)
				return
			}

			if u.ID != uid1 || u.Username != insertUser1.Username {
				t.Error("id: except", uid1, "got", u.ID)
				t.Error("username: except", insertUser1.Username, "got", u.Username)
				return
			}
		})

		t.Run("ListByUserID2", func(t *testing.T) {
			plist, ulist, err := profiles.ListByUserID2(uid2)
			if err != nil {
				t.Error(err)
				return
			}

			if len(plist) != 1 {
				t.Error("plist: except", 1, "got", len(plist))
				return
			}

			if len(ulist) != 1 {
				t.Error("ulist: except", 1, "got", len(ulist))
				return
			}

			p := &plist[0]
			u := &ulist[0]

			if p.ID != pid2 || p.UserID != uid2 || p.Key != "k2" || p.Value != "v2" {
				t.Error("id: except", pid2, "got", p.ID)
				t.Error("userid: except", uid2, "got", p.UserID)
				t.Error("key: except", "k2", "got", p.Key)
				t.Error("value: except", "v2", "got", p.Value)
				return
			}

			if u.ID != uid2 || u.Username != insertUser2.Username {
				t.Error("id: except", uid2, "got", u.ID)
				t.Error("username: except", insertUser2.Username, "got", u.Username)
				return
			}
		})

		t.Run("ListByUserID3", func(t *testing.T) {
			plist, userids, usernames, err := profiles.ListByUserID3(uid2)
			if err != nil {
				t.Error(err)
				return
			}

			if len(plist) != 1 {
				t.Error("plist: except", 1, "got", len(plist))
				return
			}

			if len(userids) != 1 {
				t.Error("userids3: except", 1, "got", len(userids))
				return
			}

			if len(usernames) != 1 {
				t.Error("usernames3: except", 1, "got", len(usernames))
				return
			}

			p := &plist[0]
			u := &AuthUser{ID: userids[0], Username: usernames[0]}

			if p.ID != pid2 || p.UserID != uid2 || p.Key != "k2" || p.Value != "v2" {
				t.Error("id: except", pid2, "got", p.ID)
				t.Error("userid: except", uid2, "got", p.UserID)
				t.Error("key: except", "k2", "got", p.Key)
				t.Error("value: except", "v2", "got", p.Value)
				return
			}

			if u.ID != uid2 || u.Username != insertUser2.Username {
				t.Error("id: except", uid2, "got", u.ID)
				t.Error("username: except", insertUser2.Username, "got", u.Username)
				return
			}
		})

		t.Run("ListByUserID4", func(t *testing.T) {
			plist, userids, usernames, err := profiles.ListByUserID4(uid2)
			if err != nil {
				t.Error(err)
				return
			}

			if len(plist) != 1 {
				t.Error("plist: except", 1, "got", len(plist))
				return
			}

			if len(userids) != 1 {
				t.Error("userids: except", 1, "got", len(userids))
				return
			}

			if len(usernames) != 1 {
				t.Error("usernames: except", 1, "got", len(usernames))
				return
			}

			p := plist[0]
			u := &AuthUser{ID: *userids[0], Username: *usernames[0]}

			if p.ID != pid2 || p.UserID != uid2 || p.Key != "k2" || p.Value != "v2" {
				t.Error("id: except", pid2, "got", p.ID)
				t.Error("userid: except", uid2, "got", p.UserID)
				t.Error("key: except", "k2", "got", p.Key)
				t.Error("value: except", "v2", "got", p.Value)
				return
			}

			if u.ID != uid2 || u.Username != insertUser2.Username {
				t.Error("id: except", uid2, "got", u.ID)
				t.Error("username: except", insertUser2.Username, "got", u.Username)
				return
			}
		})
	})
}
