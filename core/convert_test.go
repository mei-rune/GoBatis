package core_test

import (
	"context"
	"database/sql"
	"net"
	"strings"
	"testing"

	"github.com/runner-mei/GoBatis/core"
	"github.com/runner-mei/GoBatis/dialects"
	"github.com/runner-mei/GoBatis/tests"
)

func TestConvert(t *testing.T) {
	tests.Run(t, func(_ testing.TB, factory *core.SessionFactory) {

		convert := tests.NewIconvertTest(factory.SessionReference())

		t.Run("int_null", func(t *testing.T) {
			queryStr := "SELECT field0 FROM gobatis_convert1 WHERE id = ?"
			if factory.Dialect() == dialects.Postgres {
				queryStr = "SELECT field0 FROM gobatis_convert1 WHERE id = $1"
			}

			for _, test := range []interface{}{
				int8(0),
				int16(0),
				int32(0),
				int64(0),
				int(0),

				uint8(0),
				uint16(0),
				uint32(0),
				uint64(0),
				uint(0),
				float32(0),
				float64(0),
				complex64(0),
				complex128(0),

				new(int8),
				new(int16),
				new(int32),
				new(int64),
				new(int),
				new(uint8),
				new(uint16),
				new(uint32),
				new(uint64),
				new(uint),
				new(float32),
				new(float64),
				new(complex64),
				new(complex128),

				(*int)(nil),
				(*uint)(nil),
				(*float32)(nil),
				(*complex64)(nil),
				(*complex64)(nil),
				(*complex128)(nil),
			} {
				id, err := convert.InsertIntNULL(test)
				if err != nil {
					t.Error(err)
					return
				}

				var value sql.NullInt64
				err = factory.DB().QueryRowContext(context.Background(), queryStr, id).Scan(&value)
				if err != nil {
					t.Error(err)
					return
				}

				if value.Valid {
					t.Error("want null got ok - ", value.Int64)
				}
			}
		})

		t.Run("int_not_null", func(t *testing.T) {
			for _, test := range []interface{}{
				int8(0),
				int16(0),
				int32(0),
				int64(0),
				int(0),

				uint8(0),
				uint16(0),
				uint32(0),
				uint64(0),
				uint(0),
				float32(0),
				float64(0),
				complex64(0),
				complex128(0),

				new(int8),
				new(int16),
				new(int32),
				new(int64),
				new(int),
				new(uint8),
				new(uint16),
				new(uint32),
				new(uint64),
				new(uint),
				new(float32),
				new(float64),
				new(complex64),
				new(complex128),

				(*int)(nil),
				(*uint)(nil),
				(*float32)(nil),
				(*complex64)(nil),
				(*complex64)(nil),
				(*complex128)(nil),
			} {
				_, err := convert.InsertIntNotNULL(test)
				if err != nil {
					if !strings.Contains(err.Error(), "zero") && !strings.Contains(err.Error(), "nil") {
						t.Error(err)
					}
				} else {
					t.Error("want error got ok")
				}
			}
		})

		t.Run("int_1", func(t *testing.T) {
			queryStr := "SELECT field0 FROM gobatis_convert1 WHERE id = ?"
			if factory.Dialect() == dialects.Postgres {
				queryStr = "SELECT field0 FROM gobatis_convert1 WHERE id = $1"
			}
			for idx, test := range []interface{}{
				int8(1),
				int16(1),
				int32(1),
				int64(1),
				int(1),

				uint8(1),
				uint16(1),
				uint32(1),
				uint64(1),
				uint(1),
				float32(1),
				float64(1),
				// complex64(0),
				// complex128(0),
			} {
				id, err := convert.InsertIntNULL(test)
				if err != nil {
					t.Error(err)
					return
				}

				var value sql.NullInt64
				err = factory.DB().QueryRowContext(context.Background(), queryStr, id).Scan(&value)
				if err != nil {
					t.Error(idx, id, err)
					return
				}

				if !value.Valid {
					t.Error("want null got ok")
				} else if value.Int64 != 1 {
					t.Error("want 1 got ", value.Int64)
				}
			}
		})

		t.Run("string_null", func(t *testing.T) {
			queryStr := "SELECT field0 FROM gobatis_convert2 WHERE id = ?"
			if factory.Dialect() == dialects.Postgres {
				queryStr = "SELECT field0 FROM gobatis_convert2 WHERE id = $1"
			}

			for _, test := range []interface{}{
				"",
				new(string),
				nil,
				[]byte{},
				(*string)(nil),
				([]byte)(nil),
				(*[]byte)(nil),
			} {
				id, err := convert.InsertStrNULL(test)
				if err != nil {
					t.Error(err)
					return
				}

				var value sql.NullString
				err = factory.DB().QueryRowContext(context.Background(), queryStr, id).Scan(&value)
				if err != nil {
					t.Error(err)
					return
				}

				if value.Valid {
					t.Error("want null got ok")
				}
			}
		})

		t.Run("str_not_null", func(t *testing.T) {
			var a *string
			for idx, test := range []interface{}{
				"",
				new(string),
				nil,
				a,
				[]byte{},
				([]byte)(nil),
				(*[]byte)(nil),
			} {
				_, err := convert.InsertStrNotNULL(test)
				if err != nil {
					if !strings.Contains(err.Error(), "zero") && !strings.Contains(err.Error(), "nil") {
						t.Error(err)
					}
				} else {
					t.Error(idx, "want error got ok")
				}
			}
		})

		t.Run("string_1", func(t *testing.T) {
			queryStr := "SELECT field0 FROM gobatis_convert2 WHERE id = ?"
			if factory.Dialect() == dialects.Postgres {
				queryStr = "SELECT field0 FROM gobatis_convert2 WHERE id = $1"
			}
			for _, test := range []interface{}{
				"a",
			} {
				id, err := convert.InsertStrNULL(test)
				if err != nil {
					t.Error(err)
					return
				}

				var value sql.NullString
				err = factory.DB().QueryRowContext(context.Background(), queryStr, id).Scan(&value)
				if err != nil {
					t.Error(err)
					return
				}

				if !value.Valid {
					t.Error("want null got ok")
				} else if value.String != "a" {
					t.Error("want a got ", value.String)
				}
			}
		})

		t.Run("string_not_1", func(t *testing.T) {
			queryStr := "SELECT field0 FROM gobatis_convert2 WHERE id = ?"
			if factory.Dialect() == dialects.Postgres {
				queryStr = "SELECT field0 FROM gobatis_convert2 WHERE id = $1"
			}
			for _, test := range []interface{}{
				"a",
				[]byte{'a'},
			} {
				id, err := convert.InsertStrNotNULL(test)
				if err != nil {
					t.Error(err)
					return
				}

				var value sql.NullString
				err = factory.DB().QueryRowContext(context.Background(), queryStr, id).Scan(&value)
				if err != nil {
					t.Error(err)
					return
				}

				if !value.Valid {
					t.Error("want null got ok")
				} else if value.String != "a" {
					t.Error("want a got ", value.String)
				}
			}
		})

		t.Run("ipaddress", func(t *testing.T) {
			queryStr := "SELECT field0 FROM gobatis_convert2"

			_, err := factory.DB().ExecContext(context.Background(), "DELETE FROM gobatis_convert2")
			if err != nil {
				t.Error(err)
				return
			}
			_, err = factory.DB().ExecContext(context.Background(), "INSERT INTO gobatis_convert2(field0) VALUES ('192.168.1.2')")
			if err != nil {
				t.Error(err)
				return
			}

			var value net.IP
			var scan = core.MakeIPScanner("field0", &value)
			err = factory.DB().QueryRowContext(context.Background(), queryStr).Scan(scan)
			if err != nil {
				t.Error(err)
				return
			}

			if value.String() != "192.168.1.2" {
				t.Error("want '92.168.1.2' got ", value.String())
			}

			_, err = factory.DB().ExecContext(context.Background(), "DELETE FROM gobatis_convert2")
			if err != nil {
				t.Error(err)
				return
			}
			_, err = factory.DB().ExecContext(context.Background(), "INSERT INTO gobatis_convert2(field0) VALUES ('192.168.1.2/0')")
			if err != nil {
				t.Error(err)
				return
			}

			value = net.IPv4(0, 0, 0, 0)
			err = factory.DB().QueryRowContext(context.Background(), queryStr).Scan(scan)
			if err != nil {
				t.Error(err)
				return
			}

			if value.String() != "192.168.1.2" {
				t.Error("want '92.168.1.2' got ", value.String())
			}

			_, err = factory.DB().ExecContext(context.Background(), "DELETE FROM gobatis_convert2")
			if err != nil {
				t.Error(err)
				return
			}
			_, err = factory.DB().ExecContext(context.Background(), "INSERT INTO gobatis_convert2(field0) VALUES ('192.168.1.A')")
			if err != nil {
				t.Error(err)
				return
			}

			value = net.IPv4(0, 0, 0, 0)
			err = factory.DB().QueryRowContext(context.Background(), queryStr).Scan(scan)
			if err == nil {
				t.Error("want error got ok")
				return
			}

			if !strings.Contains(err.Error(), "192.168.1.A") {
				t.Error("want contains '192.168.1.A' got ", err)
			}

			_, err = factory.DB().ExecContext(context.Background(), "DELETE FROM gobatis_convert2")
			if err != nil {
				t.Error(err)
				return
			}
			_, err = factory.DB().ExecContext(context.Background(), "INSERT INTO gobatis_convert2(field0) VALUES ('192.168.1.2/2')")
			if err != nil {
				t.Error(err)
				return
			}

			value = net.IPv4(0, 0, 0, 0)
			err = factory.DB().QueryRowContext(context.Background(), queryStr).Scan(scan)
			if err == nil {
				t.Error("want error got ok")
				return
			}

			if !strings.Contains(err.Error(), "192.168.1.2/2") {
				t.Error("want contains '192.168.1.2/2' got ", err)
			}
		})

		t.Run("ipnet", func(t *testing.T) {
			queryStr := "SELECT field0 FROM gobatis_convert2"

			_, err := factory.DB().ExecContext(context.Background(), "DELETE FROM gobatis_convert2")
			if err != nil {
				t.Error(err)
				return
			}
			_, err = factory.DB().ExecContext(context.Background(), "INSERT INTO gobatis_convert2(field0) VALUES ('192.168.1.2/12')")
			if err != nil {
				t.Error(err)
				return
			}

			var value net.IPNet
			var scan = core.MakeIPNetScanner("field0", &value)
			err = factory.DB().QueryRowContext(context.Background(), queryStr).Scan(scan)
			if err != nil {
				t.Error(err)
				return
			}

			if value.String() != "192.160.0.0/12" {
				t.Error("want '192.160.0.0/12' got ", value.String())
			}

			_, err = factory.DB().ExecContext(context.Background(), "DELETE FROM gobatis_convert2")
			if err != nil {
				t.Error(err)
				return
			}
			_, err = factory.DB().ExecContext(context.Background(), "INSERT INTO gobatis_convert2(field0) VALUES ('192.168.1.A/1')")
			if err != nil {
				t.Error(err)
				return
			}

			value.IP = net.IPv4(0, 0, 0, 0)
			err = factory.DB().QueryRowContext(context.Background(), queryStr).Scan(scan)
			if err == nil {
				t.Error("want error got ok")
				return
			}

			if !strings.Contains(err.Error(), "192.168.1.A") {
				t.Error("want contains '192.168.1.A' got ", err)
			}
		})

	})
}
