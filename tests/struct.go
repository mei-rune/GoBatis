//go:generate gobatis struct.go
package tests

import (
	"database/sql/driver"
	"encoding/json"
	"net"
	"time"
	"unsafe"

	"github.com/runner-mei/GoBatis/dialects"
)

type TestA1 struct {
	TableName struct{}         `db:"gobatis_testa"`
	ID        int64            `db:"id,pk,autoincr"`
	Field0    bool             `db:"field0,null"`
	Field1    int              `db:"field1,null"`
	Field2    uint             `db:"field2,null"`
	Field3    float32          `db:"field3,null"`
	Field4    float64          `db:"field4,null"`
	Field5    string           `db:"field5,null"`
	Field6    time.Time        `db:"field6,null"`
	Field7    net.IP           `db:"field7,null"`
	Field8    net.HardwareAddr `db:"field8,null"`
	Field9    string           `db:"field9,null"`
}

type TestA2 struct {
	TableName struct{}          `db:"gobatis_testa"`
	ID        int64             `db:"id,pk,autoincr"`
	Field0    *bool             `db:"field0,null"`
	Field1    *int              `db:"field1,null"`
	Field2    *uint             `db:"field2,null"`
	Field3    *float32          `db:"field3,null"`
	Field4    *float64          `db:"field4,null"`
	Field5    *string           `db:"field5,null"`
	Field6    *time.Time        `db:"field6,null"`
	Field7    *net.IP           `db:"field7,null"`
	Field8    *net.HardwareAddr `db:"field8,null"`
	Field9    *string           `db:"field9,null"`
}

type TestA3 struct {
	TableName struct{}         `db:"gobatis_testa"`
	ID        int64            `db:"id,pk,autoincr"`
	Field0    bool             `db:"field0"`
	Field1    int              `db:"field1"`
	Field2    uint             `db:"field2"`
	Field3    float32          `db:"field3"`
	Field4    float64          `db:"field4"`
	Field5    string           `db:"field5"`
	Field6    time.Time        `db:"field6"`
	Field7    net.IP           `db:"field7"`
	Field8    net.HardwareAddr `db:"field8"`
	Field9    string           `db:"field9"`
}

type TestA4 struct {
	TableName struct{}          `db:"gobatis_testa"`
	ID        int64             `db:"id,pk,autoincr"`
	Field0    *bool             `db:"field0"`
	Field1    *int              `db:"field1"`
	Field2    *uint             `db:"field2"`
	Field3    *float32          `db:"field3"`
	Field4    *float64          `db:"field4"`
	Field5    *string           `db:"field5"`
	Field6    *time.Time        `db:"field6"`
	Field7    *net.IP           `db:"field7"`
	Field8    *net.HardwareAddr `db:"field8"`
	Field9    *string           `db:"field9"`
}

type TestB1 struct {
	TableName struct{}         `db:"gobatis_testb"`
	ID        int64            `db:"id,pk,autoincr"`
	Field0    bool             `db:"field0,notnull"`
	Field1    int              `db:"field1,notnull"`
	Field2    uint             `db:"field2,notnull"`
	Field3    float32          `db:"field3,notnull"`
	Field4    float64          `db:"field4,notnull"`
	Field5    string           `db:"field5,notnull"`
	Field6    time.Time        `db:"field6,notnull"`
	Field7    net.IP           `db:"field7,notnull"`
	Field8    net.HardwareAddr `db:"field8,notnull"`
	Field9    string           `db:"field9,notnull"`
}

type TestB2 struct {
	TableName struct{}          `db:"gobatis_testb"`
	ID        int64             `db:"id,pk,autoincr"`
	Field0    *bool             `db:"field0,notnull"`
	Field1    *int              `db:"field1,notnull"`
	Field2    *uint             `db:"field2,notnull"`
	Field3    *float32          `db:"field3,notnull"`
	Field4    *float64          `db:"field4,notnull"`
	Field5    *string           `db:"field5,notnull"`
	Field6    *time.Time        `db:"field6,notnull"`
	Field7    *net.IP           `db:"field7,notnull"`
	Field8    *net.HardwareAddr `db:"field8,notnull"`
	Field9    *string           `db:"field9,notnull"`
}

var (
	TestMAC net.HardwareAddr
	TestIP  net.IP
)

func init() {
	TestMAC, _ = net.ParseMAC("01:02:03:04:A5:A6")
	TestIP = net.ParseIP("192.168.1.1")
}

