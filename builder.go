package gobatis

import (
	"database/sql"
	"errors"
	"fmt"
	"net"
	"reflect"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/grsmv/inflect"
	"github.com/runner-mei/GoBatis/dialects"
)

var (
	tableNameLock sync.Mutex
	tableNames    = map[reflect.Type]string{}

	AutoCreatedAt              = true
	AutoUpdatedAt              = true
	UpsertSupportAutoIncrField = false
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

func notAuto(field *FieldInfo) bool {
	if field == nil {
		return false
	}

	_, ok := field.Options["notauto"]
	return ok
}

func GenerateInsertSQL(dbType Dialect, mapper *Mapper, rType reflect.Type, names []string, argTypes []reflect.Type, noReturn bool) (string, error) {
	if len(names) > 1 {
		return GenerateInsertSQL2(dbType, mapper, rType, names, noReturn)
	}

	mustPrefix := false
	if len(names) == 1 {
		if argTypes[0] == nil {
			return GenerateInsertSQL2(dbType, mapper, rType, names, noReturn)
		}
		if !isStructType(argTypes[0]) || isIgnoreStructType(argTypes[0]) {
			return GenerateInsertSQL2(dbType, mapper, rType, names, noReturn)
		}

		for _, field := range mapper.TypeMap(rType).Index {
			if field.Name == names[0] && !isSameType(field.Field.Type, argTypes[0]) {
				mustPrefix = true
				break
			}
		}
	}

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

		sb.WriteString(dbType.Quote(field.Name))
	}
	sb.WriteString(")")

	if dbType == dialects.MSSql {
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


		if  isTimeField(field) {

			if dbType == dialects.Postgres {
				sb.WriteString("now()")
			} else {
				sb.WriteString("CURRENT_TIMESTAMP")
			}
			continue
		}

		sb.WriteString("#{")
		if mustPrefix {
			sb.WriteString(names[0])
			sb.WriteString(".")
		}
		sb.WriteString(field.Name)
		sb.WriteString("}")
	}

	sb.WriteString(")")

	if dbType == dialects.Postgres {
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
			nm = strings.ToLower(nm)
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


			if isTimeField(field) {

				if !isFirst {
					sb.WriteString(", ")
				} else {
					isFirst = false
				}

				sb.WriteString(field.Name)
				continue
			}

			if _, ok := field.Options["notnull"]; ok {
				return "", errors.New("field '" + rType.Name() + "." + field.Name + "' is missing")
			}
			continue
		}

		if !isFirst {
			sb.WriteString(", ")
		} else {
			isFirst = false
		}

		sb.WriteString(dbType.Quote(field.Name))
	}
	sb.WriteString(")")

	if dbType == dialects.MSSql {
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
			if isTimeField(field) {
				if !isFirst {
					sb.WriteString(", ")
				} else {
					isFirst = false
				}

				if dbType == dialects.Postgres {
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

		if  isTimeField(field) {
			if dbType == dialects.Postgres {
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

	if dbType == dialects.Postgres {
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

func isTimeField(field *FieldInfo) bool {
	_, isCreated := field.Options["created"]
	_, isUpdated := field.Options["updated"]

	if (AutoCreatedAt && ((isCreated && isTimeType(field.Field.Type)) || (field.Name == "created_at" && !notAuto(field)))) ||
		(AutoUpdatedAt && ((isUpdated && isTimeType(field.Field.Type)) || (field.Name == "updated_at" && !notAuto(field)))) {
		return true
	}
	return false
}

func GenerateUpsertSQL(dbType Dialect, mapper *Mapper, rType reflect.Type, keyNames []string, argNames []string, argTypes []reflect.Type, noReturn bool) (string, error) {
	structType := mapper.TypeMap(rType)

	var keyFields []*FieldInfo
	if len(keyNames) == 0 {
		var incrFields []*FieldInfo
		for _, field := range structType.Index {
			if _, ok := field.Options["autoincr"]; ok {
				if _, ok := field.Options["pk"]; ok {
					incrFields = append(incrFields, field)
				}
				continue
			}
			if _, ok := field.Options["pk"]; ok {
				keyFields = append(keyFields, field)
			} else if _, ok := field.Options["unique"]; ok {
				keyFields = append(keyFields, field)
			}
		}
		if len(keyFields) == 0 {
			if len(incrFields) == 0 || !UpsertSupportAutoIncrField {
				return "", errors.New("upsert isnot generate")
			}

			keyFields = incrFields
		}
	} else {
		for idx := range keyNames {
			fi, _, err := toFieldName(structType, keyNames[idx], nil)
			if err != nil {
				return "", errors.New("upsert isnot generate, " + err.Error())
			}
			keyFields = append(keyFields, fi)
		}
	}

	if len(argNames) == 0 {
		return generateUpsertSQLForStruct(dbType, mapper, rType, keyNames, keyFields, "", noReturn)
	}

	if len(argNames) == 1 {
		var prefix string
		for _, field := range mapper.TypeMap(rType).Index {
			if field.Name == argNames[0] && !isSameType(field.Field.Type, argTypes[0]) {
				//    这里的是为下面情况的特殊处理
				//    结构为 type XXX struct { f1 int, f2  int}
				//    方法定义为 Insert(f1 *XXX) error
				//    对应 sql 为  insert into xxx (f1, f2) values(#{f1.f1}, #{f1.f2})
				//    而不是 insert into xxx (f1, f2) values(#{f1}, #{f2})
				//    因为 #{f1} 取的值为 f1 *XXX, 而不是期望的 f1.f1

				prefix = argNames[0] + "."
			}
		}
		return generateUpsertSQLForStruct(dbType, mapper, rType, keyNames, keyFields, prefix, noReturn)
	}

	tableName, err := ReadTableName(mapper, rType)
	if err != nil {
		return "", err
	}

	var insertFields, updateFields []*FieldInfo
	var originInsertNames []string
	var originUpdateNames []string

	fieldExists := func(list []*FieldInfo, field *FieldInfo) bool {
		for idx := range list {
			if list[idx] == field {
				return true
			}
		}
		return false
	}

	argExists := func(list []string, field *FieldInfo) bool {
		for idx := range list {
			if list[idx] == field.Name ||
				list[idx] == field.FieldName ||
				strings.ToLower(list[idx]) == strings.ToLower(field.Name) ||
				strings.ToLower(list[idx]) == strings.ToLower(field.FieldName) {
				return true
			}
		}
		return false
	}

	for idx, field := range keyFields {

		if len(keyNames) == 0 {
			if !argExists(argNames, field) {
				return "", errors.New("argument '" + field.Name + "' is missing")
			}
		}

		if !skipFieldForUpsert(keyFields, field, false) {
			insertFields = append(insertFields, field)
			if len(keyNames) > idx {
				originInsertNames = append(originInsertNames, keyNames[idx])
			} else {
				originInsertNames = append(originInsertNames, field.Name)
			}
		}
		if !skipFieldForUpsert(keyFields, field, true) {
			updateFields = append(updateFields, field)
			if len(keyNames) > idx {
				originUpdateNames = append(originUpdateNames, keyNames[idx])
			} else {
				originUpdateNames = append(originUpdateNames, field.Name)
			}
		}
	}

	for _, argName := range argNames {
		field, _, err := toFieldName(structType, argName, nil)
		if err != nil {
			return "", err
		}
		if fieldExists(insertFields, field) ||
			fieldExists(updateFields, field) {
			continue
		}
		if !skipFieldForUpsert(keyFields, field, false) {
			insertFields = append(insertFields, field)
			originInsertNames = append(originInsertNames, argName)
		}
		if !skipFieldForUpsert(keyFields, field, true) {
			updateFields = append(updateFields, field)
			originUpdateNames = append(originUpdateNames, argName)
		}
	}

	for _, field := range structType.Index {
		if fieldExists(insertFields, field) ||
			fieldExists(updateFields, field) {
			continue
		}

		if _, ok := field.Options["pk"]; ok {
			continue
		}

		if _, ok := field.Options["updated"]; ok || field.Name == "updated_at" {
			insertFields = append(insertFields, field)
			updateFields = append(updateFields, field)
		} else if _, ok := field.Options["created"]; ok || field.Name == "created_at" {
			insertFields = append(insertFields, field)
		}
	}

	if dbType == dialects.DM || dbType == dialects.Oracle {
		return GenerateUpsertOracle(dbType, mapper, rType, tableName, "", keyNames, keyFields, originInsertNames, insertFields, originUpdateNames, updateFields, noReturn)
	}

	if dbType == dialects.MSSql {
		return GenerateUpsertMSSQL(dbType, mapper, rType, tableName, "", keyNames, keyFields, originInsertNames, insertFields, originUpdateNames, updateFields, noReturn)
	}

	return generateUpsertSQL(dbType, mapper, rType, tableName, "", keyNames, keyFields, originInsertNames, insertFields, originUpdateNames, updateFields, noReturn)
}

func generateUpsertSQLForStruct(dbType Dialect, mapper *Mapper, rType reflect.Type, keyNames []string, keyFields []*FieldInfo, prefix string, noReturn bool) (string, error) {
	tableName, err := ReadTableName(mapper, rType)
	if err != nil {
		return "", err
	}

	var insertFields, updateFields []*FieldInfo
	for _, field := range mapper.TypeMap(rType).Index {
		if !skipFieldForUpsert(keyFields, field, false) {
			insertFields = append(insertFields, field)
		}
		if !skipFieldForUpsert(keyFields, field, true) {
			updateFields = append(updateFields, field)
		}
	}

	if dbType == dialects.DM || dbType == dialects.Oracle {
		return GenerateUpsertOracle(dbType, mapper, rType, tableName, prefix, keyNames, keyFields, nil, insertFields, nil, updateFields, noReturn)
	}

	if dbType == dialects.MSSql {
		return GenerateUpsertMSSQL(dbType, mapper, rType, tableName, prefix, keyNames, keyFields, nil, insertFields, nil, updateFields, noReturn)
	}

	return generateUpsertSQL(dbType, mapper, rType, tableName, prefix, keyNames, keyFields, nil, insertFields, nil, updateFields, noReturn)
}

func skipFieldForUpsert(keys []*FieldInfo, field *FieldInfo, isUpdated bool) bool {
	if field.Field.Name == "TableName" {
		return true
	}
	if field.Field.Anonymous {
		return true
	}

	if field.Parent != nil && len(field.Parent.Index) != 0 && !field.Parent.Field.Anonymous {
		return true
	}

	for _, fi := range keys {
		if name := strings.ToLower(fi.Name); name == strings.ToLower(field.Name) ||
			name == strings.ToLower(field.FieldName) ||
			fi.FieldName == field.FieldName {
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

	if _, ok := field.Options["updated"]; ok || field.Name == "updated_at" {
		return false
	}

	if _, ok := field.Options["created"]; ok || field.Name == "created_at" {
		if isUpdated {
			return true
		}
		return false
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

func generateUpsertSQL(dbType Dialect, mapper *Mapper, rType reflect.Type, tableName string, prefix string, keyNames []string, keyFields []*FieldInfo, originInsertNames []string, insertFields []*FieldInfo, originUpdateNames []string, updateFields []*FieldInfo, noReturn bool) (string, error) {
	var sb strings.Builder
	sb.WriteString("INSERT INTO ")
	sb.WriteString(tableName)
	sb.WriteString("(")

	for idx, field := range insertFields {
		if idx != 0 {
			sb.WriteString(", ")
		}
		sb.WriteString(dbType.Quote(field.Name))
	}
	sb.WriteString(")")

	sb.WriteString(" VALUES(")

	for idx, field := range insertFields {
		if idx != 0 {
			sb.WriteString(", ")
		}

		if isTimeField(field) {
			if dbType == dialects.Postgres {
				sb.WriteString("now()")
			} else {
				sb.WriteString("CURRENT_TIMESTAMP")
			}
			continue
		}

		sb.WriteString("#{")
		if len(originInsertNames) > idx {
			sb.WriteString(originInsertNames[idx])
		} else {
			sb.WriteString(prefix)
			sb.WriteString(field.Name)
		}
		sb.WriteString("}")
	}

	sb.WriteString(")")

	switch dbType {
	case dialects.Postgres:
		// @postgres insert into auth_users(username, phone, address, status, birth_day, created_at, updated_at)
		// values (?,?,?,?,?,CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
		// ON CONFLICT (id) DO UPDATE SET
		//   username=EXCLUDED.username, phone=EXCLUDED.phone, address=EXCLUDED.address, status=EXCLUDED.status,
		//   birth_day=EXCLUDED.birth_day, created_at=EXCLUDED.created_at, updated_at=EXCLUDED.updated_at

		sb.WriteString(" ON CONFLICT (")
		for idx, fi := range keyFields {
			if idx != 0 {
				sb.WriteString(", ")
			}
			sb.WriteString(dbType.Quote(fi.Name))
		}
		sb.WriteString(") DO")

		if len(updateFields) == 0 {
			sb.WriteString(" NOTHING ")
		} else {
			for idx, field := range updateFields {
				if idx != 0 {
					sb.WriteString(", ")
				} else {
					sb.WriteString(" UPDATE SET ")
				}

				sb.WriteString(dbType.Quote(field.Name))
				sb.WriteString("=EXCLUDED.")
				sb.WriteString(dbType.Quote(field.Name))
			}
		}

		if !noReturn {
			for _, field := range mapper.TypeMap(rType).Index {
				if _, ok := field.Options["autoincr"]; ok {
					sb.WriteString(" RETURNING ")
					sb.WriteString(dbType.Quote(field.Name))
					break
				}
			}
		}
	// case DbTypeMSSql:
	// 	// @mssql MERGE auth_users USING (
	// 	//     VALUES (?,?,?,?,?, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
	// 	// ) AS foo (username, phone, address, status, birth_day, created_at, updated_at)
	// 	// ON auth_users.username = foo.username
	// 	// WHEN MATCHED THEN
	// 	//    UPDATE SET username=foo.username, phone=foo.phone, address=foo.address, status=foo.status, birth_day=foo.birth_day, updated_at=foo.updated_at
	// 	// WHEN NOT MATCHED THEN
	// 	//    INSERT (username, phone, address, status, birth_day, created_at, updated_at)
	// 	//    VALUES (foo.username, foo.phone, foo.address, foo.status, foo.birth_day,  foo.created_at, foo.updated_at);
	// 	return "", errors.New("upsert is unimplemented for mssql")
	case dialects.Mysql:
		// @mysql insert into auth_users(username, phone, address, status, birth_day, created_at, updated_at)
		// values (?,?,?,?,?,CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
		// on duplicate key update
		//   username=values(username), phone=values(phone), address=values(address),
		//   status=values(status), birth_day=values(birth_day), updated_at=CURRENT_TIMESTAMP

		if len(updateFields) == 0 {
			return "", errors.New("empty update fields is unsupported")
		}

		sb.WriteString(" ON DUPLICATE KEY UPDATE ")
		for idx, field := range updateFields {
			if idx != 0 {
				sb.WriteString(", ")
			}
			sb.WriteString(dbType.Quote(field.Name))
			sb.WriteString("=VALUES(")
			sb.WriteString(dbType.Quote(field.Name))
			sb.WriteString(")")
		}
	default:
		return "", errors.New("upsert is unimplemented for db type - " + dbType.Name())
	}
	return sb.String(), nil
}

func GenerateUpsertOracle(dbType Dialect, mapper *Mapper, rType reflect.Type, tableName string, prefixName string, keyNames []string, keyFields []*FieldInfo, originInsertNames []string, insertFields []*FieldInfo, originUpdateNames []string, updateFields []*FieldInfo, noReturn bool) (string, error) {

	// MERGE INTO T1 USING dual ON T1.C1=1
	// WHEN MATCHED THEN UPDATE SET T1.C2='T2_1'
	// WHEN NOT MATCHED THEN INSERT (C1, C2) VALUES(1, 't2_1');

	var sb strings.Builder
	sb.WriteString("MERGE INTO ")
	sb.WriteString(tableName)
	sb.WriteString(" AS t USING dual ON ")

	for idx, fi := range keyFields {
		if idx != 0 {
			sb.WriteString(" AND ")
		}
		sb.WriteString("t.")
		sb.WriteString(dbType.Quote(fi.Name))

		sb.WriteString("= #{")
		if len(keyNames) > idx {
			sb.WriteString(keyNames[idx])
		} else {
			sb.WriteString(prefixName)
			sb.WriteString(fi.Name)
		}
		sb.WriteString("}")
	}

	if len(updateFields) > 0 {
		sb.WriteString(" WHEN MATCHED THEN UPDATE SET ")

		for idx, field := range updateFields {
			if idx != 0 {
				sb.WriteString(", ")
			}
			sb.WriteString(dbType.Quote(field.Name))
			sb.WriteString("= #{")
			if len(originUpdateNames) > idx {
				sb.WriteString(originUpdateNames[idx])
			} else {
				sb.WriteString(prefixName)
				sb.WriteString(field.Name)
			}
			sb.WriteString("}")
		}
	}

	sb.WriteString(" WHEN NOT MATCHED THEN INSERT (")

	for idx, field := range insertFields {
		if idx != 0 {
			sb.WriteString(", ")
		}
		sb.WriteString(dbType.Quote(field.Name))
	}
	sb.WriteString(") VALUES(")
	for idx, field := range insertFields {
		if idx != 0 {
			sb.WriteString(", ")
		}

		sb.WriteString("#{")
		if len(originInsertNames) > idx {
			sb.WriteString(originInsertNames[idx])
		} else {
			sb.WriteString(prefixName)
			sb.WriteString(field.Name)
		}
		sb.WriteString("}")
	}
	sb.WriteString(") ")

	// if !noReturn {
	// 	for _, field := range mapper.TypeMap(rType).Index {
	// 		if _, ok := field.Options["autoincr"]; ok {
	// 			sb.WriteString(" OUTPUT inserted.")
	// 			sb.WriteString(field.Name)
	// 			break
	// 		}
	// 	}
	// }
	return sb.String(), nil
}

func GenerateUpsertMSSQL(dbType Dialect, mapper *Mapper, rType reflect.Type, tableName string, prefixName string, keyNames []string, keyFields []*FieldInfo, originInsertNames []string, insertFields []*FieldInfo, originUpdateNames []string, updateFields []*FieldInfo, noReturn bool) (string, error) {
	// MERGE INTO t16_table AS t USING ( 
	//	   VALUES(#{f1}, #{f2}, #{f3}, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP ) 
    //   ) AS s (f1, f2, f3, created_at, updated_at ) 
    //     ON t.f1 = s.f1 
    //   WHEN MATCHED THEN UPDATE SET f2 = s.f2, f3 = s.f3, updated_at = s.updated_at 
    //   WHEN NOT MATCHED THEN INSERT (f1, f2, f3, created_at, updated_at) VALUES(s.f1, s.f2, s.f3, s.created_at, s.updated_at)  OUTPUT inserted.id


	var sb strings.Builder
	sb.WriteString("MERGE INTO ")
	sb.WriteString(tableName)
	sb.WriteString(" AS t USING ( VALUES(")
	for idx, field := range insertFields {
		if idx != 0 {
			sb.WriteString(", ")
		}

		if isTimeField(field) {
			sb.WriteString("CURRENT_TIMESTAMP")
			continue
		}

		sb.WriteString("#{")

		if len(originInsertNames) > idx {
			sb.WriteString(originInsertNames[idx])
		} else {
			sb.WriteString(prefixName)
			sb.WriteString(field.Name)
		}

		sb.WriteString("}")
	}
	sb.WriteString(" ) ) AS s (")

	for idx, field := range insertFields {
		if idx != 0 {
			sb.WriteString(", ")
		}
		sb.WriteString(dbType.Quote(field.Name))
	}
	sb.WriteString(" ) ON ")
	for idx, fi := range keyFields {
		if idx != 0 {
			sb.WriteString(" AND ")
		}
		sb.WriteString("t.")
		sb.WriteString(dbType.Quote(fi.Name))
		sb.WriteString(" = s.")
		sb.WriteString(dbType.Quote(fi.Name))
	}
	if len(updateFields) > 0 {
		sb.WriteString(" WHEN MATCHED THEN UPDATE SET ")

		for idx, field := range updateFields {
			if idx != 0 {
				sb.WriteString(", ")
			}
			sb.WriteString(dbType.Quote(field.Name))
			sb.WriteString(" = s.")
			sb.WriteString(dbType.Quote(field.Name))
		}
	}

	sb.WriteString(" WHEN NOT MATCHED THEN INSERT (")

	for idx, field := range insertFields {
		if idx != 0 {
			sb.WriteString(", ")
		}
		sb.WriteString(dbType.Quote(field.Name))
	}
	sb.WriteString(") VALUES(")
	for idx, field := range insertFields {
		if idx != 0 {
			sb.WriteString(", ")
		}

		sb.WriteString("s.")
		sb.WriteString(field.Name)
	}
	sb.WriteString(") ")

	if !noReturn {
		for _, field := range mapper.TypeMap(rType).Index {
			if _, ok := field.Options["autoincr"]; ok {
				sb.WriteString(" OUTPUT inserted.")
				sb.WriteString(dbType.Quote(field.Name))
				break
			}
		}
	}
	sb.WriteString(";")
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

		sb.WriteString(dbType.Quote(field.Name))

		if _, isUpdated := field.Options["updated"]; AutoUpdatedAt && ((isUpdated && isTimeType(field.Field.Type)) || field.Name == "updated_at") {
			if dbType == dialects.Postgres {
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

			sb.WriteString(dbType.Quote(field.Name))
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

		sb.WriteString(dbType.Quote(field.Name))

		if _, isUpdated := field.Options["updated"]; AutoUpdatedAt && ((isUpdated && isTimeType(field.Field.Type)) || field.Name == "updated_at") {
			if dbType == dialects.Postgres {
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

		sb.WriteString(dbType.Quote(field.Name))
		if dbType == dialects.Postgres {
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
		if filters[idx].Dialect != "" && NewDialect(filters[idx].Dialect) == dbType {
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
		full.WriteString(`<if test="!`)
		full.WriteString(names[forceIndex])
		full.WriteString(`">`)
	}
	full.WriteString(`UPDATE `)

	full.WriteString(tableName)
	full.WriteString(" SET ")
	full.WriteString(dbType.Quote(deletedField.Name))
	if dbType == dialects.Postgres {
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
		full.WriteString(" WHERE ")
		for idx := range exprs {
			s := strings.TrimSpace(exprs[idx])
			if idx > 0 {
				full.WriteString(" AND ")
			}
			full.WriteString(s)
		}
	}

	if forceIndex >= 0 {
		full.WriteString("</if>")
		full.WriteString(`<if test="`)
		full.WriteString(names[forceIndex])
		full.WriteString(`">`)
		full.WriteString(sb.String())
		full.WriteString("</if>")
	}
	return full.String(), nil
}

func GenerateSelectSQL(dbType Dialect, mapper *Mapper, rType reflect.Type, names []string, argTypes []reflect.Type, filters []Filter) (string, error) {
	var sb strings.Builder
	sb.WriteString("SELECT * FROM ")
	tableName, err := ReadTableName(mapper, rType)
	if err != nil {
		return "", err
	}
	sb.WriteString(tableName)

	hasOffset, hasLimit, hasOrderBy := false, false, false
	hasOffset, hasLimit, names, argTypes = removeOffsetAndLimit(names, argTypes)
	hasOrderBy, names, argTypes = removeArg(names, argTypes, "sortBy")

	exprs := toFilters(filters, dbType)
	if len(names) > 0 {
		err := generateWhere(dbType, mapper, rType, names, argTypes, exprs, StatementTypeSelect, false, &sb)
		if err != nil {
			return "", err
		}
	} else if deletedField := findDeletedField(mapper, rType); deletedField != nil {
		sb.WriteString(" WHERE ")
		sb.WriteString(dbType.Quote(deletedField.Name))
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

	if hasOrderBy {
		sb.WriteString(` <order_by by="sortBy"/>`)
	}

	if hasOffset {
		// <if test="offset &gt; 0"> OFFSET #{offset} </if>
		sb.WriteString(` <if test="offset &gt; 0"> OFFSET #{offset} </if>`)
	}
	if hasLimit {
		// <if test="limit &gt; 0"> LIMIT #{limit} </if>
		sb.WriteString(` <if test="limit &gt; 0"> LIMIT #{limit} </if>`)
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
		sb.WriteString(dbType.Quote(deletedField.Name))
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

func removeOffsetAndLimit(names []string, argTypes []reflect.Type) (bool, bool, []string, []reflect.Type) {
	hasOffset, nameCopy, argTypeCopy := removeArg(names, argTypes, "offset")
	hasLimit, nameCopy, argTypeCopy := removeArg(nameCopy, argTypeCopy, "limit")
	return hasOffset, hasLimit, nameCopy, argTypeCopy
}

func removeArg(names []string, argTypes []reflect.Type, name string) (bool, []string, []reflect.Type) {
	isExists := false

	for idx := range names {
		if names[idx] == name {
			isExists = true
		}
	}

	if !isExists {
		return false, names, argTypes
	}

	nameCopy := make([]string, 0, len(names))
	var argTypeCopy []reflect.Type
	if argTypes != nil {
		argTypeCopy = make([]reflect.Type, 0, len(argTypes))
	}

	for idx := range names {
		if names[idx] == name {
			continue
		}

		nameCopy = append(nameCopy, names[idx])
		if argTypes != nil {
			argTypeCopy = append(argTypeCopy, argTypes[idx])
		}
	}
	return isExists, nameCopy, argTypeCopy
}

func generateWhere(dbType Dialect, mapper *Mapper, rType reflect.Type, names []string, argTypes []reflect.Type, exprs []string, stmtType StatementType, isCount bool, sb *strings.Builder) error {
	// FIXME: 这里要重构，并处理好下面几种情况
	// <if/> AND xxx               -- 这里要将 AND 移到 if 中
	// xxx AND <if/>               -- 这里要将 AND 移到 if 中
	// xxx AND <if/> AND xxx       -- 这里要其中一个 AND 移到 if 中，不能两个都移到 if 中
	// xxx AND <if/> AND xxx AND <if/>
	// <if/> AND <if/>               -- 这里要将 AND 移到后一个 if 中

	var deletedField = findDeletedField(mapper, rType)
	var forceIndex = findForceArg(names, argTypes, stmtType)
	var structType = mapper.TypeMap(rType)

	isNotNull := func(name string, argType reflect.Type) (bool, error) {
		fi, isSlice, err := toFieldName(structType, name, argType)
		if err != nil {
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
	var needIFExprArray []bool
	if len(argTypes) == 0 {
		needWhereTag = false
	} else {
		needIFExprArray = make([]bool, len(argTypes))
		for idx := range argTypes {
			needIFExprArray[idx] = false

			if ok, _, _ := isValidable(argTypes[idx]); ok {
				needIFExprArray[idx] = true
				continue
			}

			if deletedField != nil && forceIndex == idx {
				continue
			}

			if notNull, err := isNotNull(names[idx], argTypes[idx]); err != nil {
				return err
			} else if notNull {
				needIFExprArray[idx] = true
			} else if strings.HasSuffix(strings.ToLower(names[idx]), "like") {
				needIFExprArray[idx] = true
			} else {
				needWhereTag = false
			}
		}
	}

	if deletedField != nil {
		if forceIndex < 0 && stmtType != StatementTypeDelete {
			needWhereTag = false
		}
	}
	if needWhereTag {
		sb.WriteString(" <where>")
	} else {
		sb.WriteString(" WHERE ")
	}

	var nameArgs, err = searchNameIndexs(exprs, names)
	if err != nil {
		return err
	}

	inNameArgs := func(args []int, a int) bool {
		for idx := range args {
			if args[idx] == a {
				return true
			}
		}
		return false
	}

	needANDExprSuffix := func(idx int) bool {

		nextStatic := false
		if (idx+1) < len(needIFExprArray) && !needIFExprArray[idx+1] {
			nextStatic = true
		} else if (len(exprs) > 0) || (deletedField != nil && stmtType != StatementTypeDelete) {
			nextStatic = true
		}

		if nextStatic {
			for i := idx - 1; i >= 0; i-- {
				if !needIFExprArray[i] {
					return false
				}
			}
			return true
		}

		return false
	}

	prefixANDExpr := false
	for idx, name := range names {
		if inNameArgs(nameArgs, idx) {
			continue
		}

		if deletedField != nil && forceIndex == idx {
			continue
		}

		var argType reflect.Type
		if argTypes != nil {
			argType = argTypes[idx]
		}

		isLike := false
		field, isArgSlice, err := toFieldName(structType, name, argType)
		if err != nil {
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
			if !prefixANDExpr {
				prefixANDExpr = true
			} else {
				sb.WriteString(` AND `)
			}

			sb.WriteString(dbType.Quote(field.Name))
			sb.WriteString(` in (<foreach collection="`)
			sb.WriteString(name)
			sb.WriteString(`" item="item" separator="," >#{item}</foreach>)`)
		} else if ok, _, _ := isValidable(argType); ok {
			sb.WriteString(`<if test="`)
			sb.WriteString(name)
			sb.WriteString(`.Valid"> `)

			if !prefixANDExpr {
				prefixANDExpr = true
			} else {
				sb.WriteString(`AND `)
			}

			sb.WriteString(dbType.Quote(field.Name))
			if isLike {
				sb.WriteString(" like ")
				sb.WriteString("<like value=\"")
				sb.WriteString(name)
				sb.WriteString("\" /> ")
			} else {
				sb.WriteString("=")
				sb.WriteString("#{")
				sb.WriteString(name)
				sb.WriteString("} ")
			}

			if needANDExprSuffix(idx) {
				sb.WriteString(`AND `)
				prefixANDExpr = false
			}
			sb.WriteString(`</if>`)
		} else if ok := IsValueRange(argType); ok {
			if !prefixANDExpr {
				prefixANDExpr = true
			} else {
				sb.WriteString(` AND`)
			}

			sb.WriteString(" (")
			sb.WriteString(dbType.Quote(field.Name))
			sb.WriteString(" BETWEEN #{")
			sb.WriteString(name)
			sb.WriteString(".Start} AND #{")
			sb.WriteString(name)
			sb.WriteString(".End})")
		} else if field.Field.Type.Kind() == reflect.Slice {
			if !prefixANDExpr {
				prefixANDExpr = true
			} else {
				sb.WriteString(` AND `)
			}
			_, jsonExists := field.Options["json"]
			if !jsonExists {
				_, jsonExists = field.Options["jsonb"]
			}
			if jsonExists {
				sb.WriteString(dbType.Quote(field.Name))
				sb.WriteString(" @> ")
				sb.WriteString("#{")
				sb.WriteString(name)
				sb.WriteString("}")
			} else {
				sb.WriteString("#{")
				sb.WriteString(name)
				sb.WriteString("} = ANY (")
				sb.WriteString(dbType.Quote(field.Name))
				sb.WriteString(")")
			}
		} else if _, ok := field.Options["notnull"]; ok {
			if field.Field.Type.Kind() == reflect.String {
				if argTypes != nil {
					if argTypes[idx].Kind() != reflect.String {
						return errors.New("arg '" + names[idx] + "' isnot expect string type, actual is " + argTypes[idx].Kind().String())
					}
				}
				sb.WriteString(`<if test="isNotEmptyString(`)
				sb.WriteString(name)
				sb.WriteString(`, true)"> `)
			} else {
				sb.WriteString(`<if test="`)
				sb.WriteString(name)
				sb.WriteString(` != 0"> `)
			}

			if !prefixANDExpr {
				prefixANDExpr = true
			} else {
				sb.WriteString(`AND `)
			}

			sb.WriteString(dbType.Quote(field.Name))
			if isLike {
				sb.WriteString(" like ")
				sb.WriteString("<like value=\"")
				sb.WriteString(name)
				sb.WriteString("\" /> ")
			} else {
				sb.WriteString("=")
				sb.WriteString("#{")
				sb.WriteString(name)
				sb.WriteString("} ")
			}

			if needANDExprSuffix(idx) {
				sb.WriteString(`AND `)
				prefixANDExpr = false
			}
			sb.WriteString(`</if>`)
		} else {
			if isLike {
				sb.WriteString(`<if test="isNotEmptyString(`)
				sb.WriteString(name)
				sb.WriteString(`, true)"> `)

				if !prefixANDExpr {
					prefixANDExpr = true
				} else {
					sb.WriteString(` AND `)
				}

				sb.WriteString(dbType.Quote(field.Name))
				sb.WriteString(" like ")
				sb.WriteString("<like value=\"")
				sb.WriteString(name)
				sb.WriteString("\" /> ")

				if needANDExprSuffix(idx) {
					sb.WriteString(`AND `)
					prefixANDExpr = false
				}

				sb.WriteString(`</if> `)
			} else {
				if !prefixANDExpr {
					prefixANDExpr = true
				} else {
					sb.WriteString(` AND `)
				}

				sb.WriteString(dbType.Quote(field.Name))
				sb.WriteString("=")
				sb.WriteString("#{")
				sb.WriteString(name)
				sb.WriteString("}")
			}
		}
	}

	for idx := range exprs {
		s := strings.TrimSpace(exprs[idx])

		if !prefixANDExpr {
			prefixANDExpr = true
		} else {
			sb.WriteString(` AND `)
		}

		sb.WriteString(s)
	}

	if deletedField != nil {
		if forceIndex >= 0 {
			if stmtType == StatementTypeSelect || stmtType == StatementTypeUpdate {
				validable := false
				if argTypes != nil {
					validable, _, _ = isValidable(argTypes[forceIndex])
				}
				if validable {
					sb.WriteString(`<if test="`)
					sb.WriteString(names[forceIndex])
					sb.WriteString(`.Valid">`)
				}

				sb.WriteString(`<if test="`)
				sb.WriteString(names[forceIndex])
				if validable {
					sb.WriteString(`.Bool"> `)
				} else {
					sb.WriteString(`"> `)
				}

				if !prefixANDExpr {
					prefixANDExpr = true
				} else {
					sb.WriteString(`AND `)
				}

				sb.WriteString(dbType.Quote(deletedField.Name))
				sb.WriteString(` IS NOT NULL </if>`)

				sb.WriteString(`<if test="!`)
				sb.WriteString(names[forceIndex])
				if validable {
					sb.WriteString(`.Bool"> `)
				} else {
					sb.WriteString(`"> `)
				}

				if !prefixANDExpr {
					prefixANDExpr = true
				} else {
					sb.WriteString(`AND `)
				}

				sb.WriteString(dbType.Quote(deletedField.Name))
				sb.WriteString(` IS NULL `)
				sb.WriteString(`</if>`)

				if validable {
					sb.WriteString(`</if>`)
				}
			}
		} else {
			if stmtType == StatementTypeSelect {
				if !prefixANDExpr {
					prefixANDExpr = true
				} else {
					sb.WriteString(` AND `)
				}
				sb.WriteString(dbType.Quote(deletedField.Name))
				sb.WriteString(" IS NULL")
			}
		}
	}

	if needWhereTag {
		sb.WriteString("</where>")
	}
	return nil
}

func searchNameIndexs(exprs, names []string) ([]int, error) {
	var nameArgs = make([]int, 0, len(exprs))
	for idx := range exprs {
		_, args, err := CompileNamedQuery(exprs[idx])
		if err != nil {
			return nil, err
		}

		if len(args) > 0 {
			for _, param := range args {
				foundIndex := -1
				for nameidx, nm := range names {
					if nm == param.Name {
						foundIndex = nameidx
						break
					}
				}

				if foundIndex < 0 {
					return nil, errors.New("param '" + param.Name + "' isnot exists in the arguments")
				}
				nameArgs = append(nameArgs, foundIndex)
			}
		}
	}
	return nameArgs, nil
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
	isTypeStr := lower == "typestr"

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

		if isTypeStr {
			if strings.ToLower(field.Name) == "type" {
				found = field
				break
			}
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
		} else if strings.HasSuffix(lower, "_list") {
			lower = strings.TrimSuffix(lower, "_list")
		} else if strings.HasSuffix(lower, "list") {
			lower = strings.TrimSuffix(lower, "list")
		} else {
			singularizeName = inflect.Singularize(name)
			if strings.HasPrefix(lower, singularizeName) {
				// handlerIDs 会变成 handlerid, 所以我修正一下
				singularizeName = name[:len(singularizeName)]
			}
			lower = strings.ToLower(singularizeName)
		}

		underscore := inflect.Underscore(singularizeName)
		if strings.HasSuffix(underscore, "_i_d") {
			underscore = strings.TrimSuffix(underscore, "_i_d") + "_id"
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

			if strings.ToLower(field.Field.Name) == underscore {
				if argType.Elem().ConvertibleTo(field.Field.Type) {
					return field, true, nil
				}
			}
			if strings.ToLower(field.Name) == underscore {
				if argType.Elem().ConvertibleTo(field.Field.Type) {
					return field, true, nil
				}
			}
		}
	}

	return nil, false, errors.New("field '" + name + "' is missing")
}

type validableTypeSpec struct {
	Typ  reflect.Type
	Name string
	Kind reflect.Kind
}

var validableTypes = []validableTypeSpec{
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

	return startAt.Type.Kind() == endAt.Type.Kind()
}

var timeType = reflect.TypeOf(time.Time{})

func isTimeType(argType reflect.Type) bool {
	return argType.AssignableTo(timeType)
}

func SqlValuePrint(value interface{}) string {
	switch value.(type) {
	case int8, int16, int32, int64, int, uint8, uint16, uint32, uint64, uint, float64, float32:
		return fmt.Sprint(value)
	case string:
		return "'" + value.(string) + "'"
	default:
		rv := reflect.ValueOf(value)
		kind := rv.Kind()
		if kind == reflect.Interface {
			rv = rv.Elem()
			kind = rv.Kind()
		}
		if kind == reflect.Int8 ||
			kind == reflect.Int16 ||
			kind == reflect.Int32 ||
			kind == reflect.Int64 ||
			kind == reflect.Int {
			return strconv.FormatInt(rv.Int(), 10)
		}
		if kind == reflect.Uint8 ||
			kind == reflect.Uint16 ||
			kind == reflect.Uint32 ||
			kind == reflect.Uint64 ||
			kind == reflect.Uint {
			return strconv.FormatUint(rv.Uint(), 10)
		}
		if kind == reflect.Float32 ||
			kind == reflect.Float64 {
			return strconv.FormatFloat(rv.Float(), 'f', -1, 10)
		}

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
		kind == reflect.Uint64 ||
		kind == reflect.Complex64 ||
		kind == reflect.Complex128
}

func isSameType(a, b reflect.Type) bool {
	if a == nil {
		return false
	}
	if b == nil {
		return false
	}
	if a.Kind() == reflect.Ptr {
		a = a.Elem()
	}

	if b.Kind() == reflect.Ptr {
		b = b.Elem()
	}

	return a == b
}

func isStructType(t reflect.Type) bool {
	kind := t.Kind()
	if kind == reflect.Ptr {
		kind = t.Elem().Kind()
	}
	return kind == reflect.Struct
}

var ignoreTypes = []reflect.Type{
	reflect.TypeOf((*time.Time)(nil)).Elem(),
	reflect.TypeOf((*net.IP)(nil)).Elem(),
	reflect.TypeOf((*net.HardwareAddr)(nil)).Elem(),
}

func isIgnoreStructType(argType reflect.Type) bool {
	if argType == nil {
		return false
	}

	if argType.Kind() != reflect.Struct {
		if argType.Kind() == reflect.Ptr {
			argType = argType.Elem()
		}
		if argType.Kind() != reflect.Struct {
			return false
		}
	}

	for _, typ := range validableTypes {
		if argType.AssignableTo(typ.Typ) {
			return true
		}
	}
	for _, typ := range ignoreTypes {
		if argType.AssignableTo(typ) {
			return true
		}
	}

	for idx := 0; idx < argType.NumField(); idx++ {
		if argType.Field(idx).Anonymous {
			if ok := isIgnoreStructType(argType.Field(idx).Type); ok {
				return true
			}
		}
	}

	pkgName := argType.PkgPath()
	if idx := strings.LastIndex(pkgName, "/"); idx >= 0 {
		pkgName = pkgName[idx+1:]
	}
	structName := pkgName + "." + argType.Name()
	for _, name := range IgnoreStructNames {
		if structName == name {
			return true
		}
	}
	return false
}

var IgnoreStructNames = []string{
	"time.Time",
	"sql.NullInt32",
	"sql.NullInt64",
	"sql.NullFloat64",
	"sql.NullString",
	"sql.NullBool",
	"sql.NullTime",
	"pq.NullTime",
	"null.Bool",
	"null.Float",
	"null.Int",
	"null.String",
	"null.Time",
}
