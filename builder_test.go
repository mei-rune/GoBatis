package gobatis_test

import (
	"database/sql"
	"fmt"
	"reflect"
	"strings"
	"testing"
	"time"

	gobatis "github.com/runner-mei/GoBatis"
	_ "github.com/runner-mei/GoBatis/tests"
)

type TimeRange struct {
	Start time.Time
	End   time.Time
}

type T1 struct {
	ID        string    `db:"id,autoincr,pk"`
	F1        string    `db:"f1"`
	F2        int       `db:"f2,null"`
	F3        int       `db:"f3,notnull"`
	F4        int       `db:"f4,<-"`
	FIgnore   int       `db:"-"`
	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
	DeletedAt time.Time `db:"deleted_at,deleted"`
}

type T1_1 struct {
	ID        string    `db:"id,autoincr,pk"`
	F1        string    `db:"f1"`
	F2        int       `db:"f2,null"`
	F3        int       `db:"f3,notnull"`
	F4        int       `db:"f4,<-"`
	FIgnore   int       `db:"-"`
	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
	DeletedAt time.Time `db:"deleted"`
}

type T1ForNoDeleted struct {
	TableName struct{}  `db:"t1_table"`
	ID        string    `db:"id,autoincr,pk"`
	F1        string    `db:"f1"`
	F2        int       `db:"f2,null"`
	F3        int       `db:"f3,notnull"`
	F4        int       `db:"f4,<-"`
	FIgnore   int       `db:"-"`
	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
}

func init() {
	gobatis.RegisterTableName(T1{}, "t1_table")
	gobatis.RegisterTableName(T1_1{}, "t1_table")
}

type T2 struct {
	TableName struct{}  `db:"t2_table"`
	ID        string    `db:"id,autoincr,pk"`
	F1        string    `db:"f1"`
	F2        int       `db:"f2"`
	FIgnore   int       `db:"-"`
	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
}

type T3 struct {
	ID        string    `db:"id,autoincr,pk"`
	F1        string    `db:"f1"`
	F2        int       `db:"f2"`
	FIgnore   int       `db:"-"`
	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
}

func (t T3) TableName() string {
	return "t3_table"
}

type T4 struct {
	T2
	F3       string `db:"f3"`
	F4       int    `db:"f4"`
	FIgnore5 int    `db:"-"`
}

type T5 struct {
	ID        string    `db:"id,autoincr"`
	F1        string    `db:"f1"`
	F2        int       `db:"f2"`
	FIgnore   int       `db:"-"`
	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
}

func (t *T5) TableName() string {
	return "t5_table"
}

type T6 struct {
	ID        string    `db:"id,autoincr"`
	F1        string    `db:"f1"`
	F2        int       `db:"f2"`
	FIgnore   int       `db:"-"`
	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
}

func init() {
	gobatis.RegisterTableName(&T6{}, "t6_table")
}

type T7 struct {
	TableName struct{} `db:"t7_table"`
	T2
	F3       string `db:"f3"`
	FIgnore4 int    `db:"-"`
}

var mapper = gobatis.CreateMapper("", nil, nil)

type EmbededStruct struct {
	E1      string `db:"e1"`
	E2      int    `db:"e2"`
	FIgnore int    `db:"-"`
}

type T8 struct {
	TableName struct{}      `db:"t8_table"`
	ID        string        `db:"id,autoincr,pk"`
	T2        EmbededStruct `db:"-"`
	F1        string        `db:"f1"`
	F2        int           `db:"f2,created"`
	FIgnore3  int           `db:"-"`
	CreatedAt time.Time     `db:"created_at"`
	UpdatedAt time.Time     `db:"updated_at"`
}

type T9 struct {
	TableName struct{}      `db:"t9_table"`
	ID        string        `db:"id,autoincr,pk"`
	T2        EmbededStruct `db:"e"`
	F1        string        `db:"f1"`
	F2        int           `db:"f2,created"`
	CreatedAt time.Time     `db:"created_at"`
	UpdatedAt time.Time     `db:"updated_at"`
}

type T10 struct {
	TableName struct{}  `db:"t10_table"`
	ID        string    `db:"id,autoincr"`
	F1        string    `db:"f_1"`
	F2        int       `db:"f2,created"`
	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
}

type T11 struct {
	TableName struct{}  `db:"t11_table"`
	ID        int       `db:"id,autoincr"`
	F1        []string  `db:"f_1"`
	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
}

type T12 struct {
	TableName struct{}  `db:"t12_table"`
	ID        int       `db:"id,autoincr"`
	F1        []string  `db:"f_1,json"`
	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
}

type T13 struct {
	TableName struct{}  `db:"t13_table"`
	ID        int       `db:"id,autoincr"`
	F1        string    `db:"f1,notnull"`
	F2        int       `db:"f2,notnull"`
	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
}

type T14 struct {
	TableName struct{} `db:"t14_table"`
	ID        int      `db:"id,autoincr"`
	F1        string   `db:"f1,notnull"`
}

type T15 struct {
	TableName struct{}               `db:"t15_table"`
	ID        int                    `db:"id,autoincr"`
	F1        map[string]interface{} `db:"f1,notnull"`
}

type T16 struct {
	TableName struct{}  `db:"t16_table"`
	ID        string    `db:"id,autoincr,pk"`
	F1        string    `db:"f1,unique"`
	F2        int       `db:"f2,null"`
	F3        int       `db:"f3,notnull"`
	F4        int       `db:"f4,<-"`
	FIgnore   int       `db:"-"`
	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
	DeletedAt time.Time `db:"deleted_at,deleted"`
}

type T16_1 struct {
	TableName struct{}  `db:"t16_table"`
	ID        string    `db:"id,autoincr,pk"`
	F1        string    `db:"f1,unique"`
	F2        int       `db:"f2,null"`
	F3        int       `db:"f3,notnull"`
	F4        int       `db:"f4,<-"`
	FIgnore   int       `db:"-"`
	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
	DeletedAt time.Time `db:"deleted"`
}

type T17 struct {
	TableName struct{} `db:"t17_table"`
	ID        string   `db:"id,autoincr,pk"`
	F1        string   `db:"f1,unique"`
}

type T18 struct {
	TableName struct{} `db:"t18_table"`
	ID        string   `db:"id,autoincr,pk"`
	F1        string   `db:"f1"`
}

type Worklog struct {
	TableName   struct{}  `json:"-" db:"worklogs"`
	ID          string    `db:"id,autoincr,pk"`
	PlanID      int64     `json:"plan_id" db:"plan_id"`
	UserID      int64     `json:"user_id" db:"user_id"`
	Description string    `json:"description" db:"description"`
	CreatedAt   time.Time `json:"created_at,omitempty" db:"created_at,created"`
}

type T19 struct {
	TableName struct{}  `db:"t19_table"`
	ID        string    `db:"id,autoincr,pk"`
	F1        string    `db:"f_1,unique"`
	F2        int       `db:"f2,null"`
	F3        int       `db:"f3,notnull"`
	F4        int       `db:"f4,<-"`
	FIgnore   int       `db:"-"`
	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
	DeletedAt time.Time `db:"deleted_at,deleted"`
}

type T19_1 struct {
	TableName struct{}  `db:"t19_table"`
	ID        string    `db:"id,autoincr,pk"`
	F1        string    `db:"f_1,unique"`
	F2        int       `db:"f2,null"`
	F3        int       `db:"f3,notnull"`
	F4        int       `db:"f4,<-"`
	FIgnore   int       `db:"-"`
	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
	DeletedAt time.Time `db:"deleted"`
}

var (
	_stringType = reflect.TypeOf(new(string)).Elem()
	_intType    = reflect.TypeOf(new(int)).Elem()
	_timeType   = reflect.TypeOf(new(time.Time)).Elem()
)

func TestTableNameOK(t *testing.T) {
	for idx, test := range []struct {
		value     interface{}
		tableName string
	}{
		{value: T1{}, tableName: "t1_table"},
		{value: &T1{}, tableName: "t1_table"},
		{value: T2{}, tableName: "t2_table"},
		{value: &T2{}, tableName: "t2_table"},
		{value: T3{}, tableName: "t3_table"},
		{value: &T3{}, tableName: "t3_table"},
		{value: T4{}, tableName: "t2_table"},
		{value: &T4{}, tableName: "t2_table"},
		{value: T5{}, tableName: "t5_table"},
		{value: &T5{}, tableName: "t5_table"},
		{value: T6{}, tableName: "t6_table"},
		{value: &T6{}, tableName: "t6_table"},
	} {
		actaul, err := gobatis.ReadTableName(mapper, reflect.TypeOf(test.value))
		if err != nil {
			t.Error(err)
			continue
		}

		if actaul != test.tableName {
			t.Error("[", idx, "] excepted is", test.tableName)
			t.Error("[", idx, "] actual   is", actaul)
		}
	}
}

func TestTableNameError(t *testing.T) {
	for idx, test := range []struct {
		value interface{}
		err   string
	}{
		{value: T7{}, err: "mult choices"},
		{value: &T7{}, err: "mult choices"},
		{value: struct{}{}, err: "missing"},
	} {
		_, err := gobatis.ReadTableName(mapper, reflect.TypeOf(test.value))
		if err == nil {
			t.Error("excepted error got ok")
			continue
		}

		if !strings.Contains(err.Error(), test.err) {
			t.Error("[", idx, "] excepted is", test.err)
			t.Error("[", idx, "] actual   is", err)
		}
	}
}


type Assoc1 struct {
	TableName struct{}               `db:"assoc_table"`
	F1        int `db:"f1,unique"`
	F2        int `db:"f2,unique"`
}

type Assoc2 struct {
	TableName struct{}               `db:"assoc_table"`
	ID        int                    `db:"id,autoincr"`
	F1        int `db:"f1,unique"`
	F2        int `db:"f2,unique"`
}

type Assoc3 struct {
	TableName struct{}               `db:"assoc_table"`
	ID        int                    `db:"id,pk,autoincr"`
	F1        int `db:"f1,unique"`
	F2        int `db:"f2,unique"`
}

type Assoc4 struct {
	TableName struct{}               `db:"assoc_table"`
	ID        int                    `db:"id,autoincr"`
	F1        int `db:"f1,unique"`
	F2        int `db:"f2,unique"`

	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
}


