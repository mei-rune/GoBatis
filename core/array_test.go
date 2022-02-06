package core_test

import (
	"reflect"
	"testing"

	"github.com/runner-mei/GoBatis/core"
	"github.com/runner-mei/GoBatis/tests"
)

func TestArray(t *testing.T) {
	tests.Run(t, func(_ testing.TB, factory *core.Session) {

		itest := tests.NewITest(factory.SessionReference())

		// InsertTestE(v *TestE) (int64, error)
		// GetTestE(id int64) (*TestE, error)
		// UpdateTestE(id int64, v *TestE) (int64, error)

		origin := &tests.TestE{Field0: []int64{123, 334}}
		id, err := itest.InsertTestE(origin)
		if err != nil {
			t.Error(err)
			return
		}

		value, err := itest.GetTestE(id)
		if err != nil {
			t.Error(err)
			return
		}

		if !reflect.DeepEqual(origin.Field0, value.Field0) {
			t.Error("want", origin.Field0, "got", value.Field0)
			return
		}
		origin.Field0[0] = 2

		_, err = itest.UpdateTestE(id, origin)
		if err != nil {
			t.Error(err)
			return
		}

		value, err = itest.GetTestE(id)
		if err != nil {
			t.Error(err)
			return
		}

		if !reflect.DeepEqual(origin.Field0, value.Field0) {
			t.Error("want", origin.Field0, "got", value.Field0)
			return
		}

		originInts := []int64{22, 33}
		id, err = itest.InsertTestE_2(originInts)
		if err != nil {
			t.Error(err)
			return
		}

		value, err = itest.GetTestE(id)
		if err != nil {
			t.Error(err)
			return
		}

		if !reflect.DeepEqual(originInts, value.Field0) {
			t.Error("want", origin.Field0, "got", value.Field0)
			return
		}
	})
}
