package st_oscar

import (
	"strings"

	_ "gitee.com/shentongdata/go-aci"
	"github.com/runner-mei/GoBatis/dialects"
)

func init() {
	dialects.SetHandleError(dialects.DriverShengtongOscar.Name(), handleError)
	// dialects.SetHandleArray(dialects.ShengtongOscar.Name(), makePQArrayValuer, makePQArrayScanner)
}

// 神通数据库(兼容 Oracle)唯一性冲突时返回的提示信息，如：
//   - 不能向索引GOBATIS_UNIQUE_TEST_PKEY中插入重复键值(ID) = (1)
//   - 违反唯一约束条件 (GOLANG.SYS_C0012345)
var uniqueViolationKeys = []string{
	"插入重复键值",
	"重复键值",
	"违反唯一约束",
	"违反唯一性约束",
	"唯一约束冲突",
	"唯一性约束冲突",
	"ORA-00001",
	"unique constraint",
	"duplicate key",
	"duplicate entry",
}

func handleError(e error) error {
	if e == nil {
		return nil
	}

	msg := e.Error()

	// "ERROR, 表或视图 \"GOBATIS_TEST_TABLE_NOT_EXISTS\" 不存在或无权访问\n"
	//     main_test.go:32: want TableNotExists
	//     main_test.go:33:  got ERROR, 表或视图 "GOBATIS_TEST_TABLE_NOT_EXISTS" 不存在或无权访问

	if strings.Contains(msg, "表或视图") &&
		strings.Contains(msg, "不存在或无权访问") {
		return dialects.ErrTableNotExists{
			Err: e,
			// Tablename: pe.TableName,
		}
	}

	if isUniqueViolation(msg) {
		return &dialects.Error{Validations: []dialects.ValidationError{
			{Code: "unique_value_already_exists", Message: msg, Columns: parseUniqueColumns(msg)},
		}, Err: e}
	}

	return e
}

func isUniqueViolation(message string) bool {
	lower := strings.ToLower(message)
	for _, key := range uniqueViolationKeys {
		if strings.Contains(lower, strings.ToLower(key)) {
			return true
		}
	}
	return false
}

// parseUniqueColumns 从唯一性冲突的提示信息中提取冲突的列名：
//   - "不能向索引XXX中插入重复键值(ID) = (1)"        => ["ID"]
//   - "违反唯一约束条件 (GOLANG.SYS_C0012345)"      => ["GOLANG.SYS_C0012345"]
func parseUniqueColumns(message string) []string {
	var columns string
	for _, keyword := range []string{"重复键值", "duplicate key value", "违反唯一"} {
		if columns = parseParenthesizedAfter(message, keyword); columns != "" {
			break
		}
	}
	if columns == "" {
		return nil
	}

	names := strings.Split(columns, ",")
	for idx := range names {
		names[idx] = strings.TrimSpace(names[idx])
	}
	return names
}

// parseParenthesizedAfter 返回 keyword 之后第一个括号中的内容。
func parseParenthesizedAfter(message, keyword string) string {
	idx := strings.Index(strings.ToLower(message), strings.ToLower(keyword))
	if idx < 0 {
		return ""
	}

	rest := message[idx+len(keyword):]
	start := strings.IndexByte(rest, '(')
	if start < 0 {
		return ""
	}

	end := strings.IndexByte(rest[start+1:], ')')
	if end < 0 {
		return ""
	}
	return strings.TrimSpace(rest[start+1 : start+1+end])
}

// func makePQArrayValuer(v interface{}) (interface{}, error) {
// 	switch a := v.(type) {
// 	case []bool:
// 		var iv = gaussdbtype.FlatArray[bool](a)
// 		return iv, nil
// 	case []float64:
// 		var iv = gaussdbtype.FlatArray[float64](a)
// 		return iv, nil
// 	case []int64:
// 		var iv = gaussdbtype.FlatArray[int64](a)
// 		return iv, nil
// 	case []string:
// 		var iv = gaussdbtype.FlatArray[string](a)
// 		return iv, nil
// 	default:
// 		return nil, errors.New("must is array, it isnot support - []bool, []float64, []int64 and []string")
// 	}
// }

// func makePQArrayScanner(name string, v interface{}) (interface{}, error) {
// 	switch a := v.(type) {
// 	case *[]bool:
// 		var iv = gaussdbtype.FlatArray[bool](*a)
// 		return &iv, nil
// 	case *[]float64:
// 		var iv = gaussdbtype.FlatArray[float64](*a)
// 		return &iv, nil
// 	case *[]int64:
// 		var iv = gaussdbtype.FlatArray[int64](*a)
// 		return &iv, nil
// 	case *[]string:
// 		var iv = gaussdbtype.FlatArray[string](*a)
// 		return &iv, nil
// 	default:
// 		return nil, errors.New("column '" + name + "' is array, it isnot support - []bool, []float64, []int64 and []string")
// 	}
// }
