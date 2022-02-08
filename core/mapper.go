package core

import (
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"reflect"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/runner-mei/GoBatis/reflectx"
)

type StructMap struct {
	Inner *reflectx.StructMap

	PrimaryKey [][]int
	Index      []*FieldInfo
	Paths      map[string]*FieldInfo
	Names      map[string]*FieldInfo
	FieldNames map[string]*FieldInfo
}

func (structType *StructMap) FieldByName(name string) *FieldInfo {
	lower := strings.ToLower(name)
	for _, field := range structType.Index {
		if field.Field.Name == name {
			return field
		}

		if field.Name == name {
			return field
		}

		if strings.ToLower(field.Field.Name) == lower {
			return field
		}

		if strings.ToLower(field.Name) == lower {
			return field
		}
	}
	return nil
}

func (s *StructMap) primaryKey() [][]int {
	var keyIndexs [][]int
	for _, field := range s.Index {
		if field.Options == nil {
			continue
		}
		if _, ok := field.Options["pk"]; !ok {
			continue
		}

		if len(field.Index) > 1 && field.Parent != nil && field.Parent.Options != nil {
			if _, ok := field.Parent.Options["-"]; ok {
				continue
			}
			if _, ok := field.Parent.Options["<-"]; ok {
				continue
			}
		}
		keyIndexs = append(keyIndexs, field.Index)
	}

	return keyIndexs
}

type Mapper struct {
	mapper *reflectx.Mapper
	cache  atomic.Value
	mutex  sync.Mutex
}

func (m *Mapper) getCache() map[reflect.Type]*StructMap {
	o := m.cache.Load()
	if o == nil {
		return nil
	}

	c, _ := o.(map[reflect.Type]*StructMap)
	return c
}

// func (m *Mapper) FieldByName(v reflect.Value, name string) reflect.Value {
// 	tm := m.TypeMap(v.Type())
// 	fi, ok := tm.Names[name]
// 	if !ok {
// 		return v
// 	}
// 	return reflectx.FieldByIndexes(v, fi.Index)
// }

// func (m *Mapper) TraversalsByNameFunc(t reflect.Type, names []string, fn func(int, *FieldInfo) error) error {
// 	tm := m.TypeMap(t)
// 	for i, name := range names {
// 		fi, _ := tm.Names[name]
// 		if err := fn(i, fi); err != nil {
// 			return err
// 		}
// 	}
// 	return nil
// }

// TypeMap returns a mapping of field strings to int slices representing
// the traversal down the struct to reach the field.
func (m *Mapper) TypeMap(t reflect.Type) *StructMap {
	t = reflectx.Deref(t)

	var cache = m.getCache()
	if cache != nil {
		mapping, ok := cache[t]
		if ok {
			return mapping
		}
	}

	m.mutex.Lock()
	defer m.mutex.Unlock()
	// double check
	cache = m.getCache()
	if cache != nil {
		mapping, ok := cache[t]
		if ok {
			return mapping
		}
		newCache := map[reflect.Type]*StructMap{}
		for key, value := range cache {
			newCache[key] = value
		}
		cache = newCache
	} else {
		cache = map[reflect.Type]*StructMap{}
	}

	mapping := getMapping(m.mapper, t)
	cache[t] = mapping
	m.cache.Store(cache)
	return mapping
}

func getMapping(mapper *reflectx.Mapper, t reflect.Type) *StructMap {
	mapping := mapper.TypeMap(t)
	info := &StructMap{
		Inner:      mapping,
		Paths:      map[string]*FieldInfo{},
		Names:      map[string]*FieldInfo{},
		FieldNames: map[string]*FieldInfo{},
	}
	for idx := range mapping.Index {
		info.Index = append(info.Index, getFeildInfo(mapping.Index[idx]))
	}

	find := func(field *reflectx.FieldInfo) *FieldInfo {
		for idx := range info.Index {
			if info.Index[idx].FieldInfo == field {
				return info.Index[idx]
			}
		}
		panic(errors.New("field '" + field.Name + "' isnot found"))
	}

	info.PrimaryKey = info.primaryKey()

	for key, field := range mapping.FieldNames {
		info.FieldNames[key] = find(field)
	}
	for key, field := range mapping.Paths {
		info.Paths[key] = find(field)
	}
	for key, field := range mapping.Names {
		info.Names[key] = find(field)
	}
	return info
}

type FieldInfo struct {
	*reflectx.FieldInfo
	RValue func(dialect Dialect, param *Param, v reflect.Value) (interface{}, error)
	LValue func(dialect Dialect, column string, v reflect.Value) (interface{}, error)
}

