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

func handleError(e error) error {
	if e == nil {
		return nil
	}

// "ERROR, 表或视图 \"GOBATIS_TEST_TABLE_NOT_EXISTS\" 不存在或无权访问\n"
//     main_test.go:32: want TableNotExists
//     main_test.go:33:  got ERROR, 表或视图 "GOBATIS_TEST_TABLE_NOT_EXISTS" 不存在或无权访问

 if strings.Contains(e.Error(), "表或视图") &&
 	strings.Contains(e.Error(), "不存在或无权访问") {
 		return dialects.ErrTableNotExists{
				Err:       e,
	 			// Tablename: pe.TableName,
			}
 }

	// if pe, ok := e.(*aci.GaussdbError); ok {
	// 	switch pe.Code {
	// 	case "23505":
	// 		detail := strings.TrimPrefix(strings.TrimPrefix(pe.Detail, "Key ("), "键值\"(")
	// 		if pidx := strings.Index(detail, ")"); pidx > 0 {
	// 			return &dialects.Error{Validations: []dialects.ValidationError{
	// 				{Code: "unique_value_already_exists", Message: pe.Detail, Columns: strings.Split(detail[:pidx], ",")},
	// 			}, Err: e}
	// 		}

	// 	case "42P01":
	// 		return dialects.ErrTableNotExists{
	// 			Err:       e,
	// 			Tablename: pe.TableName,
	// 		}

	// 	// case "23503":
	// 	//  return &Error{Validations: []ValidationError{
	// 	//    {Code: "PG.foreign_key_constraint", Message: pe.Message},
	// 	//  }, e: e}
	// 	default:
	// 		return &dialects.Error{Validations: []dialects.ValidationError{
	// 			{Code: "GaussDB." + pe.Code, Message: pe.Message, Columns: []string{pe.ColumnName}},
	// 		}, Err: e}
	// 	}
	// }
	return e
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
