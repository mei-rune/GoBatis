package core

import (
	"database/sql"
	"database/sql/driver"
	"errors"
	"fmt"
	"reflect"
	"strings"

	"github.com/runner-mei/GoBatis/reflectx"
)

var (
	_scannerInterface = reflect.TypeOf((*sql.Scanner)(nil)).Elem()
	_valuerInterface  = reflect.TypeOf((*driver.Valuer)(nil)).Elem()
)

// isScannable takes the reflect.Type and the actual dest value and returns
// whether or not it's Scannable.  Something is scannable if:
//   - it is not a struct
//   - it implements sql.Scanner
//   - it has no exported fields
func isScannable(mapper *Mapper, t reflect.Type) bool {
	if reflect.PtrTo(t).Implements(_scannerInterface) {
		return true
	}
	if t.Kind() != reflect.Struct {
		return true
	}

	if len(mapper.TypeMap(t).Index) == 0 {
		return true
	}
	return false
}

type ColumnInfoGetter interface {
	Columns() ([]string, error)
	ColumnTypes() ([]*sql.ColumnType, error)
}

// colScanner is an interface used by MapScan and SliceScan
type colScanner interface {
	ColumnInfoGetter
	Scan(dest ...interface{}) error
	Err() error
}

type rowsi interface {
	ColumnInfoGetter
	Err() error
	Next() bool
	Scan(...interface{}) error
}

// structOnlyError returns an error appropriate for type when a non-scannable
// struct is expected but something else is given
func structOnlyError(t reflect.Type) error {
	isStruct := t.Kind() == reflect.Struct
	isScanner := reflect.PtrTo(t).Implements(_scannerInterface)
	if !isStruct {
		return fmt.Errorf("expected %s but got %s", reflect.Struct, t.Kind())
	}
	if isScanner {
		return fmt.Errorf("structscan expects a struct dest but the provided struct type %s implements scanner", t.Name())
	}
	return fmt.Errorf("expected a struct, but struct %s has no exported fields", t.Name())
}

func ScanAny(dialect Dialect, mapper *Mapper, r colScanner, dest interface{}, structOnly, isUnsafe bool) error {
	return scanAny(dialect, mapper, r, dest, structOnly, isUnsafe)
}

func scanAny(dialect Dialect, mapper *Mapper, r colScanner, dest interface{}, structOnly, isUnsafe bool) error {
	if r.Err() != nil {
		return r.Err()
	}

	v := reflect.ValueOf(dest)
	if v.Kind() != reflect.Ptr {
		if mapDest, ok := dest.(map[string]interface{}); ok {
			if mapDest != nil {
				return MapScan(dialect, r, mapDest)
			}
		}

		return errors.New("must pass a pointer, not a value, to StructScan destination")
	}
	if v.IsNil() {
		return errors.New("nil pointer passed to StructScan destination")
	}

	if mapDest, ok := dest.(*map[string]interface{}); ok {
		if *mapDest == nil {
			*mapDest = map[string]interface{}{}
		}
		return MapScan(dialect, r, *mapDest)
	}

	base := reflectx.Deref(v.Type())
	scannable := isScannable(mapper, base)

	if structOnly && scannable {
		return structOnlyError(base)
	}

	columns, err := r.Columns()
	if err != nil {
		return err
	}

	if scannable && len(columns) > 1 {
		return fmt.Errorf("scannable dest type %s with >1 columns (%d) in result", base.Kind(), len(columns))
	}

	if scannable {
		return r.Scan(dest)
	}

	// if we are not unsafe and are missing fields, return an error
	fields, err := traversalsByName(mapper, v.Type(), columns, isUnsafe)
	if err != nil {
		return err
	}
	values := make([]interface{}, len(columns))

	err = fieldsByTraversal(dialect, v, columns, fields, values)
	if err != nil {
		return err
	}
	// scan into the struct field pointers and append to our results
	err = r.Scan(values...)
	if err != nil {
		return errors.New("Scan into " + toTypeName(dest) + "(" + strings.Join(columns, ",") + ") error : " + err.Error())
	}
	return nil
}

