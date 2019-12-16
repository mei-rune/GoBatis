package gobatis_test

import (
	"context"
	"testing"

	gobatis "github.com/runner-mei/GoBatis"
	"github.com/runner-mei/GoBatis/tests"
)

func TestResultsClose(t *testing.T) {
	var res gobatis.Results

	// if res.Next() {
	// 	t.Error("except error go ok")
	// }

	// var c int
	// if err := res.Scan(&c); err == nil {
	// 	t.Error("except error go ok")
	// }

	if err := res.Close(); err != nil {
		t.Error(err)
	}
}

func TestResults(t *testing.T) {
	tests.Run(t, func(_ testing.TB, factory *gobatis.SessionFactory) {
		var u tests.User
		res := factory.SelectOne(context.Background(), "selectError", u)
		err := res.Scan(&u)
		if err == nil {
			t.Error("excepted error get ok")
			return
		}

		results := factory.Select(context.Background(), "selectError", u)
		if results.Next() {
			t.Error("except error go ok")
		}

		var array []map[string]interface{}
		if err := results.ScanResults(&array); err == nil {
			t.Error("except error go ok")
		}
		if err == nil {
			t.Error("excepted error get ok")
			return
		}

		if err := results.Close(); err != nil {
			t.Error(err)
		}
	})
}
