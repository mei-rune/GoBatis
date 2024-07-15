package core

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
		if dotIndex := strings.IndexByte(name, '.'); dotIndex > 0 {
			value, ok := m[name[:dotIndex]]
			if ok {
				return readField(value, name[dotIndex+1:])
			}
		}
		return nil, ErrNotFound
	}
	return v, nil
}

func readField(value interface{}, name string) (interface{}, error) {
	if values, ok := value.(map[string]interface{}); ok {
		v, ok := values[name]
		if ok {
			return v, nil
		}
		dotIndex := strings.IndexByte(name, '.')
		if dotIndex <= 0 {
			return nil, ErrNotFound
		}
		v, ok = values[name[:dotIndex]]
		if !ok {
			return nil, ErrNotFound
		}
		return readField(v, name[dotIndex+1:])
	}

	rv := reflect.ValueOf(value)
	if rv.Kind() == reflect.Ptr {
		rv = rv.Elem()
	}

	switch rv.Kind() {
	case reflect.Struct:
		dotIndex := strings.IndexByte(name, '.')
		if dotIndex <= 0 {
			v := rv.FieldByName(name)
			if v.IsValid() {
				return v.Interface(), nil
			}
		} else {
			v := rv.FieldByName(name[:dotIndex])
			if v.IsValid() {
				return readField(v.Interface(), name[dotIndex+1:])
			}
		}
	case reflect.Map:
		v := rv.MapIndex(reflect.ValueOf(name))
		if v.IsValid() {
			return v.Interface(), nil
		}
	}
	return nil, ErrNotFound
}