func toTypeName(dest interface{}) string {
	return fmt.Sprintf("%T", dest)
}

func ScanAll(dialect Dialect, mapper *Mapper, rows rowsi, dest interface{}, structOnly, isUnsafe bool) error {
	if mapSlice, ok := dest.(*[]map[string]interface{}); ok {
		return scanMapSlice(dialect, rows, mapSlice)
	}
	return scanAll(dialect, mapper, rows, dest, structOnly, isUnsafe)
}

func scanAll(dialect Dialect, mapper *Mapper, rows rowsi, dest interface{}, structOnly, isUnsafe bool) error {
	var v, vp reflect.Value

	value := reflect.ValueOf(dest)

	// json.Unmarshal returns errors for these
	if value.Kind() != reflect.Ptr {
		return errors.New("must pass a pointer, not a value, to StructScan destination")
	}
	if value.IsNil() {
		return errors.New("nil pointer passed to StructScan destination")
	}
	direct := reflect.Indirect(value)

	var isPtr bool
	var base reflect.Type
	var scannable bool
	var add func(v reflect.Value)

	if t := reflectx.Deref(value.Type()); t.Kind() == reflect.Slice {
		isPtr = t.Elem().Kind() == reflect.Ptr
		base = reflectx.Deref(t.Elem())
		scannable = isScannable(mapper, base)
		add = func(v reflect.Value) {
			direct.Set(reflect.Append(direct, v))
		}
	} else if t.Kind() == reflect.Map {
		if direct.IsNil() {
			direct.Set(reflect.Indirect(reflect.MakeMap(t)))
		}
		isPtr = t.Elem().Kind() == reflect.Ptr
		base = reflectx.Deref(t.Elem())
		scannable = isScannable(mapper, base)

		if len(mapper.TypeMap(base).PrimaryKey) == 0 {
			return fmt.Errorf("field with pk tag isnot exists in %s", base.Name())
		}

		if len(mapper.TypeMap(base).PrimaryKey) > 1 {
			return fmt.Errorf("field with pk tag is one than more in %s", base.Name())
		}

		keyIndexs := mapper.TypeMap(base).PrimaryKey[0]
		add = func(v reflect.Value) {
			k := reflectx.FieldByIndexes(v, keyIndexs)
			if !k.IsValid() {
				panic(fmt.Errorf("pk is invalid in the %s", base.Name()))
			}
			direct.SetMapIndex(k, v)
		}
	} else {
		return fmt.Errorf("expected %s or %s but got %s", reflect.Slice, reflect.Map, t.Kind())
	}

	if structOnly && scannable {
		return structOnlyError(base)
	}

	columns, err := rows.Columns()
	if err != nil {
		return err
	}

	// if it's a base type make sure it only has 1 column;  if not return an error
	if scannable && len(columns) > 1 {
		return fmt.Errorf("non-struct dest type %s with >1 columns (%d)", base.Kind(), len(columns))
	}

	if !scannable {
		var values []interface{}

		fields, err := traversalsByName(mapper, base, columns, isUnsafe)
		if err != nil {
			return err
		}

		values = make([]interface{}, len(columns))

		for rows.Next() {
			// create a new struct type (which returns PtrTo) and indirect it
			vp = reflect.New(base)
			v = reflect.Indirect(vp)

			err = fieldsByTraversal(dialect, v, columns, fields, values)
			if err != nil {
				return err
			}

			// scan into the struct field pointers and append to our results
			err = rows.Scan(values...)
			if err != nil {
				return errors.New("Scan into " + toTypeName(dest) + "(" + strings.Join(columns, ",") + ") error : " + err.Error())
			}

			if isPtr {
				add(vp)
			} else {
				add(v)
			}
		}
	} else {
		for rows.Next() {
			vp = reflect.New(base)
			err = rows.Scan(vp.Interface())
			if err != nil {
				return err
			}
			// append
			if isPtr {
				add(vp)
			} else {
				add(reflect.Indirect(vp))
			}
		}
	}

	return rows.Err()
}