type ITest interface {
	InsertA1(v *TestA1) (int64, error)
	InsertA2(v *TestA2) (int64, error)
	InsertA3(v *TestA3) (int64, error)
	InsertA4(v *TestA4) (int64, error)
	InsertB1(v *TestB1) (int64, error)
	InsertB2(v *TestB2) (int64, error)

	InsertFail1(v *Testfail1) (int64, error)
	InsertFail2(v *Testfail2) (int64, error)
	InsertFail3(v *Testfail3) (int64, error)

	InsertC1(v *TestC1) (int64, error)
	InsertC2(v *TestC2) (int64, error)
	InsertC3(v *TestC3) (int64, error)
	InsertC4(v *TestC4) (int64, error)
	InsertC5(v *TestC5) (int64, error)
	InsertC6(v *TestC6) (int64, error)

	InsertC7(v *TestC7) (int64, error)
	InsertC8(v *TestC8) (int64, error)
	InsertC9(v *TestC9) (int64, error)

	InsertD1(v *TestD1) (int64, error)
	InsertD2(v *TestD2) (int64, error)
	InsertD3(v *TestD3) (int64, error)
	InsertD4(v *TestD4) (int64, error)

	InsertTestE(v *TestE) (int64, error)
	// @record_type TestE
	InsertTestE_2(field0 []int64) (int64, error)
	GetTestE(id int64) (*TestE, error)
	UpdateTestE(id int64, v *TestE) (int64, error)

	InsertE1(v *TestE1) (int64, error)
	InsertE2(v *TestE2) (int64, error)
	InsertE3(v *TestE3) (int64, error)
	InsertE4(v *TestE4) (int64, error)
	InsertE5(v *TestE5) (int64, error)
	InsertE6(v *TestE6) (int64, error)

	InsertF1(v *TestF1) (int64, error)
	InsertF2(v *TestF2) (int64, error)
	InsertF3(v *TestF3) (int64, error)
	InsertF4(v *TestF4) (int64, error)
	InsertF5(v *TestF5) (int64, error)
	InsertF6(v *TestF6) (int64, error)
}

type Testfail1 struct {
	TableName struct{} `db:"gobatis_testfail"`
	ID        int64    `db:"id,pk,autoincr"`
	Field0    func()   `db:"field0,notnull"`
}

type Testfail2 struct {
	TableName struct{} `db:"gobatis_testfail"`
	ID        int64    `db:"id,pk,autoincr"`
	Field0    chan int `db:"field0,notnull"`
}

type Testfail3 struct {
	TableName struct{}       `db:"gobatis_testfail"`
	ID        int64          `db:"id,pk,autoincr"`
	Field0    unsafe.Pointer `db:"field0,notnull"`
}

type TestC1 struct {
	TableName struct{}               `db:"gobatis_testc"`
	ID        int64                  `db:"id,pk,autoincr"`
	Field0    map[string]interface{} `db:"field0,null"`
}

type TestC2 struct {
	TableName struct{}               `db:"gobatis_testc"`
	ID        int64                  `db:"id,pk,autoincr"`
	Field0    map[string]interface{} `db:"field0,notnull"`
}

type TestC3 struct {
	TableName struct{}               `db:"gobatis_testc"`
	ID        int64                  `db:"id,pk,autoincr"`
	Field0    map[string]interface{} `db:"field0"`
}

type Data struct {
	A int
}

type TestC4 struct {
	TableName struct{} `db:"gobatis_testc"`
	ID        int64    `db:"id,pk,autoincr"`
	Field0    *Data    `db:"field0,null"`
}

type TestC5 struct {
	TableName struct{} `db:"gobatis_testc"`
	ID        int64    `db:"id,pk,autoincr"`
	Field0    *Data    `db:"field0,notnull"`
}

type TestC6 struct {
	TableName struct{} `db:"gobatis_testc"`
	ID        int64    `db:"id,pk,autoincr"`
	Field0    *Data    `db:"field0"`
}

type TestC7 struct {
	TableName struct{} `db:"gobatis_testc"`
	ID        int64    `db:"id,pk,autoincr"`
	Field0    Data     `db:"field0,null"`
}

type TestC8 struct {
	TableName struct{} `db:"gobatis_testc"`
	ID        int64    `db:"id,pk,autoincr"`
	Field0    Data     `db:"field0,notnull"`
}

type TestC9 struct {
	TableName struct{} `db:"gobatis_testc"`
	ID        int64    `db:"id,pk,autoincr"`
	Field0    Data     `db:"field0"`
}

type DriverData1 struct {
	A int
}

func (a DriverData1) Value() (driver.Value, error) {
	bs, err := json.Marshal(a)
	if err != nil {
		return nil, err
	}
	if TestDrv == dialects.DM.Name() {
		return string(bs), nil
	}
	return bs, nil
}

var _ driver.Valuer = DriverData1{}

type DriverData2 struct {
	A int
}

func (a *DriverData2) Value() (driver.Value, error) {
	bs, err := json.Marshal(a)
	if err != nil {
		return nil, err
	}
	if TestDrv == dialects.DM.Name() {
		return string(bs), nil
	}
	return bs, nil
}

