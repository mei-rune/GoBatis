package gobatis

import (
	"reflect"
	"text/template"
)

func isEmpty(list interface{}) bool {
	if list == nil {
		return true
	}
	rValue := reflect.ValueOf(list)
	if rValue.Kind() != reflect.Slice {
		return false
	}
	return rValue.Len() == 0
}

func isNotEmpty(list interface{}) bool {
	return !isEmpty(list)
}

var templateFuncs = template.FuncMap{
	"isLast": func(list interface{}, idx int) bool {
		if list == nil {
			return false
		}
		rValue := reflect.ValueOf(list)
		if rValue.Kind() != reflect.Slice {
			return false
		}
		return idx == (rValue.Len() - 1)
	},

	"isFirst": func(list interface{}, idx int) bool {
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
	},

	"isEmpty":    isEmpty,
	"isNotEmpty": isNotEmpty,
}
