package gobatis_test

import (
	"reflect"
	"strings"
	"testing"
	"time"

	gobatis "github.com/runner-mei/GoBatis"
)

type T1 struct {
	ID        string    `db:"id,autoincr"`
	F1        string    `db:"f1"`
	F2        int       `db:"f2"`
	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
}

func init() {
	gobatis.RegisterTableName(T1{}, "t1_table")
}

type T2 struct {
	TableName struct{}  `db:"t2_table"`
	ID        string    `db:"id,autoincr"`
	F1        string    `db:"f1"`
	F2        int       `db:"f2"`
	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
}

type T3 struct {
	ID        string    `db:"id,autoincr"`
	F1        string    `db:"f1"`
	F2        int       `db:"f2"`
	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
}

func (t T3) TableName() string {
	return "t3_table"
}

type T4 struct {
	T2
	F3 string `db:"f3"`
	F4 int    `db:"f4"`
}

type T5 struct {
	ID        string    `db:"id,autoincr"`
	F1        string    `db:"f1"`
	F2        int       `db:"f2"`
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
	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
}

func init() {
	gobatis.RegisterTableName(&T6{}, "t6_table")
}

type T7 struct {
	TableName struct{} `db:"t7_table"`
	T2
	F3 string `db:"f3"`
}

var mapper = gobatis.CreateMapper("", nil, nil)

