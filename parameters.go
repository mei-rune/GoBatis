package gobatis

import (
	"errors"
	"reflect"
	"strings"

	"upper.io/db.v3/lib/reflectx"
)

var ErrNotFound = errors.New("not found")

type Parameters interface {
	Get(name string) (interface{}, error)
	RValue(dialect Dialect, param *Param) (interface{}, error)
}

type emptyFinder struct{}

func (*emptyFinder) Get(name string) (interface{}, error) {
	return nil, ErrNotFound
}

func (*emptyFinder) RValue(dialect Dialect, param *Param) (interface{}, error) {
	return nil, ErrNotFound
}

var emptyParameters = &emptyFinder{}

type singleFinder struct {
	value interface{}
}

func (s singleFinder) Get(name string) (interface{}, error) {
	return s.value, nil
}

func (s singleFinder) RValue(dialect Dialect, param *Param) (interface{}, error) {
	return toSQLType(dialect, param, s.value)
}

type mapFinder map[string]interface{}

func (m mapFinder) Get(name string) (interface{}, error) {
	v, ok := m[name]
	if !ok {
		return nil, ErrNotFound
	}
	return v, nil
}

func (m mapFinder) RValue(dialect Dialect, param *Param) (interface{}, error) {
	value, ok := m[param.Name]
	if !ok {
		return nil, ErrNotFound
	}

	return toSQLType(dialect, param, value)
}

type structFinder struct {
	rawValue interface{}
	rValue   reflect.Value
	tm       *StructMap
}

func (sf structFinder) Get(name string) (interface{}, error) {
	fi, ok := sf.tm.Names[name]
	if !ok {
		return nil, ErrNotFound
	}

	field := reflectx.FieldByIndexesReadOnly(sf.rValue, fi.Index)
	return field.Interface(), nil
}

func (sf structFinder) RValue(dialect Dialect, param *Param) (interface{}, error) {
	fi, ok := sf.tm.Names[param.Name]
	if !ok {
		return nil, ErrNotFound
	}
	return fi.RValue(dialect, param, sf.rValue)
}

type kvFinder struct {
	mapper      *Mapper
	paramNames  []string
	paramValues []interface{}

	rValues []reflect.Value
}

func (kvf *kvFinder) Get(name string) (interface{}, error) {
	foundIdx := -1
	for idx := range kvf.paramNames {
		if kvf.paramNames[idx] == name {
			foundIdx = idx
			break
		}
	}

	if foundIdx >= 0 {
		return kvf.paramValues[foundIdx], nil
	}

	dotIndex := strings.IndexByte(name, '.')
	if dotIndex < 0 {
		if len(kvf.paramValues) != 1 {
			return nil, ErrNotFound
		}
		dotIndex = -1
		foundIdx = 0
	} else {
		fieldName := name[:dotIndex]
		for idx := range kvf.paramNames {
			if kvf.paramNames[idx] == fieldName {
				foundIdx = idx
				break
			}
		}
		if foundIdx < 0 {
			return nil, ErrNotFound
		}
	}

	if kvf.rValues == nil {
		kvf.rValues = make([]reflect.Value, len(kvf.paramValues))
	}
	rValue := kvf.rValues[foundIdx]
	if !rValue.IsValid() {
		rValue = reflect.ValueOf(kvf.paramValues[foundIdx])
		kvf.rValues[foundIdx] = rValue
	}

	tm := kvf.mapper.TypeMap(rValue.Type())
	fi, ok := tm.Names[name[dotIndex+1:]]
	if !ok {
		return nil, ErrNotFound
	}
	field := reflectx.FieldByIndexesReadOnly(rValue, fi.Index)
	return field.Interface(), nil
}

func (kvf *kvFinder) RValue(dialect Dialect, param *Param) (interface{}, error) {
	foundIdx := -1
	for idx := range kvf.paramNames {
		if kvf.paramNames[idx] == param.Name {
			foundIdx = idx
			break
		}
	}

	if foundIdx >= 0 {
		return toSQLType(dialect, param, kvf.paramValues[foundIdx])
	}

	dotIndex := strings.IndexByte(param.Name, '.')
	if dotIndex < 0 {
		if len(kvf.paramValues) != 1 {
			return nil, ErrNotFound
		}
		dotIndex = -1
		foundIdx = 0
	} else {
		fieldName := param.Name[:dotIndex]
		for idx := range kvf.paramNames {
			if kvf.paramNames[idx] == fieldName {
				foundIdx = idx
				break
			}
		}
		if foundIdx < 0 {
			return nil, ErrNotFound
		}
	}

	if kvf.rValues == nil {
		kvf.rValues = make([]reflect.Value, len(kvf.paramValues))
	}
	rValue := kvf.rValues[foundIdx]
	if !rValue.IsValid() {
		rValue = reflect.ValueOf(kvf.paramValues[foundIdx])
		kvf.rValues[foundIdx] = rValue
	}

	tm := kvf.mapper.TypeMap(rValue.Type())
	fi, ok := tm.Names[param.Name[dotIndex+1:]]
	if !ok {
		return nil, ErrNotFound
	}
	return fi.RValue(dialect, param, rValue)
}

type Context struct {
	Dialect Dialect
	Mapper  *Mapper

	ParamNames  []string
	ParamValues []interface{}

	finder Parameters
}

func (bc *Context) init() error {
	if len(bc.ParamNames) == 0 {
		if len(bc.ParamValues) <= 0 {
			bc.finder = emptyParameters
			return nil
		}

		if len(bc.ParamValues) > 1 {
			return errors.New("arguments is exceed 1")
		}

		if mapArgs, ok := bc.ParamValues[0].(map[string]interface{}); ok {
			bc.finder = mapFinder(mapArgs)
			return nil
		}

		rValue := reflect.ValueOf(bc.ParamValues[0])
		for rValue.Kind() == reflect.Ptr {
			rValue = rValue.Elem()
		}

		if rValue.Kind() == reflect.Struct {
			tm := bc.Mapper.TypeMap(rValue.Type())
			bc.finder = &structFinder{rawValue: bc.ParamValues[0], rValue: rValue, tm: tm}
			return nil
		}
	}

	bc.finder = &kvFinder{
		mapper:      bc.Mapper,
		paramNames:  bc.ParamNames,
		paramValues: bc.ParamValues,
	}

	return nil
}

func (bc *Context) Get(name string) (interface{}, error) {
	return bc.finder.Get(name)
}

func (bc *Context) RValue(param *Param) (interface{}, error) {
	return bc.finder.RValue(bc.Dialect, param)
}
