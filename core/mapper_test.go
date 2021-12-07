package core_test

import (
	"context"
	"database/sql"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/lib/pq"
	"github.com/runner-mei/GoBatis/core"
	"github.com/runner-mei/GoBatis/dialects"
	"github.com/runner-mei/GoBatis/tests"
)

func TestMapper(t *testing.T) {
	tests.Run(t, func(_ testing.TB, factory *core.SessionFactory) {
		ref := factory.SessionReference()
		itest := tests.NewITest(ref)

		abyid := `select field0,field1,field2,field3,field4,field5,field6,field7,field8 from gobatis_testa where id = $1`
		bbyid := `select field0,field1,field2,field3,field4,field5,field6,field7,field8 from gobatis_testb where id = $1`

		if factory.Dialect() != dialects.Postgres {
			abyid = `select field0,field1,field2,field3,field4,field5,field6,field7,field8 from gobatis_testa where id = ?`
			bbyid = `select field0,field1,field2,field3,field4,field5,field6,field7,field8 from gobatis_testb where id = ?`
		}

		t.Run("testa1 result is null", func(t *testing.T) {
			id, err := itest.InsertA1(&tests.TestA1{})
			if err != nil {
				t.Error(err)
				return
			}

			var Field0 sql.NullBool
			var Field1 sql.NullInt64
			var Field2 sql.NullInt64
			var Field3 sql.NullFloat64
			var Field4 sql.NullFloat64
			var Field5 sql.NullString
			var Field6 sql.NullString
			var Field7 sql.NullString
			var Field8 sql.NullString

			err = factory.DB().QueryRowContext(context.Background(), abyid, id).Scan(&Field0, &Field1, &Field2, &Field3, &Field4, &Field5, &Field6, &Field7, &Field8)
			if err != nil {
				t.Error(err)
				return
			}

			if !Field0.Valid || Field0.Bool {
				t.Error("want nil got", Field0.Bool)
			}
			if Field1.Valid {
				t.Error("want nil got", Field1.Int64)
			}
			if Field2.Valid {
				t.Error("want nil got", Field2.Int64)
			}
			if Field3.Valid {
				t.Error("want nil got", Field3.Float64)
			}
			if Field4.Valid {
				t.Error("want nil got", Field4.Float64)
			}
			if Field5.Valid {
				t.Error("want nil got", Field5.String)
			}

			if Field6.Valid {
				t.Error("want nil got", Field6.String)
			}

			if Field7.Valid {
				t.Error("want nil got", Field7.String)
			}
			if Field8.Valid {
				t.Error("want nil got", Field8.String)
			}
		})

		t.Run("testa1 result is not null", func(t *testing.T) {

			now := time.Now()
			a := &tests.TestA1{
				Field0: true,
				Field1: 1,
				Field2: 1,
				Field3: 1,
				Field4: 1,
				Field5: "1",
				Field6: now,
				Field7: tests.TestIP,
				Field8: tests.TestMAC,
			}
			id, err := itest.InsertA1(a)
			if err != nil {
				t.Error(err)
				return
			}

			var Field0 sql.NullBool
			var Field1 sql.NullInt64
			var Field2 sql.NullInt64
			var Field3 sql.NullFloat64
			var Field4 sql.NullFloat64
			var Field5 sql.NullString
			var Field6 time.Time
			var Field7 sql.NullString
			var Field8 sql.NullString

			err = factory.DB().QueryRowContext(context.Background(), abyid, id).Scan(&Field0, &Field1, &Field2, &Field3, &Field4, &Field5, &Field6, &Field7, &Field8)
			if err != nil {
				t.Error(err)
				return
			}

			if !Field0.Valid || a.Field0 != Field0.Bool {
				t.Error("want not nil got", Field0.Bool)
			}
			if !Field1.Valid || int64(a.Field1) != Field1.Int64 {
				t.Error("want not nil got", Field1.Int64)
			}
			if !Field2.Valid || int64(a.Field2) != Field2.Int64 {
				t.Error("want not nil got", Field2.Int64)
			}
			if !Field3.Valid || float64(a.Field3) != Field3.Float64 {
				t.Error("want not nil got", Field3.Float64)
			}
			if !Field4.Valid || a.Field4 != Field4.Float64 {
				t.Error("want not nil got", Field4.Float64)
			}
			if !Field5.Valid || a.Field5 != Field5.String {
				t.Error("want not nil got", Field5.String)
			}

			if !equalTime(a.Field6, Field6) {
				t.Error("want", a.Field6.Format(time.RFC3339), " got", Field6.Format(time.RFC3339))
			}

			if !Field7.Valid || a.Field7.String() != Field7.String {
				t.Error("want not nil got", Field7.String)
			}
			if !Field8.Valid || a.Field8.String() != Field8.String {
				t.Error("want not nil got", Field8.String)
			}
		})

		t.Run("testa2 result is null - 1", func(t *testing.T) {
			id, err := itest.InsertA2(&tests.TestA2{})
			if err != nil {
				t.Error(err)
				return
			}

			var Field0 sql.NullBool
			var Field1 sql.NullInt64
			var Field2 sql.NullInt64
			var Field3 sql.NullFloat64
			var Field4 sql.NullFloat64
			var Field5 sql.NullString
			var Field6 sql.NullString
			var Field7 sql.NullString
			var Field8 sql.NullString

			err = factory.DB().QueryRowContext(context.Background(), abyid, id).Scan(&Field0, &Field1, &Field2, &Field3, &Field4, &Field5, &Field6, &Field7, &Field8)
			if err != nil {
				t.Error(err)
				return
			}

			if Field0.Valid {
				t.Error("want nil got", Field0.Bool)
			}
			if Field1.Valid {
				t.Error("want nil got", Field1.Int64)
			}
			if Field2.Valid {
				t.Error("want nil got", Field2.Int64)
			}
			if Field3.Valid {
				t.Error("want nil got", Field3.Float64)
			}
			if Field4.Valid {
				t.Error("want nil got", Field4.Float64)
			}
			if Field5.Valid {
				t.Error("want nil got", Field5.String)
			}

			if Field6.Valid {
				t.Error("want nil got", Field6.String)
			}

			if Field7.Valid {
				t.Error("want nil got", Field7.String)
			}
			if Field8.Valid {
				t.Error("want nil got", Field8.String)
			}
		})

		t.Run("testa2 result is null - 2", func(t *testing.T) {
			a1 := tests.TestA1{}

			id, err := itest.InsertA2(&tests.TestA2{
				Field0: &a1.Field0,
				Field1: &a1.Field1,
				Field2: &a1.Field2,
				Field3: &a1.Field3,
				Field4: &a1.Field4,
				Field5: &a1.Field5,
				Field6: &a1.Field6,
				Field7: &a1.Field7,
				Field8: &a1.Field8,
			})
			if err != nil {
				t.Error(err)
				return
			}

			var Field0 sql.NullBool
			var Field1 sql.NullInt64
			var Field2 sql.NullInt64
			var Field3 sql.NullFloat64
			var Field4 sql.NullFloat64
			var Field5 sql.NullString
			var Field6 sql.NullString
			var Field7 sql.NullString
			var Field8 sql.NullString

			err = factory.DB().QueryRowContext(context.Background(), abyid, id).Scan(&Field0, &Field1, &Field2, &Field3, &Field4, &Field5, &Field6, &Field7, &Field8)
			if err != nil {
				t.Error(err)
				return
			}

			if !Field0.Valid || Field0.Bool {
				t.Error("want nil got", Field0.Bool)
			}
			if Field1.Valid {
				t.Error("want nil got", Field1.Int64)
			}
			if Field2.Valid {
				t.Error("want nil got", Field2.Int64)
			}
			if Field3.Valid {
				t.Error("want nil got", Field3.Float64)
			}
			if Field4.Valid {
				t.Error("want nil got", Field4.Float64)
			}
			if Field5.Valid {
				t.Error("want nil got", Field5.String)
			}

			if Field6.Valid {
				t.Error("want nil got", Field6.String)
			}

			if Field7.Valid {
				t.Error("want nil got", Field7.String)
			}
			if Field8.Valid {
				t.Error("want nil got", Field8.String)
			}
		})

		t.Run("testa2 result is not null", func(t *testing.T) {

			now := time.Now()
			a := &tests.TestA1{
				Field0: true,
				Field1: 1,
				Field2: 1,
				Field3: 1,
				Field4: 1,
				Field5: "1",
				Field6: now,
				Field7: tests.TestIP,
				Field8: tests.TestMAC,
			}

			id, err := itest.InsertA2(&tests.TestA2{
				Field0: &a.Field0,
				Field1: &a.Field1,
				Field2: &a.Field2,
				Field3: &a.Field3,
				Field4: &a.Field4,
				Field5: &a.Field5,
				Field6: &a.Field6,
				Field7: &a.Field7,
				Field8: &a.Field8,
			})
			if err != nil {
				t.Error(err)
				return
			}

			var Field0 sql.NullBool
			var Field1 sql.NullInt64
			var Field2 sql.NullInt64
			var Field3 sql.NullFloat64
			var Field4 sql.NullFloat64
			var Field5 sql.NullString
			var Field6 time.Time
			var Field7 sql.NullString
			var Field8 sql.NullString

			err = factory.DB().QueryRowContext(context.Background(), abyid, id).Scan(&Field0, &Field1, &Field2, &Field3, &Field4, &Field5, &Field6, &Field7, &Field8)
			if err != nil {
				t.Error(err)
				return
			}

			if !Field0.Valid || a.Field0 != Field0.Bool {
				t.Error("want not nil got", Field0.Bool)
			}
			if !Field1.Valid || int64(a.Field1) != Field1.Int64 {
				t.Error("want not nil got", Field1.Int64)
			}
			if !Field2.Valid || int64(a.Field2) != Field2.Int64 {
				t.Error("want not nil got", Field2.Int64)
			}
			if !Field3.Valid || float64(a.Field3) != Field3.Float64 {
				t.Error("want not nil got", Field3.Float64)
			}
			if !Field4.Valid || a.Field4 != Field4.Float64 {
				t.Error("want not nil got", Field4.Float64)
			}
			if !Field5.Valid || a.Field5 != Field5.String {
				t.Error("want not nil got", Field5.String)
			}

			if !equalTime(a.Field6, Field6) {
				t.Error("want ", a.Field6.Format(time.RFC3339), " got", Field6.Format(time.RFC3339))
			}

			if !Field7.Valid || a.Field7.String() != Field7.String {
				t.Error("want not nil got", Field7.String)
			}
			if !Field8.Valid || a.Field8.String() != Field8.String {
				t.Error("want not nil got", Field8.String)
			}
		})

		t.Run("testa3 result is zero value", func(t *testing.T) {
			skipFiled6 := false
			id, err := itest.InsertA3(&tests.TestA3{})
			if err != nil {
				if strings.Contains(err.Error(), "Error 1292: Incorrect datetime value: '0000-00-00' for column 'field6'") {
					id, err = itest.InsertA3(&tests.TestA3{Field6: time.Now()})
				}
				if err != nil {
					t.Error(err)
					return
				}
				skipFiled6 = true
			}

			var Field0 sql.NullBool
			var Field1 sql.NullInt64
			var Field2 sql.NullInt64
			var Field3 sql.NullFloat64
			var Field4 sql.NullFloat64
			var Field5 sql.NullString
			var Field6 time.Time
			var Field7 sql.NullString
			var Field8 sql.NullString

			err = factory.DB().QueryRowContext(context.Background(), abyid, id).Scan(&Field0, &Field1, &Field2, &Field3, &Field4, &Field5, &Field6, &Field7, &Field8)
			if err != nil {
				t.Error(err)
				return
			}

			if !Field0.Valid || Field0.Bool {
				t.Error("want nil got", Field0.Bool)
			}
			if !Field1.Valid || Field1.Int64 != 0 {
				t.Error("want nil got", Field1.Int64)
			}
			if !Field2.Valid || Field2.Int64 != 0 {
				t.Error("want nil got", Field2.Int64)
			}
			if !Field3.Valid || Field3.Float64 != 0 {
				t.Error("want nil got", Field3.Float64)
			}
			if !Field4.Valid || Field4.Float64 != 0 {
				t.Error("want nil got", Field4.Float64)
			}
			if !Field5.Valid || Field5.String != "" {
				t.Error("want nil got", Field5.String)
			}

			if !Field6.IsZero() {
				if !skipFiled6 {
					t.Error("want nil got", Field6)
				}
			}

			if Field7.Valid {
				t.Error("want nil got", Field7.String)
			}
			if Field8.Valid {
				t.Error("want nil got", Field8.String)
			}
		})

		t.Run("testa3 result is not null", func(t *testing.T) {
			now := time.Now()
			a := &tests.TestA3{
				Field0: true,
				Field1: 1,
				Field2: 1,
				Field3: 1,
				Field4: 1,
				Field5: "1",
				Field6: now,
				Field7: tests.TestIP,
				Field8: tests.TestMAC,
			}
			id, err := itest.InsertA3(a)
			if err != nil {
				t.Error(err)
				return
			}

			var Field0 sql.NullBool
			var Field1 sql.NullInt64
			var Field2 sql.NullInt64
			var Field3 sql.NullFloat64
			var Field4 sql.NullFloat64
			var Field5 sql.NullString
			var Field6 time.Time
			var Field7 sql.NullString
			var Field8 sql.NullString

			err = factory.DB().QueryRowContext(context.Background(), abyid, id).Scan(&Field0, &Field1, &Field2, &Field3, &Field4, &Field5, &Field6, &Field7, &Field8)
			if err != nil {
				t.Error(err)
				return
			}

			if !Field0.Valid || a.Field0 != Field0.Bool {
				t.Error("want not nil got", Field0.Bool)
			}
			if !Field1.Valid || int64(a.Field1) != Field1.Int64 {
				t.Error("want not nil got", Field1.Int64)
			}
			if !Field2.Valid || int64(a.Field2) != Field2.Int64 {
				t.Error("want not nil got", Field2.Int64)
			}
			if !Field3.Valid || float64(a.Field3) != Field3.Float64 {
				t.Error("want not nil got", Field3.Float64)
			}
			if !Field4.Valid || a.Field4 != Field4.Float64 {
				t.Error("want not nil got", Field4.Float64)
			}
			if !Field5.Valid || a.Field5 != Field5.String {
				t.Error("want not nil got", Field5.String)
			}

			if !equalTime(a.Field6, Field6) {
				t.Error("want", a.Field6.Format(time.RFC3339), " got", Field6.Format(time.RFC3339))
			}

			if !Field7.Valid || a.Field7.String() != Field7.String {
				t.Error("want not nil got", Field7.String)
			}
			if !Field8.Valid || a.Field8.String() != Field8.String {
				t.Error("want not nil got", Field8.String)
			}
		})

		t.Run("testa4 result is null - 1", func(t *testing.T) {
			id, err := itest.InsertA4(&tests.TestA4{})
			if err != nil {
				t.Error(err)
				return
			}

			var Field0 sql.NullBool
			var Field1 sql.NullInt64
			var Field2 sql.NullInt64
			var Field3 sql.NullFloat64
			var Field4 sql.NullFloat64
			var Field5 sql.NullString
			var Field6 sql.NullString
			var Field7 sql.NullString
			var Field8 sql.NullString

			err = factory.DB().QueryRowContext(context.Background(), abyid, id).Scan(&Field0, &Field1, &Field2, &Field3, &Field4, &Field5, &Field6, &Field7, &Field8)
			if err != nil {
				t.Error(err)
				return
			}

			if Field0.Valid {
				t.Error("want nil got", Field0.Bool)
			}
			if Field1.Valid {
				t.Error("want nil got", Field1.Int64)
			}
			if Field2.Valid {
				t.Error("want nil got", Field2.Int64)
			}
			if Field3.Valid {
				t.Error("want nil got", Field3.Float64)
			}
			if Field4.Valid {
				t.Error("want nil got", Field4.Float64)
			}
			if Field5.Valid {
				t.Error("want nil got", Field5.String)
			}

			if Field6.Valid {
				t.Error("want nil got", Field6.String)
			}

			if Field7.Valid {
				t.Error("want nil got", Field7.String)
			}
			if Field8.Valid {
				t.Error("want nil got", Field8.String)
			}
		})

		t.Run("testa4 result is null - 2", func(t *testing.T) {
			a1 := tests.TestA1{}

			skipFiled6 := false
			id, err := itest.InsertA4(&tests.TestA4{
				Field0: &a1.Field0,
				Field1: &a1.Field1,
				Field2: &a1.Field2,
				Field3: &a1.Field3,
				Field4: &a1.Field4,
				Field5: &a1.Field5,
				Field6: &a1.Field6,
				Field7: &a1.Field7,
				Field8: &a1.Field8,
			})
			if err != nil {
				if strings.Contains(err.Error(), "Error 1292: Incorrect datetime value: '0000-00-00' for column 'field6'") {
					a1.Field6 = time.Now()
					id, err = itest.InsertA4(&tests.TestA4{
						Field0: &a1.Field0,
						Field1: &a1.Field1,
						Field2: &a1.Field2,
						Field3: &a1.Field3,
						Field4: &a1.Field4,
						Field5: &a1.Field5,
						Field6: &a1.Field6,
						Field7: &a1.Field7,
						Field8: &a1.Field8,
					})
				}
				if err != nil {
					t.Error(err)
					return
				}
				skipFiled6 = true
			}

			var Field0 sql.NullBool
			var Field1 sql.NullInt64
			var Field2 sql.NullInt64
			var Field3 sql.NullFloat64
			var Field4 sql.NullFloat64
			var Field5 sql.NullString
			var Field6 time.Time
			var Field7 sql.NullString
			var Field8 sql.NullString

			err = factory.DB().QueryRowContext(context.Background(), abyid, id).Scan(&Field0, &Field1, &Field2, &Field3, &Field4, &Field5, &Field6, &Field7, &Field8)
			if err != nil {
				t.Error(err)
				return
			}

			if !Field0.Valid || Field0.Bool {
				t.Error("want nil got", Field0.Bool)
			}
			if !Field1.Valid || Field1.Int64 != 0 {
				t.Error("want nil got", Field1.Int64)
			}
			if !Field2.Valid || Field2.Int64 != 0 {
				t.Error("want nil got", Field2.Int64)
			}
			if !Field3.Valid || Field3.Float64 != 0 {
				t.Error("want nil got", Field3.Float64)
			}
			if !Field4.Valid || Field4.Float64 != 0 {
				t.Error("want nil got", Field4.Float64)
			}
			if !Field5.Valid || Field5.String != "" {
				t.Error("want nil got", Field5.String)
			}

			if !Field6.IsZero() {
				if !skipFiled6 {
					t.Error("want nil got", Field6)
				}
			}

			if Field7.Valid {
				t.Error("want nil got", Field7.String)
			}
			if Field8.Valid {
				t.Error("want nil got", Field8.String)
			}
		})

		t.Run("testa4 result is not null", func(t *testing.T) {
			now := time.Now()
			a := &tests.TestA1{
				Field0: true,
				Field1: 1,
				Field2: 1,
				Field3: 1,
				Field4: 1,
				Field5: "1",
				Field6: now,
				Field7: tests.TestIP,
				Field8: tests.TestMAC,
			}

			id, err := itest.InsertA4(&tests.TestA4{
				Field0: &a.Field0,
				Field1: &a.Field1,
				Field2: &a.Field2,
				Field3: &a.Field3,
				Field4: &a.Field4,
				Field5: &a.Field5,
				Field6: &a.Field6,
				Field7: &a.Field7,
				Field8: &a.Field8,
			})
			if err != nil {
				t.Error(err)
				return
			}

			var Field0 sql.NullBool
			var Field1 sql.NullInt64
			var Field2 sql.NullInt64
			var Field3 sql.NullFloat64
			var Field4 sql.NullFloat64
			var Field5 sql.NullString
			var Field6 time.Time
			var Field7 sql.NullString
			var Field8 sql.NullString

			err = factory.DB().QueryRowContext(context.Background(), abyid, id).Scan(&Field0, &Field1, &Field2, &Field3, &Field4, &Field5, &Field6, &Field7, &Field8)
			if err != nil {
				t.Error(err)
				return
			}

			if !Field0.Valid || a.Field0 != Field0.Bool {
				t.Error("want not nil got", Field0.Bool)
			}
			if !Field1.Valid || int64(a.Field1) != Field1.Int64 {
				t.Error("want not nil got", Field1.Int64)
			}
			if !Field2.Valid || int64(a.Field2) != Field2.Int64 {
				t.Error("want not nil got", Field2.Int64)
			}
			if !Field3.Valid || float64(a.Field3) != Field3.Float64 {
				t.Error("want not nil got", Field3.Float64)
			}
			if !Field4.Valid || a.Field4 != Field4.Float64 {
				t.Error("want not nil got", Field4.Float64)
			}
			if !Field5.Valid || a.Field5 != Field5.String {
				t.Error("want not nil got", Field5.String)
			}

			if !equalTime(a.Field6, Field6) {
				t.Error("want", a.Field6.Format(time.RFC3339), " got", Field6.Format(time.RFC3339))
			}

			if !Field7.Valid || a.Field7.String() != Field7.String {
				t.Error("want not nil got", Field7.String)
			}
			if !Field8.Valid || a.Field8.String() != Field8.String {
				t.Error("want not nil got", Field8.String)
			}
		})

		t.Run("TestB1 result is null", func(t *testing.T) {
			now := time.Now()

			_, err := itest.InsertB1(&tests.TestB1{
				Field0: false,
				// Field1: 1,
				// Field2: 1,
				// Field3: 1,
				// Field4: 1,
				// Field5: "1",
				// Field6: now,
				// Field7: tests.TestIP,
				// Field8: tests.TestMAC,
			})
			if err == nil {
				t.Error("want err got ok")
			} else if !strings.Contains(err.Error(), "field 'Field1") {
				t.Error("want contains 'Field1' got", err)
			}

			_, err = itest.InsertB1(&tests.TestB1{
				Field0: false,
				Field1: 1,
				// Field2: 1,
				// Field3: 1,
				// Field4: 1,
				// Field5: "1",
				// Field6: now,
				// Field7: tests.TestIP,
				// Field8: tests.TestMAC,
			})
			if err == nil {
				t.Error("want err got ok")
			} else if !strings.Contains(err.Error(), "field 'Field2") {
				t.Error("want contains 'Field2' got", err)
			}

			_, err = itest.InsertB1(&tests.TestB1{
				Field0: false,
				Field1: 1,
				Field2: 1,
				// Field3: 1,
				// Field4: 1,
				// Field5: "1",
				// Field6: now,
				// Field7: tests.TestIP,
				// Field8: tests.TestMAC,
			})
			if err == nil {
				t.Error("want err got ok")
			} else if !strings.Contains(err.Error(), "field 'Field3") {
				t.Error("want contains 'Field3' got", err)
			}

			_, err = itest.InsertB1(&tests.TestB1{
				Field0: false,
				Field1: 1,
				Field2: 1,
				Field3: 1,
				// Field4: 1,
				// Field5: "1",
				// Field6: now,
				// Field7: tests.TestIP,
				// Field8: tests.TestMAC,
			})
			if err == nil {
				t.Error("want err got ok")
			} else if !strings.Contains(err.Error(), "field 'Field4") {
				t.Error("want contains 'Field4' got", err)
			}

			_, err = itest.InsertB1(&tests.TestB1{
				Field0: false,
				Field1: 1,
				Field2: 1,
				Field3: 1,
				Field4: 1,
				// Field5: "1",
				// Field6: now,
				// Field7: tests.TestIP,
				// Field8: tests.TestMAC,
			})
			if err == nil {
				t.Error("want err got ok")
			} else if !strings.Contains(err.Error(), "field 'Field5") {
				t.Error("want contains 'Field5' got", err)
			}

			_, err = itest.InsertB1(&tests.TestB1{
				Field0: false,
				Field1: 1,
				Field2: 1,
				Field3: 1,
				Field4: 1,
				Field5: "1",
				// Field6: now,
				// Field7: tests.TestIP,
				// Field8: tests.TestMAC,
			})
			if err == nil {
				t.Error("want err got ok")
			} else if !strings.Contains(err.Error(), "field 'Field6") {
				t.Error("want contains 'Field6' got", err)
			}

			_, err = itest.InsertB1(&tests.TestB1{
				Field0: false,
				Field1: 1,
				Field2: 1,
				Field3: 1,
				Field4: 1,
				Field5: "1",
				Field6: now,
				// Field7: tests.TestIP,
				// Field8: tests.TestMAC,
			})
			if err == nil {
				t.Error("want err got ok")
			} else if !strings.Contains(err.Error(), "field 'Field7") {
				t.Error("want contains 'Field7' got", err)
			}

			_, err = itest.InsertB1(&tests.TestB1{
				Field0: false,
				Field1: 1,
				Field2: 1,
				Field3: 1,
				Field4: 1,
				Field5: "1",
				Field6: now,
				Field7: tests.TestIP,
				// Field8: tests.TestMAC,
			})
			if err == nil {
				t.Error("want err got ok")
			} else if !strings.Contains(err.Error(), "field 'Field8") {
				t.Error("want contains 'Field8' got", err)
			}
		})

		t.Run("TestB1 result is not null", func(t *testing.T) {

			now := time.Now()
			a := &tests.TestB1{
				Field0: true,
				Field1: 1,
				Field2: 1,
				Field3: 1,
				Field4: 1,
				Field5: "1",
				Field6: now,
				Field7: tests.TestIP,
				Field8: tests.TestMAC,
			}
			id, err := itest.InsertB1(a)
			if err != nil {
				t.Error(err)
				return
			}

			var Field0 sql.NullBool
			var Field1 sql.NullInt64
			var Field2 sql.NullInt64
			var Field3 sql.NullFloat64
			var Field4 sql.NullFloat64
			var Field5 sql.NullString
			var Field6 time.Time
			var Field7 sql.NullString
			var Field8 sql.NullString

			err = factory.DB().QueryRowContext(context.Background(), bbyid, id).Scan(&Field0, &Field1, &Field2, &Field3, &Field4, &Field5, &Field6, &Field7, &Field8)
			if err != nil {
				t.Error(err)
				return
			}

			if !Field0.Valid || a.Field0 != Field0.Bool {
				t.Error("want not nil got", Field0.Bool)
			}
			if !Field1.Valid || int64(a.Field1) != Field1.Int64 {
				t.Error("want not nil got", Field1.Int64)
			}
			if !Field2.Valid || int64(a.Field2) != Field2.Int64 {
				t.Error("want not nil got", Field2.Int64)
			}
			if !Field3.Valid || float64(a.Field3) != Field3.Float64 {
				t.Error("want not nil got", Field3.Float64)
			}
			if !Field4.Valid || a.Field4 != Field4.Float64 {
				t.Error("want not nil got", Field4.Float64)
			}
			if !Field5.Valid || a.Field5 != Field5.String {
				t.Error("want not nil got", Field5.String)
			}

			if !equalTime(a.Field6, Field6) {
				t.Error("want", a.Field6.Format(time.RFC3339), " got", Field6.Format(time.RFC3339))
			}

			if !Field7.Valid || a.Field7.String() != Field7.String {
				t.Error("want not nil got", Field7.String)
			}
			if !Field8.Valid || a.Field8.String() != Field8.String {
				t.Error("want not nil got", Field8.String)
			}
		})

		t.Run("testb2 result is null - 1", func(t *testing.T) {

			now := time.Now()
			a := &tests.TestB1{
				Field0: true,
				Field1: 1,
				Field2: 1,
				Field3: 1,
				Field4: 1,
				Field5: "1",
				Field6: now,
				Field7: tests.TestIP,
				Field8: tests.TestMAC,
			}

			_, err := itest.InsertB2(&tests.TestB2{})
			if err == nil {
				t.Error("want err got ok")
			} else if !strings.Contains(err.Error(), "field 'Field0") {
				t.Error("want contains 'Field0' got", err)
			}

			_, err = itest.InsertB2(&tests.TestB2{
				Field0: &a.Field0,
				// Field1: &a.Field1,
				// Field2: &a.Field2,
				// Field3: &a.Field3,
				// Field4: &a.Field4,
				// Field5: &a.Field5,
				// Field6: &a.Field6,
				// Field7: &a.Field7,
				// Field8: &a.Field8,
			})
			if err == nil {
				t.Error("want err got ok")
			} else if !strings.Contains(err.Error(), "field 'Field1") {
				t.Error("want contains 'Field1' got", err)
			}

			_, err = itest.InsertB2(&tests.TestB2{
				Field0: &a.Field0,
				Field1: &a.Field1,
				// Field2: &a.Field2,
				// Field3: &a.Field3,
				// Field4: &a.Field4,
				// Field5: &a.Field5,
				// Field6: &a.Field6,
				// Field7: &a.Field7,
				// Field8: &a.Field8,
			})
			if err == nil {
				t.Error("want err got ok")
			} else if !strings.Contains(err.Error(), "field 'Field2") {
				t.Error("want contains 'Field2' got", err)
			}

			_, err = itest.InsertB2(&tests.TestB2{
				Field0: &a.Field0,
				Field1: &a.Field1,
				Field2: &a.Field2,
				// Field3: &a.Field3,
				// Field4: &a.Field4,
				// Field5: &a.Field5,
				// Field6: &a.Field6,
				// Field7: &a.Field7,
				// Field8: &a.Field8,
			})
			if err == nil {
				t.Error("want err got ok")
			} else if !strings.Contains(err.Error(), "field 'Field3") {
				t.Error("want contains 'Field3' got", err)
			}

			_, err = itest.InsertB2(&tests.TestB2{
				Field0: &a.Field0,
				Field1: &a.Field1,
				Field2: &a.Field2,
				Field3: &a.Field3,
				// Field4: &a.Field4,
				// Field5: &a.Field5,
				// Field6: &a.Field6,
				// Field7: &a.Field7,
				// Field8: &a.Field8,
			})
			if err == nil {
				t.Error("want err got ok")
			} else if !strings.Contains(err.Error(), "field 'Field4") {
				t.Error("want contains 'Field4' got", err)
			}

			_, err = itest.InsertB2(&tests.TestB2{
				Field0: &a.Field0,
				Field1: &a.Field1,
				Field2: &a.Field2,
				Field3: &a.Field3,
				Field4: &a.Field4,
				// Field5: &a.Field5,
				// Field6: &a.Field6,
				// Field7: &a.Field7,
				// Field8: &a.Field8,
			})
			if err == nil {
				t.Error("want err got ok")
			} else if !strings.Contains(err.Error(), "field 'Field5") {
				t.Error("want contains 'Field5' got", err)
			}

			_, err = itest.InsertB2(&tests.TestB2{
				Field0: &a.Field0,
				Field1: &a.Field1,
				Field2: &a.Field2,
				Field3: &a.Field3,
				Field4: &a.Field4,
				Field5: &a.Field5,
				// Field6: &a.Field6,
				// Field7: &a.Field7,
				// Field8: &a.Field8,
			})
			if err == nil {
				t.Error("want err got ok")
			} else if !strings.Contains(err.Error(), "field 'Field6") {
				t.Error("want contains 'Field6' got", err)
			}

			_, err = itest.InsertB2(&tests.TestB2{
				Field0: &a.Field0,
				Field1: &a.Field1,
				Field2: &a.Field2,
				Field3: &a.Field3,
				Field4: &a.Field4,
				Field5: &a.Field5,
				Field6: &a.Field6,
				// Field7: &a.Field7,
				// Field8: &a.Field8,
			})
			if err == nil {
				t.Error("want err got ok")
			} else if !strings.Contains(err.Error(), "field 'Field7") {
				t.Error("want contains 'Field7' got", err)
			}

			_, err = itest.InsertB2(&tests.TestB2{
				Field0: &a.Field0,
				Field1: &a.Field1,
				Field2: &a.Field2,
				Field3: &a.Field3,
				Field4: &a.Field4,
				Field5: &a.Field5,
				Field6: &a.Field6,
				Field7: &a.Field7,
				// Field8: &a.Field8,
			})
			if err == nil {
				t.Error("want err got ok")
			} else if !strings.Contains(err.Error(), "field 'Field8") {
				t.Error("want contains 'Field8' got", err)
			}
		})

		t.Run("testb2 result is null - 2", func(t *testing.T) {
			a1 := tests.TestB1{}
			now := time.Now()
			a := &tests.TestB1{
				Field0: true,
				Field1: 1,
				Field2: 1,
				Field3: 1,
				Field4: 1,
				Field5: "1",
				Field6: now,
				Field7: tests.TestIP,
				Field8: tests.TestMAC,
			}

			// _, err := itest.InsertB2(&tests.TestB2{
			// 	Field0: &a1.Field0,
			// 	Field1: &a1.Field1,
			// 	Field2: &a1.Field2,
			// 	Field3: &a1.Field3,
			// 	Field4: &a1.Field4,
			// 	Field5: &a1.Field5,
			// 	Field6: &a1.Field6,
			// 	Field7: &a1.Field7,
			// 	Field8: &a1.Field8,
			// })
			// if err == nil {
			// 	t.Error("want err got ok")
			// } else if !strings.Contains(err.Error(), "field 'Field0") {
			// 	t.Error("want contains 'Field0' got", err)
			// }

			_, err := itest.InsertB2(&tests.TestB2{
				Field0: &a.Field0,
				Field1: &a1.Field1,
				Field2: &a1.Field2,
				Field3: &a1.Field3,
				Field4: &a1.Field4,
				Field5: &a1.Field5,
				Field6: &a1.Field6,
				Field7: &a1.Field7,
				Field8: &a1.Field8,
			})
			if err == nil {
				t.Error("want err got ok")
			} else if !strings.Contains(err.Error(), "field 'Field1") {
				t.Error("want contains 'Field1' got", err)
			}

			_, err = itest.InsertB2(&tests.TestB2{
				Field0: &a.Field0,
				Field1: &a.Field1,
				Field2: &a1.Field2,
				Field3: &a1.Field3,
				Field4: &a1.Field4,
				Field5: &a1.Field5,
				Field6: &a1.Field6,
				Field7: &a1.Field7,
				Field8: &a1.Field8,
			})
			if err == nil {
				t.Error("want err got ok")
			} else if !strings.Contains(err.Error(), "field 'Field2") {
				t.Error("want contains 'Field2' got", err)
			}

			_, err = itest.InsertB2(&tests.TestB2{
				Field0: &a.Field0,
				Field1: &a.Field1,
				Field2: &a.Field2,
				Field3: &a1.Field3,
				Field4: &a1.Field4,
				Field5: &a1.Field5,
				Field6: &a1.Field6,
				Field7: &a1.Field7,
				Field8: &a1.Field8,
			})
			if err == nil {
				t.Error("want err got ok")
			} else if !strings.Contains(err.Error(), "field 'Field3") {
				t.Error("want contains 'Field3' got", err)
			}

			_, err = itest.InsertB2(&tests.TestB2{
				Field0: &a.Field0,
				Field1: &a.Field1,
				Field2: &a.Field2,
				Field3: &a.Field3,
				Field4: &a1.Field4,
				Field5: &a1.Field5,
				Field6: &a1.Field6,
				Field7: &a1.Field7,
				Field8: &a1.Field8,
			})
			if err == nil {
				t.Error("want err got ok")
			} else if !strings.Contains(err.Error(), "field 'Field4") {
				t.Error("want contains 'Field4' got", err)
			}

			_, err = itest.InsertB2(&tests.TestB2{
				Field0: &a.Field0,
				Field1: &a.Field1,
				Field2: &a.Field2,
				Field3: &a.Field3,
				Field4: &a.Field4,
				Field5: &a1.Field5,
				Field6: &a1.Field6,
				Field7: &a1.Field7,
				Field8: &a1.Field8,
			})
			if err == nil {
				t.Error("want err got ok")
			} else if !strings.Contains(err.Error(), "field 'Field5") {
				t.Error("want contains 'Field5' got", err)
			}

			_, err = itest.InsertB2(&tests.TestB2{
				Field0: &a.Field0,
				Field1: &a.Field1,
				Field2: &a.Field2,
				Field3: &a.Field3,
				Field4: &a.Field4,
				Field5: &a.Field5,
				Field6: &a1.Field6,
				Field7: &a1.Field7,
				Field8: &a1.Field8,
			})
			if err == nil {
				t.Error("want err got ok")
			} else if !strings.Contains(err.Error(), "field 'Field6") {
				t.Error("want contains 'Field6' got", err)
			}

			_, err = itest.InsertB2(&tests.TestB2{
				Field0: &a.Field0,
				Field1: &a.Field1,
				Field2: &a.Field2,
				Field3: &a.Field3,
				Field4: &a.Field4,
				Field5: &a.Field5,
				Field6: &a.Field6,
				Field7: &a1.Field7,
				Field8: &a1.Field8,
			})
			if err == nil {
				t.Error("want err got ok")
			} else if !strings.Contains(err.Error(), "field 'Field7") {
				t.Error("want contains 'Field7' got", err)
			}

			_, err = itest.InsertB2(&tests.TestB2{
				Field0: &a.Field0,
				Field1: &a.Field1,
				Field2: &a.Field2,
				Field3: &a.Field3,
				Field4: &a.Field4,
				Field5: &a.Field5,
				Field6: &a.Field6,
				Field7: &a.Field7,
				Field8: &a1.Field8,
			})
			if err == nil {
				t.Error("want err got ok")
			} else if !strings.Contains(err.Error(), "field 'Field8") {
				t.Error("want contains 'Field8' got", err)
			}

		})

		t.Run("testb2 result is not null", func(t *testing.T) {
			now := time.Now()
			a := &tests.TestB1{
				Field0: true,
				Field1: 1,
				Field2: 1,
				Field3: 1,
				Field4: 1,
				Field5: "1",
				Field6: now,
				Field7: tests.TestIP,
				Field8: tests.TestMAC,
			}

			id, err := itest.InsertB2(&tests.TestB2{
				Field0: &a.Field0,
				Field1: &a.Field1,
				Field2: &a.Field2,
				Field3: &a.Field3,
				Field4: &a.Field4,
				Field5: &a.Field5,
				Field6: &a.Field6,
				Field7: &a.Field7,
				Field8: &a.Field8,
			})
			if err != nil {
				t.Error(err)
				return
			}

			var Field0 sql.NullBool
			var Field1 sql.NullInt64
			var Field2 sql.NullInt64
			var Field3 sql.NullFloat64
			var Field4 sql.NullFloat64
			var Field5 sql.NullString
			var Field6 time.Time
			var Field7 sql.NullString
			var Field8 sql.NullString

			err = factory.DB().QueryRowContext(context.Background(), bbyid, id).Scan(&Field0, &Field1, &Field2, &Field3, &Field4, &Field5, &Field6, &Field7, &Field8)
			if err != nil {
				t.Error(err)
				return
			}

			if !Field0.Valid || a.Field0 != Field0.Bool {
				t.Error("want not nil got", Field0.Bool)
			}
			if !Field1.Valid || int64(a.Field1) != Field1.Int64 {
				t.Error("want not nil got", Field1.Int64)
			}
			if !Field2.Valid || int64(a.Field2) != Field2.Int64 {
				t.Error("want not nil got", Field2.Int64)
			}
			if !Field3.Valid || float64(a.Field3) != Field3.Float64 {
				t.Error("want not nil got", Field3.Float64)
			}
			if !Field4.Valid || a.Field4 != Field4.Float64 {
				t.Error("want not nil got", Field4.Float64)
			}
			if !Field5.Valid || a.Field5 != Field5.String {
				t.Error("want not nil got", Field5.String)
			}

			if !equalTime(a.Field6, Field6) {
				t.Error("want", a.Field6.Format(time.RFC3339), " got", Field6.Format(time.RFC3339))
			}

			if !Field7.Valid || a.Field7.String() != Field7.String {
				t.Error("want not nil got", Field7.String)
			}
			if !Field8.Valid || a.Field8.String() != Field8.String {
				t.Error("want not nil got", Field8.String)
			}
		})

	})
}