func (fi *FieldInfo) makeRValue() func(dialect Dialect, param *Param, v reflect.Value) (interface{}, error) {
	if fi == nil {
		return func(dialect Dialect, param *Param, v reflect.Value) (interface{}, error) {
			return nil, nil
		}
	}
	typ := fi.Field.Type
	kind := typ.Kind()
	isPtr := false
	if typ.Kind() == reflect.Ptr {
		kind = typ.Elem().Kind()
		isPtr = true
	}

	switch kind {
	case reflect.Slice:
		typeElem := typ
		if typeElem.Kind() == reflect.Ptr {
			typeElem = typeElem.Elem()
		}
		if typeElem.PkgPath() != "" {
			break
		}
		if typeElem.Elem().Kind() == reflect.Struct {
			break
		}

		if typeElem == _bytesType {
			if _, ok := fi.Options["notnull"]; ok {
				return func(dialect Dialect, param *Param, v reflect.Value) (interface{}, error) {
					field := reflectx.FieldByIndexesReadOnly(v, fi.Index)
					if field.IsNil() {
						return nil, errors.New("field '" + fi.Field.Name + "' is zero value")
					}
					if isPtr {
						if field.Elem().Len() == 0 {
							return nil, errors.New("field '" + fi.Field.Name + "' is empty value")
						}
					} else {
						if field.Len() == 0 {
							return nil, errors.New("field '" + fi.Field.Name + "' is empty value")
						}
					}
					return field.Interface(), nil
				}
			}

			if _, ok := fi.Options["null"]; ok {
				return func(dialect Dialect, param *Param, v reflect.Value) (interface{}, error) {
					field := reflectx.FieldByIndexesReadOnly(v, fi.Index)
					if field.IsNil() {
						return nil, nil
					}
					if isPtr {
						if field.Elem().Len() == 0 {
							return nil, nil
						}
					} else {
						if field.Len() == 0 {
							return nil, nil
						}
					}
					return field.Interface(), nil
				}
			}
			return func(dialect Dialect, param *Param, v reflect.Value) (interface{}, error) {
				field := reflectx.FieldByIndexesReadOnly(v, fi.Index)
				return field.Interface(), nil
			}
		}

		if _, ok := fi.Options["json"]; ok {
			break
		}

		if _, ok := fi.Options["jsonb"]; ok {
			break
		}

		if _, ok := fi.Options["notnull"]; ok {
			return func(dialect Dialect, param *Param, v reflect.Value) (interface{}, error) {
				field := reflectx.FieldByIndexesReadOnly(v, fi.Index)
				if field.IsNil() {
					return nil, fmt.Errorf("field '%s' is nil", fi.Field.Name)
				}

				if isPtr {
					if field.Elem().Len() == 0 {
						return nil, errors.New("field '" + fi.Field.Name + "' is empty value")
					}
				} else {
					if field.Len() == 0 {
						return nil, errors.New("field '" + fi.Field.Name + "' is empty value")
					}
				}

				fvalue := field.Interface()
				if fvalue == nil {
					return nil, fmt.Errorf("field '%s' is nil", fi.Field.Name)
				}
				value, err := dialect.MakeArrayValuer(fvalue)
				if err != nil {
					return nil, fmt.Errorf("field '%s' to array valuer, %s", fi.Field.Name, err)
				}
				return value, nil
			}
		}

		if _, ok := fi.Options["null"]; ok {
			return func(dialect Dialect, param *Param, v reflect.Value) (interface{}, error) {
				field := reflectx.FieldByIndexesReadOnly(v, fi.Index)
				if field.IsNil() || !field.IsValid() {
					return nil, nil
				}

				if isPtr {
					if field.Elem().Len() == 0 {
						return nil, nil
					}
				} else {
					if field.Len() == 0 {
						return nil, nil
					}
				}

				fvalue := field.Interface()
				if fvalue == nil {
					return nil, nil
				}

				value, err := dialect.MakeArrayValuer(fvalue)
				if err != nil {
					return nil, fmt.Errorf("field '%s' to array valuer, %s", fi.Field.Name, err)
				}
				return value, nil
			}
		}

		return func(dialect Dialect, param *Param, v reflect.Value) (interface{}, error) {
			field := reflectx.FieldByIndexesReadOnly(v, fi.Index)
			if field.IsNil() {
				return nil, nil
			}
			fvalue := field.Interface()
			if fvalue == nil {
				return nil, nil
			}
			value, err := dialect.MakeArrayValuer(fvalue)
			if err != nil {
				return nil, fmt.Errorf("field '%s' to array valuer, %s", fi.Field.Name, err)
			}
			return value, nil
		}
	case reflect.Bool:
		if _, ok := fi.Options["notnull"]; ok && isPtr {
			return func(dialect Dialect, param *Param, v reflect.Value) (interface{}, error) {
				field := reflectx.FieldByIndexesReadOnly(v, fi.Index)
				if field.IsNil() {
					return nil, errors.New("field '" + fi.Field.Name + "' is zero value")
				}
				return field.Interface(), nil
			}
		}

		return func(dialect Dialect, param *Param, v reflect.Value) (interface{}, error) {
			field := reflectx.FieldByIndexesReadOnly(v, fi.Index)
			return field.Interface(), nil
		}
	case reflect.Int,
		reflect.Int8,
		reflect.Int16,
		reflect.Int32,
		reflect.Int64:
		if _, ok := fi.Options["notnull"]; ok {
			if isPtr {
				return func(dialect Dialect, param *Param, v reflect.Value) (interface{}, error) {
					field := reflectx.FieldByIndexesReadOnly(v, fi.Index)
					if field.IsNil() {
						return nil, errors.New("field '" + fi.Field.Name + "' is zero value")
					}
					if field.Elem().Int() == 0 {
						return nil, errors.New("field '" + fi.Field.Name + "' is zero value")
					}
					return field.Interface(), nil
				}
			}

			return func(dialect Dialect, param *Param, v reflect.Value) (interface{}, error) {
				field := reflectx.FieldByIndexesReadOnly(v, fi.Index)
				if field.Int() == 0 {
					return nil, errors.New("field '" + fi.Field.Name + "' is zero value")
				}
				return field.Interface(), nil
			}
		}
		if _, ok := fi.Options["null"]; ok {
			if isPtr {
				return func(dialect Dialect, param *Param, v reflect.Value) (interface{}, error) {
					field := reflectx.FieldByIndexesReadOnly(v, fi.Index)
					if field.IsNil() || !field.IsValid() {
						return nil, nil
					}

					if field.Elem().Int() == 0 {
						return nil, nil
					}
					return field.Interface(), nil
				}
			}

			return func(dialect Dialect, param *Param, v reflect.Value) (interface{}, error) {
				field := reflectx.FieldByIndexesReadOnly(v, fi.Index)
				if field.Int() == 0 {
					return nil, nil
				}
				return field.Interface(), nil
			}
		}
		return func(dialect Dialect, param *Param, v reflect.Value) (interface{}, error) {
			field := reflectx.FieldByIndexesReadOnly(v, fi.Index)
			return field.Interface(), nil
		}
	case reflect.Uint,
		reflect.Uint8,
		reflect.Uint16,
		reflect.Uint32,
		reflect.Uint64,
		reflect.Uintptr:
		if _, ok := fi.Options["notnull"]; ok {
			if isPtr {
				return func(dialect Dialect, param *Param, v reflect.Value) (interface{}, error) {
					field := reflectx.FieldByIndexesReadOnly(v, fi.Index)
					if field.IsNil() {
						return nil, errors.New("field '" + fi.Field.Name + "' is zero value")
					}
					if field.Elem().Uint() == 0 {
						return nil, errors.New("field '" + fi.Field.Name + "' is zero value")
					}
					return field.Interface(), nil
				}
			}

			return func(dialect Dialect, param *Param, v reflect.Value) (interface{}, error) {
				field := reflectx.FieldByIndexesReadOnly(v, fi.Index)
				if field.Uint() == 0 {
					return nil, errors.New("field '" + fi.Field.Name + "' is zero value")
				}
				return field.Interface(), nil
			}
		}

		if _, ok := fi.Options["null"]; ok {
			if isPtr {
				return func(dialect Dialect, param *Param, v reflect.Value) (interface{}, error) {
					field := reflectx.FieldByIndexesReadOnly(v, fi.Index)
					if field.IsNil() {
						return nil, nil
					}
					if field.Elem().Uint() == 0 {
						return nil, nil
					}
					return field.Interface(), nil
				}
			}

			return func(dialect Dialect, param *Param, v reflect.Value) (interface{}, error) {
				field := reflectx.FieldByIndexesReadOnly(v, fi.Index)
				if field.Uint() == 0 {
					return nil, nil
				}
				return field.Interface(), nil
			}
		}
		return func(dialect Dialect, param *Param, v reflect.Value) (interface{}, error) {
			field := reflectx.FieldByIndexesReadOnly(v, fi.Index)
			return field.Interface(), nil
		}
	case reflect.Float32,
		reflect.Float64:
		if _, ok := fi.Options["notnull"]; ok {
			if isPtr {
				return func(dialect Dialect, param *Param, v reflect.Value) (interface{}, error) {
					field := reflectx.FieldByIndexesReadOnly(v, fi.Index)
					if field.IsNil() {
						return nil, errors.New("field '" + fi.Field.Name + "' is zero value")
					}
					if field.Elem().Float() == 0 {
						return nil, errors.New("field '" + fi.Field.Name + "' is zero value")
					}
					return field.Interface(), nil
				}
			}

			return func(dialect Dialect, param *Param, v reflect.Value) (interface{}, error) {
				field := reflectx.FieldByIndexesReadOnly(v, fi.Index)
				if field.Float() == 0 {
					return nil, errors.New("field '" + fi.Field.Name + "' is zero value")
				}
				return field.Interface(), nil
			}
		}

		if _, ok := fi.Options["null"]; ok {
			if isPtr {
				return func(dialect Dialect, param *Param, v reflect.Value) (interface{}, error) {
					field := reflectx.FieldByIndexesReadOnly(v, fi.Index)
					if field.IsNil() {
						return nil, nil
					}
					if field.Elem().Float() == 0 {
						return nil, nil
					}
					return field.Interface(), nil
				}
			}

			return func(dialect Dialect, param *Param, v reflect.Value) (interface{}, error) {
				field := reflectx.FieldByIndexesReadOnly(v, fi.Index)
				if field.Float() == 0 {
					return nil, nil
				}
				return field.Interface(), nil
			}
		}
		return func(dialect Dialect, param *Param, v reflect.Value) (interface{}, error) {
			field := reflectx.FieldByIndexesReadOnly(v, fi.Index)
			return field.Interface(), nil
		}

	// case reflect.Complex64,
	// reflect.Complex128:
	case reflect.String:
		if _, ok := fi.Options["notnull"]; ok {
			if isPtr {
				return func(dialect Dialect, param *Param, v reflect.Value) (interface{}, error) {
					field := reflectx.FieldByIndexesReadOnly(v, fi.Index)
					if field.IsNil() {
						return nil, errors.New("field '" + fi.Field.Name + "' is zero value")
					}
					if field.Elem().Len() == 0 {
						return nil, errors.New("field '" + fi.Field.Name + "' is zero value")
					}
					return field.Interface(), nil
				}
			}

			return func(dialect Dialect, param *Param, v reflect.Value) (interface{}, error) {
				field := reflectx.FieldByIndexesReadOnly(v, fi.Index)
				if field.Len() == 0 {
					return nil, errors.New("field '" + fi.Field.Name + "' is zero value")
				}
				return field.Interface(), nil
			}
		}

		if _, ok := fi.Options["null"]; ok {
			if isPtr {
				return func(dialect Dialect, param *Param, v reflect.Value) (interface{}, error) {
					field := reflectx.FieldByIndexesReadOnly(v, fi.Index)
					if field.IsNil() {
						return nil, nil
					}
					if field.Elem().Len() == 0 {
						return nil, nil
					}
					return field.Interface(), nil
				}
			}

			return func(dialect Dialect, param *Param, v reflect.Value) (interface{}, error) {
				field := reflectx.FieldByIndexesReadOnly(v, fi.Index)
				if field.Len() == 0 {
					return nil, nil
				}
				return field.Interface(), nil
			}
		}

		return func(dialect Dialect, param *Param, v reflect.Value) (interface{}, error) {
			field := reflectx.FieldByIndexesReadOnly(v, fi.Index)
			return field.Interface(), nil
		}
	case reflect.Chan, reflect.Func, reflect.UnsafePointer:
		return func(dialect Dialect, param *Param, v reflect.Value) (interface{}, error) {
			field := reflectx.FieldByIndexesReadOnly(v, fi.Index)
			return nil, fmt.Errorf("field '%s' isnot a sql type got %T", fi.Field.Name, field.Interface())
		}
	}
	if typ.Implements(_valuerInterface) {
		if isPtr {
			return func(dialect Dialect, param *Param, v reflect.Value) (interface{}, error) {
				field := reflectx.FieldByIndexesReadOnly(v, fi.Index)
				if field.IsNil() {
					return nil, nil
				}
				return field.Interface(), nil
			}
		}

		return func(dialect Dialect, param *Param, v reflect.Value) (interface{}, error) {
			field := reflectx.FieldByIndexesReadOnly(v, fi.Index)
			return field.Interface(), nil
		}
	}
	if reflect.PtrTo(typ).Implements(_valuerInterface) {
		if isPtr {
			return func(dialect Dialect, param *Param, v reflect.Value) (interface{}, error) {
				field := reflectx.FieldByIndexesReadOnly(v, fi.Index)
				if field.IsNil() {
					return nil, nil
				}
				return field.Addr().Interface(), nil
			}
		}

		return func(dialect Dialect, param *Param, v reflect.Value) (interface{}, error) {
			field := reflectx.FieldByIndexesReadOnly(v, fi.Index)
			return field.Addr().Interface(), nil
		}
	}

	switch typ {
	case _timeType:
		if _, ok := fi.Options["notnull"]; ok {
			return func(dialect Dialect, param *Param, v reflect.Value) (interface{}, error) {
				field := reflectx.FieldByIndexesReadOnly(v, fi.Index)
				fvalue := field.Interface()
				if fvalue == nil {
					return nil, errors.New("field '" + fi.Field.Name + "' is zero value")
				}
				if t, _ := fvalue.(time.Time); t.IsZero() {
					return nil, errors.New("field '" + fi.Field.Name + "' is zero value")
				}
				return fvalue, nil
			}
		}
		if _, ok := fi.Options["null"]; ok {
			return func(dialect Dialect, param *Param, v reflect.Value) (interface{}, error) {
				field := reflectx.FieldByIndexesReadOnly(v, fi.Index)
				fvalue := field.Interface()
				if t, _ := fvalue.(time.Time); !t.IsZero() {
					return fvalue, nil
				}
				return nil, nil
			}
		}
		return func(dialect Dialect, param *Param, v reflect.Value) (interface{}, error) {
			field := reflectx.FieldByIndexesReadOnly(v, fi.Index)
			return field.Interface(), nil
		}
	case _timePtr:
		if _, ok := fi.Options["notnull"]; ok {
			return func(dialect Dialect, param *Param, v reflect.Value) (interface{}, error) {
				field := reflectx.FieldByIndexesReadOnly(v, fi.Index)
				if field.IsNil() {
					return nil, errors.New("field '" + fi.Field.Name + "' is zero value")
				}
				fvalue := field.Interface()
				if fvalue == nil {
					return nil, errors.New("field '" + fi.Field.Name + "' is zero value")
				}
				if t, _ := fvalue.(*time.Time); t == nil || t.IsZero() {
					return nil, errors.New("field '" + fi.Field.Name + "' is zero value")
				}
				return fvalue, nil
			}
		}

		if _, ok := fi.Options["null"]; ok {
			return func(dialect Dialect, param *Param, v reflect.Value) (interface{}, error) {
				field := reflectx.FieldByIndexesReadOnly(v, fi.Index)
				if field.IsNil() {
					return nil, nil
				}
				fvalue := field.Interface()
				if fvalue == nil {
					return nil, nil
				}
				if t, _ := fvalue.(*time.Time); t == nil || t.IsZero() {
					return nil, nil
				}
				return fvalue, nil
			}
		}
		return func(dialect Dialect, param *Param, v reflect.Value) (interface{}, error) {
			field := reflectx.FieldByIndexesReadOnly(v, fi.Index)
			if field.IsNil() {
				return nil, nil
			}
			fvalue := field.Interface()
			if fvalue == nil {
				return nil, nil
			}
			if t, _ := fvalue.(*time.Time); t == nil {
				return nil, nil
			}
			return fvalue, nil
		}
	case _ipType:

		if _, ok := fi.Options["notnull"]; ok {
			return func(dialect Dialect, param *Param, v reflect.Value) (interface{}, error) {
				field := reflectx.FieldByIndexesReadOnly(v, fi.Index)
				if field.IsNil() {
					return nil, errors.New("field '" + fi.Field.Name + "' is zero value")
				}
				fvalue := field.Interface()
				if fvalue == nil {
					return nil, errors.New("field '" + fi.Field.Name + "' is zero value")
				}
				ip, _ := fvalue.(net.IP)
				if len(ip) == 0 || ip.IsUnspecified() {
					return nil, errors.New("field '" + fi.Field.Name + "' is zero value")
				}
				return ip.String(), nil
			}
		}

		if _, ok := fi.Options["null"]; ok {
			return func(dialect Dialect, param *Param, v reflect.Value) (interface{}, error) {
				field := reflectx.FieldByIndexesReadOnly(v, fi.Index)
				if field.IsNil() {
					return nil, nil
				}
				fvalue := field.Interface()
				if fvalue == nil {
					return nil, nil
				}

				ip, _ := fvalue.(net.IP)
				if len(ip) == 0 || ip.IsUnspecified() {
					return nil, nil
				}
				return ip.String(), nil
			}
		}
		return func(dialect Dialect, param *Param, v reflect.Value) (interface{}, error) {
			field := reflectx.FieldByIndexesReadOnly(v, fi.Index)
			fvalue := field.Interface()
			if ip, _ := fvalue.(net.IP); len(ip) != 0 && !ip.IsUnspecified() {
				return ip.String(), nil
			}
			return nil, nil
		}
	case _ipPtr:
		if _, ok := fi.Options["notnull"]; ok {
			return func(dialect Dialect, param *Param, v reflect.Value) (interface{}, error) {
				field := reflectx.FieldByIndexesReadOnly(v, fi.Index)
				if field.IsNil() {
					return nil, errors.New("field '" + fi.Field.Name + "' is zero value")
				}
				fvalue := field.Interface()
				if fvalue == nil {
					return nil, errors.New("field '" + fi.Field.Name + "' is zero value")
				}

				ip, _ := fvalue.(*net.IP)
				if ip == nil || len(*ip) == 0 || ip.IsUnspecified() {
					return nil, errors.New("field '" + fi.Field.Name + "' is zero value")
				}
				return ip.String(), nil
			}
		}

		return func(dialect Dialect, param *Param, v reflect.Value) (interface{}, error) {
			field := reflectx.FieldByIndexesReadOnly(v, fi.Index)
			if field.IsNil() {
				return nil, nil
			}
			fvalue := field.Interface()
			if fvalue == nil {
				return nil, nil
			}

			ip, _ := fvalue.(*net.IP)
			if ip == nil || len(*ip) == 0 || ip.IsUnspecified() {
				return nil, nil
			}
			return ip.String(), nil
		}
	case _macType:
		if _, ok := fi.Options["notnull"]; ok {
			return func(dialect Dialect, param *Param, v reflect.Value) (interface{}, error) {
				field := reflectx.FieldByIndexesReadOnly(v, fi.Index)
				fvalue := field.Interface()
				if fvalue == nil {
					return nil, errors.New("field '" + fi.Field.Name + "' is zero value")
				}

				hwAddr, _ := fvalue.(net.HardwareAddr)
				if len(hwAddr) == 0 || isZeroHwAddr(hwAddr) {
					return nil, errors.New("field '" + fi.Field.Name + "' is zero value")
				}
				return hwAddr.String(), nil
			}
		}

		if _, ok := fi.Options["null"]; ok {
			return func(dialect Dialect, param *Param, v reflect.Value) (interface{}, error) {
				field := reflectx.FieldByIndexesReadOnly(v, fi.Index)
				if field.IsNil() {
					return nil, nil
				}
				fvalue := field.Interface()
				if fvalue == nil {
					return nil, nil
				}

				hwAddr, _ := fvalue.(net.HardwareAddr)
				if len(hwAddr) == 0 || isZeroHwAddr(hwAddr) {
					return nil, nil
				}
				return hwAddr.String(), nil
			}
		}

		return func(dialect Dialect, param *Param, v reflect.Value) (interface{}, error) {
			field := reflectx.FieldByIndexesReadOnly(v, fi.Index)
			fvalue := field.Interface()
			if fvalue == nil {
				return nil, nil
			}

			hwAddr, _ := fvalue.(net.HardwareAddr)
			if len(hwAddr) == 0 || isZeroHwAddr(hwAddr) {
				return nil, nil
			}
			return hwAddr.String(), nil
		}
	case _macPtr:
		if _, ok := fi.Options["notnull"]; ok {
			return func(dialect Dialect, param *Param, v reflect.Value) (interface{}, error) {
				field := reflectx.FieldByIndexesReadOnly(v, fi.Index)
				if field.IsNil() {
					return nil, errors.New("field '" + fi.Field.Name + "' is zero value")
				}
				fvalue := field.Interface()
				if fvalue == nil {
					return nil, errors.New("field '" + fi.Field.Name + "' is zero value")
				}

				hwAddr, _ := fvalue.(*net.HardwareAddr)
				if hwAddr == nil || len(*hwAddr) == 0 || isZeroHwAddr(*hwAddr) {
					return nil, errors.New("field '" + fi.Field.Name + "' is zero value")
				}
				return hwAddr.String(), nil
			}
		}

		return func(dialect Dialect, param *Param, v reflect.Value) (interface{}, error) {
			field := reflectx.FieldByIndexesReadOnly(v, fi.Index)
			if field.IsNil() {
				return nil, nil
			}
			fvalue := field.Interface()
			if fvalue == nil {
				return nil, nil
			}
			hwAddr, _ := fvalue.(*net.HardwareAddr)
			if hwAddr == nil || len(*hwAddr) == 0 || isZeroHwAddr(*hwAddr) {
				return nil, nil
			}
			return hwAddr.String(), nil
		}
	}

	canNil := isPtr
	if kind == reflect.Map || kind == reflect.Slice {
		canNil = true
	}

	if _, ok := fi.Options["notnull"]; ok {
		return func(dialect Dialect, param *Param, v reflect.Value) (interface{}, error) {
			field := reflectx.FieldByIndexesReadOnly(v, fi.Index)
			if canNil {
				if field.IsNil() {
					return nil, fmt.Errorf("field '%s' is nil", fi.Field.Name)
				}
			}
			fvalue := field.Interface()
			if fvalue == nil {
				return nil, fmt.Errorf("field '%s' is nil", fi.Field.Name)
			}
			bs, err := json.Marshal(fvalue)
			if err != nil {
				return nil, fmt.Errorf("field '%s' convert to json, %s", fi.Field.Name, err)
			}
			return string(bs), nil
		}
	}

	return func(dialect Dialect, param *Param, v reflect.Value) (interface{}, error) {
		field := reflectx.FieldByIndexesReadOnly(v, fi.Index)
		if canNil {
			if field.IsNil() {
				return nil, nil
			}
		}
		fvalue := field.Interface()
		if fvalue == nil {
			return nil, nil
		}
		bs, err := json.Marshal(fvalue)
		if err != nil {
			return nil, fmt.Errorf("field '%s' convert to json, %s", fi.Field.Name, err)
		}
		return string(bs), nil
	}
}

