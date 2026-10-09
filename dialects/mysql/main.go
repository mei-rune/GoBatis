package mysql

import (
	"strings"

	"github.com/go-sql-driver/mysql"
	"github.com/runner-mei/GoBatis/dialects"
)

func init() {
	dialects.SetHandleError(dialects.DriverMysql.Name(), handleError)
	dialects.SetHandleError(dialects.DriverMariadb.Name(), handleError)
	dialects.SetHandleError(dialects.DriverOceanbaseMysql.Name(), handleError)
}

func handleError(e error) error {
	if e == nil {
		return nil
	}
	//   fmt.Println("=================", fmt.Sprintf("%#v %T", e, e))

	// ================= &mysql.MySQLError{Number:0x47a, SQLState:[5]uint8{0x34, 0x32, 0x53, 0x30, 0x32}, Message:"Table 'golang.gobatis_test_table_not_exists' doesn't exist"} *mysql.MySQLError

	if pe, ok := e.(*mysql.MySQLError); ok {
		switch pe.Number {
		case 0x426: // ER_DUP_ENTRY, Duplicate entry 'xxx' for key 'idx_name'
			return &dialects.Error{Validations: []dialects.ValidationError{
				{Code: "unique_value_already_exists", Message: pe.Message, Columns: parseDuplicateEntryKey(pe.Message)},
			}, Err: e}

		case 0x47a:
			return dialects.ErrTableNotExists{
				Err: e,
				// Tablename: pe.Table,
			}
		}
	}
	return e
}

// parseDuplicateEntryKey 从 MySQL 的错误信息
// "Duplicate entry 'xxx' for key 'table.idx_name'" 中提取唯一索引名。
func parseDuplicateEntryKey(message string) []string {
	const prefix = "for key "
	idx := strings.LastIndex(message, prefix)
	if idx < 0 {
		return nil
	}

	key := strings.TrimSpace(message[idx+len(prefix):])
	if len(key) == 0 {
		return nil
	}
	if key[0] == '\'' || key[0] == '"' || key[0] == '`' {
		if end := strings.IndexByte(key[1:], key[0]); end >= 0 {
			key = key[1 : end+1]
		}
	}
	if key == "" {
		return nil
	}
	return []string{key}
}