type T8 struct {
	TableName struct{}  `db:"t8_table"`
	ID        string    `db:"id,autoincr"`
	F1        string    `db:"f1"`
	F2        int       `db:"f2,created"`
	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
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

func TestGenerateInsertSQL(t *testing.T) {
	for idx, test := range []struct {
		dbType   gobatis.Dialect
		value    interface{}
		noReturn bool
		sql      string
	}{
		{dbType: gobatis.DbTypePostgres, value: T1{}, sql: "INSERT INTO t1_table(f1, f2, created_at, updated_at) VALUES(#{f1}, #{f2}, #{created_at}, #{updated_at}) RETURNING id"},
		{dbType: gobatis.DbTypePostgres, value: &T1{}, sql: "INSERT INTO t1_table(f1, f2, created_at, updated_at) VALUES(#{f1}, #{f2}, #{created_at}, #{updated_at}) RETURNING id"},
		{dbType: gobatis.DbTypePostgres, value: T2{}, sql: "INSERT INTO t2_table(f1, f2, created_at, updated_at) VALUES(#{f1}, #{f2}, #{created_at}, #{updated_at}) RETURNING id"},
		{dbType: gobatis.DbTypePostgres, value: &T2{}, sql: "INSERT INTO t2_table(f1, f2, created_at, updated_at) VALUES(#{f1}, #{f2}, #{created_at}, #{updated_at}) RETURNING id"},
		{dbType: gobatis.DbTypePostgres, value: T3{}, sql: "INSERT INTO t3_table(f1, f2, created_at, updated_at) VALUES(#{f1}, #{f2}, #{created_at}, #{updated_at}) RETURNING id"},
		{dbType: gobatis.DbTypePostgres, value: &T3{}, sql: "INSERT INTO t3_table(f1, f2, created_at, updated_at) VALUES(#{f1}, #{f2}, #{created_at}, #{updated_at}) RETURNING id"},
		{dbType: gobatis.DbTypePostgres, value: T4{}, sql: "INSERT INTO t2_table(f3, f4, f1, f2, created_at, updated_at) VALUES(#{f3}, #{f4}, #{f1}, #{f2}, #{created_at}, #{updated_at}) RETURNING id"},
		{dbType: gobatis.DbTypePostgres, value: &T4{}, sql: "INSERT INTO t2_table(f3, f4, f1, f2, created_at, updated_at) VALUES(#{f3}, #{f4}, #{f1}, #{f2}, #{created_at}, #{updated_at}) RETURNING id"},
		{dbType: gobatis.DbTypePostgres, value: T4{}, sql: "INSERT INTO t2_table(f3, f4, f1, f2, created_at, updated_at) VALUES(#{f3}, #{f4}, #{f1}, #{f2}, #{created_at}, #{updated_at})", noReturn: true},
		{dbType: gobatis.DbTypePostgres, value: &T4{}, sql: "INSERT INTO t2_table(f3, f4, f1, f2, created_at, updated_at) VALUES(#{f3}, #{f4}, #{f1}, #{f2}, #{created_at}, #{updated_at})", noReturn: true},
		{dbType: gobatis.DbTypeMysql, value: T4{}, sql: "INSERT INTO t2_table(f3, f4, f1, f2, created_at, updated_at) VALUES(#{f3}, #{f4}, #{f1}, #{f2}, #{created_at}, #{updated_at})"},
		{dbType: gobatis.DbTypeMysql, value: &T4{}, sql: "INSERT INTO t2_table(f3, f4, f1, f2, created_at, updated_at) VALUES(#{f3}, #{f4}, #{f1}, #{f2}, #{created_at}, #{updated_at})"},
		{dbType: gobatis.DbTypePostgres, value: T8{}, sql: "INSERT INTO t8_table(f1, f2, created_at, updated_at) VALUES(#{f1}, #{f2}, #{created_at}, #{updated_at}) RETURNING id"},
		{dbType: gobatis.DbTypePostgres, value: &T8{}, sql: "INSERT INTO t8_table(f1, f2, created_at, updated_at) VALUES(#{f1}, #{f2}, #{created_at}, #{updated_at}) RETURNING id"},
	} {
		actaul, err := gobatis.GenerateInsertSQL(test.dbType,
			mapper, reflect.TypeOf(test.value), test.noReturn)
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
		mapper, reflect.TypeOf(&T7{}), false)
	if err == nil {
		t.Error("excepted error got ok")
		return
	}
}

func TestGenerateUpdateSQL(t *testing.T) {
	for idx, test := range []struct {
		dbType gobatis.Dialect
		value  interface{}
		names  []string
		sql    string
	}{
		{dbType: gobatis.DbTypePostgres, value: T1{}, sql: "UPDATE t1_table SET f1=#{f1}, f2=#{f2}, updated_at=now()"},
		{dbType: gobatis.DbTypeMysql, value: &T1{}, names: []string{"id"}, sql: "UPDATE t1_table SET f1=#{f1}, f2=#{f2}, updated_at=CURRENT_TIMESTAMP WHERE id=#{id}"},
		{dbType: gobatis.DbTypePostgres, value: T2{}, sql: "UPDATE t2_table SET f1=#{f1}, f2=#{f2}, updated_at=now()"},
		{dbType: gobatis.DbTypePostgres, value: &T2{}, names: []string{"id"}, sql: "UPDATE t2_table SET f1=#{f1}, f2=#{f2}, updated_at=now() WHERE id=#{id}"},
		{dbType: gobatis.DbTypePostgres, value: T3{}, sql: "UPDATE t3_table SET f1=#{f1}, f2=#{f2}, updated_at=now()"},
		{dbType: gobatis.DbTypePostgres, value: &T3{}, names: []string{"id"}, sql: "UPDATE t3_table SET f1=#{f1}, f2=#{f2}, updated_at=now() WHERE id=#{id}"},
		{dbType: gobatis.DbTypePostgres, value: T4{}, sql: "UPDATE t2_table SET f3=#{f3}, f4=#{f4}, f1=#{f1}, f2=#{f2}, updated_at=now()"},
		{dbType: gobatis.DbTypePostgres, value: &T4{}, names: []string{"id"}, sql: "UPDATE t2_table SET f3=#{f3}, f4=#{f4}, f1=#{f1}, f2=#{f2}, updated_at=now() WHERE id=#{id}"},
		{dbType: gobatis.DbTypePostgres, value: &T4{}, names: []string{"id", "f2"}, sql: "UPDATE t2_table SET f3=#{f3}, f4=#{f4}, f1=#{f1}, updated_at=now() WHERE id=#{id} AND f2=#{f2}"},
		{dbType: gobatis.DbTypePostgres, value: T8{}, sql: "UPDATE t8_table SET f1=#{f1}, updated_at=now()"},
		{dbType: gobatis.DbTypePostgres, value: &T8{}, names: []string{"id"}, sql: "UPDATE t8_table SET f1=#{f1}, updated_at=now() WHERE id=#{id}"},
	} {
		actaul, err := gobatis.GenerateUpdateSQL(test.dbType,
			mapper, reflect.TypeOf(test.value), test.names)
		if err != nil {
			t.Error(err)
			continue
		}

		if actaul != test.sql {
			t.Error("[", idx, "] excepted is", test.sql)
			t.Error("[", idx, "] actual   is", actaul)
		}
	}

	_, err := gobatis.GenerateUpdateSQL(gobatis.DbTypeMysql,
		mapper, reflect.TypeOf(&T7{}), []string{})
	if err == nil {
		t.Error("excepted error got ok")
		return
	}
}

func TestGenerateDeleteSQL(t *testing.T) {
	for idx, test := range []struct {
		dbType gobatis.Dialect
		value  interface{}
		names  []string
		sql    string
	}{
		{dbType: gobatis.DbTypePostgres, value: T1{}, sql: "DELETE FROM t1_table"},
		{dbType: gobatis.DbTypePostgres, value: &T1{}, names: []string{"id"}, sql: "DELETE FROM t1_table WHERE id=#{id}"},
		{dbType: gobatis.DbTypePostgres, value: &T1{}, names: []string{"id", "f1"}, sql: "DELETE FROM t1_table WHERE id=#{id} AND f1=#{f1}"},
	} {
		actaul, err := gobatis.GenerateDeleteSQL(test.dbType,
			mapper, reflect.TypeOf(test.value), test.names)
		if err != nil {
			t.Error(err)
			continue
		}

		if actaul != test.sql {
			t.Error("[", idx, "] excepted is", test.sql)
			t.Error("[", idx, "] actual   is", actaul)
		}
	}

	_, err := gobatis.GenerateDeleteSQL(gobatis.DbTypeMysql,
		mapper, reflect.TypeOf(&T7{}), []string{})
	if err == nil {
		t.Error("excepted error got ok")
		return
	}
}

func TestGenerateSelectSQL(t *testing.T) {
	for idx, test := range []struct {
		dbType gobatis.Dialect
		value  interface{}
		names  []string
		sql    string
	}{
		{dbType: gobatis.DbTypePostgres, value: T1{}, sql: "SELECT * FROM t1_table"},
		{dbType: gobatis.DbTypePostgres, value: &T1{}, names: []string{"id"}, sql: "SELECT * FROM t1_table WHERE id=#{id}"},
		{dbType: gobatis.DbTypePostgres, value: &T1{}, names: []string{"id", "f1"}, sql: "SELECT * FROM t1_table WHERE id=#{id} AND f1=#{f1}"},
	} {
		actaul, err := gobatis.GenerateSelectSQL(test.dbType,
			mapper, reflect.TypeOf(test.value), test.names)
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
		mapper, reflect.TypeOf(&T7{}), []string{})
	if err == nil {
		t.Error("excepted error got ok")
		return
	}
}

func TestGenerateCountSQL(t *testing.T) {
	for idx, test := range []struct {
		dbType gobatis.Dialect
		value  interface{}
		names  []string
		sql    string
	}{
		{dbType: gobatis.DbTypePostgres, value: T1{}, sql: "SELECT count(*) FROM t1_table"},
		{dbType: gobatis.DbTypePostgres, value: &T1{}, names: []string{"id"}, sql: "SELECT count(*) FROM t1_table WHERE id=#{id}"},
		{dbType: gobatis.DbTypePostgres, value: &T1{}, names: []string{"id", "f1"}, sql: "SELECT count(*) FROM t1_table WHERE id=#{id} AND f1=#{f1}"},
	} {
		actaul, err := gobatis.GenerateCountSQL(test.dbType,
			mapper, reflect.TypeOf(test.value), test.names)
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
		mapper, reflect.TypeOf(&T7{}), []string{})
	if err == nil {
		t.Error("excepted error got ok")
		return
	}
}