func isZeroHwAddr(hwAddr net.HardwareAddr) bool {
	for i := range hwAddr {
		if hwAddr[i] != 0 {
			return false
		}
	}
	return true
}

func (fi *FieldInfo) makeLValue() func(dialect Dialect, column string, v reflect.Value) (interface{}, error) {
	if fi == nil {
		return func(dialect Dialect, column string, v reflect.Value) (interface{}, error) {
			return emptyScan, nil
		}
	}
	typ := fi.Field.Type
	kind := typ.Kind()
	if kind == reflect.Ptr {
		kind = typ.Elem().Kind()
		if kind == reflect.Ptr {
			kind = typ.Elem().Elem().Kind()
		}
	}

	switch kind {
	case reflect.Slice:
		typeElem := typ
		if typeElem.Kind() == reflect.Ptr {
			typeElem = typeElem.Elem()
		}

		if typeElem.PkgPath() != "" {
			break
		}

		if typeElem == _bytesType {
			_, blobExists := fi.Options["blob"]
			if blobExists {
				return func(dialect Dialect, column string, v reflect.Value) (interface{}, error) {
					field := reflectx.FieldByIndexes(v, fi.Index)
					addr := field.Addr().Interface()
					if dialect.BlobSupported() {
						return dialect.NewBlob(addr.(*[]byte)), nil
					}
					return addr, nil
				}
			}

			return func(dialect Dialect, column string, v reflect.Value) (interface{}, error) {
				field := reflectx.FieldByIndexes(v, fi.Index)
				return field.Addr().Interface(), nil
			}
		}

		_, jsonExists := fi.Options["json"]
		if !jsonExists {
			_, jsonExists = fi.Options["jsonb"]
		}
		if !jsonExists {
			if typeElem.Elem().Kind() == reflect.Struct {
				jsonExists = true
			}
		}

		if jsonExists {
			_, blobExists := fi.Options["blob"]
			if blobExists {
				return func(dialect Dialect, column string, v reflect.Value) (interface{}, error) {
					field := reflectx.FieldByIndexes(v, fi.Index)
					return &scanner{name: column, value: field.Addr().Interface(), blob: dialect.NewBlob(nil)}, nil
				}
			}

			return func(dialect Dialect, column string, v reflect.Value) (interface{}, error) {
				field := reflectx.FieldByIndexes(v, fi.Index)
				return &scanner{name: column, value: field.Addr().Interface()}, nil
			}
		}
		return func(dialect Dialect, column string, v reflect.Value) (interface{}, error) {
			field := reflectx.FieldByIndexes(v, fi.Index)
			return dialect.MakeArrayScanner(fi.Name, field.Addr().Interface())
		}
	case reflect.Bool,
		reflect.Int,
		reflect.Int8,
		reflect.Int16,
		reflect.Int32,
		reflect.Int64,
		reflect.Uint,
		reflect.Uint8,
		reflect.Uint16,
		reflect.Uint32,
		reflect.Uint64,
		reflect.Uintptr,
		reflect.Float32,
		reflect.Float64,
		reflect.Complex64,
		reflect.Complex128:
		if _, ok := fi.Options["null"]; ok {
			return func(dialect Dialect, column string, v reflect.Value) (interface{}, error) {
				field := reflectx.FieldByIndexes(v, fi.Index)
				fvalue := &Nullable{Name: fi.Name, Value: field.Addr().Interface()}
				return fvalue, nil
			}
		}
		return func(dialect Dialect, column string, v reflect.Value) (interface{}, error) {
			field := reflectx.FieldByIndexes(v, fi.Index)
			return field.Addr().Interface(), nil
		}

	case reflect.String:
		_, clobExists := fi.Options["clob"]
		if clobExists {


			if _, ok := fi.Options["null"]; ok {
				return func(dialect Dialect, column string, v reflect.Value) (interface{}, error) {
					field := reflectx.FieldByIndexes(v, fi.Index)
					addr := field.Addr().Interface()

					if dialect.ClobSupported() {
						return dialect.NewClob(addr.(*string)), nil
					}
					fvalue := &Nullable{Name: fi.Name, Value: addr}
					return fvalue, nil
				}
			}

			return func(dialect Dialect, column string, v reflect.Value) (interface{}, error) {
				field := reflectx.FieldByIndexes(v, fi.Index)
				addr := field.Addr().Interface()

				if dialect.ClobSupported() {
					return dialect.NewClob(addr.(*string)), nil
				}
				return addr, nil
			}
		}

		if _, ok := fi.Options["null"]; ok {
			return func(dialect Dialect, column string, v reflect.Value) (interface{}, error) {
				field := reflectx.FieldByIndexes(v, fi.Index)
				fvalue := &Nullable{Name: fi.Name, Value: field.Addr().Interface()}
				return fvalue, nil
			}
		}

		return func(dialect Dialect, column string, v reflect.Value) (interface{}, error) {
			field := reflectx.FieldByIndexes(v, fi.Index)
			return field.Addr().Interface(), nil
		}
	case reflect.Chan, reflect.Func, reflect.UnsafePointer:
		return func(dialect Dialect, column string, v reflect.Value) (interface{}, error) {
			return nil, fmt.Errorf("cannot convert column '%s' into '%T' fail, error type %T", column, v.Interface(), typ.Name())
		}
	}

	if reflect.PtrTo(typ).Implements(_scannerInterface) {
		return func(dialect Dialect, column string, v reflect.Value) (interface{}, error) {
			field := reflectx.FieldByIndexes(v, fi.Index)
			return field.Addr().Interface(), nil
		}
	}

	switch typ {
	case _timeType:
		if _, ok := fi.Options["null"]; ok {
			return func(dialect Dialect, column string, v reflect.Value) (interface{}, error) {
				field := reflectx.FieldByIndexes(v, fi.Index)
				fvalue := &Nullable{Name: fi.Name, Value: field.Addr().Interface()}
				return fvalue, nil
			}
		}
		return func(dialect Dialect, column string, v reflect.Value) (interface{}, error) {
			field := reflectx.FieldByIndexes(v, fi.Index)
			return field.Addr().Interface(), nil
		}
	case _timePtr:
		return func(dialect Dialect, column string, v reflect.Value) (interface{}, error) {
			field := reflectx.FieldByIndexes(v, fi.Index)
			return field.Addr().Interface(), nil
		}
	case _ipType:
		return func(dialect Dialect, column string, v reflect.Value) (interface{}, error) {
			field := reflectx.FieldByIndexes(v, fi.Index)
			return &sScanner{name: column, field: field, scanFunc: scanIP}, nil
		}
	case _ipPtr:
		return func(dialect Dialect, column string, v reflect.Value) (interface{}, error) {
			field := reflectx.FieldByIndexes(v, fi.Index)
			return &sScanner{name: column, field: field, scanFunc: scanIP}, nil
		}
	case _macType:
		return func(dialect Dialect, column string, v reflect.Value) (interface{}, error) {
			field := reflectx.FieldByIndexes(v, fi.Index)
			return &sScanner{name: column, field: field, scanFunc: scanMAC}, nil
		}
	case _macPtr:
		return func(dialect Dialect, column string, v reflect.Value) (interface{}, error) {
			field := reflectx.FieldByIndexes(v, fi.Index)
			return &sScanner{name: column, field: field, scanFunc: scanMAC}, nil
		}
	}


	_, blobExists := fi.Options["blob"]
	if blobExists {
		return func(dialect Dialect, column string, v reflect.Value) (interface{}, error) {
			field := reflectx.FieldByIndexes(v, fi.Index)
			return &scanner{name: column, value: field.Addr().Interface(), blob: dialect.NewBlob(nil)}, nil
		}
	}

	return func(dialect Dialect, column string, v reflect.Value) (interface{}, error) {
		field := reflectx.FieldByIndexes(v, fi.Index)
		return &scanner{name: column, value: field.Addr().Interface()}, nil
	}
}

