package goparser

import (
	"reflect"
	"strings"
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
				  //  @reference a.b
				  //  @type select
				  //  @option k1 v1
				  //  @option k2 v2
			`,
			cfg: &SQLConfig{
				Description: "assss\r\n    abc",
				Reference: &struct {
					Interface string
					Method    string
				}{Interface: "a",
					Method: "b"},
				StatementType: "select",
				Options:       map[string]string{"k1": "v1", "k2": "v2"},
			},
		},

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
			if reflect.DeepEqual(actual.Reference, test.cfg.Reference) {
				t.Error("[Reference] actual is", actual.Reference)
				t.Error("[Reference] excepted is", test.cfg.Reference)
			}

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

func TestParseCommentFail(t *testing.T) {
	for _, test := range []struct {
		txt string
		err string
	}{
		{
			txt: `// assss
				  //    abc
				  //
				  //  @reference a.b
				  //  @type select
				  //  @option k1 v1
				  //  @option k2 v2
				  //  @default select * from abc
			`,
			err: "sql statement is unnecessary",
		},
		{
			txt: `// assss
				  //    abc
				  //
				  //  @reference a
				  //  @type select
			`,
			err: "syntex error",
		},
	} {
		coments := splitLines(test.txt)
		_, err := parseComments(coments)
		if err == nil {
			t.Error("except error got ok")
			continue
		}
		if test.err != "" && !strings.Contains(err.Error(), test.err) {
			t.Error("excepted is", err)
			t.Error("actual   is", test.err)
		}
	}
}
