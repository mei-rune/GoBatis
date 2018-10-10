package gobatis_test

import (
	"net"
	"reflect"
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

			_, _, err = users.QueryFieldNotExist3()
			if err == nil {
				t.Error("except error got ok")
				return
			}

			if !strings.Contains(err.Error(), "isnot found in the") {
				t.Error("excepted is", "isnot found in the")
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

func TestComputer(t *testing.T) {
	tests.Run(t, func(_ testing.TB, factory *gobatis.SessionFactory) {
		ref := factory.Reference()
		computers := tests.NewComputers(&ref)

		k1, err := computers.InsertKeyboard(&tests.Keyboard{Description: "k1"})
		if err != nil {
			t.Error(err)
			return
		}

		k2, err := computers.InsertKeyboard(&tests.Keyboard{Description: "k2"})
		if err != nil {
			t.Error(err)
			return
		}

		m1, err := computers.InsertMotherboard(&tests.Motherboard{Description: "m1"})
		if err != nil {
			t.Error(err)
			return
		}

		m3, err := computers.InsertMotherboard(&tests.Motherboard{Description: "m3"})
		if err != nil {
			t.Error(err)
			return
		}

		c1, err := computers.InsertComputer(&tests.Computer{Description: "c1", KeyID: k1, MotherID: m1})
		if err != nil {
			t.Error(err)
			return
		}

		c2, err := computers.InsertComputer(&tests.Computer{Description: "c2", KeyID: k2})
		if err != nil {
			t.Error(err)
			return
		}

		c3, err := computers.InsertComputer(&tests.Computer{Description: "c3", MotherID: m3})
		if err != nil {
			t.Error(err)
			return
		}

		assertComputer := func(t *testing.T, c *tests.Computer, descr string, id, keyID, motherID int64) {
			t.Helper()
			if c.Description != descr {
				t.Error("computer.description: want", descr, "got", c.Description)
			}
			if c.ID != id {
				t.Error("computer.ID: want", id, "got", c.ID)
			}
			if c.KeyID != keyID {
				t.Error("computer.keyID: want", keyID, "got", c.KeyID)
			}
			if c.MotherID != motherID {
				t.Error("computer.motherID: want", motherID, "got", c.MotherID)
			}
		}

		assertKeyboard := func(t *testing.T, k *tests.Keyboard, descr string, keyID int64) {
			t.Helper()
			if k == nil {
				t.Error("Keyboard: value is nil")
				return
			}

			if k.Description != descr {
				t.Error("Keyboard.description: want", descr, "got", k.Description)
			}

			if k.ID != keyID {
				t.Error("Keyboard.ID: want", keyID, "got", k.ID)
			}
		}

		assertMotherboard := func(t *testing.T, m *tests.Motherboard, descr string, motherID int64) {
			t.Helper()
			if m == nil {
				t.Error("Motherboard: value is nil")
				return
			}
			if m.Description != descr {
				t.Error("Motherboard.description: want", descr, "got", m.Description)
			}

			if m.ID != motherID {
				t.Error("Motherboard.ID: want", motherID, "got", m.ID)
			}
		}

		assert := func(t *testing.T, got, want interface{}, msg string) {
			t.Helper()
			if !reflect.DeepEqual(got, want) {
				t.Error(msg, ": want", want, "got", got)
			}
		}

		t.Run("computer and keyboard is all exists", func(t *testing.T) {
			rc1, rk1, err := computers.FindByID1(c1)
			if err != nil {
				t.Error(err)
				return
			}

			assertComputer(t, rc1, "c1", c1, k1, m1)
			assertKeyboard(t, rk1, "k1", k1)
		})

		t.Run("computer is exist and keyboard isnot exist", func(t *testing.T) {
			rc3, rk3, err := computers.FindByID1(c3)
			if err != nil {
				t.Error(err)
				return
			}

			assertComputer(t, rc3, "c3", c3, 0, m3)
			if rk3 != nil {
				t.Error("want rk3 is nil got", rk3)
			}
		})

		t.Run("computer, motherboard and keyboard is all exists", func(t *testing.T) {
			rc1, rk1, rm1, err := computers.FindByID3(c1)
			if err != nil {
				t.Error(err)
				return
			}

			assertComputer(t, rc1, "c1", c1, k1, m1)
			assertKeyboard(t, rk1, "k1", k1)
			assertMotherboard(t, rm1, "m1", m1)
		})

		t.Run("computer and keyboard is all exists, but motherboard isnot exist", func(t *testing.T) {
			rc2, rk2, rm2, err := computers.FindByID3(c2)
			if err != nil {
				t.Error(err)
				return
			}

			assertComputer(t, rc2, "c2", c2, k2, 0)
			assertKeyboard(t, rk2, "k2", k2)
			if rm2 != nil {
				t.Error("want rm2 is nil got", rm2)
			}
		})

		t.Run("computer and motherboard is all exists, but keyboard isnot exist", func(t *testing.T) {
			// FIXME: 临时先跳过
			t.Skip("临时先跳过")

			rc3, rk3, rm3, err := computers.FindByID3(c3)
			if err != nil {
				t.Error(err)
				return
			}

			assertComputer(t, rc3, "c3", c3, 0, m3)
			if rk3 != nil {
				t.Error("want rk3 is nil got", rk3)
			}
			assertMotherboard(t, rm3, "m3", m3)
		})

		t.Run("QueryAll2", func(t *testing.T) {
			rcs, rks, rms, err := computers.QueryAll2()
			if err != nil {
				t.Error(err)
				return
			}

			if len(rcs) != 3 {
				t.Error("want len(rcs) is 3 got", len(rcs))
				return
			}

			assertComputer(t, &rcs[0], "c1", c1, k1, m1)
			assertKeyboard(t, rks[0], "k1", k1)
			assertMotherboard(t, rms[0], "m1", m1)

			assertComputer(t, &rcs[1], "c2", c2, k2, 0)
			assertKeyboard(t, rks[1], "k2", k2)
			if rms[1] != nil {
				t.Error("want rm2 is nil got", rms[1])
			}

			assertComputer(t, &rcs[2], "c3", c3, 0, m3)
			if rks[2] != nil {
				t.Error("want rk3 is nil got", rks[2])
			}
			assertMotherboard(t, rms[2], "m3", m3)
		})

		t.Run("QueryAll3", func(t *testing.T) {
			rcs, rk_ids, rk_descrs, rm_ids, rm_descrs, err := computers.QueryAll3()
			if err != nil {
				t.Error(err)
				return
			}

			if len(rcs) != 3 {
				t.Error("want len(rcs) is 3 got", len(rcs))
				return
			}

			assertComputer(t, &rcs[0], "c1", c1, k1, m1)
			assert(t, rk_ids[0], k1, "Keyboard.ID")
			assert(t, rk_descrs[0], "k1", "Keyboard.Description")
			assert(t, rm_ids[0], m1, "Motherboard.ID")
			assert(t, rm_descrs[0], "m1", "Motherboard.Description")

			assertComputer(t, &rcs[1], "c2", c2, k2, 0)
			assert(t, rk_ids[1], k2, "Keyboard.ID")
			assert(t, rk_descrs[1], "k2", "Keyboard.Description")
			assert(t, rm_ids[1], int64(0), "Motherboard.ID")
			assert(t, rm_descrs[1], "", "Motherboard.Description")

			assertComputer(t, &rcs[2], "c3", c3, 0, m3)
			assert(t, rk_ids[2], int64(0), "Keyboard.ID")
			assert(t, rk_descrs[2], "", "Keyboard.Description")
			assert(t, rm_ids[2], m3, "Motherboard.ID")
			assert(t, rm_descrs[2], "m3", "Motherboard.Description")
		})

		t.Run("QueryAllFail1", func(t *testing.T) {
			_, _, a, _, _, err := computers.QueryAllFail1()
			if err == nil {
				t.Error(a)
				t.Error("want error got ok")
				return
			}
			if !strings.Contains(err.Error(), "Scan error on column index 5") {
				t.Error(err)
			}
		})
	})
}