var _bytesType = reflect.TypeOf((*[]byte)(nil)).Elem()
var _timeType = reflect.TypeOf((*time.Time)(nil)).Elem()
var _timePtr = reflect.TypeOf((*time.Time)(nil))
var _timePtrPtr = reflect.TypeOf((**time.Time)(nil))

var _ipType = reflect.TypeOf((*net.IP)(nil)).Elem()
var _ipPtr = reflect.TypeOf((*net.IP)(nil))
var _ipPtrPtr = reflect.TypeOf((**net.IP)(nil))

var _macType = reflect.TypeOf((*net.HardwareAddr)(nil)).Elem()
var _macPtr = reflect.TypeOf((*net.HardwareAddr)(nil))
var _macPtrPtr = reflect.TypeOf((**net.HardwareAddr)(nil))

func getFeildInfo(field *reflectx.FieldInfo) *FieldInfo {
	info := &FieldInfo{
		FieldInfo: field,
	}
	info.LValue = info.makeLValue()
	info.RValue = info.makeRValue()
	return info
}

const tagPrefix = "db"

// CreateMapper returns a valid mapper using the configured NameMapper func.
func CreateMapper(prefix string, nameMapper func(string) string, tagMapper func(string, string) []string) *Mapper {
	if nameMapper == nil {
		nameMapper = strings.ToLower
	}
	if prefix == "" {
		prefix = tagPrefix
	}
	return &Mapper{mapper: reflectx.NewMapperTagFunc(prefix, nameMapper, tagMapper)}
}

