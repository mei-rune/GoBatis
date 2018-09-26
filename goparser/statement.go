package goparser

import (
	"strings"
)

func isExceptedStatement(name string, fragments, fullnames []string) bool {
	name = strings.ToLower(name)
	for _, fragment := range fragments {
		if strings.Contains(name, fragment) {
			return true
		}
	}
	for _, fullname := range fullnames {
		if name == fullname {
			return true
		}
	}
	return false
}

func isInsertStatement(name string) bool {
	return isExceptedStatement(name, []string{
		"insert",
		"upsert",
		"add",
	}, []string{})
}

func isUpdateStatement(name string) bool {
	return isExceptedStatement(name, []string{
		"update",
		"set",
	}, []string{})
}
func isDeleteStatement(name string) bool {
	return isExceptedStatement(name, []string{
		"delete",
		"remove",
		"clear",
	}, []string{})
}
func isSelectStatement(name string) bool {
	return isExceptedStatement(name, []string{
		"select",
		"find",
		"get",
		"query",
		"list",
		"count",
		"read",
	}, []string{"id"})
}
