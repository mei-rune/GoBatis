package gobatis_test

import (
	"context"
	"database/sql"
	"strings"
	"testing"
	"time"

	gobatis "github.com/runner-mei/GoBatis"
	"github.com/runner-mei/GoBatis/tests"
)

func TestMapper(t *testing.T) {
	tests.Run(t, func(_ testing.TB, factory *gobatis.SessionFactory) {
		ref := factory.Reference()
		itest := tests.NewITest(&ref)

		abyid := `select field0,field1,field2,field3,field4,field5,field6,field7,field8 from gobatis_testa where id = $1`
		bbyid := `select field0,field1,field2,field3,field4,field5,field6,field7,field8 from gobatis_testb where id = $1`

		if factory.Dialect() != gobatis.DbTypePostgres {
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
			id, err := itest.InsertA3(&tests.TestA3{})
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
				t.Error("want nil got", Field6)
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
				t.Error("want nil got", Field6)
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
	tests.Run(t, func(_ testing.TB, factory *gobatis.SessionFactory) {
		ref := factory.Reference()
		itest := tests.NewITest(&ref)

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
	tests.Run(t, func(_ testing.TB, factory *gobatis.SessionFactory) {
		ref := factory.Reference()
		itest := tests.NewITest(&ref)

		abyid := `select field0 from gobatis_testc where id = $1`

		if factory.Dialect() != gobatis.DbTypePostgres {
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

			if !Field0.Valid || Field0.String != `{"a": "b"}` {
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

			if !Field0.Valid || Field0.String != `{"a": "b"}` {
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

			if !Field0.Valid || Field0.String != `{"a": "b"}` {
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

			if !Field0.Valid || Field0.String != `{"A": 1}` {
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

			if !Field0.Valid || Field0.String != `{"A": 1}` {
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

			if !Field0.Valid || Field0.String != `{"A": 1}` {
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

			if !Field0.Valid || Field0.String != `{"A": 0}` {
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

			if !Field0.Valid || Field0.String != `{"A": 1}` {
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

			if !Field0.Valid || Field0.String != `{"A": 1}` {
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

			if !Field0.Valid || Field0.String != `{"A": 0}` {
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

			if !Field0.Valid || Field0.String != `{"A": 1}` {
				t.Error("want nil got", Field0.String)
			}
		})

	})
}