var xormkeyTags = map[string]struct{}{
	"pk":         struct{}{},
	"autoincr":   struct{}{},
	"null":       struct{}{},
	"notnull":    struct{}{},
	"unique":     struct{}{},
	"extends":    struct{}{},
	"index":      struct{}{},
	"<-":         struct{}{},
	"->":         struct{}{},
	"created":    struct{}{},
	"updated":    struct{}{},
	"deleted":    struct{}{},
	"version":    struct{}{},
	"default":    struct{}{},
	"json":       struct{}{},
	"bit":        struct{}{},
	"tinyint":    struct{}{},
	"smallint":   struct{}{},
	"mediumint":  struct{}{},
	"int":        struct{}{},
	"integer":    struct{}{},
	"bigint":     struct{}{},
	"char":       struct{}{},
	"varchar":    struct{}{},
	"tinytext":   struct{}{},
	"text":       struct{}{},
	"mediumtext": struct{}{},
	"longtext":   struct{}{},
	"binary":     struct{}{},
	"varbinary":  struct{}{},
	"date":       struct{}{},
	"datetime":   struct{}{},
	"time":       struct{}{},
	"timestamp":  struct{}{},
	"timestampz": struct{}{},
	"real":       struct{}{},
	"float":      struct{}{},
	"double":     struct{}{},
	"decimal":    struct{}{},
	"numeric":    struct{}{},
	"tinyblob":   struct{}{},
	"blob":       struct{}{},
	"mediumblob": struct{}{},
	"longblob":   struct{}{},
	"bytea":      struct{}{},
	"bool":       struct{}{},
	"serial":     struct{}{},
	"bigserial":  struct{}{},
}

