//go:generate gobatis user.go
package tests

import (
	"bytes"
	"context"
	"math"
	"net"
	"reflect"
	"testing"
	"time"
)

type UserGroup struct {
	TableName struct{} `db:"gobatis_usergroups"`
	ID        int64    `db:"id,pk,autoincr"`
	Name      string   `db:"name"`
	UserIDs   []int64  `db:"user_ids,<-"`
}

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
	Field3      float32                `db:"field3,null"`
	Field4      float64                `db:"field4"`
	Field5      string                 `db:"field5,null"`
	Field6      time.Time              `db:"field6,null"`
	Field7      *time.Time             `db:"field7"`
	FieldBool   bool                   `db:"fieldbool,null"`
	FieldBoolP  *bool                  `db:"fieldboolp"`
	CreateTime  time.Time              `db:"create_time"`
	GroupIDs    []int64                `db:"group_ids,<-"`
}

type TestUsers interface {
	// @mysql INSERT INTO gobatis_users(name, nickname, password, description, birth, address, host_ip, host_mac, host_ip_ptr, host_mac_ptr, sex, contact_info, field1, field2, field3, field4, field5, field6, field7, fieldBool, fieldBoolP, create_time)
	// VALUES(#{name}, #{nickname}, #{password}, #{description}, #{birth}, #{address}, #{host_ip}, #{host_mac}, #{host_ip_ptr}, #{host_mac_ptr}, #{sex}, #{contact_info}, #{field1}, #{field2}, #{field3}, #{field4}, #{field5}, #{field6}, #{field7}, #{fieldBool}, #{fieldBoolP}, #{create_time})
	//
	// @mssql INSERT INTO gobatis_users(name, nickname, password, description, birth, address, host_ip, host_mac, host_ip_ptr, host_mac_ptr, sex, contact_info, field1, field2, field3, field4, field5, field6, field7, fieldBool, fieldBoolP, create_time) OUTPUT inserted.id
	// VALUES(#{name}, #{nickname}, #{password}, #{description}, #{birth}, #{address}, #{host_ip}, #{host_mac}, #{host_ip_ptr}, #{host_mac_ptr}, #{sex}, #{contact_info}, #{field1}, #{field2}, #{field3}, #{field4}, #{field5}, #{field6}, #{field7}, #{fieldBool}, #{fieldBoolP}, #{create_time})
	//
	// @default INSERT INTO gobatis_users(name, nickname, password, description, birth, address, host_ip, host_mac, host_ip_ptr, host_mac_ptr, sex, contact_info, field1, field2, field3, field4, field5, field6, field7, fieldBool, fieldBoolP, create_time)
	// VALUES(#{name}, #{nickname}, #{password}, #{description}, #{birth}, #{address}, #{host_ip}, #{host_mac}, #{host_ip_ptr}, #{host_mac_ptr}, #{sex}, #{contact_info}, #{field1}, #{field2}, #{field3}, #{field4}, #{field5}, #{field6}, #{field7}, #{fieldBool}, #{fieldBoolP}, #{create_time}) RETURNING id
	InsertByArgs(name, nickname, password, description string, birth time.Time, address string,
		host_ip net.IP, host_mac net.HardwareAddr, host_ip_ptr *net.IP, host_mac_ptr *net.HardwareAddr,
		sex string, contact_info map[string]interface{}, field1 int, field2 uint, field3 float32,
		field4 float64, field5 string, field6 time.Time, field7 *time.Time,
		fieldBool bool, fieldBoolP *bool, create_time time.Time) (int64, error)

	Insert(u *User) (int64, error)

	InsertContext(ctx context.Context, u *User) (int64, error)

	Update(id int64, u *User) (int64, error)

	UpdateContext(ctx context.Context, id int64, u *User) (int64, error)

	DeleteAll() (int64, error)

	Delete(id int64) (int64, error)

	DeleteContext(ctx context.Context, id int64) (int64, error)

	Get(id int64) (*User, error)

	GetByID(id int64) (User, error)

	GetContext(ctx context.Context, id int64) (*User, error)

	Count() (int64, error)

	CountContext(ctx context.Context) (int64, error)

	GetAllContext(ctx context.Context) ([]User, error)

	// @default SELECT * FROM gobatis_users {{if isNotEmpty .idList}} WHERE id in ({{range $i, $v :=  .idList }} {{$v}} {{if isLast $.idList $i | not }} , {{end}}{{end}}){{end}}
	Query(idList []int64) ([]User, error)

	// @default SELECT id as u_id, name, name as p_name FROM gobatis_users
	QueryFieldNotExist1() (u []User, name []string, err error)

	// @default SELECT id as u, name, name as pname FROM gobatis_users
	QueryFieldNotExist2() (u []User, name []string, err error)

	// @default SELECT id as u_id, name as u_user, name FROM gobatis_users
	QueryFieldNotExist3() (u []User, name []string, err error)

	// @option default_return_name u
	// @default SELECT id as u_id, name as name_name, name as name FROM gobatis_users
	QueryReturnDupError1() (u []User, name []string, err error)

	// @option default_return_name u
	// @default SELECT id as u_id, name as name, name as name_name FROM gobatis_users
	QueryReturnDupError2() (u []User, name []string, err error)

	// @default INSERT INTO gobatis_user_and_groups(user_id,group_id) values(#{userID}, #{groupID})
	AddToGroup(userID, groupID int64) error

	// @default SELECT * from gobatis_user_and_groups
	// <foreach collection="idList" open="WHERE id  in (" separator="," close=")"> #{item} </foreach>
	QueryByGroups(idList ...int64) ([]User, error)
}

