package gobatis

import "reflect"

func toSQLType(param *Param, value interface{}) (interface{}, error) {
	return value, nil
}

func toSQLTypeWith(param *Param, value reflect.Value) (interface{}, error) {
	return value.Interface(), nil
}

func toGOTypeWith(instance, field reflect.Value) (interface{}, error) {
	return field.Interface(), nil
}
