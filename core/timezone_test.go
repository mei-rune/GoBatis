package core_test

import (
	"database/sql"
	"testing"
	"time"

	"github.com/runner-mei/GoBatis/core"
	"github.com/runner-mei/GoBatis/dialects"
	"github.com/runner-mei/GoBatis/tests"
)

// tzOffsetQuery 返回各数据库查询自身相对于 UTC 偏移秒数的 SQL，
// 用于与 GetDbTimeZone 的结果做交叉验证。无法用单条 SQL 表达的驱动返回空串。
func tzOffsetQuery(drv string) string {
	switch drv {
	case "postgres", "pgx", "pgx/v5", "kingbase", "gaussdb", "opengauss", "polardb", "goldendb", "cockroach":
		return "SELECT EXTRACT(timezone FROM now())::int"
	case "mysql", "mariadb", "tidb":
		return "SELECT TIMESTAMPDIFF(SECOND, UTC_TIMESTAMP(), NOW())"
	case "sqlserver", "mssql":
		return "SELECT DATEDIFF(SECOND, GETUTCDATE(), GETDATE())"
	default:
		return ""
	}
}

// TestGetDbTimeZone 使用 tests.Run 提供的真实数据库连接（DriverName: TestDrv,
// DataSource: GetTestConnURL()）来验证 GetDbTimeZone 能正确返回数据库时区。
func TestGetDbTimeZone(t *testing.T) {
	tests.Run(t, func(t testing.TB, factory *core.Session) {
		db, ok := factory.DB().(*sql.DB)
		if !ok {
			t.Fatal("factory.DB() is not *sql.DB")
		}

		loc, err := dialects.GetDbTimeZone(tests.TestDrv, db)
		if err != nil {
			t.Fatalf("GetDbTimeZone(%q) failed: %v", tests.TestDrv, err)
		}
		if loc == nil {
			t.Fatal("GetDbTimeZone returned nil location")
		}

		if q := tzOffsetQuery(tests.TestDrv); q != "" {
			var wantSec int
			if err := db.QueryRow(q).Scan(&wantSec); err != nil {
				t.Fatalf("query db offset failed: %v", err)
			}
			_, gotSec := time.Date(2020, 1, 1, 12, 0, 0, 0, loc).Zone()
			if gotSec != wantSec {
				t.Fatalf("timezone offset mismatch: GetDbTimeZone=%ds, db reports=%ds", gotSec, wantSec)
			}
		}
	})
}

// TestGetDbTimeZoneUnsupported 验证不支持的驱动名会返回错误。
func TestGetDbTimeZoneUnsupported(t *testing.T) {
	tests.Run(t, func(t testing.TB, factory *core.Session) {
		db, ok := factory.DB().(*sql.DB)
		if !ok {
			t.Fatal("factory.DB() is not *sql.DB")
		}

		if _, err := dialects.GetDbTimeZone("unknown_driver", db); err == nil {
			t.Fatal("expected error for unsupported driver, got nil")
		}
	})
}