type TestUserGroups interface {
	// @mysql INSERT INTO gobatis_usergroups(name) VALUES(#{name})
	// @mssql INSERT INTO gobatis_usergroups(name) OUTPUT inserted.id VALUES(#{name})
	// @default INSERT INTO gobatis_usergroups(name) VALUES(#{name}) RETURNING id
	InsertByName(name string) (int64, error)

	Insert(u *UserGroup) (int64, error)

	Update(id int64, u *UserGroup) (int64, error)

	DeleteAll() (int64, error)

	Delete(id int64) (int64, error)

	// @type select
	// @default SELECT max(id) FROM gobatis_usergroups
	MaxID() (int64, error)

	// @default SELECT groups.id, groups.name, array_to_json(array_agg(u2g.user_id)) as user_ids
	//          FROM gobatis_usergroups as groups LEFT JOIN gobatis_user_and_groups as u2g
	//               ON groups.id = u2g.group_id
	//          WHERE groups.id = #{id}
	//          GROUP BY groups.id
	//
	// @mysql SELECT groups.id, groups.name, CONCAT('[', GROUP_CONCAT(DISTINCT u2g.user_id SEPARATOR ','), ']') as user_ids
	//          FROM gobatis_usergroups as groups LEFT JOIN gobatis_user_and_groups as u2g
	//               ON groups.id = u2g.group_id
	//          WHERE groups.id = #{id}
	//          GROUP BY groups.id
	//         -- see JSON_OBJECTAGG and JSON_ARRAYAGG
	//
	// @mssql SELECT groups.id, groups.name, CONCAT('[', STRING_AGG(CAST(u2g.user_id AS varchar(100)), ','), ']') as user_ids
	//          FROM gobatis_usergroups as groups LEFT JOIN gobatis_user_and_groups as u2g
	//               ON groups.id = u2g.group_id
	//          WHERE groups.id = #{id}
	//          GROUP BY groups.id
	//         -- see CROSS APPLY
	Get(id int64) (*UserGroup, error)

	Count() (int64, error)

	// @default INSERT INTO gobatis_user_and_groups(user_id,group_id) values(#{userID}, #{groupID})
	AddUser(groupID, userID int64) error

	// @reference TestUsers.QueryByGroups
	UsersByGroupID(id ...int64) ([]User, error)
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
