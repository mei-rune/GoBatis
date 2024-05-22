//go:build !gval && !cel
// +build !gval,!cel

package core

import (
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/Knetic/govaluate"
)

func isEmptyString(args ...interface{}) (bool, error) {
	isLike := false
	if len(args) != 1 {
		if len(args) != 2 {
			return false, errors.New("args.len() isnot 1 or 2")
		}
		rv := reflect.ValueOf(args[1])
		if rv.Kind() != reflect.Bool {
			return false, errors.New("args[1] isnot bool type")
		}
		isLike = rv.Bool()
	}
	if args[0] == nil {
		return true, nil
	}
	rv := reflect.ValueOf(args[0])
	if rv.Kind() == reflect.String {
		if rv.Len() == 0 {
			return true, nil
		}
		if isLike {
			return rv.String() == "%" || rv.String() == "%%", nil
		}
		return false, nil
	}
	return false, errors.New("value isnot string")
}

func isNil(args ...interface{}) (bool, error) {
	for idx, arg := range args {
		rv := reflect.ValueOf(arg)
		if rv.Kind() != reflect.Ptr &&
			rv.Kind() != reflect.Map &&
			rv.Kind() != reflect.Slice &&
			rv.Kind() != reflect.Interface {
			return false, errors.New("isNil: args(" + strconv.FormatInt(int64(idx), 10) + ") isnot ptr - " + rv.Kind().String())
		}

		if !rv.IsNil() {
			return false, nil
		}
	}

	return true, nil
}

func isNull(args ...interface{}) (interface{}, error) {
	if len(args) == 0 {
		return nil, errors.New("isnull() args is empty")
	}

	b, err := isNil(args...)
	if err != nil {
		return nil, err
	}

	return b, nil
}

func isZero(args ...interface{}) (bool, error) {
	if len(args) == 0 {
		return false, errors.New("isZero() args is empty")
	}

	switch v := args[0].(type) {
	case time.Time:
		return v.IsZero(), nil
	case *time.Time:
		if v == nil {
			return true, nil
		}
		return v.IsZero(), nil
	case int:
		return v == 0, nil
	case int64:
		return v == 0, nil
	case int32:
		return v == 0, nil
	case int16:
		return v == 0, nil
	case int8:
		return v == 0, nil
	case uint:
		return v == 0, nil
	case uint64:
		return v == 0, nil
	case uint32:
		return v == 0, nil
	case uint16:
		return v == 0, nil
	case uint8:
		return v == 0, nil
	}

	return false, nil
}

func isNotNull(args ...interface{}) (interface{}, error) {
	if len(args) == 0 {
		return nil, errors.New("isnotnull() args is empty")
	}

	for _, arg := range args {
		rv := reflect.ValueOf(arg)
		if rv.Kind() != reflect.Ptr &&
			rv.Kind() != reflect.Map &&
			rv.Kind() != reflect.Slice &&
			rv.Kind() != reflect.Interface {
			continue
		}

		if rv.IsNil() {
			return false, nil
		}
	}

	return true, nil
}