func TestGenerateUpsertSQL(t *testing.T) {
	for idx, test := range []struct {
		dbType    gobatis.Dialect
		value     interface{}
		keyNames  []string
		argNames  []string
		argTypes  []reflect.Type
		noReturn  bool
		sql       string
		IncrField bool
	}{
		{dbType: gobatis.Postgres, value: T16_1{}, sql: "INSERT INTO t16_table(f1, f2, f3, created_at, updated_at) VALUES(#{f1}, #{f2}, #{f3}, now(), now()) ON CONFLICT (f1) DO UPDATE SET f2=EXCLUDED.f2, f3=EXCLUDED.f3, updated_at=EXCLUDED.updated_at RETURNING id"},
		{dbType: gobatis.Postgres, value: T16{}, sql: "INSERT INTO t16_table(f1, f2, f3, created_at, updated_at) VALUES(#{f1}, #{f2}, #{f3}, now(), now()) ON CONFLICT (f1) DO UPDATE SET f2=EXCLUDED.f2, f3=EXCLUDED.f3, updated_at=EXCLUDED.updated_at RETURNING id"},
		{dbType: gobatis.Mysql, value: T16{}, sql: "INSERT INTO t16_table(f1, f2, f3, created_at, updated_at) VALUES(#{f1}, #{f2}, #{f3}, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP) ON DUPLICATE KEY UPDATE f2=VALUES(f2), f3=VALUES(f3), updated_at=VALUES(updated_at)"},
		{dbType: gobatis.MSSql, value: T16{}, sql: `MERGE INTO t16_table AS t USING ( VALUES(#{f1}, #{f2}, #{f3}, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP ) ) AS s (f1, f2, f3, created_at, updated_at ) ON t.f1 = s.f1 WHEN MATCHED THEN UPDATE SET f2 = s.f2, f3 = s.f3, updated_at = s.updated_at WHEN NOT MATCHED THEN INSERT (f1, f2, f3, created_at, updated_at) VALUES(s.f1, s.f2, s.f3, s.created_at, s.updated_at)  OUTPUT inserted.id;`},
		{dbType: gobatis.MSSql, value: T16_1{}, sql: `MERGE INTO t16_table AS t USING ( VALUES(#{f1}, #{f2}, #{f3}, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP ) ) AS s (f1, f2, f3, created_at, updated_at ) ON t.f1 = s.f1 WHEN MATCHED THEN UPDATE SET f2 = s.f2, f3 = s.f3, updated_at = s.updated_at WHEN NOT MATCHED THEN INSERT (f1, f2, f3, created_at, updated_at) VALUES(s.f1, s.f2, s.f3, s.created_at, s.updated_at)  OUTPUT inserted.id;`},
		{dbType: gobatis.Postgres, value: T18{}, sql: "INSERT INTO t18_table(id, f1) VALUES(#{id}, #{f1}) ON CONFLICT (id) DO UPDATE SET f1=EXCLUDED.f1 RETURNING id", IncrField: true},
		{dbType: gobatis.Postgres, value: T17{}, sql: "INSERT INTO t17_table(f1) VALUES(#{f1}) ON CONFLICT (f1) DO NOTHING  RETURNING id"},
		// {dbType: gobatis.Mysql, value: T17{}, sql: "INSERT INTO t17_table(f1) VALUES(#{f1}) ON DUPLICATE KEY UPDATE "},
		{dbType: gobatis.MSSql, value: T17{}, sql: `MERGE INTO t17_table AS t USING ( VALUES(#{f1} ) ) AS s (f1 ) ON t.f1 = s.f1 WHEN NOT MATCHED THEN INSERT (f1) VALUES(s.f1)  OUTPUT inserted.id;`},

		{
			dbType:   gobatis.Postgres,
			value:    T16{},
			keyNames: []string{"f1"},
			argNames: []string{"f2"},
			argTypes: []reflect.Type{_intType},
			sql:      "INSERT INTO t16_table(f1, f2, created_at, updated_at) VALUES(#{f1}, #{f2}, now(), now()) ON CONFLICT (f1) DO UPDATE SET f2=EXCLUDED.f2, updated_at=EXCLUDED.updated_at RETURNING id",
		},

		{
			dbType:   gobatis.Postgres,
			value:    T16{},
			keyNames: []string{"f1"},
			argNames: []string{"f2", "f3"},
			argTypes: []reflect.Type{_stringType, _stringType},
			sql:      "INSERT INTO t16_table(f1, f2, f3, created_at, updated_at) VALUES(#{f1}, #{f2}, #{f3}, now(), now()) ON CONFLICT (f1) DO UPDATE SET f2=EXCLUDED.f2, f3=EXCLUDED.f3, updated_at=EXCLUDED.updated_at RETURNING id",
		},

		{
			dbType:   gobatis.Postgres,
			value:    T16{},
			keyNames: []string{"f1"},
			argNames: []string{"f2", "f3", "created_at", "updated_at"},
			argTypes: []reflect.Type{_intType, _intType, _timeType, _timeType},
			sql:      "INSERT INTO t16_table(f1, f2, f3, created_at, updated_at) VALUES(#{f1}, #{f2}, #{f3}, now(), now()) ON CONFLICT (f1) DO UPDATE SET f2=EXCLUDED.f2, f3=EXCLUDED.f3, updated_at=EXCLUDED.updated_at RETURNING id",
		},
		{
			dbType:   gobatis.Postgres,
			value:    T19{},
			keyNames: []string{"f1"},
			argNames: []string{"f2", "f3", "created_at", "updated_at"},
			argTypes: []reflect.Type{_intType, _intType, _timeType, _timeType},
			sql:      "INSERT INTO t19_table(f_1, f2, f3, created_at, updated_at) VALUES(#{f1}, #{f2}, #{f3}, now(), now()) ON CONFLICT (f_1) DO UPDATE SET f2=EXCLUDED.f2, f3=EXCLUDED.f3, updated_at=EXCLUDED.updated_at RETURNING id",
		},




// type Assoc1 struct {
// 	TableName struct{}               `db:"assoc_table"`
// 	F1        int `db:"f1,unique"`
// 	F2        int `db:"f2,unique"`
// }
		{
			dbType:   gobatis.Postgres,
			value:    Assoc1{},
			keyNames: []string{},
			argNames: []string{"f1", "f2"},
			argTypes: []reflect.Type{_intType, _intType},
			sql:      "INSERT INTO assoc_table(f1, f2) VALUES(#{f1}, #{f2}) ON CONFLICT (f1, f2) DO NOTHING ",
		},


		{
			dbType:   gobatis.MSSql,
			value:    Assoc1{},
			keyNames: []string{},
			argNames: []string{"f1", "f2"},
			argTypes: []reflect.Type{_intType, _intType},
			sql:      "MERGE INTO assoc_table AS t USING ( VALUES(#{f1}, #{f2} ) ) AS s (f1, f2 ) ON t.f1 = s.f1 AND t.f2 = s.f2 WHEN NOT MATCHED THEN INSERT (f1, f2) VALUES(s.f1, s.f2) ;",
		},

		{
			dbType:   gobatis.Oracle,
			value:    Assoc1{},
			keyNames: []string{},
			argNames: []string{"f1", "f2"},
			argTypes: []reflect.Type{_intType, _intType},
			sql:      "MERGE INTO assoc_table AS t USING dual ON t.f1= #{f1} AND t.f2= #{f2} WHEN NOT MATCHED THEN INSERT (f1, f2) VALUES(#{f1}, #{f2}) ",
		},


		// {
		// 	dbType:   gobatis.Mysql,
		// 	value:    Assoc1{},
		// 	keyNames: []string{},
		// 	argNames: []string{"f1", "f2"},
		// 	argTypes: []reflect.Type{_intType, _intType},
		// 	sql:      "MERGE INTO assoc_table USING dual ON assoc_table.#{f1}=f1 AND assoc_table.f2=#{f2} WHEN NOT MATCHED THEN INSERT (f1, f2) VALUES(#{f1}, #{f2});",
		// },

		// type Assoc2 struct {
		// 	TableName struct{}               `db:"assoc_table"`
		// 	ID        int                    `db:"id,autoincr"`
		// 	F1        int `db:"f1,unique"`
		// 	F2        int `db:"f2,unique"`
		// }

		{
			dbType:   gobatis.Postgres,
			value:    Assoc2{},
			keyNames: []string{},
			argNames: []string{"f1", "f2"},
			argTypes: []reflect.Type{_intType, _intType},
			sql:      "INSERT INTO assoc_table(f1, f2) VALUES(#{f1}, #{f2}) ON CONFLICT (f1, f2) DO NOTHING  RETURNING id",
		},

		{
			dbType:   gobatis.MSSql,
			value:    Assoc2{},
			keyNames: []string{},
			argNames: []string{"f1", "f2"},
			argTypes: []reflect.Type{_intType, _intType},
			sql:      "MERGE INTO assoc_table AS t USING ( VALUES(#{f1}, #{f2} ) ) AS s (f1, f2 ) ON t.f1 = s.f1 AND t.f2 = s.f2 WHEN NOT MATCHED THEN INSERT (f1, f2) VALUES(s.f1, s.f2)  OUTPUT inserted.id;",
		},

		{
			dbType:   gobatis.Oracle,
			value:    Assoc2{},
			keyNames: []string{},
			argNames: []string{"f1", "f2"},
			argTypes: []reflect.Type{_intType, _intType},
			sql:      "MERGE INTO assoc_table AS t USING dual ON t.f1= #{f1} AND t.f2= #{f2} WHEN NOT MATCHED THEN INSERT (f1, f2) VALUES(#{f1}, #{f2}) ",
		},

		// {
		// 	dbType:   gobatis.Mysql,
		// 	value:    Assoc2{},
		// 	keyNames: []string{},
		// 	argNames: []string{"f1", "f2"},
		// 	argTypes: []reflect.Type{_intType, _intType},
		// 	sql:      "MERGE INTO assoc_table USING dual ON assoc_table.#{f1}=f1 AND assoc_table.f2=#{f2} WHEN NOT MATCHED THEN INSERT (f1, f2) VALUES(#{f1}, #{f2});",
		// },

// type Assoc3 struct {
// 	TableName struct{}               `db:"assoc_table"`
// 	ID        int                    `db:"id,pk,autoincr"`
// 	F1        int `db:"f1,unique"`
// 	F2        int `db:"f2,unique"`
// }


		{
			dbType:   gobatis.Postgres,
			value:    Assoc3{},
			keyNames: []string{},
			argNames: []string{"f1", "f2"},
			argTypes: []reflect.Type{_intType, _intType},
			sql:      "INSERT INTO assoc_table(f1, f2) VALUES(#{f1}, #{f2}) ON CONFLICT (f1, f2) DO NOTHING  RETURNING id",
		},

		{
			dbType:   gobatis.MSSql,
			value:    Assoc3{},
			keyNames: []string{},
			argNames: []string{"f1", "f2"},
			argTypes: []reflect.Type{_intType, _intType},
			sql:      "MERGE INTO assoc_table AS t USING ( VALUES(#{f1}, #{f2} ) ) AS s (f1, f2 ) ON t.f1 = s.f1 AND t.f2 = s.f2 WHEN NOT MATCHED THEN INSERT (f1, f2) VALUES(s.f1, s.f2)  OUTPUT inserted.id;",
		},

		{
			dbType:   gobatis.Oracle,
			value:    Assoc3{},
			keyNames: []string{},
			argNames: []string{"f1", "f2"},
			argTypes: []reflect.Type{_intType, _intType},
			sql:      "MERGE INTO assoc_table AS t USING dual ON t.f1= #{f1} AND t.f2= #{f2} WHEN NOT MATCHED THEN INSERT (f1, f2) VALUES(#{f1}, #{f2}) ",
		},

		// {
		// 	dbType:   gobatis.Mysql,
		// 	value:    Assoc3{},
		// 	keyNames: []string{},
		// 	argNames: []string{"f1", "f2"},
		// 	argTypes: []reflect.Type{_intType, _intType},
		// 	sql:      "MERGE INTO assoc_table USING dual ON assoc_table.#{f1}=f1 AND assoc_table.f2=#{f2} WHEN NOT MATCHED THEN INSERT (f1, f2) VALUES(#{f1}, #{f2});",
		// },

		// type Assoc4 struct {
		// 	TableName struct{}               `db:"assoc_table"`
		// 	ID        int                    `db:"id,autoincr"`
		// 	F1        int `db:"f1,unique"`
		// 	F2        int `db:"f2,unique"`

		// 	CreatedAt time.Time `db:"created_at"`
		// 	UpdatedAt time.Time `db:"updated_at"`
		// }

		{
			dbType:   gobatis.Postgres,
			value:    Assoc4{},
			keyNames: []string{},
			argNames: []string{"f1", "f2"},
			argTypes: []reflect.Type{_intType, _intType},
			sql:      "INSERT INTO assoc_table(f1, f2, created_at, updated_at) VALUES(#{f1}, #{f2}, now(), now()) ON CONFLICT (f1, f2) DO UPDATE SET updated_at=EXCLUDED.updated_at RETURNING id",
		},

		{
			dbType:   gobatis.MSSql,
			value:    Assoc4{},
			keyNames: []string{},
			argNames: []string{"f1", "f2"},
			argTypes: []reflect.Type{_intType, _intType},
			sql:      "MERGE INTO assoc_table AS t USING ( VALUES(#{f1}, #{f2}, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP ) ) AS s (f1, f2, created_at, updated_at ) ON t.f1 = s.f1 AND t.f2 = s.f2 WHEN MATCHED THEN UPDATE SET updated_at = s.updated_at WHEN NOT MATCHED THEN INSERT (f1, f2, created_at, updated_at) VALUES(s.f1, s.f2, s.created_at, s.updated_at)  OUTPUT inserted.id;",
		},

		{
			dbType:   gobatis.Oracle,
			value:    Assoc4{},
			keyNames: []string{},
			argNames: []string{"f1", "f2"},
			argTypes: []reflect.Type{_intType, _intType},
			sql:      "MERGE INTO assoc_table AS t USING dual ON t.f1= #{f1} AND t.f2= #{f2} WHEN MATCHED THEN UPDATE SET updated_at= CURRENT_TIMESTAMP WHEN NOT MATCHED THEN INSERT (f1, f2, created_at, updated_at) VALUES(#{f1}, #{f2}, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP) ",
		},

		// {
		// 	dbType:   gobatis.Mysql,
		// 	value:    Assoc4{},
		// 	keyNames: []string{},
		// 	argNames: []string{"f1", "f2"},
		// 	argTypes: []reflect.Type{_intType, _intType},
		// 	sql:      "MERGE INTO assoc_table USING dual ON assoc_table.#{f1}=f1 AND assoc_table.f2=#{f2} WHEN NOT MATCHED THEN INSERT (f1, f2) VALUES(#{f1}, #{f2});",
		// },


	} {
		old := gobatis.UpsertSupportAutoIncrField
		gobatis.UpsertSupportAutoIncrField = test.IncrField
		actaul, err := gobatis.GenerateUpsertSQL(test.dbType, mapper, reflect.TypeOf(test.value), test.keyNames, test.argNames, test.argTypes, false)
		gobatis.UpsertSupportAutoIncrField = old
		if err != nil {
			t.Error("[", idx, "]", err)
			continue
		}

		if actaul != test.sql {
			t.Error("[", idx, "] excepted is", test.sql, "|")
			t.Error("[", idx, "] actual   is", actaul, "|")
		}
	}

	for idx, test := range []struct {
		dbType   gobatis.Dialect
		value    interface{}
		keyNames []string
		argNames []string
		argTypes []reflect.Type
		noReturn bool
		err      string
	}{
		{
			dbType: gobatis.Mysql,
			value:  T17{},
			err:    "empty update fields",
		},
		{
			dbType:   gobatis.Postgres,
			value:    T16{},
			keyNames: []string{},
			argNames: []string{"f2", "f3", "created_at", "updated_at"},
			argTypes: []reflect.Type{_stringType, _intType, _intType, _timeType, _timeType},
			err:      "argument 'f1' is missing",
		},
	} {
		sql, err := gobatis.GenerateUpsertSQL(test.dbType, mapper, reflect.TypeOf(test.value), test.keyNames, test.argNames, test.argTypes, test.noReturn)
		if err == nil {
			t.Error("[", idx, "]", "want error got ok")
			t.Log(sql)
			continue
		}

		if !strings.Contains(err.Error(), test.err) {
			t.Error("[", idx, "] excepted is", test.err)
			t.Error("[", idx, "] actual   is", err.Error())
		}
	}
}

