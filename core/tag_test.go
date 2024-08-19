package core

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

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
			excepted:  []string{"a", "pk=aa", "null"},
		},
		{
			s:         "null pk(aa)",
			fieldName: "a",
			excepted:  []string{"a", "null", "pk=aa"},
		},
	} {

		ss := TagSplitForXORM(test.s, test.fieldName)
		if !reflect.DeepEqual(ss, test.excepted) {
			t.Error("excepted:", test.excepted)
			t.Error("actual  :", ss)
		}
	}
}

func TestTagSplitXormTag(t *testing.T) {
	cases := []struct {
		tag  string
		tags []tag
	}{
		{
			"not null default '2000-01-01 00:00:00' TIMESTAMP", []tag{
				{
					name: "not",
				},
				{
					name: "null",
				},
				{
					name: "default",
				},
				{
					name: "'2000-01-01 00:00:00'",
				},
				{
					name: "TIMESTAMP",
				},
			},
		},
		{
			"TEXT", []tag{
				{
					name: "TEXT",
				},
			},
		},
		{
			"default('2000-01-01 00:00:00')", []tag{
				{
					name: "default",
					params: []string{
						"'2000-01-01 00:00:00'",
					},
				},
			},
		},
		{
			"json  binary", []tag{
				{
					name: "json",
				},
				{
					name: "binary",
				},
			},
		},
		{
			"numeric(10, 2)", []tag{
				{
					name:   "numeric",
					params: []string{"10", "2"},
				},
			},
		},
		{
			"numeric(10, 2) notnull", []tag{
				{
					name:   "numeric",
					params: []string{"10", "2"},
				},
				{
					name: "notnull",
				},
			},
		},
		{
			"collate utf8mb4_bin", []tag{
				{
					name: "collate",
				},
				{
					name: "utf8mb4_bin",
				},
			},
		},
	}

	for _, kase := range cases {
		t.Run(kase.tag, func(t *testing.T) {
			tags, err := splitXormTag(kase.tag)
			assert.NoError(t, err)
			assert.EqualValues(t, len(tags), len(kase.tags))
			for i := 0; i < len(tags); i++ {
				assert.Equal(t, kase.tags[i], tags[i])
			}
		})
	}
}