var _ driver.Valuer = &DriverData2{}

type TestD1 struct {
	TableName struct{}    `db:"gobatis_testc"`
	ID        int64       `db:"id,pk,autoincr"`
	Field0    DriverData1 `db:"field0"`
}

type TestD2 struct {
	TableName struct{}     `db:"gobatis_testc"`
	ID        int64        `db:"id,pk,autoincr"`
	Field0    *DriverData1 `db:"field0"`
}

type TestD3 struct {
	TableName struct{}    `db:"gobatis_testc"`
	ID        int64       `db:"id,pk,autoincr"`
	Field0    DriverData2 `db:"field0"`
}

type TestD4 struct {
	TableName struct{}     `db:"gobatis_testc"`
	ID        int64        `db:"id,pk,autoincr"`
	Field0    *DriverData2 `db:"field0"`
}

type TestE struct {
	TableName struct{} `db:"gobatis_teste1"`
	ID        int64    `db:"id,pk,autoincr"`
	Field0    []int64  `db:"field0"`
}

type TestE1 struct {
	TableName struct{} `db:"gobatis_teste1"`
	ID        int64    `db:"id,pk,autoincr"`
	Field0    []int64  `db:"field0"`
}

type TestE2 struct {
	TableName struct{} `db:"gobatis_teste1"`
	ID        int64    `db:"id,pk,autoincr"`
	Field0    *[]int64 `db:"field0"`
}

type TestE3 struct {
	TableName struct{} `db:"gobatis_teste1"`
	ID        int64    `db:"id,pk,autoincr"`
	Field0    *[]int64 `db:"field0,null"`
}

type TestE4 struct {
	TableName struct{} `db:"gobatis_teste1"`
	ID        int64    `db:"id,pk,autoincr"`
	Field0    []int64  `db:"field0,null"`
}

type TestE5 struct {
	TableName struct{} `db:"gobatis_teste2"`
	ID        int64    `db:"id,pk,autoincr"`
	Field0    *[]int64 `db:"field0,notnull"`
}

type TestE6 struct {
	TableName struct{} `db:"gobatis_teste2"`
	ID        int64    `db:"id,pk,autoincr"`
	Field0    []int64  `db:"field0,notnull"`
}

type TestF1 struct {
	TableName struct{} `db:"gobatis_testf1"`
	ID        int64    `db:"id,pk,autoincr"`
	Field0    []byte   `db:"field0,str"`
}

type TestF2 struct {
	TableName struct{} `db:"gobatis_testf1"`
	ID        int64    `db:"id,pk,autoincr"`
	Field0    *[]byte  `db:"field0,str"`
}

type TestF3 struct {
	TableName struct{} `db:"gobatis_testf1"`
	ID        int64    `db:"id,pk,autoincr"`
	Field0    *[]byte  `db:"field0,null,str"`
}

type TestF4 struct {
	TableName struct{} `db:"gobatis_testf1"`
	ID        int64    `db:"id,pk,autoincr"`
	Field0    []byte   `db:"field0,null,str"`
}

type TestF5 struct {
	TableName struct{} `db:"gobatis_testf2"`
	ID        int64    `db:"id,pk,autoincr"`
	Field0    *[]byte  `db:"field0,notnull,str"`
}

type TestF6 struct {
	TableName struct{} `db:"gobatis_testf2"`
	ID        int64    `db:"id,pk,autoincr"`
	Field0    []byte   `db:"field0,notnull,str"`
}

type ConvertTestIntNull struct {
	TableName struct{} `db:"gobatis_convert1"`
	ID        int64    `db:"id,pk,autoincr"`
	Field0    int64    `db:"field0,null"`
}

type ConvertTestIntNotNull struct {
	TableName struct{} `db:"gobatis_convert1"`
	ID        int64    `db:"id,pk,autoincr"`
	Field0    int64    `db:"field0,notnull"`
}
type ConvertTestStrNull struct {
	TableName struct{} `db:"gobatis_convert2"`
	ID        int64    `db:"id,pk,autoincr"`
	Field0    string   `db:"field0,null"`
}

type ConvertTestStrNotNull struct {
	TableName struct{} `db:"gobatis_convert2"`
	ID        int64    `db:"id,pk,autoincr"`
	Field0    string   `db:"field0,notnull"`
}

type IconvertTest interface {
	// @record_type ConvertTestIntNull
	InsertIntNULL(field0 interface{}) (int64, error)
	// @record_type ConvertTestIntNotNull
	InsertIntNotNULL(field0 interface{}) (int64, error)

	// @record_type ConvertTestStrNull
	InsertStrNULL(field0 interface{}) (int64, error)
	// @record_type ConvertTestStrNotNull
	InsertStrNotNULL(field0 interface{}) (int64, error)
}