func defaultScanValue() (ptrValue interface{}, valueGet func() interface{}) {
	var value interface{}
	return &value, func() interface{} {
		return value
	}
}

func makeColumnValue(name string, columnType *sql.ColumnType) func() (ptrValue interface{}, valueGet func() interface{}) {
	if columnType == nil {
		return defaultScanValue
	}

	switch strings.ToLower(columnType.DatabaseTypeName()) {
	case "tinyint", "smallint", "mediumint", "int", "bigint", "integer", "biginteger", "smallserial", "serial", "bigserial", "int1", "int2", "int3", "int4", "int8":
		return func() (ptrValue interface{}, valueGet func() interface{}) {
			var value sql.NullInt64
			return &value, func() interface{} {
				if value.Valid {
					return value.Int64
				}
				return nil
			}
		}
	case "float", "float4", "float8", "double", "decimal", "numeric", "real", "double precision":
		return func() (ptrValue interface{}, valueGet func() interface{}) {
			var value sql.NullFloat64
			return &value, func() interface{} {
				if value.Valid {
					return value.Float64
				}
				return nil
			}
		}
	case "varchar", "char", "text", "tinytext", "longtext", "mediumtext", "character varying", "character":
		//if nullable, ok := columnType.Nullable()

		return func() (ptrValue interface{}, valueGet func() interface{}) {
			var value sql.NullString
			return &value, func() interface{} {
				if value.Valid {
					return value.String
				}
				return nil
			}
		}
	case "boolean", "bool", "bit":
		//if nullable, ok := columnType.Nullable()

		return func() (ptrValue interface{}, valueGet func() interface{}) {
			var value sql.NullBool
			return &value, func() interface{} {
				if value.Valid {
					return value.Bool
				}
				return nil
			}
		}
	}

	t := columnType.ScanType()
	if t != nil {
		return func() (ptrValue interface{}, valueGet func() interface{}) {
			var value = reflect.New(t)
			var nullable Nullable
			nullable.Value = value.Interface()
			return &nullable, func() interface{} {
				return value.Elem().Interface()
			}
		}
	}

	return defaultScanValue
}

func createNewRecord(rows ColumnInfoGetter, columns []string) (func() (ptrArray []interface{}, valueArray []func() interface{}), error) {
	if len(columns) == 0 {
		names, err := rows.Columns()
		if err != nil {
			return nil, err
		}
		columns = names
	}

	newRecord := func() ([]interface{}, []func() interface{}) {
		values := make([]interface{}, len(columns))
		ptrValues := make([]interface{}, len(columns))
		valueArray := make([]func() interface{}, len(columns))
		for i := range values {
			ptrValues[i] = &values[i]
			valueArray[i] = func(idx int) func() interface{} {
				return func() interface{} {
					return values[idx]
				}
			}(i)
		}
		return ptrValues, valueArray
	}
	columnTypes, err := rows.ColumnTypes()
	if err != nil {
		return nil, err
	}
	if len(columnTypes) == len(columns) {
		valueArray := make([]func() (ptrValue interface{}, valueGet func() interface{}), len(columns))

		for i := range columns {
			valueArray[i] = makeColumnValue(columns[i], columnTypes[i])
		}

		newRecord = func() ([]interface{}, []func() interface{}) {
			ptrValues := make([]interface{}, len(valueArray))
			valueGetArray := make([]func() interface{}, len(valueArray))
			for i := range valueArray {
				ptrValues[i], valueGetArray[i] = valueArray[i]()
			}
			return ptrValues, valueGetArray
		}
	}
	return newRecord, nil
}