func TestGenerateInsertSQL(t *testing.T) {
	for idx, test := range []struct {
		dbType   gobatis.Dialect
		value    interface{}
		names    []string
		argTypes []reflect.Type
		noReturn bool
		sql      string
	}{
		{dbType: gobatis.Postgres, value: T1{}, sql: "INSERT INTO t1_table(f1, f2, f3, created_at, updated_at) VALUES(#{f1}, #{f2}, #{f3}, now(), now()) RETURNING id"},
		{dbType: gobatis.Postgres, value: &T1{}, sql: "INSERT INTO t1_table(f1, f2, f3, created_at, updated_at) VALUES(#{f1}, #{f2}, #{f3}, now(), now()) RETURNING id"},
		{dbType: gobatis.Postgres, value: T2{}, sql: "INSERT INTO t2_table(f1, f2, created_at, updated_at) VALUES(#{f1}, #{f2}, now(), now()) RETURNING id"},
		{dbType: gobatis.Postgres, value: &T2{}, sql: "INSERT INTO t2_table(f1, f2, created_at, updated_at) VALUES(#{f1}, #{f2}, now(), now()) RETURNING id"},
		{dbType: gobatis.Postgres, value: T3{}, sql: "INSERT INTO t3_table(f1, f2, created_at, updated_at) VALUES(#{f1}, #{f2}, now(), now()) RETURNING id"},
		{dbType: gobatis.Postgres, value: &T3{}, sql: "INSERT INTO t3_table(f1, f2, created_at, updated_at) VALUES(#{f1}, #{f2}, now(), now()) RETURNING id"},
		{dbType: gobatis.Postgres, value: T4{}, sql: "INSERT INTO t2_table(f3, f4, f1, f2, created_at, updated_at) VALUES(#{f3}, #{f4}, #{f1}, #{f2}, now(), now()) RETURNING id"},
		{dbType: gobatis.Postgres, value: &T4{}, sql: "INSERT INTO t2_table(f3, f4, f1, f2, created_at, updated_at) VALUES(#{f3}, #{f4}, #{f1}, #{f2}, now(), now()) RETURNING id"},
		{dbType: gobatis.Postgres, value: T4{}, sql: "INSERT INTO t2_table(f3, f4, f1, f2, created_at, updated_at) VALUES(#{f3}, #{f4}, #{f1}, #{f2}, now(), now())", noReturn: true},
		{dbType: gobatis.Postgres, value: &T4{}, sql: "INSERT INTO t2_table(f3, f4, f1, f2, created_at, updated_at) VALUES(#{f3}, #{f4}, #{f1}, #{f2}, now(), now())", noReturn: true},
		{dbType: gobatis.Mysql, value: T4{}, sql: "INSERT INTO t2_table(f3, f4, f1, f2, created_at, updated_at) VALUES(#{f3}, #{f4}, #{f1}, #{f2}, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)"},
		{dbType: gobatis.Mysql, value: &T4{}, sql: "INSERT INTO t2_table(f3, f4, f1, f2, created_at, updated_at) VALUES(#{f3}, #{f4}, #{f1}, #{f2}, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)"},
		{dbType: gobatis.Postgres, value: T8{}, sql: "INSERT INTO t8_table(f1, f2, created_at, updated_at) VALUES(#{f1}, #{f2}, now(), now()) RETURNING id"},
		{dbType: gobatis.Postgres, value: &T8{}, sql: "INSERT INTO t8_table(f1, f2, created_at, updated_at) VALUES(#{f1}, #{f2}, now(), now()) RETURNING id"},
		{dbType: gobatis.Postgres, value: T9{}, sql: "INSERT INTO t9_table(e, f1, f2, created_at, updated_at) VALUES(#{e}, #{f1}, #{f2}, now(), now()) RETURNING id"},
		{dbType: gobatis.Postgres, value: &T9{}, sql: "INSERT INTO t9_table(e, f1, f2, created_at, updated_at) VALUES(#{e}, #{f1}, #{f2}, now(), now()) RETURNING id"},
		{dbType: gobatis.MSSql, value: &T4{}, sql: "INSERT INTO t2_table(f3, f4, f1, f2, created_at, updated_at) OUTPUT inserted.id VALUES(#{f3}, #{f4}, #{f1}, #{f2}, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)"},
		{dbType: gobatis.MSSql, value: &T4{}, sql: "INSERT INTO t2_table(f3, f4, f1, f2, created_at, updated_at) VALUES(#{f3}, #{f4}, #{f1}, #{f2}, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)", noReturn: true},

		{dbType: gobatis.MSSql, value: &T14{},
			sql:      "INSERT INTO t14_table(f1) VALUES(#{f1,notnull=true})",
			names:    []string{"f1"},
			argTypes: []reflect.Type{_stringType},
			noReturn: true},
		{dbType: gobatis.MSSql, value: &T14{},
			sql:      "INSERT INTO t14_table(f1) VALUES(#{f1,notnull=true})",
			names:    []string{"f1"},
			argTypes: []reflect.Type{_intType},
			noReturn: true},
		{dbType: gobatis.MSSql, value: &T14{},
			sql:      "INSERT INTO t14_table(f1) VALUES(#{f1,notnull=true})",
			names:    []string{"f1"},
			argTypes: []reflect.Type{nil},
			noReturn: true},

		{dbType: gobatis.MSSql, value: &T14{},
			sql:      "INSERT INTO t14_table(f1) VALUES(#{f1.f1})",
			names:    []string{"f1"},
			argTypes: []reflect.Type{reflect.TypeOf(new(T14)).Elem()},
			noReturn: true},
		{dbType: gobatis.MSSql, value: &T15{},
			sql:      "INSERT INTO t15_table(f1) VALUES(#{f1.f1})",
			names:    []string{"f1"},
			argTypes: []reflect.Type{reflect.TypeOf(new(T15)).Elem()},
			noReturn: true},
	} {
		actaul, err := gobatis.GenerateInsertSQL(test.dbType, mapper, reflect.TypeOf(test.value), test.names, test.argTypes, test.noReturn)
		if err != nil {
			t.Error(err)
			continue
		}

		if actaul != test.sql {
			t.Error("[", idx, "] excepted is", test.sql)
			t.Error("[", idx, "] actual   is", actaul)
		}
	}

	_, err := gobatis.GenerateInsertSQL(gobatis.Mysql,
		mapper, reflect.TypeOf(&T7{}), nil, nil, false)
	if err == nil {
		t.Error("excepted error got ok")
		return
	}
}

