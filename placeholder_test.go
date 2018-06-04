// This file is copy from https://github.com/Masterminds/squirrel
package gobatis

import (
	"strings"
	"testing"
)

func TestQuestion(t *testing.T) {
	sql := "x = ? AND y = ?"
	s, _ := Question.ReplacePlaceholders(sql)
	if sql != s {
		t.Error("excepted is", sql)
		t.Error("actual   is", s)
	}
}

func TestDollar(t *testing.T) {
	sql := "x = ? AND y = ?"
	s, _ := Dollar.ReplacePlaceholders(sql)

	if excepted := "x = $1 AND y = $2"; excepted != s {
		t.Error("excepted is", excepted)
		t.Error("actual   is", s)
	}
}

func TestPlaceholders(t *testing.T) {
	s := Placeholders(2)
	if excepted := "?,?"; excepted != s {
		t.Error("excepted is", excepted)
		t.Error("actual   is", s)
	}
}

func TestEscape(t *testing.T) {
	sql := "SELECT uuid, \"data\" #> '{tags}' AS tags FROM nodes WHERE  \"data\" -> 'tags' ??| array['?'] AND enabled = ?"
	s, _ := Dollar.ReplacePlaceholders(sql)
	excepted := "SELECT uuid, \"data\" #> '{tags}' AS tags FROM nodes WHERE  \"data\" -> 'tags' ?| array['$1'] AND enabled = $2"
	if excepted != s {
		t.Error("excepted is", excepted)
		t.Error("actual   is", s)
	}
}

func BenchmarkPlaceholdersArray(b *testing.B) {
	var count = b.N
	placeholders := make([]string, count)
	for i := 0; i < count; i++ {
		placeholders[i] = "?"
	}
	var _ = strings.Join(placeholders, ",")
}

func BenchmarkPlaceholdersStrings(b *testing.B) {
	Placeholders(b.N)
}
