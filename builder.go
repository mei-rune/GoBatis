package gobatis

import (
	"errors"
	"reflect"
	"strings"
	"sync"

	"github.com/runner-mei/GoBatis/reflectx"
)

var (
	tableNameLock sync.Mutex
	tableNames    = map[reflect.Type]string{}
)

func RegisterTableName(value interface{}, name string) {
	rType := reflect.TypeOf(value)
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

func ReadTableName(mapper *reflectx.Mapper, rType reflect.Type) (string, error) {
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

	for _, field := range mapper.TypeMap(rType).Index {
		if field.Field.Name == "TableName" {
			if tableName != "" {
				return "", errors.New("struct '" + rType.Name() + "'.TableName is mult choices")
			}
			tableName = field.Name
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

func GenerateInsertSQL(dbType int, mapper *reflectx.Mapper, rType reflect.Type, noReturn bool) (string, error) {
	var sb strings.Builder
	sb.WriteString("INSERT INTO ")
	tableName, err := ReadTableName(mapper, rType)
	if err != nil {
		return "", err
	}
	sb.WriteString(tableName)
	sb.WriteString("(")

	isFirst := true
	for _, field := range mapper.TypeMap(rType).Index {
		if field.Field.Name == "TableName" {
			continue
		}
		if _, ok := field.Options["autoincr"]; ok {
			continue
		}
		if field.Field.Anonymous {
			continue
		}

		if !isFirst {
			sb.WriteString(", ")
		} else {
			isFirst = false
		}

		sb.WriteString(field.Name)
	}
	sb.WriteString(") VALUES(")

	isFirst = true
	for _, field := range mapper.TypeMap(rType).Index {
		if field.Field.Name == "TableName" {
			continue
		}
		if _, ok := field.Options["autoincr"]; ok {
			continue
		}
		if field.Field.Anonymous {
			continue
		}

		if !isFirst {
			sb.WriteString(", ")
		} else {
			isFirst = false
		}

		sb.WriteString("#{")
		sb.WriteString(field.Name)
		sb.WriteString("}")
	}

	sb.WriteString(")")

	if dbType == DbTypePostgres {
		if !noReturn {
			for _, field := range mapper.TypeMap(rType).Index {
				if _, ok := field.Options["autoincr"]; ok {
					sb.WriteString(" RETURNING ")
					sb.WriteString(field.Name)
					break
				}
			}
		}
	}
	return sb.String(), nil
}

func GenerateUpdateSQL(dbType int, mapper *reflectx.Mapper, rType reflect.Type, names []string) (string, error) {
	var sb strings.Builder
	sb.WriteString("UPDATE ")
	tableName, err := ReadTableName(mapper, rType)
	if err != nil {
		return "", err
	}
	sb.WriteString(tableName)
	sb.WriteString(" SET ")

	isFirst := true
	for _, field := range mapper.TypeMap(rType).Index {
		if field.Field.Name == "TableName" {
			continue
		}
		if _, ok := field.Options["autoincr"]; ok {
			continue
		}

		if field.Name == "created_at" {
			continue
		}

		if field.Field.Anonymous {
			continue
		}
		found := false
		for _, name := range names {
			if name == field.Name {
				found = true
				break
			}
		}
		if found {
			continue
		}

		if !isFirst {
			sb.WriteString(", ")
		} else {
			isFirst = false
		}

		sb.WriteString(field.Name)
		if field.Name == "updated_at" {
			if dbType == DbTypePostgres {
				sb.WriteString("=now()")
			} else {
				sb.WriteString("=CURRENT_TIMESTAMP")
			}
			continue
		}
		sb.WriteString("=#{")
		sb.WriteString(field.Name)
		sb.WriteString("}")
	}

	if len(names) > 0 {
		sb.WriteString(" WHERE ")

		for idx, name := range names {
			if idx > 0 {
				sb.WriteString(" AND ")
			}
			sb.WriteString(name)
			sb.WriteString("=#{")
			sb.WriteString(name)
			sb.WriteString("}")
		}
	}
	return sb.String(), nil
}

func GenerateDeleteSQL(dbType int, mapper *reflectx.Mapper, rType reflect.Type, names []string) (string, error) {
	var sb strings.Builder
	sb.WriteString("DELETE FROM ")
	tableName, err := ReadTableName(mapper, rType)
	if err != nil {
		return "", err
	}
	sb.WriteString(tableName)

	if len(names) > 0 {
		sb.WriteString(" WHERE ")

		for idx, name := range names {
			if idx > 0 {
				sb.WriteString(" AND ")
			}
			sb.WriteString(name)
			sb.WriteString("=#{")
			sb.WriteString(name)
			sb.WriteString("}")
		}
	}
	return sb.String(), nil
}

func GenerateSelectSQL(dbType int, mapper *reflectx.Mapper, rType reflect.Type, names []string) (string, error) {
	var sb strings.Builder
	sb.WriteString("SELECT * FROM ")
	tableName, err := ReadTableName(mapper, rType)
	if err != nil {
		return "", err
	}
	sb.WriteString(tableName)

	if len(names) > 0 {
		sb.WriteString(" WHERE ")

		for idx, name := range names {
			if idx > 0 {
				sb.WriteString(" AND ")
			}
			sb.WriteString(name)
			sb.WriteString("=#{")
			sb.WriteString(name)
			sb.WriteString("}")
		}
	}
	return sb.String(), nil
}

func GenerateCountSQL(dbType int, mapper *reflectx.Mapper, rType reflect.Type, names []string) (string, error) {
	var sb strings.Builder
	sb.WriteString("SELECT count(*) FROM ")
	tableName, err := ReadTableName(mapper, rType)
	if err != nil {
		return "", err
	}
	sb.WriteString(tableName)

	if len(names) > 0 {
		sb.WriteString(" WHERE ")

		for idx, name := range names {
			if idx > 0 {
				sb.WriteString(" AND ")
			}
			sb.WriteString(name)
			sb.WriteString("=#{")
			sb.WriteString(name)
			sb.WriteString("}")
		}
	}
	return sb.String(), nil
}