func TestGenerateInsertSQL2(t *testing.T) {
	for idx, test := range []struct {
		dbType   gobatis.Dialect
		fields   []string
		value    interface{}
		noReturn bool
		sql      string
	}{
		{dbType: gobatis.Postgres, value: T1{}, fields: []string{"f1", "f2", "f3"}, sql: "INSERT INTO t1_table(f1, f2, f3, created_at, updated_at) VALUES(#{f1}, #{f2,null=true}, #{f3,notnull=true}, now(), now()) RETURNING id"},
		{dbType: gobatis.Postgres, value: &T1{}, fields: []string{"f1", "f2", "f3"}, sql: "INSERT INTO t1_table(f1, f2, f3, created_at, updated_at) VALUES(#{f1}, #{f2,null=true}, #{f3,notnull=true}, now(), now()) RETURNING id"},
		{dbType: gobatis.Postgres, value: T2{}, fields: []string{"f1", "f2"}, sql: "INSERT INTO t2_table(f1, f2, created_at, updated_at) VALUES(#{f1}, #{f2}, now(), now()) RETURNING id"},
		{dbType: gobatis.Postgres, value: &T2{}, fields: []string{"f1", "f2"}, sql: "INSERT INTO t2_table(f1, f2, created_at, updated_at) VALUES(#{f1}, #{f2}, now(), now()) RETURNING id"},
		{dbType: gobatis.Postgres, value: T3{}, fields: []string{"f1", "f2"}, sql: "INSERT INTO t3_table(f1, f2, created_at, updated_at) VALUES(#{f1}, #{f2}, now(), now()) RETURNING id"},
		{dbType: gobatis.Postgres, value: &T3{}, fields: []string{"f1", "f2"}, sql: "INSERT INTO t3_table(f1, f2, created_at, updated_at) VALUES(#{f1}, #{f2}, now(), now()) RETURNING id"},
		{dbType: gobatis.Postgres, value: T4{}, fields: []string{"f1", "f2", "f3", "f4"}, sql: "INSERT INTO t2_table(f3, f4, f1, f2, created_at, updated_at) VALUES(#{f3}, #{f4}, #{f1}, #{f2}, now(), now()) RETURNING id"},
		{dbType: gobatis.Postgres, value: &T4{}, fields: []string{"f1", "f2", "f3", "f4"}, sql: "INSERT INTO t2_table(f3, f4, f1, f2, created_at, updated_at) VALUES(#{f3}, #{f4}, #{f1}, #{f2}, now(), now()) RETURNING id"},
		{dbType: gobatis.Postgres, value: T4{}, fields: []string{"f1", "f2", "f3", "f4"}, sql: "INSERT INTO t2_table(f3, f4, f1, f2, created_at, updated_at) VALUES(#{f3}, #{f4}, #{f1}, #{f2}, now(), now())", noReturn: true},
		{dbType: gobatis.Postgres, value: &T4{}, fields: []string{"f1", "f2", "f3", "f4"}, sql: "INSERT INTO t2_table(f3, f4, f1, f2, created_at, updated_at) VALUES(#{f3}, #{f4}, #{f1}, #{f2}, now(), now())", noReturn: true},
		{dbType: gobatis.Mysql, value: T4{}, fields: []string{"f1", "f2", "f3", "f4"}, sql: "INSERT INTO t2_table(f3, f4, f1, f2, created_at, updated_at) VALUES(#{f3}, #{f4}, #{f1}, #{f2}, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)"},
		{dbType: gobatis.Mysql, value: &T4{}, fields: []string{"f1", "f2", "f3", "f4"}, sql: "INSERT INTO t2_table(f3, f4, f1, f2, created_at, updated_at) VALUES(#{f3}, #{f4}, #{f1}, #{f2}, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)"},
		{dbType: gobatis.Postgres, value: T8{}, fields: []string{"f1", "f2"}, sql: "INSERT INTO t8_table(f1, f2, created_at, updated_at) VALUES(#{f1}, #{f2}, now(), now()) RETURNING id"},
		{dbType: gobatis.Postgres, value: &T8{}, fields: []string{"f1", "f2"}, sql: "INSERT INTO t8_table(f1, f2, created_at, updated_at) VALUES(#{f1}, #{f2}, now(), now()) RETURNING id"},
		{dbType: gobatis.Postgres, value: T9{}, fields: []string{"f1", "f2", "e"}, sql: "INSERT INTO t9_table(e, f1, f2, created_at, updated_at) VALUES(#{e}, #{f1}, #{f2}, now(), now()) RETURNING id"},
		{dbType: gobatis.Postgres, value: &T9{}, fields: []string{"f1", "f2", "e"}, sql: "INSERT INTO t9_table(e, f1, f2, created_at, updated_at) VALUES(#{e}, #{f1}, #{f2}, now(), now()) RETURNING id"},
		{dbType: gobatis.MSSql, value: &T4{}, fields: []string{"f1", "f2", "f3", "f4"}, sql: "INSERT INTO t2_table(f3, f4, f1, f2, created_at, updated_at) OUTPUT inserted.id VALUES(#{f3}, #{f4}, #{f1}, #{f2}, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)"},
		{dbType: gobatis.MSSql, value: &T4{}, fields: []string{"f1", "f2", "f3", "f4"}, sql: "INSERT INTO t2_table(f3, f4, f1, f2, created_at, updated_at) VALUES(#{f3}, #{f4}, #{f1}, #{f2}, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)", noReturn: true},
		{dbType: gobatis.Postgres, value: T10{}, fields: []string{"f_1", "f2"}, sql: "INSERT INTO t10_table(f_1, f2, created_at, updated_at) VALUES(#{f_1}, #{f2}, now(), now()) RETURNING id"},
		{dbType: gobatis.Postgres, value: &T10{}, fields: []string{"f_1", "f2"}, sql: "INSERT INTO t10_table(f_1, f2, created_at, updated_at) VALUES(#{f_1}, #{f2}, now(), now()) RETURNING id"},
		{dbType: gobatis.MSSql, value: T10{}, fields: []string{"f_1", "f2"}, sql: "INSERT INTO t10_table(f_1, f2, created_at, updated_at) OUTPUT inserted.id VALUES(#{f_1}, #{f2}, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)"},
		{dbType: gobatis.MSSql, value: &T10{}, fields: []string{"f_1", "f2"}, sql: "INSERT INTO t10_table(f_1, f2, created_at, updated_at) OUTPUT inserted.id VALUES(#{f_1}, #{f2}, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)"},

		{dbType: gobatis.Postgres, value: T10{}, fields: []string{"f1", "f2"}, sql: "INSERT INTO t10_table(f_1, f2, created_at, updated_at) VALUES(#{f1}, #{f2}, now(), now()) RETURNING id"},
		{dbType: gobatis.Postgres, value: &T10{}, fields: []string{"f1", "f2"}, sql: "INSERT INTO t10_table(f_1, f2, created_at, updated_at) VALUES(#{f1}, #{f2}, now(), now()) RETURNING id"},
		{dbType: gobatis.MSSql, value: T10{}, fields: []string{"f1", "f2"}, sql: "INSERT INTO t10_table(f_1, f2, created_at, updated_at) OUTPUT inserted.id VALUES(#{f1}, #{f2}, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)"},
		{dbType: gobatis.MSSql, value: &T10{}, fields: []string{"f1", "f2"}, sql: "INSERT INTO t10_table(f_1, f2, created_at, updated_at) OUTPUT inserted.id VALUES(#{f1}, #{f2}, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)"},

		{dbType: gobatis.Postgres, value: T10{}, fields: []string{"f1", "f2", "created_at", "updated_at"}, sql: "INSERT INTO t10_table(f_1, f2, created_at, updated_at) VALUES(#{f1}, #{f2}, now(), now()) RETURNING id"},
		{dbType: gobatis.Postgres, value: &T10{}, fields: []string{"f1", "f2", "created_at", "updated_at"}, sql: "INSERT INTO t10_table(f_1, f2, created_at, updated_at) VALUES(#{f1}, #{f2}, now(), now()) RETURNING id"},
		{dbType: gobatis.MSSql, value: T10{}, fields: []string{"f1", "f2", "created_at", "updated_at"}, sql: "INSERT INTO t10_table(f_1, f2, created_at, updated_at) OUTPUT inserted.id VALUES(#{f1}, #{f2}, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)"},
		{dbType: gobatis.MSSql, value: &T10{}, fields: []string{"f1", "f2", "created_at", "updated_at"}, sql: "INSERT INTO t10_table(f_1, f2, created_at, updated_at) OUTPUT inserted.id VALUES(#{f1}, #{f2}, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)"},
	} {
		actaul, err := gobatis.GenerateInsertSQL2(test.dbType,
			mapper, reflect.TypeOf(test.value), test.fields, test.noReturn)
		if err != nil {
			t.Error(err)
			continue
		}

		if actaul != test.sql {
			t.Error("[", idx, "] excepted is", test.sql)
			t.Error("[", idx, "] actual   is", actaul)
		}
	}

	_, err := gobatis.GenerateInsertSQL2(gobatis.Postgres,
		mapper, reflect.TypeOf(&T1{}), []string{"f1", "f2", "f3", "deleted_at"}, false)
	if err == nil {
		t.Error("excepted error got ok")
		return
	}
}

func TestGenerateUpdateSQL(t *testing.T) {
	for idx, test := range []struct {
		dbType   gobatis.Dialect
		prefix   string
		value    interface{}
		names    []string
		argTypes []reflect.Type
		sql      string
	}{
		{dbType: gobatis.Postgres, value: T1{}, sql: "UPDATE t1_table SET f1=#{f1}, f2=#{f2}, f3=#{f3}, updated_at=now() WHERE id=#{id}"},
		{dbType: gobatis.Mysql, value: &T1{}, names: []string{"id"}, sql: "UPDATE t1_table SET f1=#{f1}, f2=#{f2}, f3=#{f3}, updated_at=CURRENT_TIMESTAMP WHERE id=#{id}"},
		{dbType: gobatis.Postgres, value: T2{}, sql: "UPDATE t2_table SET f1=#{f1}, f2=#{f2}, updated_at=now() WHERE id=#{id}"},
		{dbType: gobatis.Postgres, value: &T2{}, names: []string{"id"}, sql: "UPDATE t2_table SET f1=#{f1}, f2=#{f2}, updated_at=now() WHERE id=#{id}"},
		{dbType: gobatis.Postgres, value: T3{}, sql: "UPDATE t3_table SET f1=#{f1}, f2=#{f2}, updated_at=now() WHERE id=#{id}"},
		{dbType: gobatis.Postgres, value: &T3{}, names: []string{"id"}, sql: "UPDATE t3_table SET f1=#{f1}, f2=#{f2}, updated_at=now() WHERE id=#{id}"},
		{dbType: gobatis.Postgres, value: T4{}, sql: "UPDATE t2_table SET f3=#{f3}, f4=#{f4}, f1=#{f1}, f2=#{f2}, updated_at=now() WHERE id=#{id}"},
		{dbType: gobatis.Postgres, value: &T4{}, names: []string{"id"}, sql: "UPDATE t2_table SET f3=#{f3}, f4=#{f4}, f1=#{f1}, f2=#{f2}, updated_at=now() WHERE id=#{id}"},
		{dbType: gobatis.Postgres, value: &T4{}, names: []string{"id", "f2"}, sql: "UPDATE t2_table SET f3=#{f3}, f4=#{f4}, f1=#{f1}, updated_at=now() WHERE id=#{id} AND f2=#{f2}"},
		{dbType: gobatis.Postgres, value: T8{}, sql: "UPDATE t8_table SET f1=#{f1}, updated_at=now() WHERE id=#{id}"},
		{dbType: gobatis.Postgres, value: &T8{}, names: []string{"id"}, sql: "UPDATE t8_table SET f1=#{f1}, updated_at=now() WHERE id=#{id}"},
		{dbType: gobatis.Postgres, value: T9{}, sql: "UPDATE t9_table SET e=#{e}, f1=#{f1}, updated_at=now() WHERE id=#{id}"},
		{dbType: gobatis.Postgres, value: &T9{}, names: []string{"id"}, sql: "UPDATE t9_table SET e=#{e}, f1=#{f1}, updated_at=now() WHERE id=#{id}"},
		{dbType: gobatis.Postgres, prefix: "a.", value: T9{}, sql: "UPDATE t9_table SET e=#{a.e}, f1=#{a.f1}, updated_at=now() WHERE id=#{a.id}"},
		{dbType: gobatis.Postgres, prefix: "a.", value: &T9{}, names: []string{"id"}, sql: "UPDATE t9_table SET e=#{a.e}, f1=#{a.f1}, updated_at=now() WHERE id=#{id}"},
	} {
		actaul, err := gobatis.GenerateUpdateSQL(test.dbType, mapper,
			test.prefix, reflect.TypeOf(test.value), test.names, test.argTypes)
		if err != nil {
			t.Error("[", idx, "]", err)
			continue
		}

		if actaul != test.sql {
			t.Error("[", idx, "] excepted is", test.sql)
			t.Error("[", idx, "] actual   is", actaul)
		}
	}

	_, err := gobatis.GenerateUpdateSQL(gobatis.Mysql,
		mapper, "", reflect.TypeOf(&T7{}), []string{}, nil)
	if err == nil {
		t.Error("excepted error got ok")
		return
	}

	_, err = gobatis.GenerateUpdateSQL(gobatis.Mysql,
		mapper, "", reflect.TypeOf(&T12{}), []string{}, nil)
	if err == nil {
		t.Error("excepted error got ok")
		return
	}
}

