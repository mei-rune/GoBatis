package gobatis

import (
	"database/sql"
	"database/sql/driver"
	"errors"
	"fmt"
	"reflect"
	"strings"

	"github.com/runner-mei/GoBatis/reflectx"
)

var _scannerInterface = reflect.TypeOf((*sql.Scanner)(nil)).Elem()
var _valuerInterface = reflect.TypeOf((*driver.Valuer)(nil)).Elem()

// isScannable takes the reflect.Type and the actual dest value and returns
// whether or not it's Scannable.  Something is scannable if:
//   * it is not a struct
//   * it implements sql.Scanner
//   * it has no exported fields
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

// colScanner is an interface used by MapScan and SliceScan
type colScanner interface {
	Columns() ([]string, error)
	Scan(dest ...interface{}) error
	Err() error
}

type rowsi interface {
	Close() error
	Columns() ([]string, error)
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

	fields := traversalsByName(mapper, v.Type(), columns)
	// if we are not unsafe and are missing fields, return an error
	if f, err := missingFields(fields); err != nil && !isUnsafe {
		return fmt.Errorf("missing destination name %s in %T", columns[f], dest)
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

		var keyIndexs []int
		for _, field := range mapper.TypeMap(base).Names {
			if field.Options == nil {
				continue
			}
			if _, ok := field.Options["pk"]; ok {
				keyIndexs = field.Index
				break
			}
		}

		if keyIndexs == nil {
			return fmt.Errorf("field with pk tag isnot exists in %s", base.Name())
		}
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

		fields := traversalsByName(mapper, base, columns)
		// if we are not unsafe and are missing fields, return an error
		if f, err := missingFields(fields); err != nil && !isUnsafe {
			return fmt.Errorf("missing destination name %s in %T", columns[f], dest)
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

func scanMapSlice(dialect Dialect, rows rowsi, dest *[]map[string]interface{}) error {
	columns, err := rows.Columns()
	if err != nil {
		return err
	}

	for rows.Next() {
		values := make([]interface{}, len(columns))
		for i := range values {
			values[i] = new(interface{})
		}

		err = rows.Scan(values...)
		if err != nil {
			return errors.New("Scan into Map(" + strings.Join(columns, ",") + ") error : " + err.Error())
		}

		one := map[string]interface{}{}
		for i, column := range columns {
			one[column] = *(values[i].(*interface{}))
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
func MapScan(dialect Dialect, r colScanner, dest map[string]interface{}) error {
	// ignore r.started, since we needn't use reflect for anything.
	columns, err := r.Columns()
	if err != nil {
		return err
	}

	values := make([]interface{}, len(columns))
	for i := range values {
		values[i] = new(interface{})
	}

	err = r.Scan(values...)
	if err != nil {
		return errors.New("Scan into Map(" + strings.Join(columns, ",") + ") error : " + err.Error())
	}

	for i, column := range columns {
		dest[column] = *(values[i].(*interface{}))
	}

	return r.Err()
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
		fvalue, err := traversal.RValue(dialect, columns[i], v)
		if err != nil {
			return err
		}
		values[i] = fvalue
	}
	return nil
}

func traversalsByName(mapper *Mapper, t reflect.Type, columns []string) []*FieldInfo {
	tm := mapper.TypeMap(reflectx.Deref(t))
	var traversals []*FieldInfo
	for _, column := range columns {
		fi, _ := tm.Names[column]
		traversals = append(traversals, fi)
	}
	return traversals
}

func missingFields(transversals []*FieldInfo) (field int, err error) {
	for i := range transversals {
		if transversals[i] == nil {
			return i, errors.New("missing field")
		}
	}
	return 0, nil
}

func scanBasicMap(dialect Dialect, mapper *Mapper, rows rowsi, dest interface{}) error {
	value := reflect.ValueOf(dest)

	if value.Kind() != reflect.Ptr {
		return errors.New("must pass a pointer, not a value, to StructScan destination")
	}
	if value.IsNil() {
		return errors.New("nil pointer passed to StructScan destination")
	}

	t := reflectx.Deref(value.Type())
	key := t.Key()
	elem := reflectx.Deref(t.Elem())
	isPtr := t.Elem().Kind() == reflect.Ptr

	columns, err := rows.Columns()
	if err != nil {
		return err
	}
	// if it's a base type make sure it only has 1 column;  if not return an error
	if len(columns) != 2 {
		return fmt.Errorf("dest type %T with = 2 columns (%d)", dest, len(columns))
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