var expFunctions = map[string]govaluate.ExpressionFunction{
	"hasPrefix": func(args ...interface{}) (interface{}, error) {
		if len(args) != 2 {
			return nil, errors.New("hasPrefix args is invalid")
		}

		return strings.HasPrefix(args[0].(string), args[1].(string)), nil // nolint: forcetypeassert
	},
	"hasSuffix": func(args ...interface{}) (interface{}, error) {
		if len(args) != 2 {
			return nil, errors.New("hasSuffix args is invalid")
		}

		return strings.HasSuffix(args[0].(string), args[1].(string)), nil // nolint: forcetypeassert
	},
	"trimPrefix": func(args ...interface{}) (interface{}, error) {
		if len(args) != 2 {
			return nil, errors.New("hasSuffix args is invalid")
		}

		return strings.TrimPrefix(args[0].(string), args[1].(string)), nil // nolint: forcetypeassert
	},
	"trimSuffix": func(args ...interface{}) (interface{}, error) {
		if len(args) != 2 {
			return nil, errors.New("hasSuffix args is invalid")
		}

		return strings.TrimSuffix(args[0].(string), args[1].(string)), nil // nolint: forcetypeassert
	},
	"trimSpace": func(args ...interface{}) (interface{}, error) {
		if len(args) != 1 {
			return nil, errors.New("hasSuffix args is invalid")
		}
		return strings.TrimSpace(args[0].(string)), nil // nolint: forcetypeassert
	},

	"len": func(args ...interface{}) (interface{}, error) {
		if len(args) != 1 {
			return nil, errors.New("len() args isnot 1")
		}

		rv := reflect.ValueOf(args[0])
		if rv.Kind() == reflect.Slice ||
			rv.Kind() == reflect.Array ||
			rv.Kind() == reflect.Map ||
			rv.Kind() == reflect.String {
			return float64(rv.Len()), nil
		}
		return nil, errors.New("value isnot slice, array, string or map")
	},
	"isEmpty": func(args ...interface{}) (interface{}, error) {
		if len(args) != 1 {
			return nil, errors.New("len() args isnot 1")
		}

		rv := reflect.ValueOf(args[0])
		if rv.Kind() == reflect.Slice ||
			rv.Kind() == reflect.Array ||
			rv.Kind() == reflect.Map ||
			rv.Kind() == reflect.String {
			return rv.Len() == 0, nil
		}
		return nil, errors.New("value isnot slice, array, string or map")
	},
	"isNotEmpty": func(args ...interface{}) (interface{}, error) {
		if len(args) != 1 {
			return nil, errors.New("len() args isnot 1")
		}
		if args[0] == nil {
			return 0, nil
		}
		rv := reflect.ValueOf(args[0])
		if rv.Kind() == reflect.Slice ||
			rv.Kind() == reflect.Array ||
			rv.Kind() == reflect.Map ||
			rv.Kind() == reflect.String {
			return rv.Len() != 0, nil
		}
		return nil, errors.New("value isnot slice, array, string or map")
	},

	"isEmptyString": func(args ...interface{}) (interface{}, error) {
		a, err := isEmptyString(args...)
		if err != nil {
			return nil, err
		}
		return a, nil
	},

	"isZero": func(args ...interface{}) (interface{}, error) {
		a, err := isZero(args...)
		if err != nil {
			return nil, err
		}
		return a, nil
	},

	"isNotZero": func(args ...interface{}) (interface{}, error) {
		a, err := isZero(args...)
		if err != nil {
			return nil, err
		}
		return !a, nil
	},

	"isNotEmptyString": func(args ...interface{}) (interface{}, error) {
		a, err := isEmptyString(args...)
		if err != nil {
			return nil, err
		}
		return !a, nil
	},

	"isnull": isNull,
	"isNull": isNull,

	"isnotnull": isNotNull,
	"isNotNull": isNotNull,
}

func RegisterExprFunction(name string, fn func(args ...interface{}) (interface{}, error)) {
	expFunctions[name] = govaluate.ExpressionFunction(fn)
}

func ParseEvaluableExpression(s string) (Testable, error) {
	s = replaceAndOr(s)
	expr, err := govaluate.NewEvaluableExpressionWithFunctions(s, expFunctions)
	if err != nil {
		return nil, err
	}
	return govaluateTestable{test: expr}, nil
}

type govaluateTestable struct {
	test *govaluate.EvaluableExpression
}

func (gv govaluateTestable) String() string {
	return gv.test.String()
}

func (gv govaluateTestable) Test(vg TestGetter) (bool, error) {
	result, err := gv.test.Eval(vg)
	if err != nil {
		return false, err
	}

	if result == nil {
		return false, errors.New("result of test expression is nil - " + gv.test.String())
	}

	bResult, ok := result.(bool)
	if !ok {
		return false, errors.New("result of test expression isnot bool got " + fmt.Sprintf("%T", result) + " - " + gv.test.String())
	}

	return bResult, nil
}