func scanMapSlice(dialect Dialect, rows rowsi, dest *[]map[string]interface{}) error {
	columns, err := rows.Columns()
	if err != nil {
		return err
	}

	newRecord, err := createNewRecord(rows, columns)
	if err != nil {
		return err
	}

	for rows.Next() {
		ptrArray, valueArray := newRecord()

		err = rows.Scan(ptrArray...)
		if err != nil {
			return errors.New("Scan into Map(" + strings.Join(columns, ",") + ") error : " + err.Error())
		}

		one := map[string]interface{}{}
		for i, column := range columns {
			one[column] = valueArray[i]()
		}
		*dest = append(*dest, one)
	}

	return rows.Err()
}

// FIXME: StructScan was the very first bit of API in sqlx, and now unfortunately
// it doesn't really feel like it's named properly.  There is an incongruency
// between this and the way that StructScan (which might better be ScanStruct
// anyway) works on a rows object.

// StructScan all rows from an sql.Rows or an sqlx.Rows into the dest slice.
// StructScan will scan in the entire rows result, so if you do not want to
// allocate structs for the entire result, use Queryx and see sqlx.Rows.StructScan.
// If rows is sqlx.Rows, it will use its mapper, otherwise it will use the default.
func StructScan(dialect Dialect, mapper *Mapper, rows rowsi, dest interface{}, isUnsafe bool) error {
	return scanAny(dialect, mapper, rows, dest, true, isUnsafe)
}

// MapScan scans a single Row into the dest map[string]interface{}.
// Use this to get results for SQL that might not be under your control
// (for instance, if you're building an interface for an SQL server that
// executes SQL from input).  Please do not use this as a primary interface!
// This will modify the map sent to it in place, so reuse the same map with
// care.  Columns which occur more than once in the result will overwrite
// each other!
func MapScan(dialect Dialect, row colScanner, dest map[string]interface{}) error {
	columns, err := row.Columns()
	if err != nil {
		return err
	}

	newRecord, err := createNewRecord(row, columns)
	if err != nil {
		return err
	}

	ptrArray, valueArray := newRecord()
	err = row.Scan(ptrArray...)
	if err != nil {
		return errors.New("Scan into Map(" + strings.Join(columns, ",") + ") error : " + err.Error())
	}

	for i, column := range columns {
		dest[column] = valueArray[i]()
	}

	return row.Err()
}

// reflect helpers

// fieldsByName fills a values interface with fields from the passed value based
// on the traversals in int.  If ptrs is true, return addresses instead of values.
// We write this instead of using FieldsByName to save allocations and map lookups
// when iterating over many rows.  Empty traversals will get an interface pointer.
// Because of the necessity of requesting ptrs or values, it's considered a bit too
// specialized for inclusion in reflectx itself.
func fieldsByTraversal(dialect Dialect, v reflect.Value, columns []string, traversals []*FieldInfo, values []interface{}) error {
	v = reflect.Indirect(v)
	if v.Kind() != reflect.Struct {
		return errors.New("argument not a struct")
	}

	for i, traversal := range traversals {
		fvalue, err := traversal.LValue(dialect, columns[i], v)
		if err != nil {
			return err
		}
		values[i] = fvalue
	}
	return nil
}

func traversalsByName(mapper *Mapper, t reflect.Type, columns []string, isUnsafe bool) ([]*FieldInfo, error) {
	tm := mapper.TypeMap(reflectx.Deref(t))
	var traversals []*FieldInfo
	for _, column := range columns {
		fi, _ := tm.Names[column]
		if fi == nil {
			fi, _ = tm.Names[strings.ToLower(column)]
			if fi == nil {
				if strings.HasPrefix(column, "deprecated_") {
					traversals = append(traversals, emptyField)
					continue
				}
				if isUnsafe {
					traversals = append(traversals, emptyField)
					continue
				}

				// var sb strings.Builder
				// sb.WriteString(t.Name())
				// sb.WriteString("{")
				// for name := range tm.Names {
				// 	sb.WriteString(name)
				// 	sb.WriteString(",")
				// }
				// sb.WriteString("}")
				// return nil, errors.New("colunm '" + column + "' isnot found in " + sb.String())

				return nil, errors.New("missing destination name '" + column + "' in " + t.Name())
			}
		}
		traversals = append(traversals, fi)
	}
	return traversals, nil
}

