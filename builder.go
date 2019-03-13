package gobatis

import (
	"database/sql"
	"errors"
	"fmt"
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

		if _, ok := field.Options["deleted"]; ok {
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

		_, isCreated := field.Options["created"]
		_, isUpdated := field.Options["updated"]

		if (AutoCreatedAt && ((isCreated && isTimeType(field.Field.Type)) || field.Name == "created_at")) ||
			(AutoUpdatedAt && ((isUpdated && isTimeType(field.Field.Type)) || field.Name == "updated_at")) {

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

		if _, ok := field.Options["deleted"]; ok {
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
			_, isCreated := field.Options["created"]
			_, isUpdated := field.Options["updated"]

			if (isCreated && isTimeType(field.Field.Type)) || (isUpdated && isTimeType(field.Field.Type)) || "created_at" == field.Name || "updated_at" == field.Name {

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

			_, isCreated := field.Options["created"]
			_, isUpdated := field.Options["updated"]

			if (isCreated && isTimeType(field.Field.Type)) || (isUpdated && isTimeType(field.Field.Type)) || "created_at" == field.Name || "updated_at" == field.Name {
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

		_, isCreated := field.Options["created"]
		_, isUpdated := field.Options["updated"]
		if (AutoCreatedAt && ((isCreated && isTimeType(field.Field.Type)) || field.Name == "created_at")) ||
			(AutoUpdatedAt && ((isUpdated && isTimeType(field.Field.Type)) || field.Name == "updated_at")) {
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

func GenerateUpsertSQL(dbType Dialect, mapper *Mapper, rType reflect.Type, names []string, argTypes []reflect.Type, noReturn bool) (string, error) {
	var sb strings.Builder
	sb.WriteString("INSERT INTO ")
	tableName, err := ReadTableName(mapper, rType)
	if err != nil {
		return "", err
	}
	sb.WriteString(tableName)
	sb.WriteString("(")

	if len(names) == 0 {
		for _, field := range mapper.TypeMap(rType).Index {
			if _, ok := field.Options["pk"]; ok {
				names = append(names, field.Name)
			}
		}

		if len(names) == 0 {
			return "", errors.New("upsert isnot generate for the " + tableName)
		}
	}

	skip := func(field *FieldInfo, isUpdated bool) bool {
		if field.Field.Name == "TableName" {
			return true
		}
		if field.Field.Anonymous {
			return true
		}

		if field.Parent != nil && len(field.Parent.Index) != 0 && !field.Parent.Field.Anonymous {
			return true
		}

		for _, name := range names {
			if name := strings.ToLower(name); name == strings.ToLower(field.Name) ||
				name == strings.ToLower(field.FieldName) {
				if isUpdated {
					return true
				} else {
					return false
				}
			}
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

		if _, ok := field.Options["deleted"]; ok {
			return true
		}
		return false
	}

	isFirst := true
	for _, field := range mapper.TypeMap(rType).Index {
		if skip(field, false) {
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

	sb.WriteString(" VALUES(")

	isFirst = true
	for _, field := range mapper.TypeMap(rType).Index {
		if skip(field, false) {
			continue
		}

		if !isFirst {
			sb.WriteString(", ")
		} else {
			isFirst = false
		}

		_, isCreated := field.Options["created"]
		_, isUpdated := field.Options["updated"]

		if (AutoCreatedAt && ((isCreated && isTimeType(field.Field.Type)) || field.Name == "created_at")) ||
			(AutoUpdatedAt && ((isUpdated && isTimeType(field.Field.Type)) || field.Name == "updated_at")) {

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

	switch dbType {
	case DbTypePostgres:
		// @postgres insert into auth_users(username, phone, address, status, birth_day, created_at, updated_at)
		// values (?,?,?,?,?,CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
		// ON CONFLICT (id) DO UPDATE SET
		//   username=EXCLUDED.username, phone=EXCLUDED.phone, address=EXCLUDED.address, status=EXCLUDED.status,
		//   birth_day=EXCLUDED.birth_day, created_at=EXCLUDED.created_at, updated_at=EXCLUDED.updated_at

		sb.WriteString(" ON CONFLICT (")
		for idx, name := range names {
			if idx != 0 {
				sb.WriteString(", ")
			}
			sb.WriteString(name)
		}
		sb.WriteString(") DO UPDATE SET ")

		isFirst = true
		for _, field := range mapper.TypeMap(rType).Index {
			if skip(field, true) {
				continue
			}
			if !isFirst {
				sb.WriteString(", ")
			} else {
				isFirst = false
			}
			sb.WriteString(field.Name)
			sb.WriteString("=EXCLUDED.")
			sb.WriteString(field.Name)
		}
	case DbTypeMSSql:
		// @mssql MERGE auth_users USING (
		//     VALUES (?,?,?,?,?, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
		// ) AS foo (username, phone, address, status, birth_day, created_at, updated_at)
		// ON auth_users.username = foo.username
		// WHEN MATCHED THEN
		//    UPDATE SET username=foo.username, phone=foo.phone, address=foo.address, status=foo.status, birth_day=foo.birth_day, updated_at=foo.updated_at
		// WHEN NOT MATCHED THEN
		//    INSERT (username, phone, address, status, birth_day, created_at, updated_at)
		//    VALUES (foo.username, foo.phone, foo.address, foo.status, foo.birth_day,  foo.created_at, foo.updated_at);
		return "", errors.New("upsert is unimplemented for mssql")
	case DbTypeMysql:
		// @mysql insert into auth_users(username, phone, address, status, birth_day, created_at, updated_at)
		// values (?,?,?,?,?,CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
		// on duplicate key update
		//   username=values(username), phone=values(phone), address=values(address),
		//   status=values(status), birth_day=values(birth_day), updated_at=CURRENT_TIMESTAMP

		sb.WriteString(" ON DUPLICATE KEY UPDATE ")
		isFirst = true
		for _, field := range mapper.TypeMap(rType).Index {
			if skip(field, true) {
				continue
			}
			if !isFirst {
				sb.WriteString(", ")
			} else {
				isFirst = false
			}
			sb.WriteString(field.Name)
			sb.WriteString("=VALUES(")
			sb.WriteString(field.Name)
			sb.WriteString(")")
		}
	default:
		return "", errors.New("upsert is unimplemented for db type - " + dbType.Name())
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
		if _, ok := field.Options["deleted"]; ok {
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

		if _, isUpdated := field.Options["updated"]; AutoUpdatedAt && ((isUpdated && isTimeType(field.Field.Type)) || field.Name == "updated_at") {
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
		err := generateWhere(dbType, mapper, rType, names, argTypes, nil, StatementTypeUpdate, false, &sb)
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
	deletedField := findDeletedField(mapper, rType)

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

		if deletedField != nil && deletedField.Name == field.Name {
			continue
		}

		if !isFirst {
			sb.WriteString(", ")
		} else {
			isFirst = false
		}

		sb.WriteString(field.Name)

		if _, isUpdated := field.Options["updated"]; AutoUpdatedAt && ((isUpdated && isTimeType(field.Field.Type)) || field.Name == "updated_at") {
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

	err = generateWhere(dbType, mapper, rType, []string{queryName}, []reflect.Type{queryType}, nil, StatementTypeUpdate, false, &sb)
	if err != nil {
		return "", err
	}

	return sb.String(), nil
}

func findDeletedField(mapper *Mapper, rType reflect.Type) *FieldInfo {
	structType := mapper.TypeMap(rType)
	for idx := range structType.Index {
		if _, ok := structType.Index[idx].Options["deleted"]; ok {
			return structType.Index[idx]
		}
	}
	return nil
}

func findForceArg(names []string, argTypes []reflect.Type, stmtType StatementType) int {
	excepted := "force"
	if stmtType != StatementTypeDelete {
		excepted = "isDeleted"
	}

	for idx, name := range names {
		if name != excepted {
			continue
		}

		if argTypes == nil {
			return idx
		}
		if argTypes[idx].Kind() == reflect.Bool {
			return idx
		}

		if ok, kind, _ := isValidable(argTypes[idx]); ok && kind == reflect.Bool {
			return idx
		}
	}
	return -1
}

type Filter struct {
	Expression string
	Dialect    string
}

func toFilters(filters []Filter, dbType Dialect) []string {
	results := make([]string, 0, len(filters))
	for idx := range filters {
		if filters[idx].Dialect != "" && ToDbType(filters[idx].Dialect) == dbType {
			continue
		}

		results = append(results, filters[idx].Expression)
	}
	return results
}

func GenerateDeleteSQL(dbType Dialect, mapper *Mapper, rType reflect.Type, names []string, argTypes []reflect.Type, filters []Filter) (string, error) {
	var deletedField = findDeletedField(mapper, rType)
	var forceIndex = findForceArg(names, argTypes, StatementTypeDelete)

	if deletedField != nil && forceIndex >= 0 && argTypes != nil {
		validable, _, _ := isValidable(argTypes[forceIndex])
		if validable {
			return "", errors.New("argument '" + names[forceIndex] + "' is unsupported type")
		}
	}

	var sb strings.Builder
	sb.WriteString("DELETE FROM ")
	tableName, err := ReadTableName(mapper, rType)
	if err != nil {
		return "", err
	}
	sb.WriteString(tableName)

	exprs := toFilters(filters, dbType)
	if len(names) > 0 && (deletedField == nil || forceIndex < 0 || len(names) > 1) {
		err := generateWhere(dbType, mapper, rType, names, argTypes, exprs, StatementTypeDelete, false, &sb)
		if err != nil {
			return "", err
		}
	} else if len(exprs) > 0 {
		sb.WriteString(" WHERE ")
		for idx := range exprs {
			s := strings.TrimSpace(exprs[idx])
			if idx > 0 {
				sb.WriteString(" AND ")
			}
			sb.WriteString(s)
		}
	}

	if deletedField == nil {
		return sb.String(), nil
	}

	var full strings.Builder
	if forceIndex >= 0 {
		full.WriteString(`<if test="`)
		full.WriteString(names[forceIndex])
		full.WriteString(`">`)
	}
	full.WriteString(`UPDATE `)

	full.WriteString(tableName)
	full.WriteString(" SET ")
	full.WriteString(deletedField.Name)
	if dbType == DbTypePostgres {
		full.WriteString("=now() ")
	} else {
		full.WriteString("=CURRENT_TIMESTAMP ")
	}

	if len(names) > 0 && (forceIndex < 0 || len(names) > 1) {
		err := generateWhere(dbType, mapper, rType, names, argTypes, exprs, StatementTypeDelete, false, &full)
		if err != nil {
			return "", err
		}
	} else if len(exprs) > 0 {
		sb.WriteString(" WHERE ")
		for idx := range exprs {
			s := strings.TrimSpace(exprs[idx])
			if idx > 0 {
				sb.WriteString(" AND ")
			}
			sb.WriteString(s)
		}
	}

	if forceIndex >= 0 {
		full.WriteString("</if>")
		full.WriteString(`<if test="!`)
		full.WriteString(names[forceIndex])
		full.WriteString(`">`)
		full.WriteString(sb.String())
		full.WriteString("</if>")
	}
	return full.String(), nil
}

func GenerateSelectSQL(dbType Dialect, mapper *Mapper, rType reflect.Type, names []string, argTypes []reflect.Type, filters []Filter, order string) (string, error) {
	var sb strings.Builder
	sb.WriteString("SELECT * FROM ")
	tableName, err := ReadTableName(mapper, rType)
	if err != nil {
		return "", err
	}
	sb.WriteString(tableName)

	exprs := toFilters(filters, dbType)
	if len(names) > 0 {
		err := generateWhere(dbType, mapper, rType, names, argTypes, exprs, StatementTypeSelect, false, &sb)
		if err != nil {
			return "", err
		}
	} else if deletedField := findDeletedField(mapper, rType); deletedField != nil {
		sb.WriteString(" WHERE ")
		sb.WriteString(deletedField.Name)
		sb.WriteString(" IS NULL")

		for idx := range exprs {
			s := strings.TrimSpace(exprs[idx])
			sb.WriteString(" AND ")
			sb.WriteString(s)
		}
	} else if len(exprs) > 0 {
		sb.WriteString(" WHERE ")
		for idx := range exprs {
			s := strings.TrimSpace(exprs[idx])
			if idx > 0 {
				sb.WriteString(" AND ")
			}
			sb.WriteString(s)
		}
	}
	if order != "" {
		sb.WriteString(" ORDER BY ")
		sb.WriteString(order)
	}
	return sb.String(), nil
}

func GenerateCountSQL(dbType Dialect, mapper *Mapper, rType reflect.Type, names []string, argTypes []reflect.Type, filters []Filter) (string, error) {
	var sb strings.Builder
	sb.WriteString("SELECT count(*) FROM ")
	tableName, err := ReadTableName(mapper, rType)
	if err != nil {
		return "", err
	}
	sb.WriteString(tableName)

	exprs := toFilters(filters, dbType)
	if len(names) > 0 {
		err := generateWhere(dbType, mapper, rType, names, argTypes, exprs, StatementTypeSelect, false, &sb)
		if err != nil {
			return "", err
		}
	} else if deletedField := findDeletedField(mapper, rType); deletedField != nil {
		sb.WriteString(" WHERE ")
		sb.WriteString(deletedField.Name)
		sb.WriteString(" IS NULL")

		for idx := range exprs {
			s := strings.TrimSpace(exprs[idx])
			sb.WriteString(" AND ")
			sb.WriteString(s)
		}
	} else if len(exprs) > 0 {
		sb.WriteString(" WHERE ")

		for idx := range exprs {
			s := strings.TrimSpace(exprs[idx])
			if idx > 0 {
				sb.WriteString(" AND ")
			}
			sb.WriteString(s)
		}
	}
	return sb.String(), nil
}

func generateWhere(dbType Dialect, mapper *Mapper, rType reflect.Type, names []string, argTypes []reflect.Type, exprs []string, stmtType StatementType, isCount bool, sb *strings.Builder) error {
	var deletedField = findDeletedField(mapper, rType)
	var forceIndex = findForceArg(names, argTypes, stmtType)
	var structType = mapper.TypeMap(rType)

	isNotNull := func(name string, argType reflect.Type) (bool, error) {
		fi, isSlice, err := toFieldName(structType, name, argType)
		if err != nil {
			if name == "offset" || name == "limit" {
				return false, nil
			}
			fii, isSlicei, e := toFieldName(structType, strings.TrimSuffix(strings.ToLower(name), "like"), argType)
			if e != nil {
				return false, err
			}

			fi = fii
			isSlice = isSlicei
		}
		if !isSlice {
			_, ok := fi.Options["notnull"]
			return ok, nil
		}
		return false, nil
	}
	needWhereTag := true
	if len(argTypes) == 0 {
		needWhereTag = false
	} else {
		for idx := range argTypes {
			if ok, _, _ := isValidable(argTypes[idx]); !ok {
				if deletedField == nil || forceIndex != idx {
					if notNull, err := isNotNull(names[idx], argTypes[idx]); err != nil {
						return err
					} else if !notNull {
						needWhereTag = false
						break
					}
				}
			}

		}
	}

	if needWhereTag {
		sb.WriteString(" <where>")
	} else {
		sb.WriteString(" WHERE ")
	}

	var nameArgs = make([]string, 0, len(exprs))
	for idx := range exprs {
		_, args, err := compileNamedQuery(exprs[idx])
		if err != nil {
			return err
		}

		if len(args) > 0 {
			for _, param := range args {
				found := false
				for _, nm := range names {
					if nm == param.Name {
						found = true
						break
					}
				}

				if !found {
					return errors.New("param '" + param.Name + "' isnot exists in the arguments")
				}
				nameArgs = append(nameArgs, param.Name)
			}
		}
	}

	inNameArgs := func(args []string, name string) bool {
		for idx := range args {
			if args[idx] == name {
				return true
			}
		}
		return false
	}

	hasOffset := false
	hasLimit := false
	isFirst := true
	for idx, name := range names {
		if inNameArgs(nameArgs, name) {
			continue
		}

		if deletedField != nil && forceIndex == idx {
			if stmtType == StatementTypeDelete {
				continue
			}

			if stmtType == StatementTypeSelect || stmtType == StatementTypeUpdate {
				validable := false
				if argTypes != nil {
					validable, _, _ = isValidable(argTypes[idx])
				}
				if validable {
					sb.WriteString(`<if test="`)
					sb.WriteString(name)
					sb.WriteString(`.Valid">`)
				}

				sb.WriteString(`<if test="`)
				sb.WriteString(name)
				if validable {
					sb.WriteString(`.Bool"> `)
				} else {
					sb.WriteString(`"> `)
				}
				if isFirst {
					isFirst = false
				} else {
					sb.WriteString(`AND `)
				}

				sb.WriteString(deletedField.Name)
				sb.WriteString(` IS NOT NULL </if>`)

				sb.WriteString(`<if test="!`)
				sb.WriteString(name)
				if validable {
					sb.WriteString(`.Bool"> `)
				} else {
					sb.WriteString(`"> `)
				}
				if isFirst {
					isFirst = false
				} else {
					sb.WriteString(`AND `)
				}

				sb.WriteString(deletedField.Name)
				sb.WriteString(` IS NULL `)
				sb.WriteString(`</if>`)

				if validable {
					sb.WriteString(`</if>`)
				}
				continue
			}
		}

		var argType reflect.Type
		if argTypes != nil {
			argType = argTypes[idx]
		}

		isLike := false
		field, isArgSlice, err := toFieldName(structType, name, argType)
		if err != nil {
			if stmtType == StatementTypeSelect {
				if name == "offset" {
					hasOffset = true
					continue
				}
				if name == "limit" {
					hasLimit = true
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
			if isFirst {
				isFirst = false
			} else {
				sb.WriteString(` AND `)
			}

			sb.WriteString(field.Name)
			sb.WriteString(` in (<foreach collection="`)
			sb.WriteString(name)
			sb.WriteString(`" item="item" separator="," >#{item}</foreach>)`)
		} else if ok, _, _ := isValidable(argType); ok {
			sb.WriteString(`<if test="`)
			sb.WriteString(name)
			sb.WriteString(`.Valid"> `)

			if isFirst {
				isFirst = false
			} else {
				sb.WriteString(`AND `)
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
			sb.WriteString(`</if>`)
		} else if ok := IsValueRange(argType); ok {
			if isFirst {
				isFirst = false
			} else {
				sb.WriteString(` AND`)
			}

			sb.WriteString(" (")
			sb.WriteString(field.Name)
			sb.WriteString(" BETWEEN #{")
			sb.WriteString(name)
			sb.WriteString(".Start} AND #{")
			sb.WriteString(name)
			sb.WriteString(".End}) ")
		} else if field.Field.Type.Kind() == reflect.Slice {
			if isFirst {
				isFirst = false
			} else {
				sb.WriteString(` AND `)
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
		} else if _, ok := field.Options["notnull"]; ok {
			if field.Field.Type.Kind() == reflect.String {
				sb.WriteString(`<if test="isNotEmpty(`)
				sb.WriteString(name)
				sb.WriteString(`)"> `)
			} else {
				sb.WriteString(`<if test="`)
				sb.WriteString(name)
				sb.WriteString(` != 0"> `)
			}
			if isFirst {
				isFirst = false
			} else {
				sb.WriteString(`AND `)
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
			sb.WriteString(`</if>`)
		} else {
			if isFirst {
				isFirst = false
			} else {
				sb.WriteString(` AND `)
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
		}
	}

	for idx := range exprs {
		s := strings.TrimSpace(exprs[idx])
		if isFirst {
			isFirst = false
		} else {
			sb.WriteString(` AND `)
		}

		sb.WriteString(s)
	}

	if stmtType == StatementTypeSelect {
		if forceIndex < 0 && deletedField != nil {
			if isFirst {
				isFirst = false
			} else {
				sb.WriteString(` AND `)
			}
			sb.WriteString(deletedField.Name)
			sb.WriteString(" IS NULL")
		}
	}

	if hasOffset {
		// <if test="offset &gt; 0"> OFFSET #{offset} </if>
		sb.WriteString(`<if test="offset &gt; 0"> OFFSET #{offset} </if>`)
	}
	if hasLimit {
		// <if test="limit &gt; 0"> LIMIT #{limit} </if>
		sb.WriteString(`<if test="limit &gt; 0"> LIMIT #{limit} </if>`)
	}

	if needWhereTag {
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
	Kind reflect.Kind
}{
	{reflect.TypeOf((*sql.NullBool)(nil)).Elem(), "Bool", reflect.Bool},
	{reflect.TypeOf((*sql.NullInt64)(nil)).Elem(), "Int64", reflect.Int64},
	{reflect.TypeOf((*sql.NullFloat64)(nil)).Elem(), "Float64", reflect.Float64},
	{reflect.TypeOf((*sql.NullString)(nil)).Elem(), "String", reflect.String},
}

func isValidable(argType reflect.Type) (bool, reflect.Kind, string) {
	if argType == nil {
		return false, reflect.Invalid, ""
	}

	if argType.Kind() != reflect.Struct {
		return false, reflect.Invalid, ""
	}

	for _, typ := range validableTypes {
		if argType.AssignableTo(typ.Typ) {
			return true, typ.Kind, typ.Name
		}
	}

	for idx := 0; idx < argType.NumField(); idx++ {
		if argType.Field(idx).Anonymous {
			if ok, kind, name := isValidable(argType.Field(idx).Type); ok {
				return true, kind, name
			}
		}
	}
	return false, reflect.Invalid, ""
}

func IsValueRange(argType reflect.Type) bool {
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

	startAt, ok := argType.FieldByName("Start")
	if !ok || !(isTimeType(startAt.Type) || isNumberType(startAt.Type)) {
		return false
	}

	endAt, ok := argType.FieldByName("End")
	if !ok || !(isTimeType(endAt.Type) || isNumberType(endAt.Type)) {
		return false
	}

	return true
}

var timeType = reflect.TypeOf(time.Time{})

func isTimeType(argType reflect.Type) bool {
	return argType.AssignableTo(timeType)
}

func SqlValuePrint(value interface{}) string {
	switch value.(type) {
	case int16, int32, int64, int, uint16, uint32, uint64, uint, float64, float32:
		return fmt.Sprint(value)
	default:
		return fmt.Sprintf("%q", value)
	}
}

func isNumberType(argType reflect.Type) bool {
	kind := argType.Kind()
	return kind == reflect.Float32 ||
		kind == reflect.Float64 ||
		kind == reflect.Int ||
		kind == reflect.Int8 ||
		kind == reflect.Int16 ||
		kind == reflect.Int32 ||
		kind == reflect.Int64 ||
		kind == reflect.Uint ||
		kind == reflect.Uint8 ||
		kind == reflect.Uint16 ||
		kind == reflect.Uint32 ||
		kind == reflect.Uint64
}