func equalTime(a, b time.Time) bool {
	return a.Format("2006-01-02") == b.Format("2006-01-02")
}

func TestMapperFail(t *testing.T) {
	tests.Run(t, func(_ testing.TB, factory *core.SessionFactory) {
		ref := factory.SessionReference()
		itest := tests.NewITest(ref)

		_, err := itest.InsertFail1(&tests.Testfail1{})
		if err == nil {
			t.Error("want err got ok")
		} else if !strings.Contains(err.Error(), "field 'Field0") {
			t.Error("want contains 'Field0' got", err)
		}

		_, err = itest.InsertFail2(&tests.Testfail2{})
		if err == nil {
			t.Error("want err got ok")
		} else if !strings.Contains(err.Error(), "field 'Field0") {
			t.Error("want contains 'Field0' got", err)
		}

		_, err = itest.InsertFail3(&tests.Testfail3{})
		if err == nil {
			t.Error("want err got ok")
		} else if !strings.Contains(err.Error(), "field 'Field0") {
			t.Error("want contains 'Field0' got", err)
		}
	})
}

func TestMapperC(t *testing.T) {
	tests.Run(t, func(_ testing.TB, factory *core.SessionFactory) {
		ref := factory.SessionReference()
		itest := tests.NewITest(ref)

		abyid := `select field0 from gobatis_testc where id = $1`

		if factory.Dialect() != dialects.Postgres {
			abyid = `select field0 from gobatis_testc where id = ?`
		}

		t.Run("testc1 result is null", func(t *testing.T) {
			id, err := itest.InsertC1(&tests.TestC1{})
			if err != nil {
				t.Error(err)
				return
			}

			var Field0 sql.NullString
			err = factory.DB().QueryRowContext(context.Background(), abyid, id).Scan(&Field0)
			if err != nil {
				t.Error(err)
				return
			}

			if Field0.Valid {
				t.Error("want nil got", Field0.String)
			}
		})

		t.Run("testc1 result is not null", func(t *testing.T) {
			id, err := itest.InsertC1(&tests.TestC1{Field0: map[string]interface{}{"a": "b"}})
			if err != nil {
				t.Error(err)
				return
			}

			var Field0 sql.NullString
			err = factory.DB().QueryRowContext(context.Background(), abyid, id).Scan(&Field0)
			if err != nil {
				t.Error(err)
				return
			}

			if !Field0.Valid || Field0.String != `{"a":"b"}` {
				t.Error("want nil got", Field0.String)
			}
		})

		t.Run("testc1 result is json fail", func(t *testing.T) {
			_, err := itest.InsertC1(&tests.TestC1{Field0: map[string]interface{}{"a": func() {}}})
			if err == nil {
				t.Error("want err got ok")
			} else if !strings.Contains(err.Error(), "field 'Field0") {
				t.Error("want contains 'Field0' got", err)
			}
		})

		t.Run("testc2 result is null", func(t *testing.T) {
			_, err := itest.InsertC2(&tests.TestC2{})
			if err == nil {
				t.Error("want err got ok")
			} else if !strings.Contains(err.Error(), "field 'Field0") {
				t.Error("want contains 'Field0' got", err)
			}
		})

		t.Run("testc2 result is not null", func(t *testing.T) {
			id, err := itest.InsertC2(&tests.TestC2{Field0: map[string]interface{}{"a": "b"}})
			if err != nil {
				t.Error(err)
				return
			}

			var Field0 sql.NullString
			err = factory.DB().QueryRowContext(context.Background(), abyid, id).Scan(&Field0)
			if err != nil {
				t.Error(err)
				return
			}

			if !Field0.Valid || Field0.String != `{"a":"b"}` {
				t.Error("want nil got", Field0.String)
			}
		})

		t.Run("testc2 result is json fail", func(t *testing.T) {
			_, err := itest.InsertC2(&tests.TestC2{Field0: map[string]interface{}{"a": func() {}}})
			if err == nil {
				t.Error("want err got ok")
			} else if !strings.Contains(err.Error(), "field 'Field0") {
				t.Error("want contains 'Field0' got", err)
			}
		})

		t.Run("testc3 result is null", func(t *testing.T) {
			id, err := itest.InsertC3(&tests.TestC3{})
			if err != nil {
				t.Error(err)
				return
			}

			var Field0 sql.NullString
			err = factory.DB().QueryRowContext(context.Background(), abyid, id).Scan(&Field0)
			if err != nil {
				t.Error(err)
				return
			}

			if Field0.Valid {
				t.Error("want nil got", Field0.String)
			}
		})

		t.Run("testc3 result is not null", func(t *testing.T) {
			id, err := itest.InsertC3(&tests.TestC3{Field0: map[string]interface{}{"a": "b"}})
			if err != nil {
				t.Error(err)
				return
			}

			var Field0 sql.NullString
			err = factory.DB().QueryRowContext(context.Background(), abyid, id).Scan(&Field0)
			if err != nil {
				t.Error(err)
				return
			}

			if !Field0.Valid || Field0.String != `{"a":"b"}` {
				t.Error("want nil got", Field0.String)
			}
		})

		t.Run("testc3 result is json fail", func(t *testing.T) {
			_, err := itest.InsertC3(&tests.TestC3{Field0: map[string]interface{}{"a": func() {}}})
			if err == nil {
				t.Error("want err got ok")
			} else if !strings.Contains(err.Error(), "field 'Field0") {
				t.Error("want contains 'Field0' got", err)
			}
		})

		t.Run("testc4 result is null", func(t *testing.T) {
			id, err := itest.InsertC4(&tests.TestC4{})
			if err != nil {
				t.Error(err)
				return
			}

			var Field0 sql.NullString
			err = factory.DB().QueryRowContext(context.Background(), abyid, id).Scan(&Field0)
			if err != nil {
				t.Error(err)
				return
			}

			if Field0.Valid {
				t.Error("want nil got", Field0.String)
			}
		})

		t.Run("testc4 result is not null", func(t *testing.T) {
			id, err := itest.InsertC4(&tests.TestC4{Field0: &tests.Data{A: 1}})
			if err != nil {
				t.Error(err)
				return
			}

			var Field0 sql.NullString
			err = factory.DB().QueryRowContext(context.Background(), abyid, id).Scan(&Field0)
			if err != nil {
				t.Error(err)
				return
			}

			if !Field0.Valid || Field0.String != `{"A":1}` {
				t.Error("want nil got", Field0.String)
			}
		})

		t.Run("testc5 result is null", func(t *testing.T) {
			_, err := itest.InsertC5(&tests.TestC5{})
			if err == nil {
				t.Error("want err got ok")
			} else if !strings.Contains(err.Error(), "field 'Field0") {
				t.Error("want contains 'Field0' got", err)
			}
		})

		t.Run("testc5 result is not null", func(t *testing.T) {
			id, err := itest.InsertC5(&tests.TestC5{Field0: &tests.Data{A: 1}})
			if err != nil {
				t.Error(err)
				return
			}

			var Field0 sql.NullString
			err = factory.DB().QueryRowContext(context.Background(), abyid, id).Scan(&Field0)
			if err != nil {
				t.Error(err)
				return
			}

			if !Field0.Valid || Field0.String != `{"A":1}` {
				t.Error("want nil got", Field0.String)
			}
		})

		t.Run("testc6 result is null", func(t *testing.T) {
			id, err := itest.InsertC6(&tests.TestC6{})
			if err != nil {
				t.Error(err)
				return
			}

			var Field0 sql.NullString
			err = factory.DB().QueryRowContext(context.Background(), abyid, id).Scan(&Field0)
			if err != nil {
				t.Error(err)
				return
			}

			if Field0.Valid {
				t.Error("want nil got", Field0.String)
			}
		})

		t.Run("testc6 result is not null", func(t *testing.T) {
			id, err := itest.InsertC6(&tests.TestC6{Field0: &tests.Data{A: 1}})
			if err != nil {
				t.Error(err)
				return
			}

			var Field0 sql.NullString
			err = factory.DB().QueryRowContext(context.Background(), abyid, id).Scan(&Field0)
			if err != nil {
				t.Error(err)
				return
			}

			if !Field0.Valid || Field0.String != `{"A":1}` {
				t.Error("want nil got", Field0.String)
			}
		})

		t.Run("TestC7 result is null", func(t *testing.T) {
			id, err := itest.InsertC7(&tests.TestC7{})
			if err != nil {
				t.Error(err)
				return
			}

			var Field0 sql.NullString
			err = factory.DB().QueryRowContext(context.Background(), abyid, id).Scan(&Field0)
			if err != nil {
				t.Error(err)
				return
			}

			if !Field0.Valid || Field0.String != `{"A":0}` {
				t.Error("want nil got", Field0.String)
			}
		})

		t.Run("TestC7 result is not null", func(t *testing.T) {
			id, err := itest.InsertC7(&tests.TestC7{Field0: tests.Data{A: 1}})
			if err != nil {
				t.Error(err)
				return
			}

			var Field0 sql.NullString
			err = factory.DB().QueryRowContext(context.Background(), abyid, id).Scan(&Field0)
			if err != nil {
				t.Error(err)
				return
			}

			if !Field0.Valid || Field0.String != `{"A":1}` {
				t.Error("want nil got", Field0.String)
			}
		})

		// t.Run("TestC8 result is null", func(t *testing.T) {
		// 	_, err := itest.InsertC8(&tests.TestC8{})
		// 	if err == nil {
		// 		t.Error("want err got ok")
		// 	} else if !strings.Contains(err.Error(), "field 'Field0") {
		// 		t.Error("want contains 'Field0' got", err)
		// 	}
		// })

		t.Run("TestC8 result is not null", func(t *testing.T) {
			id, err := itest.InsertC8(&tests.TestC8{Field0: tests.Data{A: 1}})
			if err != nil {
				t.Error(err)
				return
			}

			var Field0 sql.NullString
			err = factory.DB().QueryRowContext(context.Background(), abyid, id).Scan(&Field0)
			if err != nil {
				t.Error(err)
				return
			}

			if !Field0.Valid || Field0.String != `{"A":1}` {
				t.Error("want nil got", Field0.String)
			}
		})

		t.Run("TestC9 result is null", func(t *testing.T) {
			id, err := itest.InsertC9(&tests.TestC9{})
			if err != nil {
				t.Error(err)
				return
			}

			var Field0 sql.NullString
			err = factory.DB().QueryRowContext(context.Background(), abyid, id).Scan(&Field0)
			if err != nil {
				t.Error(err)
				return
			}

			if !Field0.Valid || Field0.String != `{"A":0}` {
				t.Error("want nil got", Field0.String)
			}
		})

		t.Run("TestC9 result is not null", func(t *testing.T) {
			id, err := itest.InsertC9(&tests.TestC9{Field0: tests.Data{A: 1}})
			if err != nil {
				t.Error(err)
				return
			}

			var Field0 sql.NullString
			err = factory.DB().QueryRowContext(context.Background(), abyid, id).Scan(&Field0)
			if err != nil {
				t.Error(err)
				return
			}

			if !Field0.Valid || Field0.String != `{"A":1}` {
				t.Error("want nil got", Field0.String)
			}
		})

		t.Run("TestD1", func(t *testing.T) {
			id, err := itest.InsertD1(&tests.TestD1{})
			if err != nil {
				t.Error(err)
				return
			}

			var Field0 sql.NullString
			err = factory.DB().QueryRowContext(context.Background(), abyid, id).Scan(&Field0)
			if err != nil {
				t.Error(err)
				return
			}

			if !Field0.Valid || Field0.String != `{"A":0}` {
				t.Error("want nil got", Field0.String)
			}
		})

		t.Run("TestD2 is null", func(t *testing.T) {
			id, err := itest.InsertD2(&tests.TestD2{})
			if err != nil {
				t.Error(err)
				return
			}

			var Field0 sql.NullString
			err = factory.DB().QueryRowContext(context.Background(), abyid, id).Scan(&Field0)
			if err != nil {
				t.Error(err)
				return
			}

			if Field0.Valid {
				t.Error("want nil got", Field0.String)
			}
		})

		t.Run("TestD2 is not null", func(t *testing.T) {
			id, err := itest.InsertD2(&tests.TestD2{Field0: &tests.DriverData1{A: 1}})
			if err != nil {
				t.Error(err)
				return
			}

			var Field0 sql.NullString
			err = factory.DB().QueryRowContext(context.Background(), abyid, id).Scan(&Field0)
			if err != nil {
				t.Error(err)
				return
			}

			if !Field0.Valid || Field0.String != `{"A":1}` {
				t.Error("want nil got", Field0.String)
			}
		})

		t.Run("TestD3", func(t *testing.T) {
			id, err := itest.InsertD3(&tests.TestD3{})
			if err != nil {
				t.Error(err)
				return
			}

			var Field0 sql.NullString
			err = factory.DB().QueryRowContext(context.Background(), abyid, id).Scan(&Field0)
			if err != nil {
				t.Error(err)
				return
			}

			if !Field0.Valid || Field0.String != `{"A":0}` {
				t.Error("want nil got", Field0.String)
			}
		})

		t.Run("TestD4 is null", func(t *testing.T) {
			id, err := itest.InsertD4(&tests.TestD4{})
			if err != nil {
				t.Error(err)
				return
			}

			var Field0 sql.NullString
			err = factory.DB().QueryRowContext(context.Background(), abyid, id).Scan(&Field0)
			if err != nil {
				t.Error(err)
				return
			}

			if Field0.Valid {
				t.Error("want nil got", Field0.String)
			}
		})

		t.Run("TestD4 is not null", func(t *testing.T) {
			id, err := itest.InsertD4(&tests.TestD4{Field0: &tests.DriverData2{A: 1}})
			if err != nil {
				t.Error(err)
				return
			}

			var Field0 sql.NullString
			err = factory.DB().QueryRowContext(context.Background(), abyid, id).Scan(&Field0)
			if err != nil {
				t.Error(err)
				return
			}

			if !Field0.Valid || Field0.String != `{"A":1}` {
				t.Error("want nil got", Field0.String)
			}
		})

	})
}

