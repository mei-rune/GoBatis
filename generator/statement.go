package generator

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
	}, []string{})
}

func isUpdateStatement(name string) bool {
	return isExceptedStatement(name, []string{
		"update",
	}, []string{})
}
func isDeleteStatement(name string) bool {
	return isExceptedStatement(name, []string{
		"delete",
		"remove",
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
	}, []string{"id"})
}
