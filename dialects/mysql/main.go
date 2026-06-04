package mysql

import (
	"github.com/go-sql-driver/mysql"
	"github.com/runner-mei/GoBatis/dialects"
)

func init() {
	dialects.SetHandleError(dialects.Mysql.Name(), handleError)
	dialects.SetHandleError(dialects.Mariadb.Name(), handleError)
}

func handleError(e error) error {
	if e == nil {
		return nil
	}
	//   fmt.Println("=================", fmt.Sprintf("%#v %T", e, e))

	// ================= &mysql.MySQLError{Number:0x47a, SQLState:[5]uint8{0x34, 0x32, 0x53, 0x30, 0x32}, Message:"Table 'golang.gobatis_test_table_not_exists' doesn't exist"} *mysql.MySQLError

	if pe, ok := e.(*mysql.MySQLError); ok {
		switch pe.Number {
		// case "23505":
		//   detail := strings.TrimPrefix(strings.TrimPrefix(pe.Detail, "Key ("), "键值\"(")
		//   if pidx := strings.Index(detail, ")"); pidx > 0 {
		//     return &dialects.Error{Validations: []dialects.ValidationError{
		//       {Code: "unique_value_already_exists", Message: pe.Detail, Columns: strings.Split(detail[:pidx], ",")},
		//     }, e: e}
		//   }

		case 0x47a:
			return dialects.ErrTableNotExists{
				Err: e,
				// Tablename: pe.Table,
			}

			// case "23503":
			//  return &Error{Validations: []ValidationError{
			//    {Code: "PG.foreign_key_constraint", Message: pe.Message},
			//  }, e: e}
			// default:
			//   return &dialects.Error{Validations: []dialects.ValidationError{
			//     {Code: "PG." + pe.Code.Name(), Message: pe.Message, Columns: []string{pe.Column}},
			//   }, e: e}
		}
	}
	return e
}
