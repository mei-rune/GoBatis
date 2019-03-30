package gobatis_test

import (
	"database/sql"
	"reflect"
	"strings"
	"testing"
	"time"

	gobatis "github.com/runner-mei/GoBatis"
)

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

func TestGenerateUpsertSQL(t *testing.T) {
	for idx, test := range []struct {
		dbType   gobatis.Dialect
		value    interface{}
		names    []string
		argTypes []reflect.Type
		noReturn bool
		sql      string
	}{
		{dbType: gobatis.DbTypePostgres, value: T16{}, sql: "INSERT INTO t16_table(f1, f2, f3, created_at, updated_at) VALUES(#{f1}, #{f2}, #{f3}, now(), now()) ON CONFLICT (f1) DO UPDATE SET f2=EXCLUDED.f2, f3=EXCLUDED.f3, created_at=EXCLUDED.created_at, updated_at=EXCLUDED.updated_at RETURNING id"},
		{dbType: gobatis.DbTypeMysql, value: T16{}, sql: "INSERT INTO t16_table(f1, f2, f3, created_at, updated_at) VALUES(#{f1}, #{f2}, #{f3}, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP) ON DUPLICATE KEY UPDATE f2=VALUES(f2), f3=VALUES(f3), created_at=VALUES(created_at), updated_at=VALUES(updated_at)"},
		{dbType: gobatis.DbTypeMSSql, value: T16{}, sql: `MERGE INTO t16_table AS t USING ( VALUES(#{f1}, #{f2}, #{f3}, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP ) AS s (f1, f2, f3, created_at, updated_at ) ON t.f1 = s.f1 WHEN MATCHED THEN UPDATE SET f2 = s.f2, f3 = s.f3, created_at = s.created_at, updated_at = s.updated_at WHEN NOT MATCHED THEN INSERT (f1, f2, f3, created_at, updated_at) VALUES(s.f1, s.f2, s.f3, s.created_at, s.updated_at)  OUTPUT inserted.id`},
	} {
		actaul, err := gobatis.GenerateUpsertSQL(test.dbType, mapper, reflect.TypeOf(test.value), nil, nil, false)
		if err != nil {
			t.Error("[", idx, "]", err)
			continue
		}

		if actaul != test.sql {
			t.Error("[", idx, "] excepted is", test.sql)
			t.Error("[", idx, "] actual   is", actaul)
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
		{dbType: gobatis.DbTypePostgres, value: T1{}, sql: "INSERT INTO t1_table(f1, f2, f3, created_at, updated_at) VALUES(#{f1}, #{f2}, #{f3}, now(), now()) RETURNING id"},
		{dbType: gobatis.DbTypePostgres, value: &T1{}, sql: "INSERT INTO t1_table(f1, f2, f3, created_at, updated_at) VALUES(#{f1}, #{f2}, #{f3}, now(), now()) RETURNING id"},
		{dbType: gobatis.DbTypePostgres, value: T2{}, sql: "INSERT INTO t2_table(f1, f2, created_at, updated_at) VALUES(#{f1}, #{f2}, now(), now()) RETURNING id"},
		{dbType: gobatis.DbTypePostgres, value: &T2{}, sql: "INSERT INTO t2_table(f1, f2, created_at, updated_at) VALUES(#{f1}, #{f2}, now(), now()) RETURNING id"},
		{dbType: gobatis.DbTypePostgres, value: T3{}, sql: "INSERT INTO t3_table(f1, f2, created_at, updated_at) VALUES(#{f1}, #{f2}, now(), now()) RETURNING id"},
		{dbType: gobatis.DbTypePostgres, value: &T3{}, sql: "INSERT INTO t3_table(f1, f2, created_at, updated_at) VALUES(#{f1}, #{f2}, now(), now()) RETURNING id"},
		{dbType: gobatis.DbTypePostgres, value: T4{}, sql: "INSERT INTO t2_table(f3, f4, f1, f2, created_at, updated_at) VALUES(#{f3}, #{f4}, #{f1}, #{f2}, now(), now()) RETURNING id"},
		{dbType: gobatis.DbTypePostgres, value: &T4{}, sql: "INSERT INTO t2_table(f3, f4, f1, f2, created_at, updated_at) VALUES(#{f3}, #{f4}, #{f1}, #{f2}, now(), now()) RETURNING id"},
		{dbType: gobatis.DbTypePostgres, value: T4{}, sql: "INSERT INTO t2_table(f3, f4, f1, f2, created_at, updated_at) VALUES(#{f3}, #{f4}, #{f1}, #{f2}, now(), now())", noReturn: true},
		{dbType: gobatis.DbTypePostgres, value: &T4{}, sql: "INSERT INTO t2_table(f3, f4, f1, f2, created_at, updated_at) VALUES(#{f3}, #{f4}, #{f1}, #{f2}, now(), now())", noReturn: true},
		{dbType: gobatis.DbTypeMysql, value: T4{}, sql: "INSERT INTO t2_table(f3, f4, f1, f2, created_at, updated_at) VALUES(#{f3}, #{f4}, #{f1}, #{f2}, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)"},
		{dbType: gobatis.DbTypeMysql, value: &T4{}, sql: "INSERT INTO t2_table(f3, f4, f1, f2, created_at, updated_at) VALUES(#{f3}, #{f4}, #{f1}, #{f2}, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)"},
		{dbType: gobatis.DbTypePostgres, value: T8{}, sql: "INSERT INTO t8_table(f1, f2, created_at, updated_at) VALUES(#{f1}, #{f2}, now(), now()) RETURNING id"},
		{dbType: gobatis.DbTypePostgres, value: &T8{}, sql: "INSERT INTO t8_table(f1, f2, created_at, updated_at) VALUES(#{f1}, #{f2}, now(), now()) RETURNING id"},
		{dbType: gobatis.DbTypePostgres, value: T9{}, sql: "INSERT INTO t9_table(e, f1, f2, created_at, updated_at) VALUES(#{e}, #{f1}, #{f2}, now(), now()) RETURNING id"},
		{dbType: gobatis.DbTypePostgres, value: &T9{}, sql: "INSERT INTO t9_table(e, f1, f2, created_at, updated_at) VALUES(#{e}, #{f1}, #{f2}, now(), now()) RETURNING id"},
		{dbType: gobatis.DbTypeMSSql, value: &T4{}, sql: "INSERT INTO t2_table(f3, f4, f1, f2, created_at, updated_at) OUTPUT inserted.id VALUES(#{f3}, #{f4}, #{f1}, #{f2}, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)"},
		{dbType: gobatis.DbTypeMSSql, value: &T4{}, sql: "INSERT INTO t2_table(f3, f4, f1, f2, created_at, updated_at) VALUES(#{f3}, #{f4}, #{f1}, #{f2}, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)", noReturn: true},

		{dbType: gobatis.DbTypeMSSql, value: &T14{},
			sql:      "INSERT INTO t14_table(f1) VALUES(#{f1,notnull=true})",
			names:    []string{"f1"},
			argTypes: []reflect.Type{reflect.TypeOf(new(string)).Elem()},
			noReturn: true},
		{dbType: gobatis.DbTypeMSSql, value: &T14{},
			sql:      "INSERT INTO t14_table(f1) VALUES(#{f1,notnull=true})",
			names:    []string{"f1"},
			argTypes: []reflect.Type{reflect.TypeOf(new(int)).Elem()},
			noReturn: true},
		{dbType: gobatis.DbTypeMSSql, value: &T14{},
			sql:      "INSERT INTO t14_table(f1) VALUES(#{f1,notnull=true})",
			names:    []string{"f1"},
			argTypes: []reflect.Type{nil},
			noReturn: true},

		{dbType: gobatis.DbTypeMSSql, value: &T14{},
			sql:      "INSERT INTO t14_table(f1) VALUES(#{f1.f1})",
			names:    []string{"f1"},
			argTypes: []reflect.Type{reflect.TypeOf(new(T14)).Elem()},
			noReturn: true},
		{dbType: gobatis.DbTypeMSSql, value: &T15{},
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

	_, err := gobatis.GenerateInsertSQL(gobatis.DbTypeMysql,
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
		{dbType: gobatis.DbTypePostgres, value: T1{}, fields: []string{"f1", "f2", "f3"}, sql: "INSERT INTO t1_table(f1, f2, f3, created_at, updated_at) VALUES(#{f1}, #{f2,null=true}, #{f3,notnull=true}, now(), now()) RETURNING id"},
		{dbType: gobatis.DbTypePostgres, value: &T1{}, fields: []string{"f1", "f2", "f3"}, sql: "INSERT INTO t1_table(f1, f2, f3, created_at, updated_at) VALUES(#{f1}, #{f2,null=true}, #{f3,notnull=true}, now(), now()) RETURNING id"},
		{dbType: gobatis.DbTypePostgres, value: T2{}, fields: []string{"f1", "f2"}, sql: "INSERT INTO t2_table(f1, f2, created_at, updated_at) VALUES(#{f1}, #{f2}, now(), now()) RETURNING id"},
		{dbType: gobatis.DbTypePostgres, value: &T2{}, fields: []string{"f1", "f2"}, sql: "INSERT INTO t2_table(f1, f2, created_at, updated_at) VALUES(#{f1}, #{f2}, now(), now()) RETURNING id"},
		{dbType: gobatis.DbTypePostgres, value: T3{}, fields: []string{"f1", "f2"}, sql: "INSERT INTO t3_table(f1, f2, created_at, updated_at) VALUES(#{f1}, #{f2}, now(), now()) RETURNING id"},
		{dbType: gobatis.DbTypePostgres, value: &T3{}, fields: []string{"f1", "f2"}, sql: "INSERT INTO t3_table(f1, f2, created_at, updated_at) VALUES(#{f1}, #{f2}, now(), now()) RETURNING id"},
		{dbType: gobatis.DbTypePostgres, value: T4{}, fields: []string{"f1", "f2", "f3", "f4"}, sql: "INSERT INTO t2_table(f3, f4, f1, f2, created_at, updated_at) VALUES(#{f3}, #{f4}, #{f1}, #{f2}, now(), now()) RETURNING id"},
		{dbType: gobatis.DbTypePostgres, value: &T4{}, fields: []string{"f1", "f2", "f3", "f4"}, sql: "INSERT INTO t2_table(f3, f4, f1, f2, created_at, updated_at) VALUES(#{f3}, #{f4}, #{f1}, #{f2}, now(), now()) RETURNING id"},
		{dbType: gobatis.DbTypePostgres, value: T4{}, fields: []string{"f1", "f2", "f3", "f4"}, sql: "INSERT INTO t2_table(f3, f4, f1, f2, created_at, updated_at) VALUES(#{f3}, #{f4}, #{f1}, #{f2}, now(), now())", noReturn: true},
		{dbType: gobatis.DbTypePostgres, value: &T4{}, fields: []string{"f1", "f2", "f3", "f4"}, sql: "INSERT INTO t2_table(f3, f4, f1, f2, created_at, updated_at) VALUES(#{f3}, #{f4}, #{f1}, #{f2}, now(), now())", noReturn: true},
		{dbType: gobatis.DbTypeMysql, value: T4{}, fields: []string{"f1", "f2", "f3", "f4"}, sql: "INSERT INTO t2_table(f3, f4, f1, f2, created_at, updated_at) VALUES(#{f3}, #{f4}, #{f1}, #{f2}, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)"},
		{dbType: gobatis.DbTypeMysql, value: &T4{}, fields: []string{"f1", "f2", "f3", "f4"}, sql: "INSERT INTO t2_table(f3, f4, f1, f2, created_at, updated_at) VALUES(#{f3}, #{f4}, #{f1}, #{f2}, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)"},
		{dbType: gobatis.DbTypePostgres, value: T8{}, fields: []string{"f1", "f2"}, sql: "INSERT INTO t8_table(f1, f2, created_at, updated_at) VALUES(#{f1}, #{f2}, now(), now()) RETURNING id"},
		{dbType: gobatis.DbTypePostgres, value: &T8{}, fields: []string{"f1", "f2"}, sql: "INSERT INTO t8_table(f1, f2, created_at, updated_at) VALUES(#{f1}, #{f2}, now(), now()) RETURNING id"},
		{dbType: gobatis.DbTypePostgres, value: T9{}, fields: []string{"f1", "f2", "e"}, sql: "INSERT INTO t9_table(e, f1, f2, created_at, updated_at) VALUES(#{e}, #{f1}, #{f2}, now(), now()) RETURNING id"},
		{dbType: gobatis.DbTypePostgres, value: &T9{}, fields: []string{"f1", "f2", "e"}, sql: "INSERT INTO t9_table(e, f1, f2, created_at, updated_at) VALUES(#{e}, #{f1}, #{f2}, now(), now()) RETURNING id"},
		{dbType: gobatis.DbTypeMSSql, value: &T4{}, fields: []string{"f1", "f2", "f3", "f4"}, sql: "INSERT INTO t2_table(f3, f4, f1, f2, created_at, updated_at) OUTPUT inserted.id VALUES(#{f3}, #{f4}, #{f1}, #{f2}, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)"},
		{dbType: gobatis.DbTypeMSSql, value: &T4{}, fields: []string{"f1", "f2", "f3", "f4"}, sql: "INSERT INTO t2_table(f3, f4, f1, f2, created_at, updated_at) VALUES(#{f3}, #{f4}, #{f1}, #{f2}, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)", noReturn: true},
		{dbType: gobatis.DbTypePostgres, value: T10{}, fields: []string{"f_1", "f2"}, sql: "INSERT INTO t10_table(f_1, f2, created_at, updated_at) VALUES(#{f_1}, #{f2}, now(), now()) RETURNING id"},
		{dbType: gobatis.DbTypePostgres, value: &T10{}, fields: []string{"f_1", "f2"}, sql: "INSERT INTO t10_table(f_1, f2, created_at, updated_at) VALUES(#{f_1}, #{f2}, now(), now()) RETURNING id"},
		{dbType: gobatis.DbTypeMSSql, value: T10{}, fields: []string{"f_1", "f2"}, sql: "INSERT INTO t10_table(f_1, f2, created_at, updated_at) OUTPUT inserted.id VALUES(#{f_1}, #{f2}, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)"},
		{dbType: gobatis.DbTypeMSSql, value: &T10{}, fields: []string{"f_1", "f2"}, sql: "INSERT INTO t10_table(f_1, f2, created_at, updated_at) OUTPUT inserted.id VALUES(#{f_1}, #{f2}, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)"},

		{dbType: gobatis.DbTypePostgres, value: T10{}, fields: []string{"f1", "f2"}, sql: "INSERT INTO t10_table(f_1, f2, created_at, updated_at) VALUES(#{f1}, #{f2}, now(), now()) RETURNING id"},
		{dbType: gobatis.DbTypePostgres, value: &T10{}, fields: []string{"f1", "f2"}, sql: "INSERT INTO t10_table(f_1, f2, created_at, updated_at) VALUES(#{f1}, #{f2}, now(), now()) RETURNING id"},
		{dbType: gobatis.DbTypeMSSql, value: T10{}, fields: []string{"f1", "f2"}, sql: "INSERT INTO t10_table(f_1, f2, created_at, updated_at) OUTPUT inserted.id VALUES(#{f1}, #{f2}, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)"},
		{dbType: gobatis.DbTypeMSSql, value: &T10{}, fields: []string{"f1", "f2"}, sql: "INSERT INTO t10_table(f_1, f2, created_at, updated_at) OUTPUT inserted.id VALUES(#{f1}, #{f2}, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)"},

		{dbType: gobatis.DbTypePostgres, value: T10{}, fields: []string{"f1", "f2", "created_at", "updated_at"}, sql: "INSERT INTO t10_table(f_1, f2, created_at, updated_at) VALUES(#{f1}, #{f2}, now(), now()) RETURNING id"},
		{dbType: gobatis.DbTypePostgres, value: &T10{}, fields: []string{"f1", "f2", "created_at", "updated_at"}, sql: "INSERT INTO t10_table(f_1, f2, created_at, updated_at) VALUES(#{f1}, #{f2}, now(), now()) RETURNING id"},
		{dbType: gobatis.DbTypeMSSql, value: T10{}, fields: []string{"f1", "f2", "created_at", "updated_at"}, sql: "INSERT INTO t10_table(f_1, f2, created_at, updated_at) OUTPUT inserted.id VALUES(#{f1}, #{f2}, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)"},
		{dbType: gobatis.DbTypeMSSql, value: &T10{}, fields: []string{"f1", "f2", "created_at", "updated_at"}, sql: "INSERT INTO t10_table(f_1, f2, created_at, updated_at) OUTPUT inserted.id VALUES(#{f1}, #{f2}, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)"},
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

	_, err := gobatis.GenerateInsertSQL2(gobatis.DbTypePostgres,
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
		{dbType: gobatis.DbTypePostgres, value: T1{}, sql: "UPDATE t1_table SET f1=#{f1}, f2=#{f2}, f3=#{f3}, updated_at=now() WHERE id=#{id}"},
		{dbType: gobatis.DbTypeMysql, value: &T1{}, names: []string{"id"}, sql: "UPDATE t1_table SET f1=#{f1}, f2=#{f2}, f3=#{f3}, updated_at=CURRENT_TIMESTAMP WHERE id=#{id}"},
		{dbType: gobatis.DbTypePostgres, value: T2{}, sql: "UPDATE t2_table SET f1=#{f1}, f2=#{f2}, updated_at=now() WHERE id=#{id}"},
		{dbType: gobatis.DbTypePostgres, value: &T2{}, names: []string{"id"}, sql: "UPDATE t2_table SET f1=#{f1}, f2=#{f2}, updated_at=now() WHERE id=#{id}"},
		{dbType: gobatis.DbTypePostgres, value: T3{}, sql: "UPDATE t3_table SET f1=#{f1}, f2=#{f2}, updated_at=now() WHERE id=#{id}"},
		{dbType: gobatis.DbTypePostgres, value: &T3{}, names: []string{"id"}, sql: "UPDATE t3_table SET f1=#{f1}, f2=#{f2}, updated_at=now() WHERE id=#{id}"},
		{dbType: gobatis.DbTypePostgres, value: T4{}, sql: "UPDATE t2_table SET f3=#{f3}, f4=#{f4}, f1=#{f1}, f2=#{f2}, updated_at=now() WHERE id=#{id}"},
		{dbType: gobatis.DbTypePostgres, value: &T4{}, names: []string{"id"}, sql: "UPDATE t2_table SET f3=#{f3}, f4=#{f4}, f1=#{f1}, f2=#{f2}, updated_at=now() WHERE id=#{id}"},
		{dbType: gobatis.DbTypePostgres, value: &T4{}, names: []string{"id", "f2"}, sql: "UPDATE t2_table SET f3=#{f3}, f4=#{f4}, f1=#{f1}, updated_at=now() WHERE id=#{id} AND f2=#{f2}"},
		{dbType: gobatis.DbTypePostgres, value: T8{}, sql: "UPDATE t8_table SET f1=#{f1}, updated_at=now() WHERE id=#{id}"},
		{dbType: gobatis.DbTypePostgres, value: &T8{}, names: []string{"id"}, sql: "UPDATE t8_table SET f1=#{f1}, updated_at=now() WHERE id=#{id}"},
		{dbType: gobatis.DbTypePostgres, value: T9{}, sql: "UPDATE t9_table SET e=#{e}, f1=#{f1}, updated_at=now() WHERE id=#{id}"},
		{dbType: gobatis.DbTypePostgres, value: &T9{}, names: []string{"id"}, sql: "UPDATE t9_table SET e=#{e}, f1=#{f1}, updated_at=now() WHERE id=#{id}"},
		{dbType: gobatis.DbTypePostgres, prefix: "a.", value: T9{}, sql: "UPDATE t9_table SET e=#{a.e}, f1=#{a.f1}, updated_at=now() WHERE id=#{a.id}"},
		{dbType: gobatis.DbTypePostgres, prefix: "a.", value: &T9{}, names: []string{"id"}, sql: "UPDATE t9_table SET e=#{a.e}, f1=#{a.f1}, updated_at=now() WHERE id=#{id}"},
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

	_, err := gobatis.GenerateUpdateSQL(gobatis.DbTypeMysql,
		mapper, "", reflect.TypeOf(&T7{}), []string{}, nil)
	if err == nil {
		t.Error("excepted error got ok")
		return
	}

	_, err = gobatis.GenerateUpdateSQL(gobatis.DbTypeMysql,
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
		{dbType: gobatis.DbTypePostgres, value: T1{}, query: "id", values: []string{"f1", "f2", "f3", "deleted_at"}, sql: "UPDATE t1_table SET f1=#{f1}, f2=#{f2,null=true}, f3=#{f3,notnull=true}, updated_at=now() WHERE id=#{id}"},
		{dbType: gobatis.DbTypePostgres, value: T1{}, query: "id", values: []string{"f1", "f2", "f3"}, sql: "UPDATE t1_table SET f1=#{f1}, f2=#{f2,null=true}, f3=#{f3,notnull=true}, updated_at=now() WHERE id=#{id}"},
		{dbType: gobatis.DbTypeMysql, value: &T1{}, query: "id", values: []string{"f1", "f2", "f3"}, sql: "UPDATE t1_table SET f1=#{f1}, f2=#{f2,null=true}, f3=#{f3,notnull=true}, updated_at=CURRENT_TIMESTAMP WHERE id=#{id}"},
		{dbType: gobatis.DbTypePostgres, value: T10{}, query: "id", values: []string{"f1", "f2"}, sql: "UPDATE t10_table SET f_1=#{f1}, f2=#{f2}, updated_at=now() WHERE id=#{id}"},
		{dbType: gobatis.DbTypePostgres, value: T10{}, query: "id", values: []string{}, sql: "UPDATE t10_table SET updated_at=now() WHERE id=#{id}"},
		{dbType: gobatis.DbTypeMysql, value: T10{}, query: "id", values: []string{"f1", "f2"}, sql: "UPDATE t10_table SET f_1=#{f1}, f2=#{f2}, updated_at=CURRENT_TIMESTAMP WHERE id=#{id}"},
		{dbType: gobatis.DbTypeMysql, value: T10{}, query: "id", values: []string{"f1", "f2", "updatedAt"}, sql: "UPDATE t10_table SET f_1=#{f1}, f2=#{f2}, updated_at=CURRENT_TIMESTAMP WHERE id=#{id}"},
		{dbType: gobatis.DbTypeMysql, value: T10{}, query: "id", values: []string{}, sql: "UPDATE t10_table SET updated_at=CURRENT_TIMESTAMP WHERE id=#{id}"},
		{dbType: gobatis.DbTypePostgres, value: &T10{}, query: "id", values: []string{"f1", "f2"}, sql: "UPDATE t10_table SET f_1=#{f1}, f2=#{f2}, updated_at=now() WHERE id=#{id}"},
		{dbType: gobatis.DbTypePostgres, value: T10{}, query: "id", values: []string{"f_1", "f2"}, sql: "UPDATE t10_table SET f_1=#{f_1}, f2=#{f2}, updated_at=now() WHERE id=#{id}"},
		{dbType: gobatis.DbTypePostgres, value: &T10{}, query: "id", values: []string{"f_1", "f2"}, sql: "UPDATE t10_table SET f_1=#{f_1}, f2=#{f2}, updated_at=now() WHERE id=#{id}"},

		{dbType: gobatis.DbTypePostgres, value: &T10{}, query: "id", queryType: reflect.TypeOf(new(int64)).Elem(), values: []string{"f_1", "f2"}, sql: "UPDATE t10_table SET f_1=#{f_1}, f2=#{f2}, updated_at=now() WHERE id=#{id}"},
		{dbType: gobatis.DbTypePostgres, value: &T10{}, query: "id", queryType: reflect.TypeOf(new(sql.NullInt64)).Elem(), values: []string{"f_1", "f2"}, sql: "UPDATE t10_table SET f_1=#{f_1}, f2=#{f2}, updated_at=now() <where><if test=\"id.Valid\"> id=#{id} </if></where>"},
		{dbType: gobatis.DbTypePostgres, value: &T10{}, query: "id", queryType: reflect.TypeOf([]int64{}), values: []string{"f_1", "f2"}, sql: "UPDATE t10_table SET f_1=#{f_1}, f2=#{f2}, updated_at=now() WHERE id in (<foreach collection=\"id\" item=\"item\" separator=\",\" >#{item}</foreach>)"},
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

	_, err := gobatis.GenerateUpdateSQL2(gobatis.DbTypeMysql,
		mapper, reflect.TypeOf(&T10{}), nil, "id", []string{"f33"})
	if err == nil {
		t.Error("excepted error got ok")
		return
	}

	_, err = gobatis.GenerateUpdateSQL2(gobatis.DbTypeMysql,
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
		{dbType: gobatis.DbTypePostgres, value: T1ForNoDeleted{}, sql: "DELETE FROM t1_table"},
		{dbType: gobatis.DbTypePostgres, value: &T1ForNoDeleted{}, names: []string{"id"}, sql: "DELETE FROM t1_table WHERE id=#{id}"},
		{dbType: gobatis.DbTypePostgres, value: &T1ForNoDeleted{}, names: []string{"id", "f1"}, sql: "DELETE FROM t1_table WHERE id=#{id} AND f1=#{f1}"},

		{dbType: gobatis.DbTypePostgres, value: &T1ForNoDeleted{}, names: []string{"id", "f1"},
			argTypes: []reflect.Type{reflect.TypeOf(new(int64)).Elem(), reflect.TypeOf(new(string)).Elem()},
			sql:      "DELETE FROM t1_table WHERE id=#{id} AND f1=#{f1}"},
		{dbType: gobatis.DbTypePostgres, value: &T1ForNoDeleted{}, names: []string{"id", "f1"},
			argTypes: []reflect.Type{reflect.TypeOf([]int64{}), reflect.TypeOf(new(string)).Elem()},
			sql:      "DELETE FROM t1_table WHERE id in (<foreach collection=\"id\" item=\"item\" separator=\",\" >#{item}</foreach>) AND f1=#{f1}"},
		{dbType: gobatis.DbTypePostgres, value: &T1ForNoDeleted{}, names: []string{"id", "f1"},
			argTypes: []reflect.Type{reflect.TypeOf([]int64{}), reflect.TypeOf([]string{})},
			sql:      "DELETE FROM t1_table WHERE id in (<foreach collection=\"id\" item=\"item\" separator=\",\" >#{item}</foreach>) AND f1 in (<foreach collection=\"f1\" item=\"item\" separator=\",\" >#{item}</foreach>)"},

		{dbType: gobatis.DbTypePostgres, value: &T1ForNoDeleted{}, names: []string{"id", "f1"},
			argTypes: []reflect.Type{reflect.TypeOf(new(sql.NullInt64)).Elem(), reflect.TypeOf(new(string)).Elem()},
			sql:      "DELETE FROM t1_table WHERE <if test=\"id.Valid\"> id=#{id} </if> AND f1=#{f1}"},

		{dbType: gobatis.DbTypePostgres, value: &T1ForNoDeleted{}, names: []string{"id", "f1"},
			argTypes: []reflect.Type{reflect.TypeOf(new(int64)).Elem(), reflect.TypeOf(new(sql.NullString)).Elem()},
			sql:      "DELETE FROM t1_table WHERE id=#{id}<if test=\"f1.Valid\"> AND f1=#{f1} </if>"},

		{dbType: gobatis.DbTypePostgres, value: &T1ForNoDeleted{}, names: []string{"id", "f1"},
			argTypes: []reflect.Type{reflect.TypeOf(new(sql.NullInt64)).Elem(), reflect.TypeOf(new(sql.NullString)).Elem()},
			sql:      "DELETE FROM t1_table <where><if test=\"id.Valid\"> id=#{id} </if><if test=\"f1.Valid\"> AND f1=#{f1} </if></where>"},

		{dbType: gobatis.DbTypePostgres, value: T1{}, sql: "UPDATE t1_table SET deleted_at=now() "},
		{dbType: gobatis.DbTypePostgres, value: &T1{}, names: []string{"id"}, sql: "UPDATE t1_table SET deleted_at=now()  WHERE id=#{id}"},
		{dbType: gobatis.DbTypePostgres, value: &T1{}, names: []string{"id", "f1"}, sql: "UPDATE t1_table SET deleted_at=now()  WHERE id=#{id} AND f1=#{f1}"},

		{dbType: gobatis.DbTypePostgres, value: &T1{}, names: []string{"id", "f1"},
			argTypes: []reflect.Type{reflect.TypeOf(new(int64)).Elem(), reflect.TypeOf(new(string)).Elem()},
			sql:      "UPDATE t1_table SET deleted_at=now()  WHERE id=#{id} AND f1=#{f1}"},
		{dbType: gobatis.DbTypePostgres, value: &T1{}, names: []string{"id", "f1"},
			argTypes: []reflect.Type{reflect.TypeOf([]int64{}), reflect.TypeOf(new(string)).Elem()},
			sql:      "UPDATE t1_table SET deleted_at=now()  WHERE id in (<foreach collection=\"id\" item=\"item\" separator=\",\" >#{item}</foreach>) AND f1=#{f1}"},
		{dbType: gobatis.DbTypePostgres, value: &T1{}, names: []string{"id", "f1"},
			argTypes: []reflect.Type{reflect.TypeOf([]int64{}), reflect.TypeOf([]string{})},
			sql:      "UPDATE t1_table SET deleted_at=now()  WHERE id in (<foreach collection=\"id\" item=\"item\" separator=\",\" >#{item}</foreach>) AND f1 in (<foreach collection=\"f1\" item=\"item\" separator=\",\" >#{item}</foreach>)"},

		{dbType: gobatis.DbTypePostgres, value: &T1{}, names: []string{"id", "f1"},
			argTypes: []reflect.Type{reflect.TypeOf(new(sql.NullInt64)).Elem(), reflect.TypeOf(new(string)).Elem()},
			sql:      "UPDATE t1_table SET deleted_at=now()  WHERE <if test=\"id.Valid\"> id=#{id} </if> AND f1=#{f1}"},

		{dbType: gobatis.DbTypePostgres, value: &T1{}, names: []string{"id", "f1"},
			argTypes: []reflect.Type{reflect.TypeOf(new(int64)).Elem(), reflect.TypeOf(new(sql.NullString)).Elem()},
			sql:      "UPDATE t1_table SET deleted_at=now()  WHERE id=#{id}<if test=\"f1.Valid\"> AND f1=#{f1} </if>"},

		{dbType: gobatis.DbTypePostgres, value: &T1{}, names: []string{"id", "f1"},
			argTypes: []reflect.Type{reflect.TypeOf(new(sql.NullInt64)).Elem(), reflect.TypeOf(new(sql.NullString)).Elem()},
			sql:      "UPDATE t1_table SET deleted_at=now()  <where><if test=\"id.Valid\"> id=#{id} </if><if test=\"f1.Valid\"> AND f1=#{f1} </if></where>"},

		{dbType: gobatis.DbTypePostgres, value: T1{}, names: []string{"force"},
			argTypes: []reflect.Type{reflect.TypeOf(new(bool)).Elem()},
			sql:      `<if test="force">UPDATE t1_table SET deleted_at=now() </if><if test="!force">DELETE FROM t1_table</if>`},

		{dbType: gobatis.DbTypePostgres, value: T1{}, names: []string{"force"},
			argTypes: []reflect.Type{reflect.TypeOf(new(sql.NullBool)).Elem()},
			err:      "unsupported type"},

		{dbType: gobatis.DbTypePostgres, value: T1{}, names: []string{"id", "force"},
			argTypes: []reflect.Type{reflect.TypeOf(new(int64)).Elem(), reflect.TypeOf(new(bool)).Elem()},
			sql:      `<if test="force">UPDATE t1_table SET deleted_at=now()  WHERE id=#{id}</if><if test="!force">DELETE FROM t1_table WHERE id=#{id}</if>`},

		{dbType: gobatis.DbTypePostgres, value: T1{}, names: []string{"id", "force"},
			argTypes: []reflect.Type{reflect.TypeOf(new(int64)).Elem(), reflect.TypeOf(new(sql.NullBool)).Elem()},
			err:      "unsupported type"},

		{dbType: gobatis.DbTypePostgres, value: T1ForNoDeleted{}, names: []string{"id"},
			filters: []gobatis.Filter{{Expression: "id>#{id}"}},
			sql:     "DELETE FROM t1_table WHERE id>#{id}"},
		{dbType: gobatis.DbTypePostgres, value: &T1ForNoDeleted{}, names: []string{"id", "f1"},
			argTypes: []reflect.Type{reflect.TypeOf(new(int64)).Elem(), reflect.TypeOf(new(string)).Elem()},
			filters:  []gobatis.Filter{{Expression: "id>#{id}"}},
			sql:      "DELETE FROM t1_table WHERE f1=#{f1} AND id>#{id}"},
		{dbType: gobatis.DbTypePostgres, value: T1{}, names: []string{"id"},
			filters: []gobatis.Filter{{Expression: "id>#{id}"}},
			sql:     "UPDATE t1_table SET deleted_at=now()  WHERE id>#{id}"},
		{dbType: gobatis.DbTypePostgres, value: &T1{}, names: []string{"id", "f1"},
			argTypes: []reflect.Type{reflect.TypeOf(new(int64)).Elem(), reflect.TypeOf(new(string)).Elem()},
			filters:  []gobatis.Filter{{Expression: "id>#{id}"}},
			sql:      "UPDATE t1_table SET deleted_at=now()  WHERE f1=#{f1} AND id>#{id}"},
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

	_, err := gobatis.GenerateDeleteSQL(gobatis.DbTypeMysql,
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
		order    string
		sql      string
	}{
		{dbType: gobatis.DbTypePostgres, value: T1{}, sql: "SELECT * FROM t1_table WHERE deleted_at IS NULL"},
		{dbType: gobatis.DbTypePostgres, value: &T1{}, names: []string{"id"}, sql: "SELECT * FROM t1_table WHERE id=#{id} AND deleted_at IS NULL"},
		{dbType: gobatis.DbTypePostgres, value: &T1{}, names: []string{"id", "f1"}, sql: "SELECT * FROM t1_table WHERE id=#{id} AND f1=#{f1} AND deleted_at IS NULL"},
		{dbType: gobatis.DbTypePostgres, value: &T1{}, names: []string{"id", "f1", "offset", "limit"},
			sql: "SELECT * FROM t1_table WHERE id=#{id} AND f1=#{f1} AND deleted_at IS NULL<if test=\"offset &gt; 0\"> OFFSET #{offset} </if><if test=\"limit &gt; 0\"> LIMIT #{limit} </if>"},

		{dbType: gobatis.DbTypePostgres, value: &T1{}, names: []string{"id", "f1"},
			argTypes: []reflect.Type{reflect.TypeOf(new(int64)).Elem(), reflect.TypeOf(new(string)).Elem()},
			sql:      "SELECT * FROM t1_table WHERE id=#{id} AND f1=#{f1} AND deleted_at IS NULL"},
		{dbType: gobatis.DbTypePostgres, value: &T1{}, names: []string{"id", "f1"},
			argTypes: []reflect.Type{reflect.TypeOf([]int64{}), reflect.TypeOf(new(string)).Elem()},
			sql:      "SELECT * FROM t1_table WHERE id in (<foreach collection=\"id\" item=\"item\" separator=\",\" >#{item}</foreach>) AND f1=#{f1} AND deleted_at IS NULL"},
		{dbType: gobatis.DbTypePostgres, value: &T1{}, names: []string{"id", "f1"},
			argTypes: []reflect.Type{reflect.TypeOf([]int64{}), reflect.TypeOf([]string{})},
			sql:      "SELECT * FROM t1_table WHERE id in (<foreach collection=\"id\" item=\"item\" separator=\",\" >#{item}</foreach>) AND f1 in (<foreach collection=\"f1\" item=\"item\" separator=\",\" >#{item}</foreach>) AND deleted_at IS NULL"},

		{dbType: gobatis.DbTypePostgres, value: &T1{}, names: []string{"id", "f1"},
			argTypes: []reflect.Type{reflect.TypeOf(new(sql.NullInt64)).Elem(), reflect.TypeOf(new(string)).Elem()},
			sql:      "SELECT * FROM t1_table WHERE <if test=\"id.Valid\"> id=#{id} </if> AND f1=#{f1} AND deleted_at IS NULL"},

		{dbType: gobatis.DbTypePostgres, value: &T1{}, names: []string{"id", "f1"},
			argTypes: []reflect.Type{reflect.TypeOf(new(int64)).Elem(), reflect.TypeOf(new(sql.NullString)).Elem()},
			sql:      "SELECT * FROM t1_table WHERE id=#{id}<if test=\"f1.Valid\"> AND f1=#{f1} </if> AND deleted_at IS NULL"},

		{dbType: gobatis.DbTypePostgres, value: &T1{}, names: []string{"id", "f1"},
			argTypes: []reflect.Type{reflect.TypeOf(new(sql.NullInt64)).Elem(), reflect.TypeOf(new(sql.NullString)).Elem()},
			sql:      "SELECT * FROM t1_table <where><if test=\"id.Valid\"> id=#{id} </if><if test=\"f1.Valid\"> AND f1=#{f1} </if> AND deleted_at IS NULL</where>"},

		{dbType: gobatis.DbTypePostgres, value: &T1{}, names: []string{"id", "f1", "isDeleted"},
			argTypes: []reflect.Type{reflect.TypeOf(new(sql.NullInt64)).Elem(), reflect.TypeOf(new(sql.NullString)).Elem(), reflect.TypeOf(new(bool)).Elem()},
			sql:      `SELECT * FROM t1_table <where><if test="id.Valid"> id=#{id} </if><if test="f1.Valid"> AND f1=#{f1} </if><if test="isDeleted"> AND deleted_at IS NOT NULL </if><if test="!isDeleted"> AND deleted_at IS NULL </if></where>`},

		{dbType: gobatis.DbTypePostgres, value: &T1{}, names: []string{"id", "f1", "isDeleted"},
			argTypes: []reflect.Type{reflect.TypeOf(new(sql.NullInt64)).Elem(), reflect.TypeOf(new(sql.NullString)).Elem(), reflect.TypeOf(new(sql.NullBool)).Elem()},
			sql:      `SELECT * FROM t1_table <where><if test="id.Valid"> id=#{id} </if><if test="f1.Valid"> AND f1=#{f1} </if><if test="isDeleted.Valid"><if test="isDeleted.Bool"> AND deleted_at IS NOT NULL </if><if test="!isDeleted.Bool"> AND deleted_at IS NULL </if></if></where>`},

		{dbType: gobatis.DbTypePostgres, value: T1ForNoDeleted{}, sql: "SELECT * FROM t1_table"},
		{dbType: gobatis.DbTypePostgres, value: T1ForNoDeleted{}, order: "id ASC", sql: "SELECT * FROM t1_table ORDER BY id ASC"},
		{dbType: gobatis.DbTypePostgres, value: &T1ForNoDeleted{}, names: []string{"id"}, sql: "SELECT * FROM t1_table WHERE id=#{id}"},
		{dbType: gobatis.DbTypePostgres, value: &T1ForNoDeleted{}, names: []string{"id", "f1"}, sql: "SELECT * FROM t1_table WHERE id=#{id} AND f1=#{f1}"},
		{dbType: gobatis.DbTypePostgres, value: &T1ForNoDeleted{}, names: []string{"id", "f1", "offset", "limit"},
			sql: "SELECT * FROM t1_table WHERE id=#{id} AND f1=#{f1}<if test=\"offset &gt; 0\"> OFFSET #{offset} </if><if test=\"limit &gt; 0\"> LIMIT #{limit} </if>"},

		{dbType: gobatis.DbTypePostgres, value: &T1ForNoDeleted{}, names: []string{"id", "f1"},
			argTypes: []reflect.Type{reflect.TypeOf(new(int64)).Elem(), reflect.TypeOf(new(string)).Elem()},
			sql:      "SELECT * FROM t1_table WHERE id=#{id} AND f1=#{f1}"},
		{dbType: gobatis.DbTypePostgres, value: &T1ForNoDeleted{}, names: []string{"id", "f1"},
			argTypes: []reflect.Type{reflect.TypeOf([]int64{}), reflect.TypeOf(new(string)).Elem()},
			sql:      "SELECT * FROM t1_table WHERE id in (<foreach collection=\"id\" item=\"item\" separator=\",\" >#{item}</foreach>) AND f1=#{f1}"},
		{dbType: gobatis.DbTypePostgres, value: &T1ForNoDeleted{}, names: []string{"id", "f1"},
			argTypes: []reflect.Type{reflect.TypeOf([]int64{}), reflect.TypeOf([]string{})},
			sql:      "SELECT * FROM t1_table WHERE id in (<foreach collection=\"id\" item=\"item\" separator=\",\" >#{item}</foreach>) AND f1 in (<foreach collection=\"f1\" item=\"item\" separator=\",\" >#{item}</foreach>)"},

		{dbType: gobatis.DbTypePostgres, value: &T1ForNoDeleted{}, names: []string{"id", "f1"},
			argTypes: []reflect.Type{reflect.TypeOf(new(sql.NullInt64)).Elem(), reflect.TypeOf(new(string)).Elem()},
			sql:      "SELECT * FROM t1_table WHERE <if test=\"id.Valid\"> id=#{id} </if> AND f1=#{f1}"},

		{dbType: gobatis.DbTypePostgres, value: &T1ForNoDeleted{}, names: []string{"id", "f1"},
			argTypes: []reflect.Type{reflect.TypeOf(new(int64)).Elem(), reflect.TypeOf(new(sql.NullString)).Elem()},
			sql:      "SELECT * FROM t1_table WHERE id=#{id}<if test=\"f1.Valid\"> AND f1=#{f1} </if>"},

		{dbType: gobatis.DbTypePostgres, value: &T1ForNoDeleted{}, names: []string{"id", "f1"},
			argTypes: []reflect.Type{reflect.TypeOf(new(sql.NullInt64)).Elem(), reflect.TypeOf(new(sql.NullString)).Elem()},
			sql:      "SELECT * FROM t1_table <where><if test=\"id.Valid\"> id=#{id} </if><if test=\"f1.Valid\"> AND f1=#{f1} </if></where>"},

		{dbType: gobatis.DbTypePostgres, value: T1ForNoDeleted{}, names: []string{"id"},
			filters: []gobatis.Filter{{Expression: "id>#{id}"}},
			sql:     "SELECT * FROM t1_table WHERE id>#{id}"},
		{dbType: gobatis.DbTypePostgres, value: &T1ForNoDeleted{}, names: []string{"id", "f1"},
			argTypes: []reflect.Type{reflect.TypeOf(new(int64)).Elem(), reflect.TypeOf(new(string)).Elem()},
			filters:  []gobatis.Filter{{Expression: "id>#{id}"}},
			sql:      "SELECT * FROM t1_table WHERE f1=#{f1} AND id>#{id}"},
		{dbType: gobatis.DbTypePostgres, value: T1{}, names: []string{"id"},
			filters: []gobatis.Filter{{Expression: "id>#{id}"}},
			sql:     "SELECT * FROM t1_table WHERE id>#{id} AND deleted_at IS NULL"},
		{dbType: gobatis.DbTypePostgres, value: &T1{}, names: []string{"id", "f1"},
			argTypes: []reflect.Type{reflect.TypeOf(new(int64)).Elem(), reflect.TypeOf(new(string)).Elem()},
			filters:  []gobatis.Filter{{Expression: "id>#{id}"}},
			sql:      "SELECT * FROM t1_table WHERE f1=#{f1} AND id>#{id} AND deleted_at IS NULL"},

		{dbType: gobatis.DbTypePostgres,
			value:    T1ForNoDeleted{},
			names:    []string{"offset", "limit"},
			argTypes: []reflect.Type{reflect.TypeOf(new(int64)).Elem(), reflect.TypeOf(new(int64)).Elem()},
			sql:      `SELECT * FROM t1_table<if test="offset &gt; 0"> OFFSET #{offset} </if><if test="limit &gt; 0"> LIMIT #{limit} </if>`},

		{dbType: gobatis.DbTypePostgres,
			value:    T13{},
			names:    []string{"f1", "f2"},
			argTypes: []reflect.Type{reflect.TypeOf(new(string)).Elem(), reflect.TypeOf(new(int64)).Elem()},
			sql:      `SELECT * FROM t13_table <where><if test="isNotEmpty(f1)"> f1=#{f1} </if><if test="f2 != 0"> AND f2=#{f2} </if></where>`},
	} {

		actaul, err := gobatis.GenerateSelectSQL(test.dbType,
			mapper, reflect.TypeOf(test.value), test.names, test.argTypes, test.filters, test.order)
		if err != nil {
			t.Error(err)
			continue
		}

		if actaul != test.sql {
			t.Error("[", idx, "] excepted is", test.sql)
			t.Error("[", idx, "] actual   is", actaul)
		}
	}

	_, err := gobatis.GenerateSelectSQL(gobatis.DbTypeMysql,
		mapper, reflect.TypeOf(&T7{}), []string{}, nil, nil, "")
	if err == nil {
		t.Error("excepted error got ok")
		return
	}
}

func TestGenerateCountSQL(t *testing.T) {
	for idx, test := range []struct {
		dbType   gobatis.Dialect
		value    interface{}
		names    []string
		argTypes []reflect.Type
		filters  []gobatis.Filter
		sql      string
	}{
		{dbType: gobatis.DbTypePostgres, value: T1{}, sql: "SELECT count(*) FROM t1_table WHERE deleted_at IS NULL"},
		{dbType: gobatis.DbTypePostgres, value: &T1{}, names: []string{"id"}, sql: "SELECT count(*) FROM t1_table WHERE id=#{id} AND deleted_at IS NULL"},
		{dbType: gobatis.DbTypePostgres, value: &T1{}, names: []string{"id", "f1"}, sql: "SELECT count(*) FROM t1_table WHERE id=#{id} AND f1=#{f1} AND deleted_at IS NULL"},
		{dbType: gobatis.DbTypePostgres, value: &T1{}, names: []string{"f1Like"}, sql: "SELECT count(*) FROM t1_table WHERE f1 like #{f1Like} AND deleted_at IS NULL"},
		{dbType: gobatis.DbTypePostgres, value: &T1{}, names: []string{"id", "f1"},
			argTypes: []reflect.Type{reflect.TypeOf(new(int64)).Elem(), reflect.TypeOf(new(string)).Elem()},
			sql:      "SELECT count(*) FROM t1_table WHERE id=#{id} AND f1=#{f1} AND deleted_at IS NULL"},
		{dbType: gobatis.DbTypePostgres, value: &T1{}, names: []string{"id", "f1"},
			argTypes: []reflect.Type{reflect.TypeOf([]int64{}), reflect.TypeOf(new(string)).Elem()},
			sql:      "SELECT count(*) FROM t1_table WHERE id in (<foreach collection=\"id\" item=\"item\" separator=\",\" >#{item}</foreach>) AND f1=#{f1} AND deleted_at IS NULL"},
		{dbType: gobatis.DbTypePostgres, value: &T1{}, names: []string{"id", "f1"},
			argTypes: []reflect.Type{reflect.TypeOf([]int64{}), reflect.TypeOf([]string{})},
			sql:      "SELECT count(*) FROM t1_table WHERE id in (<foreach collection=\"id\" item=\"item\" separator=\",\" >#{item}</foreach>) AND f1 in (<foreach collection=\"f1\" item=\"item\" separator=\",\" >#{item}</foreach>) AND deleted_at IS NULL"},

		{dbType: gobatis.DbTypePostgres, value: &T1{}, names: []string{"id", "f1"},
			argTypes: []reflect.Type{reflect.TypeOf(new(sql.NullInt64)).Elem(), reflect.TypeOf(new(string)).Elem()},
			sql:      "SELECT count(*) FROM t1_table WHERE <if test=\"id.Valid\"> id=#{id} </if> AND f1=#{f1} AND deleted_at IS NULL"},

		{dbType: gobatis.DbTypePostgres, value: &T1{}, names: []string{"id", "f1"},
			argTypes: []reflect.Type{reflect.TypeOf(new(int64)).Elem(), reflect.TypeOf(new(sql.NullString)).Elem()},
			sql:      "SELECT count(*) FROM t1_table WHERE id=#{id}<if test=\"f1.Valid\"> AND f1=#{f1} </if> AND deleted_at IS NULL"},

		{dbType: gobatis.DbTypePostgres, value: &T1{}, names: []string{"id", "f1"},
			argTypes: []reflect.Type{reflect.TypeOf(new(sql.NullInt64)).Elem(), reflect.TypeOf(new(sql.NullString)).Elem()},
			sql:      "SELECT count(*) FROM t1_table <where><if test=\"id.Valid\"> id=#{id} </if><if test=\"f1.Valid\"> AND f1=#{f1} </if> AND deleted_at IS NULL</where>"},

		{dbType: gobatis.DbTypePostgres, value: &T1{}, names: []string{"f1Like"},
			argTypes: []reflect.Type{reflect.TypeOf(new(string)).Elem()},
			sql:      "SELECT count(*) FROM t1_table WHERE f1 like #{f1Like} AND deleted_at IS NULL"},

		{dbType: gobatis.DbTypePostgres, value: &T1{}, names: []string{"f1Like"},
			argTypes: []reflect.Type{reflect.TypeOf(new(sql.NullString)).Elem()},
			sql:      "SELECT count(*) FROM t1_table <where><if test=\"f1Like.Valid\"> f1 like #{f1Like} </if> AND deleted_at IS NULL</where>"},

		{dbType: gobatis.DbTypePostgres, value: &T1{}, names: []string{"f1Like"},
			argTypes: []reflect.Type{reflect.TypeOf(new(struct{ sql.NullString })).Elem()},
			sql:      "SELECT count(*) FROM t1_table <where><if test=\"f1Like.Valid\"> f1 like #{f1Like} </if> AND deleted_at IS NULL</where>"},

		{dbType: gobatis.DbTypePostgres, value: &T1{}, names: []string{"created_at"},
			argTypes: []reflect.Type{reflect.TypeOf(struct {
				Start, End time.Time
			}{})},
			sql: "SELECT count(*) FROM t1_table WHERE  (created_at BETWEEN #{created_at.Start} AND #{created_at.End})  AND deleted_at IS NULL"},

		{dbType: gobatis.DbTypePostgres, value: &T11{}, names: []string{"f1"},
			argTypes: []reflect.Type{reflect.TypeOf(new(string)).Elem()},
			sql:      "SELECT count(*) FROM t11_table WHERE #{f1} = ANY (f_1)"},

		{dbType: gobatis.DbTypePostgres, value: &T12{}, names: []string{"f1"},
			argTypes: []reflect.Type{reflect.TypeOf(new(string)).Elem()},
			sql:      "SELECT count(*) FROM t12_table WHERE f_1 @> #{f1}"},

		{dbType: gobatis.DbTypePostgres, value: T1ForNoDeleted{}, sql: "SELECT count(*) FROM t1_table"},
		{dbType: gobatis.DbTypePostgres, value: &T1ForNoDeleted{}, names: []string{"id"}, sql: "SELECT count(*) FROM t1_table WHERE id=#{id}"},
		{dbType: gobatis.DbTypePostgres, value: &T1ForNoDeleted{}, names: []string{"id", "f1"}, sql: "SELECT count(*) FROM t1_table WHERE id=#{id} AND f1=#{f1}"},
		{dbType: gobatis.DbTypePostgres, value: &T1ForNoDeleted{}, names: []string{"f1Like"}, sql: "SELECT count(*) FROM t1_table WHERE f1 like #{f1Like}"},
		{dbType: gobatis.DbTypePostgres, value: &T1ForNoDeleted{}, names: []string{"id", "f1"},
			argTypes: []reflect.Type{reflect.TypeOf(new(int64)).Elem(), reflect.TypeOf(new(string)).Elem()},
			sql:      "SELECT count(*) FROM t1_table WHERE id=#{id} AND f1=#{f1}"},
		{dbType: gobatis.DbTypePostgres, value: &T1ForNoDeleted{}, names: []string{"id", "f1"},
			argTypes: []reflect.Type{reflect.TypeOf([]int64{}), reflect.TypeOf(new(string)).Elem()},
			sql:      "SELECT count(*) FROM t1_table WHERE id in (<foreach collection=\"id\" item=\"item\" separator=\",\" >#{item}</foreach>) AND f1=#{f1}"},
		{dbType: gobatis.DbTypePostgres, value: &T1ForNoDeleted{}, names: []string{"id", "f1"},
			argTypes: []reflect.Type{reflect.TypeOf([]int64{}), reflect.TypeOf([]string{})},
			sql:      "SELECT count(*) FROM t1_table WHERE id in (<foreach collection=\"id\" item=\"item\" separator=\",\" >#{item}</foreach>) AND f1 in (<foreach collection=\"f1\" item=\"item\" separator=\",\" >#{item}</foreach>)"},

		{dbType: gobatis.DbTypePostgres, value: &T1ForNoDeleted{}, names: []string{"id", "f1"},
			argTypes: []reflect.Type{reflect.TypeOf(new(sql.NullInt64)).Elem(), reflect.TypeOf(new(string)).Elem()},
			sql:      "SELECT count(*) FROM t1_table WHERE <if test=\"id.Valid\"> id=#{id} </if> AND f1=#{f1}"},

		{dbType: gobatis.DbTypePostgres, value: &T1ForNoDeleted{}, names: []string{"id", "f1"},
			argTypes: []reflect.Type{reflect.TypeOf(new(int64)).Elem(), reflect.TypeOf(new(sql.NullString)).Elem()},
			sql:      "SELECT count(*) FROM t1_table WHERE id=#{id}<if test=\"f1.Valid\"> AND f1=#{f1} </if>"},

		{dbType: gobatis.DbTypePostgres, value: &T1ForNoDeleted{}, names: []string{"id", "f1"},
			argTypes: []reflect.Type{reflect.TypeOf(new(sql.NullInt64)).Elem(), reflect.TypeOf(new(sql.NullString)).Elem()},
			sql:      "SELECT count(*) FROM t1_table <where><if test=\"id.Valid\"> id=#{id} </if><if test=\"f1.Valid\"> AND f1=#{f1} </if></where>"},

		{dbType: gobatis.DbTypePostgres, value: &T1ForNoDeleted{}, names: []string{"f1Like"},
			argTypes: []reflect.Type{reflect.TypeOf(new(sql.NullString)).Elem()},
			sql:      "SELECT count(*) FROM t1_table <where><if test=\"f1Like.Valid\"> f1 like #{f1Like} </if></where>"},

		{dbType: gobatis.DbTypePostgres, value: &T1ForNoDeleted{}, names: []string{"f1Like"},
			argTypes: []reflect.Type{reflect.TypeOf(new(struct{ sql.NullString })).Elem()},
			sql:      "SELECT count(*) FROM t1_table <where><if test=\"f1Like.Valid\"> f1 like #{f1Like} </if></where>"},

		{dbType: gobatis.DbTypePostgres, value: &T1ForNoDeleted{}, names: []string{"created_at"},
			argTypes: []reflect.Type{reflect.TypeOf(struct {
				Start, End time.Time
			}{})},
			sql: "SELECT count(*) FROM t1_table WHERE  (created_at BETWEEN #{created_at.Start} AND #{created_at.End}) "},

		{dbType: gobatis.DbTypePostgres, value: T1ForNoDeleted{}, names: []string{"id"},
			filters: []gobatis.Filter{{Expression: "id>#{id}"}},
			sql:     "SELECT count(*) FROM t1_table WHERE id>#{id}"},
		{dbType: gobatis.DbTypePostgres, value: &T1ForNoDeleted{}, names: []string{"id", "f1"},
			argTypes: []reflect.Type{reflect.TypeOf(new(int64)).Elem(), reflect.TypeOf(new(string)).Elem()},
			filters:  []gobatis.Filter{{Expression: "id>#{id}"}},
			sql:      "SELECT count(*) FROM t1_table WHERE f1=#{f1} AND id>#{id}"},
		{dbType: gobatis.DbTypePostgres, value: T1{}, names: []string{"id"},
			filters: []gobatis.Filter{{Expression: "id>#{id}"}},
			sql:     "SELECT count(*) FROM t1_table WHERE id>#{id} AND deleted_at IS NULL"},
		{dbType: gobatis.DbTypePostgres, value: &T1{}, names: []string{"id", "f1"},
			argTypes: []reflect.Type{reflect.TypeOf(new(int64)).Elem(), reflect.TypeOf(new(string)).Elem()},
			filters:  []gobatis.Filter{{Expression: "id>#{id}"}},
			sql:      "SELECT count(*) FROM t1_table WHERE f1=#{f1} AND id>#{id} AND deleted_at IS NULL"},
	} {
		actaul, err := gobatis.GenerateCountSQL(test.dbType,
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

	_, err := gobatis.GenerateCountSQL(gobatis.DbTypeMysql,
		mapper, reflect.TypeOf(&T7{}), []string{}, nil, nil)
	if err == nil {
		t.Error("excepted error got ok")
		return
	}

	_, err = gobatis.GenerateCountSQL(gobatis.DbTypePostgres,
		mapper, reflect.TypeOf(&T1{}), []string{"f1Like"},
		[]reflect.Type{reflect.TypeOf([]string{})}, nil)
	if err == nil {
		t.Error("excepted error got ok")
		return
	}
	_, err = gobatis.GenerateCountSQL(gobatis.DbTypePostgres,
		mapper, reflect.TypeOf(&T1{}), []string{"f2Like"},
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
