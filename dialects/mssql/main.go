package mssql

import (
	"errors"
	"strings"

	"github.com/microsoft/go-mssqldb"
	"github.com/runner-mei/GoBatis/dialects"
)

func init() {
	dialects.SetHandleError(dialects.DriverMSSql.Name(), handleError)
}

func handleError(e error) error {
	if e == nil {
		return nil
	}

	pe, ok := asMssqlError(e)
	if !ok {
		return e
	}

	switch pe.Number {
	// 2627: Violation of PRIMARY KEY/UNIQUE KEY constraint 'xxx'
	// 2601: Cannot insert duplicate key row in object 'xxx' with unique index 'yyy'
	case 2627, 2601:
		return &dialects.Error{Validations: []dialects.ValidationError{
			{Code: "unique_value_already_exists", Message: pe.Message, Columns: parseConstraintName(pe.Message)},
		}, Err: e}

	// 208: Invalid object name 'xxx'.
	case 208:
		return dialects.ErrTableNotExists{
			Err:       e,
			Tablename: parseObjectName(pe.Message),
		}
	}
	return e
}

// asMssqlError mssql 的错误可能是 Error 本身，也可能是被 ServerError 包装之后的。
func asMssqlError(e error) (mssql.Error, bool) {
	var value mssql.Error
	if errors.As(e, &value) {
		return value, true
	}

	var ptr *mssql.Error
	if errors.As(e, &ptr) && ptr != nil {
		return *ptr, true
	}
	return mssql.Error{}, false
}

// parseConstraintName 从 "Violation of PRIMARY KEY constraint 'PK_x'." 或
// "with unique index 'UQ_x'." 中提取主键/唯一约束的名称。
func parseConstraintName(message string) []string {
	for _, keyword := range []string{"constraint", "约束", "unique index", "index", "唯一索引", "索引"} {
		if name := parseQuotedAfter(message, keyword); name != "" {
			return []string{name}
		}
	}
	return nil
}

// parseObjectName 从 "Invalid object name 'dbo.xxx'." 中提取表名。
func parseObjectName(message string) string {
	return parseQuotedAfter(message, "")
}

// parseQuotedAfter 返回 keyword 之后第一个引号（'、"、`）中的内容。
func parseQuotedAfter(message, keyword string) string {
	idx := strings.Index(strings.ToLower(message), strings.ToLower(keyword))
	if idx < 0 {
		return ""
	}

	rest := message[idx+len(keyword):]
	start := strings.IndexAny(rest, "'\"`")
	if start < 0 {
		return ""
	}

	quote := rest[start]
	if end := strings.IndexByte(rest[start+1:], quote); end >= 0 {
		return rest[start+1 : start+1+end]
	}
	return ""
}
