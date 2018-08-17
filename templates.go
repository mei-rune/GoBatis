package gobatis

import (
	"reflect"
	"text/template"
)

var TemplateFuncs = template.FuncMap{}

func init() {
	isEmpty := func(list interface{}) bool {
		if list == nil {
			return true
		}
		rValue := reflect.ValueOf(list)
		if rValue.Kind() != reflect.Slice {
			return false
		}
		return rValue.Len() == 0
	}

	isNotEmpty := func(list interface{}) bool {
		return !isEmpty(list)
	}

	isLast := func(list interface{}, idx int) bool {
		if list == nil {
			return false
		}
		rValue := reflect.ValueOf(list)
		if rValue.Kind() != reflect.Slice {
			return false
		}
		return idx == (rValue.Len() - 1)
	}

	isNotLast := func(list interface{}, idx int) bool {
		return !isLast(list, idx)
	}

	isFirst := func(list interface{}, idx int) bool {
		if list == nil {
			return false
		}
		rValue := reflect.ValueOf(list)
		if rValue.Kind() != reflect.Slice {
			return false
		}
		if rValue.Len() == 0 {
			return false
		}
		return idx == 0
	}

	isNotFirst := func(list interface{}, idx int) bool {
		return !isFirst(list, idx)
	}

	TemplateFuncs["isLast"] = isLast
	TemplateFuncs["isNotLast"] = isNotLast
	TemplateFuncs["isFirst"] = isFirst
	TemplateFuncs["isNotFirst"] = isNotFirst
	TemplateFuncs["isEmpty"] = isEmpty
	TemplateFuncs["isNotEmpty"] = isNotEmpty
}
