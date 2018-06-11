package gobatis

import (
	"errors"
	"reflect"
	"sync"

	"github.com/runner-mei/GoBatis/reflectx"
)

type StructInfo struct {
	Inner *reflectx.StructMap

	Index []*FieldInfo
	Paths map[string]*FieldInfo
	Names map[string]*FieldInfo
}

type Mapper struct {
	mapper *reflectx.Mapper
	cache  map[reflect.Type]*StructInfo
	mutex  sync.Mutex
}

// TypeMap returns a mapping of field strings to int slices representing
// the traversal down the struct to reach the field.
func (m *Mapper) TypeMap(t reflect.Type) *StructInfo {
	t = reflectx.Deref(t)

	m.mutex.Lock()
	mapping, ok := m.cache[t]
	if !ok {
		mapping = getMapping(m.mapper, t)
		m.cache[t] = mapping
	}
	m.mutex.Unlock()
	return mapping
}
func getMapping(mapper *reflectx.Mapper, t reflect.Type) *StructInfo {
	mapping := mapper.TypeMap(t)
	info := &StructInfo{
		Inner: mapping,
		Paths: map[string]*FieldInfo{},
		Names: map[string]*FieldInfo{},
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
	RValue func(column string, v reflect.Value) (interface{}, error)
}

func (fi *FieldInfo) makeRValue() func(column string, v reflect.Value) (interface{}, error) {
	if fi == nil {
		return func(column string, v reflect.Value) (interface{}, error) {
			return new(interface{}), nil
		}
	}

	if _, ok := fi.Options["null"]; ok {
		return func(column string, v reflect.Value) (interface{}, error) {
			f := reflectx.FieldByIndexes(v, fi.Index)
			fvalue, err := toGOTypeWith(column, v, f.Addr())
			return &nullScanner{name: fi.Name,
				value: fvalue}, err
		}
	}

	return func(column string, v reflect.Value) (interface{}, error) {
		f := reflectx.FieldByIndexes(v, fi.Index)
		return toGOTypeWith(column, v, f.Addr())
	}
}

func getFeildInfo(field *reflectx.FieldInfo) *FieldInfo {
	info := &FieldInfo{
		FieldInfo: field,
	}
	info.RValue = info.makeRValue()
	return info
}
