package gobatis

import (
	"database/sql"
	"errors"
	"reflect"
	"strings"
	"sync"

	"github.com/grsmv/inflect"
)

var (
	tableNameLock sync.Mutex
	tableNames    = map[reflect.Type]string{}

	AutoCreatedAt = true
	AutoUpdatedAt = true
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

func GenerateInsertSQL(dbType Dialect, mapper *Mapper, rType reflect.Type, noReturn bool) (string, error) {
	var sb strings.Builder
	sb.WriteString("INSERT INTO ")
	tableName, err := ReadTableName(mapper, rType)
	if err != nil {
		return "", err
	}
	sb.WriteString(tableName)
	sb.WriteString("(")

	skip := func(field *FieldInfo) bool {
		if field.Field.Name == "TableName" {
			return true
		}
		if field.Field.Anonymous {
			return true
		}

		if field.Parent != nil && len(field.Parent.Index) != 0 && !field.Parent.Field.Anonymous {
			return true
		}

		if _, ok := field.Options["autoincr"]; ok {
			return true
		}

		if _, ok := field.Options["<-"]; ok {
			return true
		}
		return false
	}

	isFirst := true
	for _, field := range mapper.TypeMap(rType).Index {
		if skip(field) {
			continue
		}
		if !isFirst {
			sb.WriteString(", ")
		} else {
			isFirst = false
		}

		sb.WriteString(field.Name)
	}
	sb.WriteString(")")

	if dbType == DbTypeMSSql {
		if !noReturn {
			for _, field := range mapper.TypeMap(rType).Index {
				if _, ok := field.Options["autoincr"]; ok {
					sb.WriteString(" OUTPUT inserted.")
					sb.WriteString(field.Name)
					break
				}
			}
		}
	}

	sb.WriteString(" VALUES(")

	isFirst = true
	for _, field := range mapper.TypeMap(rType).Index {
		if skip(field) {
			continue
		}

		if !isFirst {
			sb.WriteString(", ")
		} else {
			isFirst = false
		}

		if (AutoCreatedAt && field.Name == "created_at") || (AutoUpdatedAt && field.Name == "updated_at") {
			if dbType == DbTypePostgres {
				sb.WriteString("now()")
			} else {
				sb.WriteString("CURRENT_TIMESTAMP")
			}
			continue
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

func GenerateInsertSQL2(dbType Dialect, mapper *Mapper, rType reflect.Type, fields []string, noReturn bool) (string, error) {
	var sb strings.Builder
	sb.WriteString("INSERT INTO ")
	tableName, err := ReadTableName(mapper, rType)
	if err != nil {
		return "", err
	}
	sb.WriteString(tableName)
	sb.WriteString("(")

	skip := func(field *FieldInfo) bool {
		if field.Field.Name == "TableName" {
			return true
		}
		if field.Field.Anonymous {
			return true
		}

		if field.Parent != nil && len(field.Parent.Index) != 0 && !field.Parent.Field.Anonymous {
			return true
		}

		if _, ok := field.Options["autoincr"]; ok {
			return true
		}

		if _, ok := field.Options["<-"]; ok {
			return true
		}
		return false
	}

	isFirst := true
	for _, field := range mapper.TypeMap(rType).Index {
		foundIndex := -1
		for fidx, nm := range fields {
			nm := strings.ToLower(nm)
			if nm == strings.ToLower(field.Name) {
				foundIndex = fidx
				break
			}

			if nm == strings.ToLower(field.Field.Name) {
				foundIndex = fidx
				break
			}
		}
		if skip(field) {
			if foundIndex >= 0 {
				return "", errors.New("field '" + fields[foundIndex] + "' cannot present")
			}
			continue
		}

		if foundIndex < 0 {
			if "created_at" == field.Name || "updated_at" == field.Name {

				if !isFirst {
					sb.WriteString(", ")
				} else {
					isFirst = false
				}

				sb.WriteString(field.Name)
				continue
			}

			if _, ok := field.Options["notnull"]; ok {
				return "", errors.New("field '" + fields[foundIndex] + "' is missing")
			}
			continue
		}

		if !isFirst {
			sb.WriteString(", ")
		} else {
			isFirst = false
		}

		sb.WriteString(field.Name)
	}
	sb.WriteString(")")

	if dbType == DbTypeMSSql {
		if !noReturn {
			for _, field := range mapper.TypeMap(rType).Index {
				if _, ok := field.Options["autoincr"]; ok {
					sb.WriteString(" OUTPUT inserted.")
					sb.WriteString(field.Name)
					break
				}
			}
		}
	}

	sb.WriteString(" VALUES(")

	isFirst = true
	for _, field := range mapper.TypeMap(rType).Index {
		if skip(field) {
			continue
		}

		foundIndex := -1
		for fidx, nm := range fields {
			nm := strings.ToLower(nm)
			if nm == strings.ToLower(field.Name) {
				foundIndex = fidx
				break
			}

			if nm == strings.ToLower(field.Field.Name) {
				foundIndex = fidx
				break
			}
		}
		if foundIndex < 0 {

			if "created_at" == field.Name || "updated_at" == field.Name {
				if !isFirst {
					sb.WriteString(", ")
				} else {
					isFirst = false
				}

				if dbType == DbTypePostgres {
					sb.WriteString("now()")
				} else {
					sb.WriteString("CURRENT_TIMESTAMP")
				}
				continue
			}

			continue
		}

		if !isFirst {
			sb.WriteString(", ")
		} else {
			isFirst = false
		}

		if (AutoCreatedAt && field.Name == "created_at") || (AutoUpdatedAt && field.Name == "updated_at") {
			if dbType == DbTypePostgres {
				sb.WriteString("now()")
			} else {
				sb.WriteString("CURRENT_TIMESTAMP")
			}
			continue
		}

		sb.WriteString("#{")
		sb.WriteString(fields[foundIndex])
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

func GenerateUpdateSQL(dbType Dialect, mapper *Mapper, prefix string, rType reflect.Type, names []string) (string, error) {
	var sb strings.Builder
	sb.WriteString("UPDATE ")
	tableName, err := ReadTableName(mapper, rType)
	if err != nil {
		return "", err
	}
	sb.WriteString(tableName)
	sb.WriteString(" SET ")

	structType := mapper.TypeMap(rType)
	isFirst := true
	for _, field := range structType.Index {
		if field.Field.Name == "TableName" {
			continue
		}

		if field.Field.Anonymous {
			continue
		}

		if field.Parent != nil && len(field.Parent.Index) != 0 && !field.Parent.Field.Anonymous {
			continue
		}

		if field.Name == "created_at" {
			continue
		}

		if _, ok := field.Options["<-"]; ok {
			continue
		}

		if _, ok := field.Options["autoincr"]; ok {
			continue
		}
		if _, ok := field.Options["created"]; ok {
			continue
		}

		found := false
		for _, name := range names {
			if strings.ToLower(name) == strings.ToLower(field.Name) {
				found = true
				break
			}

			if strings.ToLower(name) == strings.ToLower(field.Field.Name) {
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

		if prefix != "" {
			sb.WriteString(prefix)
		}
		sb.WriteString(field.Name)
		sb.WriteString("}")
	}

	if len(names) > 0 {
		sb.WriteString(" WHERE ")

		for idx, name := range names {
			if idx > 0 {
				sb.WriteString(" AND ")
			}
			fieldName, _, err := toFieldName(structType, name, nil)
			if err != nil {
				return "", err
			}
			sb.WriteString(fieldName)
			sb.WriteString("=#{")
			sb.WriteString(name)
			sb.WriteString("}")
		}
	}
	return sb.String(), nil
}

func GenerateUpdateSQL2(dbType Dialect, mapper *Mapper, rType, queryType reflect.Type, queryName string, values []string) (string, error) {
	var sb strings.Builder
	sb.WriteString("UPDATE ")
	tableName, err := ReadTableName(mapper, rType)
	if err != nil {
		return "", err
	}
	sb.WriteString(tableName)
	sb.WriteString(" SET ")

	structType := mapper.TypeMap(rType)
	isFirst := true
	for _, fieldName := range values {
		var field *FieldInfo
		for idx := range structType.Index {
			if strings.ToLower(fieldName) == strings.ToLower(structType.Index[idx].Name) {
				field = structType.Index[idx]
				break
			}

			if strings.ToLower(fieldName) == strings.ToLower(structType.Index[idx].Field.Name) {
				field = structType.Index[idx]
				break
			}
		}
		if field == nil {
			return "", errors.New("field '" + fieldName + "' isnot exists in the " + rType.Name())
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
		sb.WriteString(fieldName)
		sb.WriteString("}")
	}

	for _, field := range structType.Index {
		if field.Name != "updated_at" {
			continue
		}

		alreadyExist := false
		for _, fieldName := range values {
			if strings.ToLower(fieldName) == strings.ToLower(field.Name) {
				alreadyExist = true
				break
			}
			if strings.ToLower(fieldName) == strings.ToLower(field.Field.Name) {
				alreadyExist = true
				break
			}
		}

		if alreadyExist {
			continue
		}

		if !isFirst {
			sb.WriteString(", ")
		} else {
			isFirst = false
		}

		sb.WriteString(field.Name)
		if dbType == DbTypePostgres {
			sb.WriteString("=now()")
		} else {
			sb.WriteString("=CURRENT_TIMESTAMP")
		}
	}

	fieldName, isSlice, err := toFieldName(structType, queryName, queryType)
	if err != nil {
		return "", err
	}

	if isSlice {
		sb.WriteString(" WHERE ")
		sb.WriteString(fieldName)
		sb.WriteString(` in (<foreach collection="`)
		sb.WriteString(queryName)
		sb.WriteString(`" item="item" separator="," >#{item}</foreach>)`)
	} else if ok, _ := isValidable(queryType); ok {
		sb.WriteString(`<if test="`)
		sb.WriteString(queryName)
		sb.WriteString(`.Valid">`)
		sb.WriteString(" WHERE ")
		sb.WriteString(fieldName)
		sb.WriteString("=#{")
		sb.WriteString(queryName)
		sb.WriteString("}")
		sb.WriteString(`</if>`)
	} else {
		sb.WriteString(" WHERE ")
		sb.WriteString(fieldName)
		sb.WriteString("=#{")
		sb.WriteString(queryName)
		sb.WriteString("}")
	}
	return sb.String(), nil
}

func GenerateDeleteSQL(dbType Dialect, mapper *Mapper, rType reflect.Type, names []string, argTypes []reflect.Type) (string, error) {
	var sb strings.Builder
	sb.WriteString("DELETE FROM ")
	tableName, err := ReadTableName(mapper, rType)
	if err != nil {
		return "", err
	}
	sb.WriteString(tableName)

	if len(names) > 0 {
		sb.WriteString(" WHERE ")

		structType := mapper.TypeMap(rType)
		for idx, name := range names {
			var argType reflect.Type
			if argTypes != nil {
				argType = argTypes[idx]
			}

			fieldName, isSlice, err := toFieldName(structType, name, argType)
			if err != nil {
				return "", err
			}
			if isSlice {
				if idx > 0 {
					sb.WriteString(" AND ")
				}

				sb.WriteString(fieldName)
				sb.WriteString(` in (<foreach collection="`)
				sb.WriteString(name)
				sb.WriteString(`" item="item" separator="," >#{item}</foreach>)`)
			} else if ok, _ := isValidable(argType); ok {
				sb.WriteString(`<if test="`)
				sb.WriteString(name)
				sb.WriteString(`.Valid">`)
				sb.WriteString(" AND ")
				sb.WriteString(fieldName)
				sb.WriteString("=#{")
				sb.WriteString(name)
				sb.WriteString("}")
				sb.WriteString(`</if>`)
			} else {
				if idx > 0 {
					sb.WriteString(" AND ")
				}
				sb.WriteString(fieldName)
				sb.WriteString("=#{")
				sb.WriteString(name)
				sb.WriteString("}")
			}
		}
	}
	return sb.String(), nil
}

func GenerateSelectSQL(dbType Dialect, mapper *Mapper, rType reflect.Type, names []string, argTypes []reflect.Type) (string, error) {
	var sb strings.Builder
	sb.WriteString("SELECT * FROM ")
	tableName, err := ReadTableName(mapper, rType)
	if err != nil {
		return "", err
	}
	sb.WriteString(tableName)

	if len(names) > 0 {
		sb.WriteString(" WHERE ")

		structType := mapper.TypeMap(rType)
		for idx, name := range names {
			var argType reflect.Type
			if argTypes != nil {
				argType = argTypes[idx]
			}

			fieldName, isSlice, err := toFieldName(structType, name, argType)
			if err != nil {
				return "", err
			}

			if isSlice {
				if idx > 0 {
					sb.WriteString(" AND ")
				}
				sb.WriteString(fieldName)
				sb.WriteString(` in (<foreach collection="`)
				sb.WriteString(name)
				sb.WriteString(`" item="item" separator="," >#{item}</foreach>)`)
			} else if ok, _ := isValidable(argType); ok {
				sb.WriteString(`<if test="`)
				sb.WriteString(name)
				sb.WriteString(`.Valid">`)
				sb.WriteString(" AND ")
				sb.WriteString(fieldName)
				sb.WriteString("=#{")
				sb.WriteString(name)
				sb.WriteString("}")
				sb.WriteString(`</if>`)
			} else {
				if idx > 0 {
					sb.WriteString(" AND ")
				}
				sb.WriteString(fieldName)
				sb.WriteString("=#{")
				sb.WriteString(name)
				sb.WriteString("}")
			}
		}
	}
	return sb.String(), nil
}

func GenerateCountSQL(dbType Dialect, mapper *Mapper, rType reflect.Type, names []string, argTypes []reflect.Type) (string, error) {
	var sb strings.Builder
	sb.WriteString("SELECT count(*) FROM ")
	tableName, err := ReadTableName(mapper, rType)
	if err != nil {
		return "", err
	}
	sb.WriteString(tableName)

	if len(names) > 0 {
		sb.WriteString(" WHERE ")

		structType := mapper.TypeMap(rType)
		for idx, name := range names {
			var argType reflect.Type
			if argTypes != nil {
				argType = argTypes[idx]
			}
			fieldName, isSlice, err := toFieldName(structType, name, argType)
			if err != nil {
				return "", err
			}
			if isSlice {
				if idx > 0 {
					sb.WriteString(" AND ")
				}
				sb.WriteString(fieldName)
				sb.WriteString(` in (<foreach collection="`)
				sb.WriteString(name)
				sb.WriteString(`" item="item" separator="," >#{item}</foreach>)`)
			} else if ok, _ := isValidable(argType); ok {
				sb.WriteString(`<if test="`)
				sb.WriteString(name)
				sb.WriteString(`.Valid">`)
				sb.WriteString(" AND ")
				sb.WriteString(fieldName)
				sb.WriteString("=#{")
				sb.WriteString(name)
				sb.WriteString("}")
				sb.WriteString(`</if>`)
			} else {
				if idx > 0 {
					sb.WriteString(" AND ")
				}
				sb.WriteString(fieldName)
				sb.WriteString("=#{")
				sb.WriteString(name)
				sb.WriteString("}")
			}
		}
	}
	return sb.String(), nil
}

func ToFieldName(mapper *Mapper, rType reflect.Type, name string, argType reflect.Type) (string, bool, error) {
	structType := mapper.TypeMap(rType)
	return toFieldName(structType, name, argType)
}

func toFieldName(structType *StructMap, name string, argType reflect.Type) (string, bool, error) {
	lower := strings.ToLower(name)
	for _, field := range structType.Index {
		if field.Field.Name == name {
			return field.Name, false, nil
		}

		if field.Name == name {
			return field.Name, false, nil
		}

		if strings.ToLower(field.Field.Name) == lower {
			return field.Name, false, nil
		}

		if strings.ToLower(field.Name) == lower {
			return field.Name, false, nil
		}
	}

	if argType != nil && argType.Kind() == reflect.Slice {

		var singularizeName string
		if lower == "ids" || lower == "id_list" || lower == "idlist" {
			singularizeName = "ID"
			lower = "id"
		} else {
			singularizeName = inflect.Singularize(name)
			lower = strings.ToLower(singularizeName)
		}
		for _, field := range structType.Index {
			if field.Field.Name == singularizeName {
				if argType.Elem().ConvertibleTo(field.Field.Type) {
					return field.Name, true, nil
				}
			}

			if field.Name == singularizeName {
				if argType.Elem().ConvertibleTo(field.Field.Type) {
					return field.Name, true, nil
				}
			}

			if strings.ToLower(field.Field.Name) == lower {
				if argType.Elem().ConvertibleTo(field.Field.Type) {
					return field.Name, true, nil
				}
			}
			if strings.ToLower(field.Name) == lower {
				if argType.Elem().ConvertibleTo(field.Field.Type) {
					return field.Name, true, nil
				}
			}
		}

	}
	return "", false, errors.New("field '" + name + "' is missing")
}

var validableTypes = []struct {
	Typ  reflect.Type
	Name string
}{
	{reflect.TypeOf((*sql.NullBool)(nil)).Elem(), "Bool"},
	{reflect.TypeOf((*sql.NullInt64)(nil)).Elem(), "Int64"},
	{reflect.TypeOf((*sql.NullFloat64)(nil)).Elem(), "Float64"},
	{reflect.TypeOf((*sql.NullString)(nil)).Elem(), "String"},
}

func isValidable(argType reflect.Type) (bool, string) {
	if argType == nil {
		return false, ""
	}

	if argType.Kind() != reflect.Struct {
		return false, ""
	}

	for _, typ := range validableTypes {
		if argType.AssignableTo(typ.Typ) {
			return true, typ.Name
		}
	}

	for idx := 0; idx < argType.NumField(); idx++ {
		if argType.Field(idx).Anonymous {
			if ok, name := isValidable(argType.Field(idx).Type); ok {
				return true, name
			}
		}
	}
	return false, ""
}
