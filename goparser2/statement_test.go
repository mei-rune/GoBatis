package goparser2

import (
	"testing"
)

func TestStatement(t *testing.T) {
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
}