func (m mapFinder) RValue(dialect Dialect, param *Param) (interface{}, error) {
	value, err := m.Get(param.Name)
	if err != nil {
		return nil, err
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
	Parameters  Parameters
	mapper      *Mapper
	paramNames  []string
	paramValues []interface{}

	cachedValues []reflect.Value
}

type valueGetter interface {
	value(value interface{}) (interface{}, error)
	field(fi *FieldInfo, rv reflect.Value) (interface{}, error)
}

type directGetter struct{}

func (v directGetter) value(value interface{}) (interface{}, error) {
	return value, nil
}

func (v directGetter) field(fi *FieldInfo, rv reflect.Value) (interface{}, error) {
	field := reflectx.FieldByIndexesReadOnly(rv, fi.Index)
	return field.Interface(), nil
}

type rvalueGetter struct {
	dialect Dialect
	param   *Param
}

func (v rvalueGetter) value(value interface{}) (interface{}, error) {
	return toSQLType(v.dialect, v.param, value)
}

func (v rvalueGetter) field(fi *FieldInfo, rv reflect.Value) (interface{}, error) {
	return fi.RValue(v.dialect, v.param, rv)
}

var directgetter = directGetter{}

func (kvf *kvFinder) Get(name string) (interface{}, error) {
	return kvf.get(name, directgetter)
}

func (kvf *kvFinder) RValue(dialect Dialect, param *Param) (interface{}, error) {
	return kvf.get(param.Name, rvalueGetter{dialect: dialect, param: param})
}

func (kvf *kvFinder) getByIndex(foundIdx int) (reflect.Value, bool) {
	if kvf.cachedValues == nil {
		kvf.cachedValues = make([]reflect.Value, len(kvf.paramValues))
	}
	rValue := kvf.cachedValues[foundIdx]
	if !rValue.IsValid() {
		value := kvf.paramValues[foundIdx]
		if value == nil {
			return reflect.Value{}, false
		}
		rValue = reflect.ValueOf(value)
		kvf.cachedValues[foundIdx] = rValue
	}
	return rValue, true
}

func (kvf *kvFinder) get(name string, getter valueGetter) (interface{}, error) {
	foundIdx := -1
	for idx := range kvf.paramNames {
		if kvf.paramNames[idx] == name {
			foundIdx = idx
			break
		}
	}

	if foundIdx >= 0 {
		return getter.value(kvf.paramValues[foundIdx])
	}

	if kvf.Parameters != nil {
		value, err := kvf.Parameters.Get(name)
		if err == nil {
			return value, nil
		}

		if err != ErrNotFound {
			return nil, err
		}

		return getter.value(value)
	}

	dotIndex := strings.IndexByte(name, '.')
	if dotIndex < 0 {
		//    这里的是为下面情况的特殊处理
		//    结构为 type XXX struct { f1 int, f2  int}
		//    方法定义为 Insert(x *XXX) error
		//    对应 sql 为  insert into xxx (f1, f2) values(#{f1}, #{f2})

		if len(kvf.paramValues) != 1 {
			return nil, ErrNotFound
		}
		dotIndex = -1
		foundIdx = 0

		rValue, ok := kvf.getByIndex(foundIdx)
		if !ok {
			return nil, ErrNotFound // errors.New("canot read param '" + name[dotIndex+1:] + "',  param '" + name[:dotIndex+1] + "' is nil")
		}

		return kvf.getFieldValue(name[dotIndex+1:], rValue, getter)
	}

	fieldName := name[:dotIndex]
	for idx := range kvf.paramNames {
		if kvf.paramNames[idx] == fieldName {
			foundIdx = idx
			break
		}
	}
	if foundIdx >= 0 {
		rValue, ok := kvf.getByIndex(foundIdx)
		if !ok {
			return nil, ErrNotFound // errors.New("canot read param '" + name[dotIndex+1:] + "',  param '" + name[:dotIndex+1] + "' is nil")
		}
		return kvf.getFieldValue(name[dotIndex+1:], rValue, getter)
	}

	if kvf.Parameters == nil {
		return nil, ErrNotFound
	}

	value, err := kvf.Parameters.Get(fieldName)
	if err != nil {
		return nil, err
	}
	if value == nil {
		return nil, ErrNotFound
	}

	return kvf.getFieldValue(name[dotIndex+1:], reflect.ValueOf(value), getter)
}

func (kvf *kvFinder) getFieldValue(fieldName string, rValue reflect.Value, getter valueGetter) (interface{}, error) {
	kind := rValue.Kind()
	if kind == reflect.Ptr {
		kind = rValue.Type().Elem().Kind()

		if rValue.IsNil() {
			return nil, ErrNotFound // errors.New("canot read param '" + name[:dotIndex+1] + "',  param '" + name[:dotIndex+1] + "' is nil")
		}
	}

	if kind == reflect.Map {
		if rValue.IsNil() {
			return nil, ErrNotFound // errors.New("canot read param '" + name[:dotIndex+1] + "',  param '" + name[:dotIndex+1] + "' is nil")
		}

		value := rValue.MapIndex(reflect.ValueOf(fieldName))
		if !value.IsValid() {
			return getter.value(nil)
			// return nil, ErrNotFound //errors.New("canot read param '" + name[:dotIndex+1] + "',  param '" + name[:dotIndex+1] + "' is nil")
		}
		return getter.value(value.Interface())
	}

	// 注意这里的代码请看上面的注释
	if kind != reflect.Struct {
		return nil, ErrNotFound // errors.New("canot read param '" + name + "',  param '" + name + "' is nil")
	}

	tm := kvf.mapper.TypeMap(rValue.Type())

	fi, ok := tm.FieldNames[fieldName]
	if !ok {
		fi, ok = tm.Names[fieldName]
		if !ok {
			return nil, ErrNotFound
		}
	}

	return getter.field(fi, rValue)
}

type constantWrapper struct {
	constants map[string]interface{}
	finder    Parameters
}

func (w constantWrapper) Get(name string) (interface{}, error) {
	v, e := w.finder.Get(name)
	if e == nil {
		return v, nil
	}
	if name == "constants" {
		return w.constants, nil
	}
	if strings.HasPrefix(name, "constants.") {
		cv, ok := w.constants[strings.TrimPrefix(name, "constants.")]
		if ok {
			return cv, nil
		}
	}
	return nil, e
}

func (w constantWrapper) RValue(dialect Dialect, param *Param) (interface{}, error) {
	v, e := w.finder.RValue(dialect, param)
	if e == nil {
		return v, nil
	}
	if param.Name == "constants" {
		return w.constants, nil
	}
	if strings.HasPrefix(param.Name, "constants.") {
		cv, ok := w.constants[strings.TrimPrefix(param.Name, "constants.")]
		if ok {
			return toSQLType(dialect, param, cv)
		}
	}
	return nil, e
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

func NewContext(constants map[string]interface{}, dialect Dialect, mapper *Mapper, paramNames []string, paramValues []interface{}) (*Context, error) {
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

	if constants != nil {
		ctx.finder = constantWrapper{constants: constants, finder: ctx.finder}
	}
	return ctx, nil
}
