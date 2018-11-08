package goparser

import (
	"strings"
)

func isExceptedStatement(name string, prefixs, suffixs, fullnames []string) bool {
	name = strings.ToLower(name)
	for _, prefix := range prefixs {
		if strings.HasPrefix(name, prefix) {
			return true
		}
	}

	for _, suffix := range suffixs {
		if strings.HasSuffix(name, suffix) {
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
	}, nil, nil)
}

func isUpdateStatement(name string) bool {
	return isExceptedStatement(name, []string{
		"set",
		"update",
	}, nil, nil)
}
func isDeleteStatement(name string) bool {
	return isExceptedStatement(name, []string{
		"delete",
		"remove",
		"clear",
	}, nil, nil)
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
	}, []string{"count"}, []string{"id", "all", "names", "titles"})
}
