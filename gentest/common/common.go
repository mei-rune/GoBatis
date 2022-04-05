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