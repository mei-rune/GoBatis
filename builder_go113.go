//go:build go1.13
// +build go1.13

package gobatis

import (
	"database/sql"
	"reflect"
)

func init() {
	validableTypes = append(validableTypes, validableTypeSpec{reflect.TypeOf((*sql.NullInt32)(nil)).Elem(), "Int32", reflect.Int32})
	validableTypes = append(validableTypes, validableTypeSpec{reflect.TypeOf((*sql.NullTime)(nil)).Elem(), "Time", reflect.Struct})
}
