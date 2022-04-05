package gentest

import (
	"context"
	"time"
	"database/sql"

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

type Request struct {
	TableName   struct{} `json:"-" xorm:"tpt_itsm_requests"`
	ID          int64    `json:"id" xorm:"id pk autoincr"`
	Uuid        string   `json:"uuid" xorm:"uuid unique null"`
	Title       string   `json:"title" xorm:"title notnull"`
	Description string   `json:"description,omitempty" xorm:"description"`
	TypeID      int64    `json:"type_id,omitempty" xorm:"request_type_id"`
	// 3.9 版本中增加了这个 hidden 字段，为了开发方便，让数据库可以兼容，我在这里加添了这个字段
	XXXHidden  bool      `json:"hidden,omitempty" xorm:"hidden <- null"`
	OverdueAt  time.Time `json:"overdue_at,omitempty" xorm:"overdue_at null"`
	PriorityID int64     `json:"priority_id,omitempty" xorm:"priority_id null"`
	SLAID      int64     `json:"sla_id,omitempty" xorm:"sla_id null"`
	CreatorID  int64     `json:"creator_id,omitempty" xorm:"creator_id null"`
	// ReporterFromRequesterID int64                  `json:"reporter_from_requester_id,omitempty" xorm:"reporter_from_requester_id null"`
	ReporterFromUserID int64                  `json:"reporter_from_user_id,omitempty" xorm:"reporter_from_user_id null"`
	RequestSolution    string                 `json:"request_solution,omitempty" xorm:"request_solution null"`
	SolutionID         int64                  `json:"solution_id,omitempty" xorm:"solution_id null"`
	ProjectID          int64                  `json:"project_id,omitempty" xorm:"project_id null"`
	AttentionUserIDs   []int64                `json:"attention_user_ids" xorm:"attention_user_ids null"`
	WorkflowID         int64                  `json:"workflow_id" xorm:"workflow_id null"`
	Attributes         map[string]interface{} `json:"attributes" xorm:"attributes jsonb"`
	ClosedReason       string                 `json:"closed_reason,omitempty" xorm:"closed_reason null"`
	ClosedAt           time.Time              `json:"closed_at,omitempty" xorm:"closed_at null"`
	Deleted            time.Time              `json:"deleted" xorm:"deleted null"`
	CreatedAt          time.Time              `json:"created_at,omitempty" xorm:"created_at created"`
	UpdatedAt          time.Time              `json:"updated_at,omitempty" xorm:"updated_at updated"`
	// AssetIDs           []int64                `json:"asset_ids,omitempty" xorm:"asset_ids <-"`
	// TechnicianIDs           []int64                `json:"technician_ids" xorm:"technician_ids jsonb null"`
	// ChangeID                int64                  `json:"change_id,omitempty" xorm:"change_id null"`

}

type Requests interface {
	Insert(request *Request) (int64, error)

	FindByID(id int64) (*Request, error)

	// @type update
	CloseByID(id int64, closedAt time.Time, closedReason string) (int64, error)
}

type SystemInformation struct {
	TableName  struct{} `json:"-" xorm:"tpt_scan_result_sysinfo"`
	TaskID     int64    `json:"task_id" xorm:"task_id null"`
	InstanceID int64    `json:"id" xorm:"instance_id pk"`
	*common.Sysinfo
	SampleError string    `json:"error" xorm:"sample_error null"`
	SampledAt   time.Time `json:"sampled_at" xorm:"sampled_at"`
}

type SystemInformationDao interface {
	UpserttErrorOnInstanceID(ctx context.Context, instanceID int64, taskID sql.NullInt64, sampleError string, sampledAt time.Time) error
	Upsert(ctx context.Context, sysinfo *SystemInformation) error
	QueryByInstance(ctx context.Context, instanceID int64) (*SystemInformation, error)
	FindByID(ctx context.Context, instanceID int64) (*SystemInformation, error)
}