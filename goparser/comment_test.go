package goparser

import (
	"reflect"
	"testing"
)

func TestParseComments(t *testing.T) {
	for _, test := range []struct {
		txt string
		cfg *SQLConfig
	}{
		{
			txt: `// assss
				  //    abc
				  //
				  //  @type select
				  //  @option k1 v1
				  //  @option k2 v2
				  //  @mysql select * from a
				  //  @postgres select 1
				  //  @default select * from abc
			`,
			cfg: &SQLConfig{
				Description:   "assss\r\n    abc",
				StatementType: "select",
				DefaultSQL:    "select * from abc",
				Options:       map[string]string{"k1": "v1", "k2": "v2"},
				Dialects: map[string]string{"mysql": "select * from a",
					"postgres": "select 1",
				},
			},
		},
	} {
		coments := splitLines(test.txt)
		actual, err := parseComments(coments)
		if err != nil {
			t.Error(err)
			continue
		}

		if !reflect.DeepEqual(actual, test.cfg) {
			if actual.Description != test.cfg.Description {
				t.Error("[Description] actual is", actual.Description)
				t.Error("[Description] excepted is", test.cfg.Description)
			}
			if actual.StatementType != test.cfg.StatementType {
				t.Error("[StatementType] actual is", actual.StatementType)
				t.Error("[StatementType] excepted is", test.cfg.StatementType)
			}
			if actual.DefaultSQL != test.cfg.DefaultSQL {
				t.Error("[DefaultSQL] actual is", actual.DefaultSQL)
				t.Error("[DefaultSQL] excepted is", test.cfg.DefaultSQL)
			}
			if !reflect.DeepEqual(actual.Options, test.cfg.Options) {
				t.Error("[Options] actual is", actual.Options)
				t.Error("[Options] excepted is", test.cfg.Options)
			}
			if !reflect.DeepEqual(actual.Dialects, test.cfg.Dialects) {
				t.Error("[Dialects] actual is", actual.Dialects)
				t.Error("[Dialects] excepted is", test.cfg.Dialects)
			}
			t.Error("actual is", *actual)
			t.Error("excepted is", *test.cfg)
		}
	}
}
