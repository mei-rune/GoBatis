package gobatis

import (
	"errors"
	"reflect"
	"strings"

	"github.com/runner-mei/GoBatis/reflectx"
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
	fi, ok := sf.tm.FieldNames[name]
	if !ok {
		fi, ok = sf.tm.Names[name]
		if !ok {
			return nil, ErrNotFound
		}
	}

	field := reflectx.FieldByIndexesReadOnly(sf.rValue, fi.Index)
	return field.Interface(), nil
}

func (sf structFinder) RValue(dialect Dialect, param *Param) (interface{}, error) {
	fi, ok := sf.tm.FieldNames[param.Name]
	if !ok {
		fi, ok = sf.tm.Names[param.Name]
		if !ok {
			return nil, ErrNotFound
		}
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

	fi, ok := tm.FieldNames[name[dotIndex+1:]]
	if !ok {
		fi, ok = tm.Names[name[dotIndex+1:]]
		if !ok {
			return nil, ErrNotFound
		}
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
	fi, ok := tm.FieldNames[param.Name[dotIndex+1:]]
	if !ok {
		fi, ok = tm.Names[param.Name[dotIndex+1:]]
		if !ok {
			return nil, ErrNotFound
		}
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

func (bc *Context) Get(name string) (interface{}, error) {
	return bc.finder.Get(name)
}

func (bc *Context) RValue(param *Param) (interface{}, error) {
	return bc.finder.RValue(bc.Dialect, param)
}

func NewContext(dialect Dialect, mapper *Mapper, paramNames []string, paramValues []interface{}) (*Context, error) {
	ctx := &Context{
		Dialect:     dialect,
		Mapper:      mapper,
		ParamNames:  paramNames,
		ParamValues: paramValues,
	}

	if len(paramNames) == 0 {
		if len(paramValues) > 1 {
			return nil, errors.New("arguments is exceed 1")
		}

		if len(paramValues) <= 0 {
			ctx.finder = emptyParameters
		} else if mapArgs, ok := paramValues[0].(map[string]interface{}); ok {
			ctx.finder = mapFinder(mapArgs)
		} else {
			rValue := reflect.ValueOf(paramValues[0])
			for rValue.Kind() == reflect.Ptr {
				rValue = rValue.Elem()
			}

			if rValue.Kind() == reflect.Struct {
				tm := ctx.Mapper.TypeMap(rValue.Type())
				ctx.finder = &structFinder{rawValue: paramValues[0], rValue: rValue, tm: tm}
			} else {
				ctx.finder = singleFinder{value: paramValues[0]}
			}
		}
	} else {
		ctx.finder = &kvFinder{
			mapper:      ctx.Mapper,
			paramNames:  paramNames,
			paramValues: paramValues,
		}
	}
	return ctx, nil
}