func TestGenerateUpdateSQL2(t *testing.T) {
	for idx, test := range []struct {
		dbType    gobatis.Dialect
		value     interface{}
		queryType reflect.Type
		query     string
		values    []string
		sql       string
	}{
		{dbType: gobatis.Postgres, value: T1{}, query: "id", values: []string{"f1", "f2", "f3", "deleted_at"}, sql: "UPDATE t1_table SET f1=#{f1}, f2=#{f2,null=true}, f3=#{f3,notnull=true}, updated_at=now() WHERE id=#{id}"},
		{dbType: gobatis.Postgres, value: T1{}, query: "id", values: []string{"f1", "f2", "f3"}, sql: "UPDATE t1_table SET f1=#{f1}, f2=#{f2,null=true}, f3=#{f3,notnull=true}, updated_at=now() WHERE id=#{id}"},
		{dbType: gobatis.Mysql, value: &T1{}, query: "id", values: []string{"f1", "f2", "f3"}, sql: "UPDATE t1_table SET f1=#{f1}, f2=#{f2,null=true}, f3=#{f3,notnull=true}, updated_at=CURRENT_TIMESTAMP WHERE id=#{id}"},
		{dbType: gobatis.Postgres, value: T10{}, query: "id", values: []string{"f1", "f2"}, sql: "UPDATE t10_table SET f_1=#{f1}, f2=#{f2}, updated_at=now() WHERE id=#{id}"},
		{dbType: gobatis.Postgres, value: T10{}, query: "id", values: []string{}, sql: "UPDATE t10_table SET updated_at=now() WHERE id=#{id}"},
		{dbType: gobatis.Mysql, value: T10{}, query: "id", values: []string{"f1", "f2"}, sql: "UPDATE t10_table SET f_1=#{f1}, f2=#{f2}, updated_at=CURRENT_TIMESTAMP WHERE id=#{id}"},
		{dbType: gobatis.Mysql, value: T10{}, query: "id", values: []string{"f1", "f2", "updatedAt"}, sql: "UPDATE t10_table SET f_1=#{f1}, f2=#{f2}, updated_at=CURRENT_TIMESTAMP WHERE id=#{id}"},
		{dbType: gobatis.Mysql, value: T10{}, query: "id", values: []string{}, sql: "UPDATE t10_table SET updated_at=CURRENT_TIMESTAMP WHERE id=#{id}"},
		{dbType: gobatis.Postgres, value: &T10{}, query: "id", values: []string{"f1", "f2"}, sql: "UPDATE t10_table SET f_1=#{f1}, f2=#{f2}, updated_at=now() WHERE id=#{id}"},
		{dbType: gobatis.Postgres, value: T10{}, query: "id", values: []string{"f_1", "f2"}, sql: "UPDATE t10_table SET f_1=#{f_1}, f2=#{f2}, updated_at=now() WHERE id=#{id}"},
		{dbType: gobatis.Postgres, value: &T10{}, query: "id", values: []string{"f_1", "f2"}, sql: "UPDATE t10_table SET f_1=#{f_1}, f2=#{f2}, updated_at=now() WHERE id=#{id}"},

		{dbType: gobatis.Postgres, value: &T10{}, query: "id", queryType: reflect.TypeOf(new(int64)).Elem(), values: []string{"f_1", "f2"}, sql: "UPDATE t10_table SET f_1=#{f_1}, f2=#{f2}, updated_at=now() WHERE id=#{id}"},
		{dbType: gobatis.Postgres, value: &T10{}, query: "id", queryType: reflect.TypeOf(new(sql.NullInt64)).Elem(), values: []string{"f_1", "f2"}, sql: "UPDATE t10_table SET f_1=#{f_1}, f2=#{f2}, updated_at=now() <where><if test=\"id.Valid\"> id=#{id} </if></where>"},
		{dbType: gobatis.Postgres, value: &T10{}, query: "id", queryType: reflect.TypeOf([]int64{}), values: []string{"f_1", "f2"}, sql: "UPDATE t10_table SET f_1=#{f_1}, f2=#{f2}, updated_at=now() WHERE id in (<foreach collection=\"id\" item=\"item\" separator=\",\" >#{item}</foreach>)"},
	} {
		actaul, err := gobatis.GenerateUpdateSQL2(test.dbType,
			mapper, reflect.TypeOf(test.value), test.queryType, test.query, test.values)
		if err != nil {
			t.Error(err)
			continue
		}

		if actaul != test.sql {
			t.Error("[", idx, "] excepted is", test.sql)
			t.Error("[", idx, "] actual   is", actaul)
		}
	}

	_, err := gobatis.GenerateUpdateSQL2(gobatis.Mysql,
		mapper, reflect.TypeOf(&T10{}), nil, "id", []string{"f33"})
	if err == nil {
		t.Error("excepted error got ok")
		return
	}

	_, err = gobatis.GenerateUpdateSQL2(gobatis.Mysql,
		mapper, reflect.TypeOf(&T10{}), nil, "f23", []string{"f33"})
	if err == nil {
		t.Error("excepted error got ok")
		return
	}
}

func TestGenerateDeleteSQL(t *testing.T) {
	for idx, test := range []struct {
		dbType   gobatis.Dialect
		value    interface{}
		names    []string
		argTypes []reflect.Type
		filters  []gobatis.Filter
		sql      string
		err      string
	}{
		{dbType: gobatis.Postgres, value: T1ForNoDeleted{}, sql: "DELETE FROM t1_table"},
		{dbType: gobatis.Postgres, value: &T1ForNoDeleted{}, names: []string{"id"}, sql: "DELETE FROM t1_table WHERE id=#{id}"},
		{dbType: gobatis.Postgres, value: &T1ForNoDeleted{}, names: []string{"id", "f1"}, sql: "DELETE FROM t1_table WHERE id=#{id} AND f1=#{f1}"},

		{dbType: gobatis.Postgres, value: &T1ForNoDeleted{}, names: []string{"id", "f1"},
			argTypes: []reflect.Type{reflect.TypeOf(new(int64)).Elem(), _stringType},
			sql:      "DELETE FROM t1_table WHERE id=#{id} AND f1=#{f1}"},
		{dbType: gobatis.Postgres, value: &T1ForNoDeleted{}, names: []string{"id", "f1"},
			argTypes: []reflect.Type{reflect.TypeOf([]int64{}), _stringType},
			sql:      "DELETE FROM t1_table WHERE id in (<foreach collection=\"id\" item=\"item\" separator=\",\" >#{item}</foreach>) AND f1=#{f1}"},
		{dbType: gobatis.Postgres, value: &T1ForNoDeleted{}, names: []string{"id", "f1"},
			argTypes: []reflect.Type{reflect.TypeOf([]int64{}), reflect.TypeOf([]string{})},
			sql:      "DELETE FROM t1_table WHERE id in (<foreach collection=\"id\" item=\"item\" separator=\",\" >#{item}</foreach>) AND f1 in (<foreach collection=\"f1\" item=\"item\" separator=\",\" >#{item}</foreach>)"},

		{dbType: gobatis.Postgres, value: &T1ForNoDeleted{}, names: []string{"id", "f1"},
			argTypes: []reflect.Type{reflect.TypeOf(new(sql.NullInt64)).Elem(), _stringType},
			sql:      "DELETE FROM t1_table WHERE <if test=\"id.Valid\"> id=#{id} AND </if>f1=#{f1}"},

		{dbType: gobatis.Postgres, value: &T1ForNoDeleted{}, names: []string{"id", "f1"},
			argTypes: []reflect.Type{reflect.TypeOf(new(int64)).Elem(), reflect.TypeOf(new(sql.NullString)).Elem()},
			sql:      "DELETE FROM t1_table WHERE id=#{id}<if test=\"f1.Valid\"> AND f1=#{f1} </if>"},

		{dbType: gobatis.Postgres, value: &T1ForNoDeleted{}, names: []string{"id", "f1"},
			argTypes: []reflect.Type{reflect.TypeOf(new(sql.NullInt64)).Elem(), reflect.TypeOf(new(sql.NullString)).Elem()},
			sql:      "DELETE FROM t1_table <where><if test=\"id.Valid\"> id=#{id} </if><if test=\"f1.Valid\"> AND f1=#{f1} </if></where>"},

		{dbType: gobatis.Postgres, value: T1_1{}, sql: "UPDATE t1_table SET deleted=now() "},
		{dbType: gobatis.Postgres, value: T1{}, sql: "UPDATE t1_table SET deleted_at=now() "},
		{dbType: gobatis.Postgres, value: &T1{}, names: []string{"id"}, sql: "UPDATE t1_table SET deleted_at=now()  WHERE id=#{id}"},
		{dbType: gobatis.Postgres, value: &T1{}, names: []string{"id", "f1"}, sql: "UPDATE t1_table SET deleted_at=now()  WHERE id=#{id} AND f1=#{f1}"},

		{dbType: gobatis.Postgres, value: &T1{}, names: []string{"id", "f1"},
			argTypes: []reflect.Type{reflect.TypeOf(new(int64)).Elem(), _stringType},
			sql:      "UPDATE t1_table SET deleted_at=now()  WHERE id=#{id} AND f1=#{f1}"},
		{dbType: gobatis.Postgres, value: &T1{}, names: []string{"id", "f1"},
			argTypes: []reflect.Type{reflect.TypeOf([]int64{}), _stringType},
			sql:      "UPDATE t1_table SET deleted_at=now()  WHERE id in (<foreach collection=\"id\" item=\"item\" separator=\",\" >#{item}</foreach>) AND f1=#{f1}"},
		{dbType: gobatis.Postgres, value: &T1{}, names: []string{"id", "f1"},
			argTypes: []reflect.Type{reflect.TypeOf([]int64{}), reflect.TypeOf([]string{})},
			sql:      "UPDATE t1_table SET deleted_at=now()  WHERE id in (<foreach collection=\"id\" item=\"item\" separator=\",\" >#{item}</foreach>) AND f1 in (<foreach collection=\"f1\" item=\"item\" separator=\",\" >#{item}</foreach>)"},

		{dbType: gobatis.Postgres, value: &T1{}, names: []string{"id", "f1"},
			argTypes: []reflect.Type{reflect.TypeOf(new(sql.NullInt64)).Elem(), _stringType},
			sql:      "UPDATE t1_table SET deleted_at=now()  WHERE <if test=\"id.Valid\"> id=#{id} AND </if>f1=#{f1}"},

		{dbType: gobatis.Postgres, value: &T1{}, names: []string{"id", "f1"},
			argTypes: []reflect.Type{reflect.TypeOf(new(int64)).Elem(), reflect.TypeOf(new(sql.NullString)).Elem()},
			sql:      "UPDATE t1_table SET deleted_at=now()  WHERE id=#{id}<if test=\"f1.Valid\"> AND f1=#{f1} </if>"},

		{dbType: gobatis.Postgres, value: &T1{}, names: []string{"id", "f1"},
			argTypes: []reflect.Type{reflect.TypeOf(new(sql.NullInt64)).Elem(), reflect.TypeOf(new(sql.NullString)).Elem()},
			sql:      "UPDATE t1_table SET deleted_at=now()  <where><if test=\"id.Valid\"> id=#{id} </if><if test=\"f1.Valid\"> AND f1=#{f1} </if></where>"},

		{dbType: gobatis.Postgres, value: T1{}, names: []string{"force"},
			argTypes: []reflect.Type{reflect.TypeOf(new(bool)).Elem()},
			sql:      `<if test="!force">UPDATE t1_table SET deleted_at=now() </if><if test="force">DELETE FROM t1_table</if>`},

		{dbType: gobatis.Postgres, value: T1{}, names: []string{"force"},
			argTypes: []reflect.Type{reflect.TypeOf(new(sql.NullBool)).Elem()},
			err:      "unsupported type"},

		{dbType: gobatis.Postgres, value: T1{}, names: []string{"id", "force"},
			argTypes: []reflect.Type{reflect.TypeOf(new(int64)).Elem(), reflect.TypeOf(new(bool)).Elem()},
			sql:      `<if test="!force">UPDATE t1_table SET deleted_at=now()  WHERE id=#{id}</if><if test="force">DELETE FROM t1_table WHERE id=#{id}</if>`},

		{dbType: gobatis.Postgres, value: T1{}, names: []string{"id", "force"},
			argTypes: []reflect.Type{reflect.TypeOf(new(int64)).Elem(), reflect.TypeOf(new(sql.NullBool)).Elem()},
			err:      "unsupported type"},

		{dbType: gobatis.Postgres, value: T1ForNoDeleted{}, names: []string{"id"},
			filters: []gobatis.Filter{{Expression: "id>#{id}"}},
			sql:     "DELETE FROM t1_table WHERE id>#{id}"},
		{dbType: gobatis.Postgres, value: &T1ForNoDeleted{}, names: []string{"id", "f1"},
			argTypes: []reflect.Type{reflect.TypeOf(new(int64)).Elem(), _stringType},
			filters:  []gobatis.Filter{{Expression: "id>#{id}"}},
			sql:      "DELETE FROM t1_table WHERE f1=#{f1} AND id>#{id}"},
		{dbType: gobatis.Postgres, value: T1{}, names: []string{"id"},
			filters: []gobatis.Filter{{Expression: "id>#{id}"}},
			sql:     "UPDATE t1_table SET deleted_at=now()  WHERE id>#{id}"},
		{dbType: gobatis.Postgres, value: &T1{}, names: []string{"id", "f1"},
			argTypes: []reflect.Type{reflect.TypeOf(new(int64)).Elem(), _stringType},
			filters:  []gobatis.Filter{{Expression: "id>#{id}"}},
			sql:      "UPDATE t1_table SET deleted_at=now()  WHERE f1=#{f1} AND id>#{id}"},

		{dbType: gobatis.Postgres, value: &T1{},
			filters: []gobatis.Filter{{Expression: "f1 = #{f1}"}, {Expression: "id = #{id}"}},
			sql:     "UPDATE t1_table SET deleted_at=now()  WHERE f1 = #{f1} AND id = #{id}"},

		{dbType: gobatis.MSSql, value: &T1{},
			filters: []gobatis.Filter{{Expression: "f1 = #{f1}"}, {Expression: "id = #{id}"}},
			sql:     "UPDATE t1_table SET deleted_at=CURRENT_TIMESTAMP  WHERE f1 = #{f1} AND id = #{id}"},

		{dbType: gobatis.Postgres, value: &T14{},
			filters: []gobatis.Filter{{Expression: "f1 = #{f1}"}, {Expression: "id = #{id}"}},
			sql:     "DELETE FROM t14_table WHERE f1 = #{f1} AND id = #{id}"},
	} {
		actaul, err := gobatis.GenerateDeleteSQL(test.dbType,
			mapper, reflect.TypeOf(test.value), test.names, test.argTypes, test.filters)
		if err != nil {
			if test.err != "" {
				if strings.Contains(err.Error(), test.err) {
					continue
				}
			}
			t.Error(err)
			continue
		}

		if test.err != "" {
			t.Error("want return error got ok")
			continue
		}

		if actaul != test.sql {
			t.Error("[", idx, "] excepted is", test.sql)
			t.Error("[", idx, "] actual   is", actaul)
		}
	}

	_, err := gobatis.GenerateDeleteSQL(gobatis.Mysql,
		mapper, reflect.TypeOf(&T7{}), []string{}, nil, nil)
	if err == nil {
		t.Error("excepted error got ok")
		return
	}
}

