package gobatis

import (
	"database/sql"
	"errors"
	"reflect"
	"strings"
	"sync"
	"time"

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

		if _, ok := field.Options["-"]; ok {
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

		if _, ok := field.Options["-"]; ok {
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
				return "", errors.New("field '" + field.Name + "' is missing")
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
		if field.Options != nil {
			if _, ok := field.Options["null"]; ok {
				sb.WriteString(",null=true")
			} else if _, ok := field.Options["notnull"]; ok {
				sb.WriteString(",notnull=true")
			}
		}
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

func GenerateUpdateSQL(dbType Dialect, mapper *Mapper, prefix string, rType reflect.Type, names []string, argTypes []reflect.Type) (string, error) {
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

		if _, ok := field.Options["-"]; ok {
			continue
		}

		if _, ok := field.Options["<-"]; ok {
			continue
		}

		if _, ok := field.Options["autoincr"]; ok {
			continue
		}
		if _, ok := field.Options["pk"]; ok {
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
		err := generateWhere(dbType, mapper, rType, names, argTypes, false, &sb)
		if err != nil {
			return "", err
		}
	} else {
		isFirst = true
		for _, field := range structType.Index {
			if field.Field.Name == "TableName" {
				continue
			}

			if _, ok := field.Options["pk"]; !ok {
				continue
			}

			if isFirst {
				isFirst = false
				sb.WriteString(" WHERE ")
			} else {
				sb.WriteString(" AND ")
			}

			sb.WriteString(field.Name)
			sb.WriteString("=#{")
			if prefix != "" {
				sb.WriteString(prefix)
			}
			sb.WriteString(field.Name)
			sb.WriteString("}")
		}

		if isFirst {
			return "", errors.New("primary key isnot found")
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
		if field.Options != nil {
			if _, ok := field.Options["null"]; ok {
				sb.WriteString(",null=true")
			} else if _, ok := field.Options["notnull"]; ok {
				sb.WriteString(",notnull=true")
			}
		}
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

	err = generateWhere(dbType, mapper, rType, []string{queryName}, []reflect.Type{queryType}, false, &sb)
	if err != nil {
		return "", err
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
		err := generateWhere(dbType, mapper, rType, names, argTypes, false, &sb)
		if err != nil {
			return "", err
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
		err := generateWhere(dbType, mapper, rType, names, argTypes, true, &sb)
		if err != nil {
			return "", err
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
		err := generateWhere(dbType, mapper, rType, names, argTypes, false, &sb)
		if err != nil {
			return "", err
		}
	}
	return sb.String(), nil
}

func generateWhere(dbType Dialect, mapper *Mapper, rType reflect.Type, names []string, argTypes []reflect.Type, isSelect bool, sb *strings.Builder) error {

	isAllValidable := true
	if len(argTypes) == 0 {
		isAllValidable = false
	} else {
		for idx := range argTypes {
			if ok, _ := isValidable(argTypes[idx]); !ok {
				isAllValidable = false
				break
			}
		}
	}

	if isAllValidable {
		sb.WriteString(" <where>")
	} else {
		sb.WriteString(" WHERE ")
	}
	structType := mapper.TypeMap(rType)
	isLastValidable := false
	for idx, name := range names {
		var argType reflect.Type
		if argTypes != nil {
			argType = argTypes[idx]
		}

		isLike := false
		field, isArgSlice, err := toFieldName(structType, name, argType)
		if err != nil {
			if isSelect {
				if name == "offset" {
					// <if test="offset &gt; 0"> OFFSET #{offset} </if>
					sb.WriteString(`<if test="offset &gt; 0"> OFFSET #{offset} </if>`)
					continue
				}
				if name == "limit" {
					// <if test="limit &gt; 0"> LIMIT #{limit} </if>
					sb.WriteString(`<if test="limit &gt; 0"> LIMIT #{limit} </if>`)
					continue
				}
			}
			if !strings.HasSuffix(strings.ToLower(name), "like") {
				return err
			}

			field, isArgSlice, err = toFieldName(structType, name[:len(name)-len("like")], argType)
			if err != nil {
				return err
			}
			if isArgSlice {
				return errors.New("'" + name + "' must cannot is a slice, like array is unsupported")
			}
			if field.Field.Type.Kind() != reflect.String {
				return errors.New("'" + name + "' must cannot is a string, like array is unsupported")
			}
			isLike = true
		}
		if isArgSlice {
			if idx > 0 {
				sb.WriteString(" AND ")
			}

			sb.WriteString(field.Name)
			sb.WriteString(` in (<foreach collection="`)
			sb.WriteString(name)
			sb.WriteString(`" item="item" separator="," >#{item}</foreach>)`)
			isLastValidable = false
		} else if ok, _ := isValidable(argType); ok {
			sb.WriteString(`<if test="`)
			sb.WriteString(name)
			sb.WriteString(`.Valid">`)
			if idx > 0 && !isLastValidable {
				sb.WriteString(" AND ")
			} else {
				sb.WriteString(" ")
			}
			sb.WriteString(field.Name)
			if isLike {
				sb.WriteString(" like ")
			} else {
				sb.WriteString("=")
			}
			sb.WriteString("#{")
			sb.WriteString(name)
			sb.WriteString("} ")

			if idx != (len(names) - 1) {
				sb.WriteString("AND ")
			}
			sb.WriteString(`</if>`)
			isLastValidable = true
		} else if ok := IsTimeRange(argType); ok {
			if idx > 0 && !isLastValidable {
				sb.WriteString(" AND (")
			} else {
				sb.WriteString(" (")
			}
			sb.WriteString(field.Name)
			sb.WriteString(" BETWEEN #{")
			sb.WriteString(name)
			sb.WriteString(".StartAt} AND #{")
			sb.WriteString(name)
			sb.WriteString(".EndAt}) ")
			if idx != (len(names) - 1) {
				sb.WriteString("AND ")
			}
			isLastValidable = false
		} else if field.Field.Type.Kind() == reflect.Slice {
			if idx > 0 && !isLastValidable {
				sb.WriteString(" AND ")
			}
			_, jsonExists := field.Options["json"]
			if !jsonExists {
				_, jsonExists = field.Options["jsonb"]
			}
			if jsonExists {
				sb.WriteString(field.Name)
				sb.WriteString(" @> ")
				sb.WriteString("#{")
				sb.WriteString(name)
				sb.WriteString("}")
			} else {
				sb.WriteString("#{")
				sb.WriteString(name)
				sb.WriteString("} = ANY (")
				sb.WriteString(field.Name)
				sb.WriteString(")")
			}
			isLastValidable = false
		} else {
			if idx > 0 && !isLastValidable {
				sb.WriteString(" AND ")
			}
			sb.WriteString(field.Name)
			if isLike {
				sb.WriteString(" like ")
			} else {
				sb.WriteString("=")
			}
			sb.WriteString("#{")
			sb.WriteString(name)
			sb.WriteString("}")
			isLastValidable = false
		}
	}

	if isAllValidable {
		sb.WriteString("</where>")
	}
	return nil
}

func ToFieldName(mapper *Mapper, rType reflect.Type, name string, argType reflect.Type) (*FieldInfo, bool, error) {
	structType := mapper.TypeMap(rType)
	return toFieldName(structType, name, argType)
}

func toFieldName(structType *StructMap, name string, argType reflect.Type) (*FieldInfo, bool, error) {
	isSlice := false
	if argType != nil && argType.Kind() == reflect.Slice {
		isSlice = true
	}

	lower := strings.ToLower(name)
	var found *FieldInfo
	for _, field := range structType.Index {
		if field.Field.Name == name {
			found = field
			break
		}

		if field.Name == name {
			found = field
			break
		}

		if strings.ToLower(field.Field.Name) == lower {
			found = field
			break
		}

		if strings.ToLower(field.Name) == lower {
			found = field
			break
		}
	}

	if found != nil {
		if isSlice {
			if !argType.Elem().ConvertibleTo(found.Field.Type) {
				isSlice = false
			}
		}
		return found, isSlice, nil
	}

	if isSlice {
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
					return field, true, nil
				}
			}

			if field.Name == singularizeName {
				if argType.Elem().ConvertibleTo(field.Field.Type) {
					return field, true, nil
				}
			}

			if strings.ToLower(field.Field.Name) == lower {
				if argType.Elem().ConvertibleTo(field.Field.Type) {
					return field, true, nil
				}
			}
			if strings.ToLower(field.Name) == lower {
				if argType.Elem().ConvertibleTo(field.Field.Type) {
					return field, true, nil
				}
			}
		}

	}
	return nil, false, errors.New("field '" + name + "' is missing")
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

func IsTimeRange(argType reflect.Type) bool {
	if argType == nil {
		return false
	}
	if argType.Kind() == reflect.Ptr {
		argType = argType.Elem()
	}
	if argType.Kind() != reflect.Struct {
		return false
	}

	if argType.NumField() != 2 {
		if argType.NumField() != 3 {
			return false
		}

		rangeField, ok := argType.FieldByName("Range")
		if !ok {
			return false
		}
		if rangeField.Type.NumField() != 0 {
			return false
		}
	}

	startAt, ok := argType.FieldByName("StartAt")
	if !ok || !isTimeType(startAt.Type) {
		return false
	}

	endAt, ok := argType.FieldByName("EndAt")
	if !ok || !isTimeType(endAt.Type) {
		return false
	}

	return true
}

var timeType = reflect.TypeOf(time.Time{})

func isTimeType(argType reflect.Type) bool {
	return argType.AssignableTo(timeType)
}