func scanStringMap(dialect Dialect, mapper *Mapper, rows rowsi, dest map[string]interface{}) error {
	columns, err := rows.Columns()
	if err != nil {
		return err
	}
	if len(columns) != 2 {
		return fmt.Errorf("dest type %T want 2 columns go %d columns", dest, len(columns))
	}
	if dest == nil {
		return errors.New("nil map passed to scanStringMap destination")
	}
	for rows.Next() {
		var key sql.NullString
		var value interface{}
		err = rows.Scan(&key, &value)
		if err != nil {
			return err
		}
		if value != nil || key.Valid {
			dest[key.String] = value
		}
	}
	return rows.Err()
}

func scanStringInt64Map(dialect Dialect, mapper *Mapper, rows rowsi, dest map[string]int64) error {
	columns, err := rows.Columns()
	if err != nil {
		return err
	}
	if len(columns) != 2 {
		return fmt.Errorf("dest type %T want 2 columns go %d columns", dest, len(columns))
	}
	if dest == nil {
		return errors.New("nil map passed to scanStringInt64Map destination")
	}
	for rows.Next() {
		var key sql.NullString
		var value sql.NullInt64
		err = rows.Scan(&key, &value)
		if err != nil {
			return err
		}
		if value.Valid {
			dest[key.String] = value.Int64
		}
	}
	return rows.Err()
}

func scanBasicMap(dialect Dialect, mapper *Mapper, rows rowsi, dest interface{}) error {
	switch mapResult := dest.(type) {
	case map[string]interface{}:
		return scanStringMap(dialect, mapper, rows, mapResult)
	case *map[string]interface{}:
		if dest == nil {
			return errors.New("nil pointer passed to scanBasicMap destination")
		}
		if *mapResult == nil {
			*mapResult = map[string]interface{}{}
		}
		return scanStringMap(dialect, mapper, rows, *mapResult)
	case map[string]int64:
		return scanStringInt64Map(dialect, mapper, rows, mapResult)
	case *map[string]int64:
		if mapResult == nil {
			return errors.New("nil pointer passed to scanBasicMap destination")
		}
		if *mapResult == nil {
			*mapResult = map[string]int64{}
		}
		return scanStringInt64Map(dialect, mapper, rows, *mapResult)
	}

	value := reflect.ValueOf(dest)

	if value.Kind() != reflect.Ptr {
		return errors.New("must pass a pointer, not a value, to StructScan destination")
	}
	if value.IsNil() {
		return errors.New("nil pointer passed to MapScan destination")
	}

	t := reflectx.Deref(value.Type())
	key := t.Key()
	elem := reflectx.Deref(t.Elem())
	isPtr := t.Elem().Kind() == reflect.Ptr

	columns, err := rows.Columns()
	if err != nil {
		return err
	}

	if len(columns) != 2 {
		return fmt.Errorf("dest type %T want 2 columns go %d columns", dest, len(columns))
	}

	direct := reflect.Indirect(value)
	if direct.IsNil() {
		direct.Set(reflect.Indirect(reflect.MakeMap(t)))
	}

	for rows.Next() {
		vkey := reflect.New(key)
		velem := reflect.New(elem)
		err = rows.Scan(vkey.Interface(), velem.Interface())
		if err != nil {
			return err
		}

		// append
		if isPtr {
			direct.SetMapIndex(reflect.Indirect(vkey), velem)
		} else {
			direct.SetMapIndex(reflect.Indirect(vkey), reflect.Indirect(velem))
		}
	}

	return rows.Err()
}
