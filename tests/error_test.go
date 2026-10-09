package tests

import (
	"context"
	"testing"

	gobatis "github.com/runner-mei/GoBatis"
	"github.com/runner-mei/GoBatis/dialects"
)

const uniqueTestTablename = "gobatis_unique_test"

// TestUniqueViolation 用当前配置的数据库(见 session.go 的 dbDrv/dbURL)
// 建一张表，验证主键、唯一索引冲突能被 dialects.IsRecordAlreadyExists 识别出来。
func TestUniqueViolation(t *testing.T) {
	Run(t, func(t testing.TB, factory *gobatis.SessionFactory) {
		ctx := context.Background()
		db := factory.DB()
		placeholder := factory.Dialect().Placeholder()

		t.Log("database is", factory.DriverName())

		// Oracle、神通等数据库不支持 DROP TABLE IF EXISTS
		if _, err := db.ExecContext(ctx, "DROP TABLE "+uniqueTestTablename); err != nil {
			t.Log("drop table:", err)
		}

		if _, err := db.ExecContext(ctx, "CREATE TABLE "+uniqueTestTablename+
			" (id int NOT NULL, name varchar(50) NOT NULL, PRIMARY KEY (id), UNIQUE (name))"); err != nil {
			t.Error("create table:", err)
			return
		}
		defer func() {
			if _, err := db.ExecContext(ctx, "DROP TABLE "+uniqueTestTablename); err != nil {
				t.Log("drop table:", err)
			}
		}()

		insertSQL := "INSERT INTO " + uniqueTestTablename + " (id, name) VALUES (" +
			placeholder.Format(0) + ", " + placeholder.Format(1) + ")"
		insert := func(id int, name string) error {
			_, err := db.ExecContext(ctx, insertSQL, id, name)
			return err
		}

		if err := insert(1, "gobatis-1"); err != nil {
			t.Error("insert the first row:", err)
			return
		}

		// 1. 主键冲突
		pkErr := insert(1, "gobatis-pk")
		if pkErr == nil {
			t.Error("want primary key violation, got ok")
		} else {
			t.Log("[pk]", pkErr)
			if !dialects.IsRecordAlreadyExists(factory.Dialect(), pkErr) {
				t.Error("[pk] want RecordAlreadyExists")
				t.Error("[pk]  got", pkErr)
			}
		}

		// 2. 唯一索引冲突
		uniqueErr := insert(2, "gobatis-1")
		if uniqueErr == nil {
			t.Error("want unique key violation, got ok")
		} else {
			t.Log("[unique]", uniqueErr)
			if !dialects.IsRecordAlreadyExists(factory.Dialect(), uniqueErr) {
				t.Error("[unique] want RecordAlreadyExists")
				t.Error("[unique]  got", uniqueErr)
			}
		}
	})
}
