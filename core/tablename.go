package core

import (
	"errors"
	"reflect"
	"sync"
)

var (
	tableNameLock sync.Mutex
	tableNames    = map[reflect.Type]string{}
)

func RegisterTableName(value interface{}, name string) {
	var rType reflect.Type

	if t, ok := value.(reflect.Type); ok {
		rType = t
	} else if v, ok := value.(reflect.Value); ok {
		rType = v.Type()
	} else {
		rType = reflect.TypeOf(value)
	}
	if rType.Kind() == reflect.Ptr {
		rType = rType.Elem()
	}

	tableNameLock.Lock()
	defer tableNameLock.Unlock()
	tableNames[rType] = name
}

type TableNameInterface interface {
	TableName() string
}

var _tableNameInterface = reflect.TypeOf((*TableNameInterface)(nil)).Elem()

type TableName struct{}

func MustReadTableName(mapper *Mapper, rType reflect.Type) string {
	name, err := ReadTableName(mapper, rType)
	if err != nil {
		panic(err)
	}
	return name
}

func ReadTableName(mapper *Mapper, rType reflect.Type) (string, error) {
	if rType.Kind() == reflect.Ptr {
		rType = rType.Elem()
	}

	var tableName string
	tableNameLock.Lock()
	tableName = tableNames[rType]
	tableNameLock.Unlock()

	if tableName != "" {
		return tableName, nil
	}

	if mapper == nil {
		field, ok := rType.FieldByName("TableName")
		if ok {
			tableName = field.Tag.Get("xorm")
			if tableName == "" {
				tableName = field.Tag.Get("db")
			}
		}
	} else {
		for _, field := range mapper.TypeMap(rType).Index {
			if field.Field.Name == "TableName" {
				if tableName != "" {
					return "", errors.New("struct '" + rType.Name() + "'.TableName is mult choices")
				}
				tableName = field.Name
			}
		}
	}

	if tableName != "" {
		RegisterTableName(reflect.New(rType), tableName)
		return tableName, nil
	}

	if rType.Implements(_tableNameInterface) {
		method := reflect.New(rType).MethodByName("TableName")
		results := method.Call([]reflect.Value{})
		tableName = results[0].Interface().(string)
		RegisterTableName(reflect.New(rType), tableName)
		return tableName, nil
	}
	if pType := reflect.PtrTo(rType); pType.Implements(_tableNameInterface) {
		method := reflect.New(rType).MethodByName("TableName")
		results := method.Call([]reflect.Value{})
		tableName = results[0].Interface().(string)
		RegisterTableName(reflect.New(rType), tableName)
		return tableName, nil
	}

	return "", errors.New("struct '" + rType.Name() + "' TableName is missing")
}