func TestMapperE(t *testing.T) {
	tests.Run(t, func(_ testing.TB, factory *core.SessionFactory) {
		ref := factory.SessionReference()
		itest := tests.NewITest(ref)

		makeScanner := func(value interface{}) interface{} {
			return pq.Array(value)
		}
		abyid := `select field0 from gobatis_teste1 where id = $1`
		if factory.Dialect() != dialects.Postgres {
			abyid = `select field0 from gobatis_teste1 where id = ?`

			makeScanner = func(value interface{}) interface{} {
				return core.MakJSONScanner("field0", value)
			}
		}

		t.Run("teste1 result is null", func(t *testing.T) {
			id, err := itest.InsertE1(&tests.TestE1{})
			if err != nil {
				t.Error(err)
				return
			}

			var Field0 []int64
			err = factory.DB().QueryRowContext(context.Background(), abyid, id).Scan(makeScanner(&Field0))
			if err != nil {
				t.Error(err)
				return
			}

			if Field0 != nil || len(Field0) != 0 {
				t.Error("want nil got", Field0)
			}
		})

		t.Run("teste1 result is empty", func(t *testing.T) {
			id, err := itest.InsertE1(&tests.TestE1{Field0: []int64{}})
			if err != nil {
				t.Error(err)
				return
			}

			var Field0 []int64
			err = factory.DB().QueryRowContext(context.Background(), abyid, id).Scan(makeScanner(&Field0))
			if err != nil {
				t.Error(err)
				return
			}

			if Field0 == nil {
				t.Error("want not nil got", Field0)
			}

			if len(Field0) != 0 {
				t.Error("want empty got", Field0)
			}
		})

		t.Run("teste1 result is not null", func(t *testing.T) {
			id, err := itest.InsertE1(&tests.TestE1{Field0: []int64{123, 456}})
			if err != nil {
				t.Error(err)
				return
			}

			var Field0 []int64
			err = factory.DB().QueryRowContext(context.Background(), abyid, id).Scan(makeScanner(&Field0))
			if err != nil {
				t.Error(err)
				return
			}

			if !reflect.DeepEqual(Field0, []int64{123, 456}) {
				t.Error("want nil got", Field0)
			}
		})

		t.Run("teste2 result is null", func(t *testing.T) {
			id, err := itest.InsertE2(&tests.TestE2{})
			if err != nil {
				t.Error(err)
				return
			}

			var Field0 []int64
			err = factory.DB().QueryRowContext(context.Background(), abyid, id).Scan(makeScanner(&Field0))
			if err != nil {
				t.Error(err)
				return
			}

			if Field0 != nil {
				t.Error("want nil got", Field0)
			}
		})

		t.Run("teste2 result is empty", func(t *testing.T) {
			value := []int64{}
			id, err := itest.InsertE2(&tests.TestE2{Field0: &value})
			if err != nil {
				t.Error(err)
				return
			}

			var Field0 []int64
			err = factory.DB().QueryRowContext(context.Background(), abyid, id).Scan(makeScanner(&Field0))
			if err != nil {
				t.Error(err)
				return
			}

			if Field0 == nil {
				t.Error("want not nil got", Field0)
			}

			if len(Field0) != 0 {
				t.Error("want empty got", Field0)
			}
		})

		t.Run("teste2 result is not null", func(t *testing.T) {
			value := []int64([]int64{123, 456})
			id, err := itest.InsertE2(&tests.TestE2{Field0: &value})
			if err != nil {
				t.Error(err)
				return
			}

			var Field0 []int64
			err = factory.DB().QueryRowContext(context.Background(), abyid, id).Scan(makeScanner(&Field0))
			if err != nil {
				t.Error(err)
				return
			}

			if !reflect.DeepEqual(Field0, []int64{123, 456}) {
				t.Error("want nil got", Field0)
			}
		})

		t.Run("teste3 result is null", func(t *testing.T) {
			id, err := itest.InsertE3(&tests.TestE3{})
			if err != nil {
				t.Error(err)
				return
			}

			var Field0 []int64
			err = factory.DB().QueryRowContext(context.Background(), abyid, id).Scan(makeScanner(&Field0))
			if err != nil {
				t.Error(err)
				return
			}

			if Field0 != nil {
				t.Error("want nil got", Field0)
			}
		})

		t.Run("teste3 result is empty", func(t *testing.T) {
			value := []int64{}
			id, err := itest.InsertE3(&tests.TestE3{Field0: &value})
			if err != nil {
				t.Error(err)
				return
			}

			var Field0 []int64
			err = factory.DB().QueryRowContext(context.Background(), abyid, id).Scan(makeScanner(&Field0))
			if err != nil {
				t.Error(err)
				return
			}

			if Field0 != nil {
				t.Error("want nil got", Field0)
			}
		})

		t.Run("teste3 result is not null", func(t *testing.T) {
			value := []int64([]int64{123, 456})
			id, err := itest.InsertE3(&tests.TestE3{Field0: &value})
			if err != nil {
				t.Error(err)
				return
			}

			var Field0 []int64
			err = factory.DB().QueryRowContext(context.Background(), abyid, id).Scan(makeScanner(&Field0))
			if err != nil {
				t.Error(err)
				return
			}

			if !reflect.DeepEqual(Field0, []int64{123, 456}) {
				t.Error("want nil got", Field0)
			}
		})

		t.Run("teste4 result is null", func(t *testing.T) {
			id, err := itest.InsertE4(&tests.TestE4{})
			if err != nil {
				t.Error(err)
				return
			}

			var Field0 []int64
			err = factory.DB().QueryRowContext(context.Background(), abyid, id).Scan(makeScanner(&Field0))
			if err != nil {
				t.Error(err)
				return
			}

			if Field0 != nil {
				t.Error("want nil got", Field0)
			}
		})

		t.Run("teste4 result is empty", func(t *testing.T) {
			value := []int64{}
			id, err := itest.InsertE4(&tests.TestE4{Field0: value})
			if err != nil {
				t.Error(err)
				return
			}

			var Field0 []int64
			err = factory.DB().QueryRowContext(context.Background(), abyid, id).Scan(makeScanner(&Field0))
			if err != nil {
				t.Error(err)
				return
			}

			if Field0 != nil {
				t.Error("want nil got", Field0)
			}
		})

		t.Run("teste4 result is not null", func(t *testing.T) {
			value := []int64([]int64{123, 456})
			id, err := itest.InsertE4(&tests.TestE4{Field0: value})
			if err != nil {
				t.Error(err)
				return
			}

			var Field0 []int64
			err = factory.DB().QueryRowContext(context.Background(), abyid, id).Scan(makeScanner(&Field0))
			if err != nil {
				t.Error(err)
				return
			}

			if !reflect.DeepEqual(Field0, []int64{123, 456}) {
				t.Error("want nil got", Field0)
			}
		})

		abyid = `select field0 from gobatis_teste2 where id = $1`
		if factory.Dialect() != dialects.Postgres {
			abyid = `select field0 from gobatis_teste2 where id = ?`
		}

		t.Run("teste5 result is null", func(t *testing.T) {
			_, err := itest.InsertE5(&tests.TestE5{})
			if err == nil {
				t.Error("want err got ok")
			} else if !strings.Contains(err.Error(), "field 'Field0") {
				t.Error("want contains 'Field0' got", err)
			}
		})

		t.Run("teste5 result is empty", func(t *testing.T) {
			value := []int64{}
			_, err := itest.InsertE5(&tests.TestE5{Field0: &value})
			if err == nil {
				t.Error("want err got ok")
			} else if !strings.Contains(err.Error(), "field 'Field0") {
				t.Error("want contains 'Field0' got", err)
			}
		})

		t.Run("teste5 result is not null", func(t *testing.T) {
			value := []int64{123, 456}
			id, err := itest.InsertE5(&tests.TestE5{Field0: &value})
			if err != nil {
				t.Error(err)
				return
			}

			var Field0 []int64
			err = factory.DB().QueryRowContext(context.Background(), abyid, id).Scan(makeScanner(&Field0))
			if err != nil {
				t.Error(err)
				return
			}

			if !reflect.DeepEqual(Field0, []int64{123, 456}) {
				t.Error("want nil got", Field0)
			}
		})

		t.Run("teste6 result is null", func(t *testing.T) {
			_, err := itest.InsertE6(&tests.TestE6{})
			if err == nil {
				t.Error("want err got ok")
			} else if !strings.Contains(err.Error(), "field 'Field0") {
				t.Error("want contains 'Field0' got", err)
			}
		})

		t.Run("teste6 result is empty", func(t *testing.T) {
			value := []int64{}
			_, err := itest.InsertE6(&tests.TestE6{Field0: value})
			if err == nil {
				t.Error("want err got ok")
			} else if !strings.Contains(err.Error(), "field 'Field0") {
				t.Error("want contains 'Field0' got", err)
			}
		})

		t.Run("teste6 result is not null", func(t *testing.T) {
			value := []int64([]int64{123, 456})
			id, err := itest.InsertE6(&tests.TestE6{Field0: value})
			if err != nil {
				t.Error(err)
				return
			}

			var Field0 []int64
			err = factory.DB().QueryRowContext(context.Background(), abyid, id).Scan(makeScanner(&Field0))
			if err != nil {
				t.Error(err)
				return
			}

			if !reflect.DeepEqual(Field0, []int64{123, 456}) {
				t.Error("want nil got", Field0)
			}
		})
	})
}