func TestGenerateSelectSQL(t *testing.T) {
	for idx, test := range []struct {
		dbType   gobatis.Dialect
		value    interface{}
		names    []string
		argTypes []reflect.Type
		filters  []gobatis.Filter
		// order    string
		sql string
	}{
		{dbType: gobatis.Postgres, value: T1{}, sql: "SELECT * FROM t1_table WHERE deleted_at IS NULL"},
		{dbType: gobatis.Postgres, value: &T1{}, names: []string{"id"}, sql: "SELECT * FROM t1_table WHERE id=#{id} AND deleted_at IS NULL"},
		{dbType: gobatis.Postgres, value: &T1{}, names: []string{"id", "f1"}, sql: "SELECT * FROM t1_table WHERE id=#{id} AND f1=#{f1} AND deleted_at IS NULL"},
		{dbType: gobatis.Postgres, value: &T1{}, names: []string{"id", "f1", "offset", "limit"},
			sql: "SELECT * FROM t1_table WHERE id=#{id} AND f1=#{f1} AND deleted_at IS NULL <pagination offset=\"offset\" limit=\"limit\" />"},

		{dbType: gobatis.Postgres, value: &T1{}, names: []string{"id", "f1"},
			argTypes: []reflect.Type{reflect.TypeOf(new(int64)).Elem(), _stringType},
			sql:      "SELECT * FROM t1_table WHERE id=#{id} AND f1=#{f1} AND deleted_at IS NULL"},
		{dbType: gobatis.Postgres, value: &T1{}, names: []string{"id", "f1"},
			argTypes: []reflect.Type{reflect.TypeOf([]int64{}), _stringType},
			sql:      "SELECT * FROM t1_table WHERE id in (<foreach collection=\"id\" item=\"item\" separator=\",\" >#{item}</foreach>) AND f1=#{f1} AND deleted_at IS NULL"},
		{dbType: gobatis.Postgres, value: &T1{}, names: []string{"id", "f1"},
			argTypes: []reflect.Type{reflect.TypeOf([]int64{}), reflect.TypeOf([]string{})},
			sql:      "SELECT * FROM t1_table WHERE id in (<foreach collection=\"id\" item=\"item\" separator=\",\" >#{item}</foreach>) AND f1 in (<foreach collection=\"f1\" item=\"item\" separator=\",\" >#{item}</foreach>) AND deleted_at IS NULL"},

		{dbType: gobatis.Postgres, value: &T1{}, names: []string{"id", "f1"},
			argTypes: []reflect.Type{reflect.TypeOf(new(sql.NullInt64)).Elem(), _stringType},
			sql:      "SELECT * FROM t1_table WHERE <if test=\"id.Valid\"> id=#{id} AND </if>f1=#{f1} AND deleted_at IS NULL"},

		{dbType: gobatis.Postgres, value: &T1{}, names: []string{"id", "f1"},
			argTypes: []reflect.Type{reflect.TypeOf(new(int64)).Elem(), reflect.TypeOf(new(sql.NullString)).Elem()},
			sql:      "SELECT * FROM t1_table WHERE id=#{id}<if test=\"f1.Valid\"> AND f1=#{f1} </if> AND deleted_at IS NULL"},

		{dbType: gobatis.Postgres, value: &T1{}, names: []string{"id", "f1"},
			argTypes: []reflect.Type{reflect.TypeOf(new(sql.NullInt64)).Elem(), reflect.TypeOf(new(sql.NullString)).Elem()},
			sql:      "SELECT * FROM t1_table WHERE <if test=\"id.Valid\"> id=#{id} AND </if><if test=\"f1.Valid\"> f1=#{f1} AND </if>deleted_at IS NULL"},

		{dbType: gobatis.Postgres, value: &T1{}, names: []string{"id", "f1", "isDeleted"},
			argTypes: []reflect.Type{reflect.TypeOf(new(sql.NullInt64)).Elem(), reflect.TypeOf(new(sql.NullString)).Elem(), reflect.TypeOf(new(bool)).Elem()},
			sql:      `SELECT * FROM t1_table <where><if test="id.Valid"> id=#{id} AND </if><if test="f1.Valid"> f1=#{f1} AND </if><if test="isDeleted"> deleted_at IS NOT NULL </if><if test="!isDeleted"> deleted_at IS NULL </if></where>`},

		{dbType: gobatis.Postgres, value: &T1{}, names: []string{"id", "f1", "isDeleted"},
			argTypes: []reflect.Type{reflect.TypeOf(new(sql.NullInt64)).Elem(), reflect.TypeOf(new(sql.NullString)).Elem(), reflect.TypeOf(new(sql.NullBool)).Elem()},
			sql:      `SELECT * FROM t1_table <where><if test="id.Valid"> id=#{id} AND </if><if test="f1.Valid"> f1=#{f1} AND </if><if test="isDeleted.Valid"><if test="isDeleted.Bool"> deleted_at IS NOT NULL </if><if test="!isDeleted.Bool"> deleted_at IS NULL </if></if></where>`},

		{dbType: gobatis.Postgres, value: T1ForNoDeleted{}, sql: "SELECT * FROM t1_table"},
		// {dbType: gobatis.Postgres, value: T1ForNoDeleted{}, order: "id ASC", sql: "SELECT * FROM t1_table ORDER BY id ASC"},
		{dbType: gobatis.Postgres, value: &T1ForNoDeleted{}, names: []string{"id"}, sql: "SELECT * FROM t1_table WHERE id=#{id}"},
		{dbType: gobatis.Postgres, value: &T1ForNoDeleted{}, names: []string{"id", "f1"}, sql: "SELECT * FROM t1_table WHERE id=#{id} AND f1=#{f1}"},
		{dbType: gobatis.Postgres, value: &T1ForNoDeleted{}, names: []string{"id", "f1", "offset", "limit"},
			sql: "SELECT * FROM t1_table WHERE id=#{id} AND f1=#{f1} <pagination offset=\"offset\" limit=\"limit\" />"},

		{dbType: gobatis.Postgres, value: &T1ForNoDeleted{}, names: []string{"id", "f1"},
			argTypes: []reflect.Type{reflect.TypeOf(new(int64)).Elem(), _stringType},
			sql:      "SELECT * FROM t1_table WHERE id=#{id} AND f1=#{f1}"},
		{dbType: gobatis.Postgres, value: &T1ForNoDeleted{}, names: []string{"id", "f1"},
			argTypes: []reflect.Type{reflect.TypeOf([]int64{}), _stringType},
			sql:      "SELECT * FROM t1_table WHERE id in (<foreach collection=\"id\" item=\"item\" separator=\",\" >#{item}</foreach>) AND f1=#{f1}"},
		{dbType: gobatis.Postgres, value: &T1ForNoDeleted{}, names: []string{"id", "f1"},
			argTypes: []reflect.Type{reflect.TypeOf([]int64{}), reflect.TypeOf([]string{})},
			sql:      "SELECT * FROM t1_table WHERE id in (<foreach collection=\"id\" item=\"item\" separator=\",\" >#{item}</foreach>) AND f1 in (<foreach collection=\"f1\" item=\"item\" separator=\",\" >#{item}</foreach>)"},

		{dbType: gobatis.Postgres, value: &T1ForNoDeleted{}, names: []string{"id", "f1"},
			argTypes: []reflect.Type{reflect.TypeOf(new(sql.NullInt64)).Elem(), _stringType},
			sql:      "SELECT * FROM t1_table WHERE <if test=\"id.Valid\"> id=#{id} AND </if>f1=#{f1}"},

		{dbType: gobatis.Postgres, value: &T1ForNoDeleted{}, names: []string{"id", "f1"},
			argTypes: []reflect.Type{reflect.TypeOf(new(int64)).Elem(), reflect.TypeOf(new(sql.NullString)).Elem()},
			sql:      "SELECT * FROM t1_table WHERE id=#{id}<if test=\"f1.Valid\"> AND f1=#{f1} </if>"},

		{dbType: gobatis.Postgres, value: &T1ForNoDeleted{}, names: []string{"id", "f1"},
			argTypes: []reflect.Type{reflect.TypeOf(new(sql.NullInt64)).Elem(), reflect.TypeOf(new(sql.NullString)).Elem()},
			sql:      "SELECT * FROM t1_table <where><if test=\"id.Valid\"> id=#{id} </if><if test=\"f1.Valid\"> AND f1=#{f1} </if></where>"},

		{dbType: gobatis.Postgres, value: T1ForNoDeleted{}, names: []string{"id"},
			filters: []gobatis.Filter{{Expression: "id>#{id}"}},
			sql:     "SELECT * FROM t1_table WHERE id>#{id}"},
		{dbType: gobatis.Postgres, value: &T1ForNoDeleted{}, names: []string{"id", "f1"},
			argTypes: []reflect.Type{reflect.TypeOf(new(int64)).Elem(), _stringType},
			filters:  []gobatis.Filter{{Expression: "id>#{id}"}},
			sql:      "SELECT * FROM t1_table WHERE f1=#{f1} AND id>#{id}"},
		{dbType: gobatis.Postgres, value: T1{}, names: []string{"id"},
			filters: []gobatis.Filter{{Expression: "id>#{id}"}},
			sql:     "SELECT * FROM t1_table WHERE id>#{id} AND deleted_at IS NULL"},
		{dbType: gobatis.Postgres, value: &T1{}, names: []string{"id", "f1"},
			argTypes: []reflect.Type{reflect.TypeOf(new(int64)).Elem(), _stringType},
			filters:  []gobatis.Filter{{Expression: "id>#{id}"}},
			sql:      "SELECT * FROM t1_table WHERE f1=#{f1} AND id>#{id} AND deleted_at IS NULL"},

		{dbType: gobatis.Postgres,
			value:    T1ForNoDeleted{},
			names:    []string{"offset", "limit"},
			argTypes: []reflect.Type{reflect.TypeOf(new(int64)).Elem(), reflect.TypeOf(new(int64)).Elem()},
			sql:      `SELECT * FROM t1_table <pagination offset="offset" limit="limit" />`},

		{dbType: gobatis.Postgres,
			value:    T13{},
			names:    []string{"f1", "f2"},
			argTypes: []reflect.Type{_stringType, reflect.TypeOf(new(int64)).Elem()},
			sql:      `SELECT * FROM t13_table <where><if test="isNotEmptyString(f1, true)"> f1=#{f1} </if><if test="f2 != 0"> AND f2=#{f2} </if></where>`},

		{dbType: gobatis.Postgres, value: &T1{},
			names:    []string{"sortBy"},
			argTypes: []reflect.Type{_stringType},
			filters:  []gobatis.Filter{{Expression: "f1 = #{f1}"}},
			sql:      "SELECT * FROM t1_table WHERE deleted_at IS NULL AND f1 = #{f1} <order_by by=\"sortBy\"/>"},

		{dbType: gobatis.Postgres, value: &T1{},
			filters: []gobatis.Filter{{Expression: "f1 = #{f1}"}},
			sql:     "SELECT * FROM t1_table WHERE deleted_at IS NULL AND f1 = #{f1}"},

		{dbType: gobatis.Postgres, value: &T14{},
			filters: []gobatis.Filter{{Expression: "f1 = #{f1}"}, {Expression: "id = #{id}"}},
			sql:     "SELECT * FROM t14_table WHERE f1 = #{f1} AND id = #{id}"},

		{dbType: gobatis.Postgres,
			value: Worklog{},
			names: []string{"planID", "userID", "descriptionLike", "createdAt"},
			argTypes: []reflect.Type{
				reflect.TypeOf(sql.NullInt64{}),
				reflect.TypeOf(sql.NullInt64{}),
				_stringType,
				reflect.TypeOf(TimeRange{}),
			},
			sql: `SELECT * FROM worklogs WHERE <if test="planID.Valid"> plan_id=#{planID} </if><if test="userID.Valid"> AND user_id=#{userID} </if><if test="isNotEmptyString(descriptionLike, true)">  AND description like <like value="descriptionLike" /> AND </if>  <value-range field="created_at" value="createdAt" />`,
		},
		{
			dbType: gobatis.Postgres,
			value:  Worklog{},
			names:  []string{"planID", "createdAt", "userID"},
			argTypes: []reflect.Type{
				reflect.TypeOf(sql.NullInt64{}),
				reflect.TypeOf(TimeRange{}),
				reflect.TypeOf(sql.NullInt64{}),
			},

			sql: `SELECT * FROM worklogs WHERE <if test="planID.Valid"> plan_id=#{planID} AND </if> <value-range field="created_at" value="createdAt" /><if test="userID.Valid"> AND user_id=#{userID} </if>`,
		},
		{
			dbType: gobatis.Postgres,
			value:  Worklog{},
			names:  []string{"createdAt", "planID", "userID"},
			argTypes: []reflect.Type{
				reflect.TypeOf(TimeRange{}),
				reflect.TypeOf(sql.NullInt64{}),
				reflect.TypeOf(sql.NullInt64{}),
			},
			sql: `SELECT * FROM worklogs WHERE  <value-range field="created_at" value="createdAt" /><if test="planID.Valid"> AND plan_id=#{planID} </if><if test="userID.Valid"> AND user_id=#{userID} </if>`,
		},

		{dbType: gobatis.Postgres, value: T1{}, names: []string{"id", "isDeleted"},
			argTypes: []reflect.Type{reflect.TypeOf(new(int64)).Elem(), reflect.TypeOf(new(sql.NullBool)).Elem()},
			sql:      `SELECT * FROM t1_table WHERE id=#{id}<if test="isDeleted.Valid"><if test="isDeleted.Bool"> AND deleted_at IS NOT NULL </if><if test="!isDeleted.Bool"> AND deleted_at IS NULL </if></if>`},

		{dbType: gobatis.Postgres, value: T1{}, names: []string{"f3", "isDeleted"},
			argTypes: []reflect.Type{reflect.TypeOf(new(int)).Elem(), reflect.TypeOf(new(sql.NullBool)).Elem()},
			sql:      `SELECT * FROM t1_table <where><if test="f3 != 0"> f3=#{f3} AND </if><if test="isDeleted.Valid"><if test="isDeleted.Bool"> deleted_at IS NOT NULL </if><if test="!isDeleted.Bool"> deleted_at IS NULL </if></if></where>`},
	} {

		actaul, err := gobatis.GenerateSelectSQL(test.dbType,
			mapper, reflect.TypeOf(test.value), test.names, test.argTypes, test.filters)
		if err != nil {
			t.Error(err)
			continue
		}

		if actaul != test.sql {
			t.Error("[", idx, "] excepted is", test.sql)
			t.Error("[", idx, "] actual   is", actaul)
		}
	}

	_, err := gobatis.GenerateSelectSQL(gobatis.Mysql,
		mapper, reflect.TypeOf(&T7{}), []string{}, nil, nil)
	if err == nil {
		t.Error("excepted error got ok")
		return
	}
}

