//go:build !cel && !govaluate
// +build !cel,!govaluate

package core

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/PaesslerAG/gval"
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

// func strlen(args ...interface{}) (interface{}, error) {
// 	s, err := as.String(args[0])
// 	if err != nil {
// 		return nil, err
// 	}
// 	return (float64)(len(s)), nil
// }

// func toUint64(ctx context.Context, args ...interface{}) (interface{}, error) {
// 	return as.Uint64(args[0])
// }

// func toInt64(ctx context.Context, args ...interface{}) (interface{}, error) {
// 	return as.Int64(args[0])
// }

// func toFloat64(ctx context.Context, args ...interface{}) (interface{}, error) {
// 	return as.Float64(args[0])
// }

// func toString(ctx context.Context, args ...interface{}) (interface{}, error) {
// 	return as.String(args[0])
// }

var expFunctions = []gval.Language{
	// gval.Function("strlen", strlen),
	// gval.Function("toUint64", toUint64),
	// gval.Function("toInt64", toInt64),
	// gval.Function("toFloat64", toFloat64),
	// gval.Function("toString", toString),
	gval.Function("hasPrefix", func(args ...interface{}) (interface{}, error) {
		if len(args) != 2 {
			return nil, errors.New("hasPrefix args is invalid")
		}

		return strings.HasPrefix(args[0].(string), args[1].(string)), nil // nolint: forcetypeassert
	}),
	gval.Function("hasSuffix", func(args ...interface{}) (interface{}, error) {
		if len(args) != 2 {
			return nil, errors.New("hasSuffix args is invalid")
		}

		return strings.HasSuffix(args[0].(string), args[1].(string)), nil // nolint: forcetypeassert
	}),
	gval.Function("trimPrefix", func(args ...interface{}) (interface{}, error) {
		if len(args) != 2 {
			return nil, errors.New("hasSuffix args is invalid")
		}

		return strings.TrimPrefix(args[0].(string), args[1].(string)), nil // nolint: forcetypeassert
	}),
	gval.Function("trimSuffix", func(args ...interface{}) (interface{}, error) {
		if len(args) != 2 {
			return nil, errors.New("hasSuffix args is invalid")
		}

		return strings.TrimSuffix(args[0].(string), args[1].(string)), nil // nolint: forcetypeassert
	}),
	gval.Function("trimSpace", func(args ...interface{}) (interface{}, error) {
		if len(args) != 1 {
			return nil, errors.New("hasSuffix args is invalid")
		}
		return strings.TrimSpace(args[0].(string)), nil // nolint: forcetypeassert
	}),
	gval.Function("len", func(args ...interface{}) (interface{}, error) {
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
	}),
	gval.Function("isEmpty", func(args ...interface{}) (interface{}, error) {
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
	}),
	gval.Function("isNotEmpty", func(args ...interface{}) (interface{}, error) {
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
	}),
	gval.Function("isEmptyString", func(args ...interface{}) (interface{}, error) {
		a, err := isEmptyString(args...)
		if err != nil {
			return nil, err
		}
		return a, nil
	}),
	gval.Function("isZero", func(args ...interface{}) (interface{}, error) {
		a, err := isZero(args...)
		if err != nil {
			return nil, err
		}
		return a, nil
	}),
	gval.Function("isNotZero", func(args ...interface{}) (interface{}, error) {
		a, err := isZero(args...)
		if err != nil {
			return nil, err
		}
		return !a, nil
	}),
	gval.Function("isNotEmptyString", func(args ...interface{}) (interface{}, error) {
		a, err := isEmptyString(args...)
		if err != nil {
			return nil, err
		}
		return !a, nil
	}),
	gval.Function("isnull", isNull),
	gval.Function("isNull", isNull),
	gval.Function("isnotnull", isNotNull),
	gval.Function("isNotNull", isNotNull),
}

func RegisterExprFunction(name string, fn func(args ...interface{}) (interface{}, error)) {
	expFunctions = append(expFunctions, gval.Function(name, fn))
}

type exprEvaluable struct {
	program func(c context.Context, parameter interface{}) (interface{}, error)
	str     string
}

func (eval exprEvaluable) String() string {
	return eval.str
}

type gvalSelector struct {
	get TestGetter
}

func (gs gvalSelector) SelectGVal(c context.Context, key string) (interface{}, error) {
	return gs.get.Get(key)
}

var _ gval.Selector = gvalSelector{}

func (eval exprEvaluable) Test(parameter TestGetter) (bool, error) {
	result, err := eval.program(context.Background(), gvalSelector{get: parameter})
	if err != nil {
		return false, err
	}

	if result == nil {
		return false, errors.New("result of test expression is nil - " + eval.str)
	}

	bResult, ok := result.(bool)
	if !ok {
		return false, errors.New("result of test expression isnot bool got " + fmt.Sprintf("%T", result) + " - " + eval.str)
	}

	return bResult, nil
}

func ParseEvaluableExpression(exprStr string) (Testable, error) {
	exprStr = replaceAndOr(exprStr)
	eval, err := gval.Full(expFunctions...).NewEvaluable(exprStr)
	if err != nil {
		return nil, errors.New("expr '" + exprStr + "' is invalid, " + err.Error())
	}
	return exprEvaluable{
		program: eval,
		str:     exprStr,
	}, nil
}
