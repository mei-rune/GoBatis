package gobatis

import (
	"errors"
	"reflect"
	"strings"
	"sync"
	"sync/atomic"

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
	cache  atomic.Value
	mutex  sync.Mutex
}

func (m *Mapper) getCache() map[reflect.Type]*StructInfo {
	o := m.cache.Load()
	if o == nil {
		return nil
	}

	c, _ := o.(map[reflect.Type]*StructInfo)
	return c
}

func (m *Mapper) FieldByName(v reflect.Value, name string) reflect.Value {
	tm := m.TypeMap(v.Type())
	fi, ok := tm.Names[name]
	if !ok {
		return v
	}
	return reflectx.FieldByIndexes(v, fi.Index)
}

func (m *Mapper) TraversalsByNameFunc(t reflect.Type, names []string, fn func(int, []int) error) error {
	tm := m.TypeMap(t)
	for i, name := range names {
		fi, ok := tm.Names[name]
		if !ok {
			if err := fn(i, nil); err != nil {
				return err
			}
		} else {
			if err := fn(i, fi.Index); err != nil {
				return err
			}
		}
	}
	return nil
}

// TypeMap returns a mapping of field strings to int slices representing
// the traversal down the struct to reach the field.
func (m *Mapper) TypeMap(t reflect.Type) *StructInfo {
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
		newCache := map[reflect.Type]*StructInfo{}
		for key, value := range cache {
			newCache[key] = value
		}
		cache = newCache
	} else {
		cache = map[reflect.Type]*StructInfo{}
	}

	mapping := getMapping(m.mapper, t)
	cache[t] = mapping
	m.cache.Store(cache)
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

const tagPrefix = "db"

// CreateMapper returns a valid mapper using the configured NameMapper func.
func CreateMapper(prefix string, nameMapper func(string) string, tagMapper func(string) []string) *Mapper {
	if nameMapper == nil {
		nameMapper = strings.ToLower
	}
	if prefix == "" {
		prefix = tagPrefix
	}
	return &Mapper{mapper: reflectx.NewMapperTagFunc(prefix, nameMapper, tagMapper)}
}