func TestGenerateCountSQL(t *testing.T) {
	for idx, test := range []struct {
		id       string
		dbType   gobatis.Dialect
		value    interface{}
		names    []string
		argTypes []reflect.Type
		filters  []gobatis.Filter
		sql      string
	}{
		{id: "0", dbType: gobatis.Postgres, value: T1{}, sql: "SELECT count(*) FROM t1_table WHERE deleted_at IS NULL"},
		{id: "1", dbType: gobatis.Postgres, value: &T1{}, names: []string{"id"}, sql: "SELECT count(*) FROM t1_table WHERE id=#{id} AND deleted_at IS NULL"},
		{dbType: gobatis.Postgres, value: &T1{}, names: []string{"id", "f1"}, sql: "SELECT count(*) FROM t1_table WHERE id=#{id} AND f1=#{f1} AND deleted_at IS NULL"},
		{dbType: gobatis.Postgres, value: &T1{}, names: []string{"f1Like"}, sql: "SELECT count(*) FROM t1_table WHERE <if test=\"isNotEmptyString(f1Like, true)\"> f1 like <like value=\"f1Like\" /> AND </if> deleted_at IS NULL"},
		{dbType: gobatis.Postgres, value: &T1{}, names: []string{"id", "f1"},
			argTypes: []reflect.Type{reflect.TypeOf(new(int64)).Elem(), _stringType},
			sql:      "SELECT count(*) FROM t1_table WHERE id=#{id} AND f1=#{f1} AND deleted_at IS NULL"},
		{dbType: gobatis.Postgres, value: &T1{}, names: []string{"id", "f1"},
			argTypes: []reflect.Type{reflect.TypeOf([]int64{}), _stringType},
			sql:      "SELECT count(*) FROM t1_table WHERE id in (<foreach collection=\"id\" item=\"item\" separator=\",\" >#{item}</foreach>) AND f1=#{f1} AND deleted_at IS NULL"},
		{dbType: gobatis.Postgres, value: &T1{}, names: []string{"id", "f1"},
			argTypes: []reflect.Type{reflect.TypeOf([]int64{}), reflect.TypeOf([]string{})},
			sql:      "SELECT count(*) FROM t1_table WHERE id in (<foreach collection=\"id\" item=\"item\" separator=\",\" >#{item}</foreach>) AND f1 in (<foreach collection=\"f1\" item=\"item\" separator=\",\" >#{item}</foreach>) AND deleted_at IS NULL"},

		{dbType: gobatis.Postgres, value: &T1{}, names: []string{"id", "f1"},
			argTypes: []reflect.Type{reflect.TypeOf(new(sql.NullInt64)).Elem(), _stringType},
			sql:      "SELECT count(*) FROM t1_table WHERE <if test=\"id.Valid\"> id=#{id} AND </if>f1=#{f1} AND deleted_at IS NULL"},

		{dbType: gobatis.Postgres, value: &T1{}, names: []string{"id", "f1"},
			argTypes: []reflect.Type{reflect.TypeOf(new(int64)).Elem(), reflect.TypeOf(new(sql.NullString)).Elem()},
			sql:      "SELECT count(*) FROM t1_table WHERE id=#{id}<if test=\"f1.Valid\"> AND f1=#{f1} </if> AND deleted_at IS NULL"},

		{dbType: gobatis.Postgres, value: &T1{}, names: []string{"id", "f1"},
			argTypes: []reflect.Type{reflect.TypeOf(new(sql.NullInt64)).Elem(), reflect.TypeOf(new(sql.NullString)).Elem()},
			sql:      "SELECT count(*) FROM t1_table WHERE <if test=\"id.Valid\"> id=#{id} AND </if><if test=\"f1.Valid\"> f1=#{f1} AND </if>deleted_at IS NULL"},

		{dbType: gobatis.Postgres, value: &T1{}, names: []string{"f1Like"},
			argTypes: []reflect.Type{_stringType},
			sql:      "SELECT count(*) FROM t1_table WHERE <if test=\"isNotEmptyString(f1Like, true)\"> f1 like <like value=\"f1Like\" /> AND </if> deleted_at IS NULL"},

		{dbType: gobatis.Postgres, value: &T1{}, names: []string{"f1Like"},
			argTypes: []reflect.Type{reflect.TypeOf(new(sql.NullString)).Elem()},
			sql:      "SELECT count(*) FROM t1_table WHERE <if test=\"f1Like.Valid\"> f1 like <like value=\"f1Like\" /> AND </if>deleted_at IS NULL"},

		{dbType: gobatis.Postgres, value: &T1{}, names: []string{"f1Like"},
			argTypes: []reflect.Type{reflect.TypeOf(new(struct{ sql.NullString })).Elem()},
			sql:      "SELECT count(*) FROM t1_table WHERE <if test=\"f1Like.Valid\"> f1 like <like value=\"f1Like\" /> AND </if>deleted_at IS NULL"},

		{dbType: gobatis.Postgres, value: &T1{}, names: []string{"created_at"},
			argTypes: []reflect.Type{reflect.TypeOf(struct {
				Start, End time.Time
			}{})},
			sql: "SELECT count(*) FROM t1_table WHERE  <value-range field=\"created_at\" value=\"created_at\" /> AND deleted_at IS NULL"},

		{dbType: gobatis.Postgres, value: &T11{}, names: []string{"f1"},
			argTypes: []reflect.Type{_stringType},
			sql:      "SELECT count(*) FROM t11_table WHERE #{f1} = ANY (f_1)"},

		{dbType: gobatis.Postgres, value: &T12{}, names: []string{"f1"},
			argTypes: []reflect.Type{_stringType},
			sql:      "SELECT count(*) FROM t12_table WHERE f_1 @> #{f1}"},

		{dbType: gobatis.Postgres, value: T1ForNoDeleted{}, sql: "SELECT count(*) FROM t1_table"},
		{dbType: gobatis.Postgres, value: &T1ForNoDeleted{}, names: []string{"id"}, sql: "SELECT count(*) FROM t1_table WHERE id=#{id}"},
		{dbType: gobatis.Postgres, value: &T1ForNoDeleted{}, names: []string{"id", "f1"}, sql: "SELECT count(*) FROM t1_table WHERE id=#{id} AND f1=#{f1}"},

		{dbType: gobatis.Postgres, value: &T1ForNoDeleted{}, names: []string{"f1Like"},
			argTypes: []reflect.Type{_stringType},
			sql:      "SELECT count(*) FROM t1_table <where><if test=\"isNotEmptyString(f1Like, true)\"> f1 like <like value=\"f1Like\" /> </if> </where>"},

		{dbType: gobatis.Postgres, value: &T14{}, names: []string{"f1Like"},
			argTypes: []reflect.Type{_stringType},
			sql:      "SELECT count(*) FROM t14_table <where><if test=\"isNotEmptyString(f1Like, true)\"> f1 like <like value=\"f1Like\" /> </if></where>"},
		/////////////////////////

		{dbType: gobatis.Postgres, value: &T1ForNoDeleted{}, names: []string{"id", "f1"},
			argTypes: []reflect.Type{reflect.TypeOf(new(int64)).Elem(), _stringType},
			sql:      "SELECT count(*) FROM t1_table WHERE id=#{id} AND f1=#{f1}"},
		{dbType: gobatis.Postgres, value: &T1ForNoDeleted{}, names: []string{"id", "f1"},
			argTypes: []reflect.Type{reflect.TypeOf([]int64{}), _stringType},
			sql:      "SELECT count(*) FROM t1_table WHERE id in (<foreach collection=\"id\" item=\"item\" separator=\",\" >#{item}</foreach>) AND f1=#{f1}"},
		{dbType: gobatis.Postgres, value: &T1ForNoDeleted{}, names: []string{"id", "f1"},
			argTypes: []reflect.Type{reflect.TypeOf([]int64{}), reflect.TypeOf([]string{})},
			sql:      "SELECT count(*) FROM t1_table WHERE id in (<foreach collection=\"id\" item=\"item\" separator=\",\" >#{item}</foreach>) AND f1 in (<foreach collection=\"f1\" item=\"item\" separator=\",\" >#{item}</foreach>)"},

		{dbType: gobatis.Postgres, value: &T1ForNoDeleted{}, names: []string{"id", "f1"},
			argTypes: []reflect.Type{reflect.TypeOf(new(sql.NullInt64)).Elem(), _stringType},
			sql:      "SELECT count(*) FROM t1_table WHERE <if test=\"id.Valid\"> id=#{id} AND </if>f1=#{f1}"},

		{dbType: gobatis.Postgres, value: &T1ForNoDeleted{}, names: []string{"id", "f1"},
			argTypes: []reflect.Type{reflect.TypeOf(new(int64)).Elem(), reflect.TypeOf(new(sql.NullString)).Elem()},
			sql:      "SELECT count(*) FROM t1_table WHERE id=#{id}<if test=\"f1.Valid\"> AND f1=#{f1} </if>"},

		{dbType: gobatis.Postgres, value: &T1ForNoDeleted{}, names: []string{"id", "f1"},
			argTypes: []reflect.Type{reflect.TypeOf(new(sql.NullInt64)).Elem(), reflect.TypeOf(new(sql.NullString)).Elem()},
			sql:      "SELECT count(*) FROM t1_table <where><if test=\"id.Valid\"> id=#{id} </if><if test=\"f1.Valid\"> AND f1=#{f1} </if></where>"},

		{dbType: gobatis.Postgres, value: &T1ForNoDeleted{}, names: []string{"f1Like"},
			argTypes: []reflect.Type{reflect.TypeOf(new(sql.NullString)).Elem()},
			sql:      "SELECT count(*) FROM t1_table <where><if test=\"f1Like.Valid\"> f1 like <like value=\"f1Like\" /> </if></where>"},

		{dbType: gobatis.Postgres, value: &T1ForNoDeleted{}, names: []string{"f1Like"},
			argTypes: []reflect.Type{reflect.TypeOf(new(struct{ sql.NullString })).Elem()},
			sql:      "SELECT count(*) FROM t1_table <where><if test=\"f1Like.Valid\"> f1 like <like value=\"f1Like\" /> </if></where>"},

		{dbType: gobatis.Postgres, value: &T1ForNoDeleted{}, names: []string{"created_at"},
			argTypes: []reflect.Type{reflect.TypeOf(struct {
				Start, End time.Time
			}{})},
			sql: "SELECT count(*) FROM t1_table WHERE  <value-range field=\"created_at\" value=\"created_at\" />"},

		{dbType: gobatis.Postgres, value: T1ForNoDeleted{}, names: []string{"id"},
			filters: []gobatis.Filter{{Expression: "id>#{id}"}},
			sql:     "SELECT count(*) FROM t1_table WHERE id>#{id}"},
		{dbType: gobatis.Postgres, value: &T1ForNoDeleted{}, names: []string{"id", "f1"},
			argTypes: []reflect.Type{reflect.TypeOf(new(int64)).Elem(), _stringType},
			filters:  []gobatis.Filter{{Expression: "id>#{id}"}},
			sql:      "SELECT count(*) FROM t1_table WHERE f1=#{f1} AND id>#{id}"},
		{dbType: gobatis.Postgres, value: T1{}, names: []string{"id"},
			filters: []gobatis.Filter{{Expression: "id>#{id}"}},
			sql:     "SELECT count(*) FROM t1_table WHERE id>#{id} AND deleted_at IS NULL"},
		{dbType: gobatis.Postgres, value: &T1{}, names: []string{"id", "f1"},
			argTypes: []reflect.Type{reflect.TypeOf(new(int64)).Elem(), _stringType},
			filters:  []gobatis.Filter{{Expression: "id>#{id}"}},
			sql:      "SELECT count(*) FROM t1_table WHERE f1=#{f1} AND id>#{id} AND deleted_at IS NULL"},

		// test:  <if/> and xxx
		{dbType: gobatis.Postgres, value: &T1ForNoDeleted{}, names: []string{"f3", "f1"},
			argTypes: []reflect.Type{_stringType, _stringType},
			sql:      "SELECT count(*) FROM t1_table WHERE <if test=\"f3 != 0\"> f3=#{f3} AND </if>f1=#{f1}"},

		// test:  xxx and <if/> and xxx
		{dbType: gobatis.Postgres, value: &T1ForNoDeleted{}, names: []string{"id", "f3", "f1"},
			argTypes: []reflect.Type{reflect.TypeOf(new(int64)).Elem(), _stringType, _stringType},
			sql:      "SELECT count(*) FROM t1_table WHERE id=#{id}<if test=\"f3 != 0\"> AND f3=#{f3} </if> AND f1=#{f1}"},

		// test:  <if/> and xxx (xxx is deleted)
		{dbType: gobatis.Postgres, value: &T1{}, names: []string{"id"},
			argTypes: []reflect.Type{reflect.TypeOf(new(sql.NullInt64)).Elem()},
			sql:      "SELECT count(*) FROM t1_table WHERE <if test=\"id.Valid\"> id=#{id} AND </if>deleted_at IS NULL"},

		// test:  xxx <if/> and xxx (xxx is deleted)
		{dbType: gobatis.Postgres, value: &T1{}, names: []string{"id", "f1"},
			argTypes: []reflect.Type{reflect.TypeOf(new(int64)).Elem(), reflect.TypeOf(new(sql.NullString)).Elem()},
			sql:      "SELECT count(*) FROM t1_table WHERE id=#{id}<if test=\"f1.Valid\"> AND f1=#{f1} </if> AND deleted_at IS NULL"},

		{dbType: gobatis.Postgres, value: &T1{},
			filters: []gobatis.Filter{{Expression: "f1 = #{f1}"}},
			sql:     "SELECT count(*) FROM t1_table WHERE deleted_at IS NULL AND f1 = #{f1}"},

		{dbType: gobatis.Postgres, value: &T14{},
			filters: []gobatis.Filter{{Expression: "f1 = #{f1}"}, {Expression: "id = #{id}"}},
			sql:     "SELECT count(*) FROM t14_table WHERE f1 = #{f1} AND id = #{id}"},
	} {
		fmt.Println("test", idx)
		actaul, err := gobatis.GenerateCountSQL(test.dbType,
			mapper, reflect.TypeOf(test.value), test.names, test.argTypes, test.filters)
		if err != nil {
			t.Error(test.id, err)
			continue
		}

		if actaul != test.sql {
			t.Error("[", test.id, idx, "] excepted is", test.sql)
			t.Error("[", test.id, idx, "] actual   is", actaul)
		}
	}

	_, err := gobatis.GenerateCountSQL(gobatis.Mysql,
		mapper, reflect.TypeOf(&T7{}), []string{}, nil, nil)
	if err == nil {
		t.Error("excepted error got ok")
		return
	}

	_, err = gobatis.GenerateCountSQL(gobatis.Postgres,
		mapper, reflect.TypeOf(&T1{}), []string{"f1Like"},
		[]reflect.Type{reflect.TypeOf([]string{})}, nil)
	if err == nil {
		t.Error("excepted error got ok")
		return
	}
	_, err = gobatis.GenerateCountSQL(gobatis.Postgres,
		mapper, reflect.TypeOf(&T1{}), []string{"f2Like"},
		[]reflect.Type{reflect.TypeOf(new(int64)).Elem()}, nil)
	if err == nil {
		t.Error("excepted error got ok")
		return
	}

	_, err = gobatis.GenerateCountSQL(gobatis.Postgres,
		mapper, reflect.TypeOf(&T14{}), []string{"f1Like"},
		[]reflect.Type{reflect.TypeOf(new(int64)).Elem()}, nil)
	if err == nil {
		t.Error("excepted error got ok")
		return
	}
}

