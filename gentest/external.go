package gentest

import (
	"context"
	"time"

	"github.com/runner-mei/GoBatis/gentest/common"
)

type IPAddress struct {
	TableName  struct{} `json:"-" xorm:"tpt_scan_result_addresses"`
	TaskID     int64    `json:"task_id" xorm:"task_id null"`
	InstanceID int64    `json:"id" xorm:"instance_id"`
	*common.IPAddress
	SampledAt time.Time `json:"sampled_at" xorm:"sampled_at"`
}

type IPAddressDao interface {
	Insert(ctx context.Context, address *IPAddress) error

	UpsertOnInstanceIdOnAddress(ctx context.Context, value *IPAddress) error

	QueryByInstance(ctx context.Context, instanceID int64) ([]IPAddress, error)

	DeleteByInstance(ctx context.Context, instanceID int64) (int64, error)

	DeleteByKey(ctx context.Context, instanceID int64, address string) error
}
