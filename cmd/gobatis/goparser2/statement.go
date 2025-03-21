package goparser2

import (
	"strings"

	gobatis "github.com/runner-mei/GoBatis"
)

func isExceptedStatement(name string, prefixs, suffixs, fullnames []string) bool {
	lowerName := strings.ToLower(name)
	for _, prefix := range prefixs {
		if strings.HasPrefix(lowerName, prefix) {
			return true
		}
		if strings.HasPrefix(name, prefix) {
			return true
		}
	}

	for _, suffix := range suffixs {
		if strings.HasSuffix(lowerName, suffix) {
			return true
		}
		if strings.HasSuffix(name, suffix) {
			return true
		}
	}
	for _, fullname := range fullnames {
		if lowerName == fullname {
			return true
		}
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
	}, []string{
		"insert",
		"upsert",
		"add",
		"create",
	})
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
	}, []string{
		"set",
		"update",
		"write",
	})
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
	}, []string{
		"delete",
		"remove",
		"clear",
	})
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
		"Has",
		"search",
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
		"search",
	}, []string{"id", "all", "names", "titles"})
}

func GetStatementType(name string) gobatis.StatementType {
	if isInsertStatement(name) {
		return gobatis.StatementTypeInsert
	}
	if isDeleteStatement(name) {
		return gobatis.StatementTypeDelete
	}
	if isUpdateStatement(name) {
		return gobatis.StatementTypeUpdate
	}
	if isSelectStatement(name) {
		return gobatis.StatementTypeSelect
	}
	return gobatis.StatementTypeNone
}