func TestIsValueRange(t *testing.T) {
	if !gobatis.IsValueRange(reflect.TypeOf(struct {
		Start, End time.Time
	}{})) {
		t.Error("want true got false")
		return
	}
	if !gobatis.IsValueRange(reflect.TypeOf(struct {
		Range      struct{}
		Start, End time.Time
	}{})) {
		t.Error("want true got false")
		return
	}
	if gobatis.IsValueRange(reflect.TypeOf(struct {
		Range struct {
			A int64
		}
		Start, End time.Time
	}{})) {
		t.Error("want true got false")
		return
	}
	if gobatis.IsValueRange(reflect.TypeOf(struct {
		abc        struct{}
		Start, End time.Time
	}{})) {
		t.Error("want true got false")
		return
	}
	if !gobatis.IsValueRange(reflect.TypeOf(struct {
		Start, End int64
	}{})) {
		t.Error("want true got false")
	}
	if gobatis.IsValueRange(reflect.TypeOf(struct {
		Start time.Time
		End   int64
	}{})) {
		t.Error("want true got false")
		return
	}
	if gobatis.IsValueRange(reflect.TypeOf(struct {
		Start int64
		End   time.Time
	}{})) {
		t.Error("want true got false")
		return
	}
	if gobatis.IsValueRange(reflect.TypeOf(struct {
		End time.Time
	}{})) {
		t.Error("want true got false")
		return
	}
	if gobatis.IsValueRange(reflect.TypeOf(struct {
		a   int
		End time.Time
	}{})) {
		t.Error("want true got false")
		return
	}
	if gobatis.IsValueRange(reflect.TypeOf(struct {
		a     int
		Start time.Time
	}{})) {
		t.Error("want true got false")
		return
	}
}
