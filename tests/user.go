//go:generate gobatis user.go
package tests

import (
	"bytes"
	"math"
	"net"
	"reflect"
	"testing"
	"time"
)

type User struct {
	TableName   struct{}               `db:"gobatis_users"`
	ID          int64                  `db:"id,pk,autoincr"`
	Name        string                 `db:"name"`
	Nickname    string                 `db:"nickname"`
	Password    string                 `db:"password"`
	Description string                 `db:"description"`
	Birth       time.Time              `db:"birth"`
	Address     string                 `db:"address"`
	HostIP      net.IP                 `db:"host_ip"`
	HostMAC     net.HardwareAddr       `db:"host_mac"`
	HostIPPtr   *net.IP                `db:"host_ip_ptr"`
	HostMACPtr  *net.HardwareAddr      `db:"host_mac_ptr"`
	Sex         string                 `db:"sex"`
	ContactInfo map[string]interface{} `db:"contact_info"`
	Field1      int                    `db:"field1,null"`
	Field2      uint                   `db:"field2,null"`
	Field3      float64                `db:"field3,null"`
	Field4      float64                `db:"field4"`
	Field5      string                 `db:"field5,null"`
	Field6      time.Time              `db:"field6,null"`
	CreateTime  time.Time              `db:"create_time"`
}

type TestUsers interface {
	Insert(u *User) (int64, error)

	Update(id int64, u *User) (int64, error)

	DeleteAll() (int64, error)

	Delete(id int64) (int64, error)

	Get(id int64) (*User, error)

	Count() (int64, error)

	// @default SELECT * FROM gobatis_users {{if isNotEmpty .idList}} WHERE id in ({{range $i, $v :=  .idList }} {{$v}} {{if isLast $.idList $i | not }} , {{end}}{{end}}){{end}}
	Query(idList []int64) ([]User, error)

	// @default SELECT id as u_id, name, name as p_name FROM gobatis_users
	QueryFieldNotExist() (u []User, name []string, err error)

	// @option default_return_name u
	// @default SELECT id as u_id, name as name_name, name as name FROM gobatis_users
	QueryReturnDupError1() (u []User, name []string, err error)

	// @option default_return_name u
	// @default SELECT id as u_id, name as name, name as name_name FROM gobatis_users
	QueryReturnDupError2() (u []User, name []string, err error)
}

func AssertUser(t testing.TB, excepted, actual User) {
	if helper, ok := t.(interface {
		Helper()
	}); ok {
		helper.Helper()
	}
	if excepted.ID != actual.ID {
		t.Error("[ID] excepted is", excepted.ID)
		t.Error("[ID] actual   is", actual.ID)
	}
	if excepted.Name != actual.Name {
		t.Error("[Name] excepted is", excepted.Name)
		t.Error("[Name] actual   is", actual.Name)
	}
	if excepted.Nickname != actual.Nickname {
		t.Error("[Nickname] excepted is", excepted.Nickname)
		t.Error("[Nickname] actual   is", actual.Nickname)
	}
	if excepted.Password != actual.Password {
		t.Error("[Password] excepted is", excepted.Password)
		t.Error("[Password] actual   is", actual.Password)
	}
	if excepted.Description != actual.Description {
		t.Error("[Description] excepted is", excepted.Description)
		t.Error("[Description] actual   is", actual.Description)
	}
	if excepted.Address != actual.Address {
		t.Error("[Address] excepted is", excepted.Address)
		t.Error("[Address] actual   is", actual.Address)
	}

	if len(excepted.HostIP) != 0 || len(actual.HostIP) != 0 {
		if !bytes.Equal(excepted.HostIP, actual.HostIP) {
			t.Error("[HostIP] excepted is", excepted.HostIP)
			t.Error("[HostIP] actual   is", actual.HostIP)
		}
	}
	if len(excepted.HostMAC) != 0 || len(actual.HostMAC) != 0 {
		if !bytes.Equal(excepted.HostMAC, actual.HostMAC) {
			t.Error("[HostMAC] excepted is", excepted.HostMAC)
			t.Error("[HostMAC] actual   is", actual.HostMAC)
		}
	}

	if (excepted.HostIPPtr != nil && len(*excepted.HostIPPtr) != 0) ||
		(actual.HostIPPtr != nil && len(*actual.HostIPPtr) != 0) {
		if !reflect.DeepEqual(excepted.HostIPPtr, actual.HostIPPtr) {
			t.Error("[HostIPPtr] excepted is", excepted.HostIPPtr)
			t.Error("[HostIPPtr] actual   is", actual.HostIPPtr)
		}
	}

	if (excepted.HostMACPtr != nil && len(*excepted.HostMACPtr) != 0) ||
		(actual.HostMACPtr != nil && len(*actual.HostMACPtr) != 0) {
		if !reflect.DeepEqual(excepted.HostMACPtr, actual.HostMACPtr) {
			t.Error("[HostMACPtr] excepted is", excepted.HostMACPtr)
			t.Error("[HostMACPtr] actual   is", actual.HostMACPtr)
		}
	}
	if excepted.Sex != actual.Sex {
		t.Error("[Sex] excepted is", excepted.Sex)
		t.Error("[Sex] actual   is", actual.Sex)
	}
	if len(excepted.ContactInfo) != len(actual.ContactInfo) ||
		!reflect.DeepEqual(excepted.ContactInfo, actual.ContactInfo) {
		t.Error("[ContactInfo] excepted is", excepted.ContactInfo)
		t.Error("[ContactInfo] actual   is", actual.ContactInfo)
	}

	if excepted.Birth.Format("2006-01-02") != actual.Birth.Format("2006-01-02") {
		t.Error("[Birth] excepted is", excepted.Birth.Format("2006-01-02"))
		t.Error("[Birth] actual   is", actual.Birth.Format("2006-01-02"))
	}
	if math.Abs(excepted.CreateTime.Sub(actual.CreateTime).Seconds()) > 2 {
		t.Error("[CreateTime] excepted is", excepted.CreateTime.Format(time.RFC1123))
		t.Error("[CreateTime] actual   is", actual.CreateTime.Format(time.RFC1123))
	}
}
