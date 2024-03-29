package goparser2

import (
	"testing"

	gobatis "github.com/runner-mei/GoBatis"
)

func TestParse2Statement(t *testing.T) {
	for _, test := range []struct {
		name     string
		excepted bool
	}{
		{"insert", true},
		{"upsert", true},
		{"inserta", true},
		{"upsertb", true},
		{"a", false},
	} {
		if test.excepted != isInsertStatement(test.name) {
			t.Error(test.name, "is error")
		}
	}

	for _, test := range []struct {
		name     string
		excepted bool
	}{
		{"update", true},
		{"updatea", true},
		{"a", false},
	} {
		if test.excepted != isUpdateStatement(test.name) {
			t.Error(test.name, "is error")
		}
	}

	for _, test := range []struct {
		name     string
		excepted bool
	}{
		{"delete", true},
		{"remove", true},
		{"deletea", true},
		{"removeb", true},
		{"a", false},
	} {
		if test.excepted != isDeleteStatement(test.name) {
			t.Error(test.name, "is error")
		}
	}

	for _, test := range []struct {
		name     string
		excepted bool
	}{
		{"select", true},
		{"find", true},
		{"get", true},
		{"query", true},
		{"id", true},
		{"a", false},
	} {
		if test.excepted != isSelectStatement(test.name) {
			t.Error(test.name, "is error")
		}
	}

	for _, test := range []struct {
		name     string
		excepted gobatis.StatementType
	}{
		{"DeleteAsset", gobatis.StatementTypeDelete}, // 这个含有 set 可能会被识别为 update
	} {
		if test.excepted != GetStatementType(test.name) {
			t.Error(test.name, "is error")
		}
	}
}
