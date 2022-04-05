package common

import "time"

type TimeRange struct {
	Start time.Time
	End   time.Time
}



type IPAddress struct {
	IfIndex int64  `json:"if_index,omitempty" xorm:"if_index null"`
	Address string `json:"address,omitempty" xorm:"address"`
}

type Sysinfo struct {
	Description  string `json:"descr" xorm:"description null"`
	Oid          string `json:"sysoid" xorm:"sysoid null"`
	Uptime       int64  `json:"uptime" xorm:"uptime null"`
	Contact      string `json:"contact" xorm:"contact  null"`
	Name         string `json:"name" xorm:"name null"`
	Location     string `json:"location" xorm:"location null"`
	Services     int    `json:"services" xorm:"services null"`
	OS           string `json:"os,omitempty" xorm:"-"`
	Manufacturer string `json:"manufacturer,omitempty" xorm:"-"`
	Catalog      int    `json:"catalog" xorm:"-"`
	OldOid       string `json:"oid,omitempty" xorm:"-"`
}