func TestMapperF(t *testing.T) {
	tests.Run(t, func(_ testing.TB, factory *core.SessionFactory) {
		ref := factory.SessionReference()
		itest := tests.NewITest(ref)

		tablename := "gobatis_testf1"
		abyid := `select field0 from ` + tablename + ` where id = $1`
		if factory.Dialect() != dialects.Postgres {
			abyid = `select field0 from ` + tablename + ` where id = ?`
		}

		t.Run("testf1 result is null", func(t *testing.T) {
			id, err := itest.InsertF1(&tests.TestF1{})
			if err != nil {
				t.Error(err)
				return
			}

			var Field0 []byte
			err = factory.DB().QueryRowContext(context.Background(), abyid, id).Scan(&Field0)
			if err != nil {
				t.Error(err)
				return
			}
			// if Field0 == nil {
			// 	t.Error("want nil got", Field0)
			// }
			if len(Field0) != 0 {
				t.Error("want empty got", Field0)
			}
		})

		t.Run("testf1 result is empty", func(t *testing.T) {
			id, err := itest.InsertF1(&tests.TestF1{Field0: []byte{}})
			if err != nil {
				t.Error(err)
				return
			}

			var Field0 []byte
			err = factory.DB().QueryRowContext(context.Background(), abyid, id).Scan(&Field0)
			if err != nil {
				t.Error(err)
				return
			}

			if Field0 == nil {
				t.Error("want not nil got", Field0)
			}

			if len(Field0) != 0 {
				t.Error("want empty got", Field0)
			}
		})

		t.Run("testf1 result is not null", func(t *testing.T) {
			id, err := itest.InsertF1(&tests.TestF1{Field0: []byte(`{"a":"b"}`)})
			if err != nil {
				t.Error(err)
				return
			}

			var Field0 []byte
			err = factory.DB().QueryRowContext(context.Background(), abyid, id).Scan(&Field0)
			if err != nil {
				t.Error(err)
				return
			}

			if string(Field0) != `{"a":"b"}` {
				t.Error("want nil got", string(Field0))
			}
		})

		t.Run("testf2 result is null", func(t *testing.T) {
			id, err := itest.InsertF2(&tests.TestF2{})
			if err != nil {
				t.Error(err)
				return
			}

			var Field0 []byte
			err = factory.DB().QueryRowContext(context.Background(), abyid, id).Scan(&Field0)
			if err != nil {
				t.Error(err)
				return
			}

			if Field0 != nil {
				t.Error("want nil got", Field0)
			}
		})

		t.Run("testf2 result is not null", func(t *testing.T) {
			value := []byte(`{"a":"b"}`)
			id, err := itest.InsertF2(&tests.TestF2{Field0: &value})
			if err != nil {
				t.Error(err)
				return
			}

			var Field0 []byte
			err = factory.DB().QueryRowContext(context.Background(), abyid, id).Scan(&Field0)
			if err != nil {
				t.Error(err)
				return
			}

			if string(Field0) != `{"a":"b"}` {
				t.Error("want nil got", string(Field0))
			}
		})

		t.Run("testf3 result is null", func(t *testing.T) {
			id, err := itest.InsertF3(&tests.TestF3{})
			if err != nil {
				t.Error(err)
				return
			}

			var Field0 []byte
			err = factory.DB().QueryRowContext(context.Background(), abyid, id).Scan(&Field0)
			if err != nil {
				t.Error(err)
				return
			}

			if Field0 != nil {
				t.Error("want nil got", Field0)
			}
		})

		t.Run("testf3 result is empty", func(t *testing.T) {
			value := []byte{}
			id, err := itest.InsertF3(&tests.TestF3{Field0: &value})
			if err != nil {
				t.Error(err)
				return
			}

			var Field0 []byte
			err = factory.DB().QueryRowContext(context.Background(), abyid, id).Scan(&Field0)
			if err != nil {
				t.Error(err)
				return
			}

			if Field0 != nil {
				t.Error("want nil got", Field0)
			}
		})

		t.Run("testf3 result is not null", func(t *testing.T) {
			value := []byte(`{"a":"b"}`)
			id, err := itest.InsertF3(&tests.TestF3{Field0: &value})
			if err != nil {
				t.Error(err)
				return
			}

			var Field0 []byte
			err = factory.DB().QueryRowContext(context.Background(), abyid, id).Scan(&Field0)
			if err != nil {
				t.Error(err)
				return
			}

			if string(Field0) != `{"a":"b"}` {
				t.Error("want nil got", string(Field0))
			}
		})

		t.Run("testf4 result is null", func(t *testing.T) {
			id, err := itest.InsertF4(&tests.TestF4{})
			if err != nil {
				t.Error(err)
				return
			}

			var Field0 []byte
			err = factory.DB().QueryRowContext(context.Background(), abyid, id).Scan(&Field0)
			if err != nil {
				t.Error(err)
				return
			}

			if Field0 != nil {
				t.Error("want nil got", Field0)
			}
		})

		t.Run("testf4 result is empty", func(t *testing.T) {
			value := []byte{}
			id, err := itest.InsertF4(&tests.TestF4{Field0: value})
			if err != nil {
				t.Error(err)
				return
			}

			var Field0 []byte
			err = factory.DB().QueryRowContext(context.Background(), abyid, id).Scan(&Field0)
			if err != nil {
				t.Error(err)
				return
			}

			if Field0 != nil {
				t.Error("want nil got", Field0)
			}
		})

		t.Run("testf4 result is not null", func(t *testing.T) {
			value := []byte(`{"a":"b"}`)
			id, err := itest.InsertF4(&tests.TestF4{Field0: value})
			if err != nil {
				t.Error(err)
				return
			}

			var Field0 []byte
			err = factory.DB().QueryRowContext(context.Background(), abyid, id).Scan(&Field0)
			if err != nil {
				t.Error(err)
				return
			}

			if string(Field0) != `{"a":"b"}` {
				t.Error("want nil got", string(Field0))
			}
		})

		abyid = `select field0 from gobatis_testf2 where id = $1`
		if factory.Dialect() != dialects.Postgres {
			abyid = `select field0 from gobatis_testf2 where id = ?`
		}

		t.Run("testf5 result is null", func(t *testing.T) {
			_, err := itest.InsertF5(&tests.TestF5{})
			if err == nil {
				t.Error("want err got ok")
			} else if !strings.Contains(err.Error(), "field 'Field0") {
				t.Error("want contains 'Field0' got", err)
			}
		})

		t.Run("testf5 result is empty", func(t *testing.T) {
			value := []byte{}
			_, err := itest.InsertF5(&tests.TestF5{Field0: &value})
			if err == nil {
				t.Error("want err got ok")
			} else if !strings.Contains(err.Error(), "field 'Field0") {
				t.Error("want contains 'Field0' got", err)
			}
		})

		t.Run("testf5 result is not null", func(t *testing.T) {
			value := []byte(`{"a":"b"}`)
			id, err := itest.InsertF5(&tests.TestF5{Field0: &value})
			if err != nil {
				t.Error(err)
				return
			}

			var Field0 []byte
			err = factory.DB().QueryRowContext(context.Background(), abyid, id).Scan(&Field0)
			if err != nil {
				t.Error(err)
				return
			}

			if string(Field0) != `{"a":"b"}` {
				t.Error("want nil got", string(Field0))
			}
		})

		t.Run("testf6 result is null", func(t *testing.T) {
			_, err := itest.InsertF6(&tests.TestF6{})
			if err == nil {
				t.Error("want err got ok")
			} else if !strings.Contains(err.Error(), "field 'Field0") {
				t.Error("want contains 'Field0' got", err)
			}
		})

		t.Run("testf6 result is empty", func(t *testing.T) {
			value := []byte{}
			_, err := itest.InsertF6(&tests.TestF6{Field0: value})
			if err == nil {
				t.Error("want err got ok")
			} else if !strings.Contains(err.Error(), "field 'Field0") {
				t.Error("want contains 'Field0' got", err)
			}
		})

		t.Run("testf6 result is not null", func(t *testing.T) {
			value := []byte(`{"a":"b"}`)
			id, err := itest.InsertF6(&tests.TestF6{Field0: value})
			if err != nil {
				t.Error(err)
				return
			}

			var Field0 []byte
			err = factory.DB().QueryRowContext(context.Background(), abyid, id).Scan(&Field0)
			if err != nil {
				t.Error(err)
				return
			}

			if string(Field0) != `{"a":"b"}` {
				t.Error("want nil got", string(Field0))
			}
		})
	})
}

func TestTagSplitForXORM(t *testing.T) {

	for _, test := range []struct {
		s         string
		fieldName string
		excepted  []string
	}{
		{
			s:         "a pk null",
			fieldName: "a",
			excepted:  []string{"a", "pk", "null"},
		},
		{
			s:         "pk null",
			fieldName: "a",
			excepted:  []string{"a", "pk", "null"},
		},
		{
			s:         "pk(aa) null",
			fieldName: "a",
			excepted:  []string{"a", "pk(aa)", "null"},
		},
		{
			s:         "null pk(aa)",
			fieldName: "a",
			excepted:  []string{"a", "null", "pk(aa)"},
		},
	} {

		ss := core.TagSplitForXORM(test.s, test.fieldName)
		if !reflect.DeepEqual(ss, test.excepted) {
			t.Error("excepted:", test.excepted)
			t.Error("actual  :", ss)
		}
	}
}
