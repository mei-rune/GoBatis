package gobatis

import (
	"reflect"
	"strings"
	"testing"
)

type FinderInnerTest struct {
	A int `db:"a"`
}

type FinderTest struct {
	F1 string           `db:"f1"`
	F2 *FinderInnerTest `db:"f2"`
}

func TestFinder(t *testing.T) {
	constants := map[string]interface{}{}
	var mapper = CreateMapper("", nil, nil)

	t.Run("1", func(t *testing.T) {
		paramNames := []string{
			"a",
			"b",
			"c",
			"d",
		}
		paramValues := []interface{}{
			2,
			&FinderTest{
				F1: "a",
			},
			nil,
			map[string]interface{}{
				"f1": "f1value",
			},
		}

		ctx, err := NewContext(constants, DbTypePostgres, mapper, paramNames, paramValues)
		if err != nil {
			t.Error(err)
			return
		}

		for _, test := range []struct {
			name  string
			value interface{}
			err   string
		}{
			{name: "a", value: 2},
			{name: "b.f1", value: "a"},
			// {name: "b.f2.a", value: "a"},
			{name: "c.a", err: "not found"},
			{name: "d.f1", value: "f1value"},
			{name: "d.f_not_exists", value: nil},
		} {
			value, err := ctx.Get(test.name)
			if err != nil {
				if test.err == "" {
					t.Error(err)
					continue
				}
				if !strings.Contains(err.Error(), test.err) {
					t.Error(err)
					continue
				}
				continue
			}

			if !reflect.DeepEqual(value, test.value) {
				t.Errorf("want %T %v", test.value, test.value)
				t.Errorf("got  %T %v", value, value)
			}
		}
	})

	t.Run("2", func(t *testing.T) {
		paramNames := []string{
			"a",
		}
		paramValues := []interface{}{
			2,
		}

		ctx, err := NewContext(constants, DbTypePostgres, mapper, paramNames, paramValues)
		if err != nil {
			t.Error(err)
			return
		}

		for _, test := range []struct {
			name  string
			value interface{}
			err   string
		}{
			{name: "a", value: 2},
			{name: "c.a", err: "not found"},
		} {
			value, err := ctx.Get(test.name)
			if err != nil {
				if test.err == "" {
					t.Error(err)
					continue
				}
				if !strings.Contains(err.Error(), test.err) {
					t.Error(err)
					continue
				}
				continue
			}

			if !reflect.DeepEqual(value, test.value) {
				t.Errorf("want %T %v", test.value, test.value)
				t.Errorf("got  %T %v", value, value)
			}
		}
	})
}