func TagSplitForXORM(s string, fieldName string) []string {
	parts := strings.Fields(s)
	if len(parts) == 0 {
		return parts
	}
	name := parts[0]
	idx := strings.IndexByte(name, '(')
	if idx >= 0 {
		name = name[:idx]
	}

	if _, ok := xormkeyTags[name]; !ok {
		return parts
	}

	for i := 1; i < len(parts); i++ {
		name := parts[i]

		idx := strings.IndexByte(name, '(')
		if idx >= 0 {
			name = name[:idx]
		}

		if _, ok := xormkeyTags[name]; !ok {
			tmp := parts[i]
			parts[i] = parts[0]
			parts[0] = tmp
			return parts
		}
	}

	clone := make([]string, len(parts)+1)
	clone[0] = fieldName
	copy(clone[1:], parts)
	return clone
}

var emptyField = &FieldInfo{
	FieldInfo: &reflectx.FieldInfo{},
	RValue: func(dialect Dialect, param *Param, v reflect.Value) (interface{}, error) {
		return nil, errors.New("emptyField is unsupported")
	},
	LValue: func(dialect Dialect, column string, v reflect.Value) (interface{}, error) {
		return emptyScan, nil
	},
}

func ReadTableFields(mapper *Mapper, instance reflect.Type) ([]string, error) {
	st := mapper.TypeMap(instance)
	columns := make([]string, 0, len(st.Index))
	for _, field := range st.Index {
		columns = append(columns, field.Name)
	}
	return columns, nil
}
