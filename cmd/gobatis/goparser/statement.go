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
		"create",
	}, []string{
		"insert",
		"upsert",
		"add",
		"create",
	}, nil)
}

func isUpdateStatement(name string) bool {
	return isExceptedStatement(name, []string{
		"set",
		"update",
		"write",
	}, []string{
		"set",
		"update",
		"write",
	}, nil)
}
func isDeleteStatement(name string) bool {
	return isExceptedStatement(name, []string{
		"delete",
		"remove",
		"clear",
		"unset",
	}, []string{
		"delete",
		"remove",
		"clear",
	}, nil)
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
		"load",
		"statby",
		"statsby",
		"foreach",
		"exist",
	}, []string{
		"select",
		"find",
		"get",
		"query",
		"list",
		"count",
		"read",
		"load",
		"stat",
		"stats", 
		"count", 
		"foreach",
		"exist",
		"exists",
	}, []string{"id", "all", "names", "titles"})
}
