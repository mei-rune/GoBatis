package core

import (
	"database/sql"
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"time"
	"unicode"

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

func isZero(args ...interface{}) (interface{}, error) {
	if len(args) == 0 {
		return nil, errors.New("isnull() args is empty")
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

type sqlPrinter struct {
	ctx    *Context
	sb     strings.Builder
	params []interface{}
	err    error
}

func (printer *sqlPrinter) addPlaceholder() {
	sql := printer.ctx.Dialect.Placeholder().Format(len(printer.params))
	printer.sb.WriteString(sql)
}

func (printer *sqlPrinter) addPlaceholderAndParam(value interface{}) {
	printer.addPlaceholder()
	printer.params = append(printer.params, value)
}

func (printer *sqlPrinter) Clone() *sqlPrinter {
	return &sqlPrinter{
		ctx:    printer.ctx,
		params: printer.params,
		err:    printer.err,
	}
}

var elseExpr = elseExpression{}

type elseExpression struct {
	test string
}

func (e elseExpression) String() string {
	return "<else/>"
}

func (e elseExpression) writeTo(printer *sqlPrinter) {
	panic("这个不能作为")
}

func toElse(s sqlExpression) (elseExpression, bool) {
	expr, ok := s.(elseExpression)
	return expr, ok
}

type sqlExpression interface {
	String() string
	writeTo(printer *sqlPrinter)
}

func newRawExpression(content string) (sqlExpression, error) {
	fragments, bindParams, err := CompileNamedQuery(content)
	if err != nil {
		return nil, err
	}

	if len(bindParams) == 0 {
		return rawString(content), nil
	}

	return &rawStringWithParams{
		rawSQL:     content,
		fragments:  fragments,
		bindParams: bindParams,
	}, nil
}

type rawString string

func (rss rawString) String() string {
	return string(rss)
}

func (rss rawString) writeTo(printer *sqlPrinter) {
	printer.sb.WriteString(string(rss))
}

type rawStringWithParams struct {
	rawSQL     string
	fragments  []string
	bindParams Params
}

func (rs *rawStringWithParams) String() string {
	return rs.rawSQL
}

func (stmt *rawStringWithParams) writeTo(printer *sqlPrinter) {
	sql := Placeholders(printer.ctx.Dialect.Placeholder(),
		stmt.fragments, stmt.bindParams, len(printer.params))
	sqlParams, err := bindNamedQuery(stmt.bindParams, printer.ctx)
	if err != nil {
		printer.err = err
		return
	}
	printer.sb.WriteString(sql)
	if len(sqlParams) != 0 {
		printer.params = append(printer.params, sqlParams...)
	}
}

type evalParameters struct {
	ctx *Context
}

func (eval evalParameters) Get(name string) (interface{}, error) {
	value, err := eval.ctx.Get(name)
	if err == nil {
		return value, nil
	}
	if err == ErrNotFound {
		return nil, nil
	}
	return nil, err
}

type ifExpression struct {
	test                            *govaluate.EvaluableExpression
	trueExpression, falseExpression sqlExpression
}

func (ifExpr ifExpression) String() string {
	var sb strings.Builder
	sb.WriteString("<if test=\"")
	sb.WriteString(ifExpr.test.String())
	sb.WriteString("\">")
	if ifExpr.trueExpression != nil {
		sb.WriteString(ifExpr.trueExpression.String())
	}
	if ifExpr.falseExpression != nil {
		sb.WriteString("<else/>")
		sb.WriteString(ifExpr.falseExpression.String())
	}
	sb.WriteString("</if>")
	return sb.String()
}

func (ifExpr ifExpression) writeTo(printer *sqlPrinter) {
	bResult, err := isOK(ifExpr.test, printer)
	if err != nil {
		printer.err = err
		return
	}

	if bResult {
		if ifExpr.trueExpression != nil {
			ifExpr.trueExpression.writeTo(printer)
		}
	} else {
		if ifExpr.falseExpression != nil {
			ifExpr.falseExpression.writeTo(printer)
		}
	}
}

func isOK(test *govaluate.EvaluableExpression, printer *sqlPrinter) (bool, error) {
	result, err := test.Eval(evalParameters{ctx: printer.ctx})
	if err != nil {
		return false, err
	}

	if result == nil {
		return false, errors.New("result of if expression  is nil - " + test.String())
	}

	bResult, ok := result.(bool)
	if !ok {
		return false, errors.New("result of if expression isnot bool got " + fmt.Sprintf("%T", result) + " - " + test.String())
	}

	return bResult, nil
}

func newIFExpression(test string, segements []sqlExpression) (sqlExpression, error) {
	if test == "" {
		return nil, errors.New("if test is empty")
	}
	if len(segements) == 0 {
		return nil, errors.New("if content is empty")
	}

	expr, err := govaluate.NewEvaluableExpressionWithFunctions(test, expFunctions)
	if err != nil {
		return nil, errors.New("expression '" + test + "' is invalid: " + err.Error())
	}

	ifExpr := ifExpression{test: expr}

	elseIndex := -1
	for idx := range segements {
		if _, ok := toElse(segements[idx]); ok {
			elseIndex = idx
			break
		}
	}

	var trueExprs, falseExprs []sqlExpression
	if elseIndex >= 0 {
		trueExprs = segements[:elseIndex]
		falseExprs = segements[elseIndex+1:]
	} else {
		trueExprs = segements
	}

	if len(trueExprs) == 1 {
		ifExpr.trueExpression = trueExprs[0]
	} else if len(trueExprs) > 1 {
		ifExpr.trueExpression = expressionArray(trueExprs)
	}

	if len(falseExprs) == 1 {
		ifExpr.falseExpression = falseExprs[0]
	} else if len(falseExprs) > 1 {
		ifExpr.falseExpression = expressionArray(falseExprs)
	}

	return ifExpr, nil
}

type choseExpression struct {
	el xmlChoseElement

	when      []whenExpression
	otherwise sqlExpression
}

func (chose *choseExpression) String() string {
	var sb strings.Builder
	sb.WriteString("<chose>")
	for idx := range chose.when {
		sb.WriteString(chose.when[idx].String())
	}
	sb.WriteString("<otherwise>")
	sb.WriteString(chose.otherwise.String())
	sb.WriteString("</otherwise>")
	sb.WriteString("</chose>")
	return chose.el.String()
}

func (chose *choseExpression) writeTo(printer *sqlPrinter) {
	for idx := range chose.when {
		bResult, err := isOK(chose.when[idx].test, printer)
		if err != nil {
			printer.err = err
			return
		}

		if bResult {
			chose.when[idx].expression.writeTo(printer)
			return
		}
	}

	if chose.otherwise != nil {
		chose.otherwise.writeTo(printer)
	}
}

type whenExpression struct {
	test       *govaluate.EvaluableExpression
	expression sqlExpression
}

func (ifExpr whenExpression) String() string {
	var sb strings.Builder
	sb.WriteString("<when test=\"")
	sb.WriteString(ifExpr.test.String())
	sb.WriteString("\">")
	sb.WriteString(ifExpr.expression.String())
	sb.WriteString("</when>")
	return sb.String()
}

// func (ifExpr whenExpression) writeTo(printer *sqlPrinter) {
// 	bResult, err := isOK(ifExpr.test, printer)
// 	if err != nil {
// 		printer.err = err
// 		return
// 	}

// 	if bResult {
// 		ifExpr.expression.writeTo(printer)
// 	}
// }

func newChoseExpression(el xmlChoseElement) (sqlExpression, error) {
	var when []whenExpression

	for idx := range el.when {
		if el.when[idx].test == "" {
			return nil, errors.New("when test is empty")
		}
		if el.when[idx].content == nil {
			return nil, errors.New("when content is empty")
		}

		expr, err := govaluate.NewEvaluableExpressionWithFunctions(el.when[idx].test, expFunctions)
		if err != nil {
			return nil, errors.New("expression '" + el.when[idx].test + "' is invalid: " + err.Error())
		}

		when = append(when, whenExpression{test: expr, expression: el.when[idx].content})
	}

	return &choseExpression{
		el:        el,
		when:      when,
		otherwise: el.otherwise,
	}, nil
}

type forEachExpression struct {
	el        xmlForEachElement
	segements []sqlExpression
}

func (foreach *forEachExpression) String() string {
	return foreach.el.String()
}

func (foreach *forEachExpression) execOne(printer *sqlPrinter, key, value interface{}) {
	newPrinter := printer.Clone()
	ctx := *printer.ctx
	newPrinter.ctx = &ctx
	newPrinter.ctx.finder = &kvFinder{
		mapper:      printer.ctx.Mapper,
		paramNames:  []string{foreach.el.item, foreach.el.index},
		paramValues: []interface{}{value, key},
	}

	for idx := range foreach.segements {
		foreach.segements[idx].writeTo(newPrinter)
		if newPrinter.err != nil {
			break
		}
	}

	printer.sb.WriteString(newPrinter.sb.String())
	printer.params = newPrinter.params
	printer.err = newPrinter.err
}

func (foreach *forEachExpression) writeTo(printer *sqlPrinter) {
	collection, err := printer.ctx.Get(foreach.el.collection)
	if err != nil {
		if err != ErrNotFound {
			printer.err = err
		}
		return
	}

	if collection == nil {
		return
	}

	switch array := collection.(type) {
	case []int:
		if len(array) == 0 {
			return
		}

		printer.sb.WriteString(foreach.el.openTag)
		for idx := range array {
			if idx > 0 {
				printer.sb.WriteString(foreach.el.separatorTag)
			}
			foreach.execOne(printer, idx, array[idx])
		}
		printer.sb.WriteString(foreach.el.closeTag)
	case []int64:
		if len(array) == 0 {
			return
		}

		printer.sb.WriteString(foreach.el.openTag)
		for idx := range array {
			if idx > 0 {
				printer.sb.WriteString(foreach.el.separatorTag)
			}
			foreach.execOne(printer, idx, array[idx])
		}
		printer.sb.WriteString(foreach.el.closeTag)
	case []interface{}:
		if len(array) == 0 {
			return
		}

		printer.sb.WriteString(foreach.el.openTag)
		for idx := range array {
			if idx > 0 {
				printer.sb.WriteString(foreach.el.separatorTag)
			}
			foreach.execOne(printer, idx, array[idx])
		}
		printer.sb.WriteString(foreach.el.closeTag)

	case map[string]interface{}:
		if len(array) == 0 {
			return
		}

		printer.sb.WriteString(foreach.el.openTag)
		isFirst := true
		for key, value := range array {
			if isFirst {
				isFirst = false
			} else {
				printer.sb.WriteString(foreach.el.separatorTag)
			}
			foreach.execOne(printer, key, value)
		}
		printer.sb.WriteString(foreach.el.closeTag)
	default:
		rv := reflect.ValueOf(collection)
		if rv.Kind() == reflect.Slice || rv.Kind() == reflect.Array {
			aLen := rv.Len()
			if aLen <= 0 {
				return
			}

			printer.sb.WriteString(foreach.el.openTag)
			for idx := 0; idx < aLen; idx++ {
				if idx > 0 {
					printer.sb.WriteString(foreach.el.separatorTag)
				}
				foreach.execOne(printer, idx, rv.Index(idx).Interface())
			}
			printer.sb.WriteString(foreach.el.closeTag)
			return
		}

		if rv.Kind() == reflect.Map {
			keys := rv.MapKeys()
			if len(keys) <= 0 {
				return
			}

			printer.sb.WriteString(foreach.el.openTag)
			for idx := range keys {
				if idx > 0 {
					printer.sb.WriteString(foreach.el.separatorTag)
				}

				foreach.execOne(printer, keys[idx].Interface(), rv.MapIndex(keys[idx]).Interface())
			}
			printer.sb.WriteString(foreach.el.closeTag)
			return
		}

		printer.err = errors.New(fmt.Sprintf("collection isnot slice, array or map, actual is %T - ", collection) + foreach.String())
	}
}

func newForEachExpression(el xmlForEachElement) (sqlExpression, error) {
	if len(el.contents) == 0 {
		return nil, errors.New("contents of foreach is empty")
	}

	if el.collection == "" {
		return nil, errors.New("collection of foreach is empty")
	}

	if el.index == "" {
		el.index = "index"
	}
	if el.item == "" {
		el.item = "item"
	}

	return &forEachExpression{el: el, segements: el.contents}, nil
}

type whereExpression struct {
	expressions expressionArray
}

func (where *whereExpression) String() string {
	var sb strings.Builder
	sb.WriteString("<where>")
	for idx := range where.expressions {
		sb.WriteString(where.expressions[idx].String())
	}
	sb.WriteString("</where>")
	return sb.String()
}

func (where *whereExpression) writeTo(printer *sqlPrinter) {
	old := printer.sb.String()
	oldLen := printer.sb.Len()
	printer.sb.WriteString(" WHERE ")
	whereStart := printer.sb.Len()
	where.expressions.writeTo(printer)
	if printer.err != nil {
		return
	}
	if printer.sb.Len() == whereStart {
		printer.sb.Reset()
		printer.sb.WriteString(old)
	} else {
		full := printer.sb.String()
		s := full[whereStart:]
		s = strings.TrimSpace(s)
		if len(s) < 4 {
			s = strings.ToLower(s)
			if s == "or" || s == "and" {
				printer.sb.Reset()
				printer.sb.WriteString(full[:oldLen])
			}
			return
		}

		c0 := s[0]
		c1 := s[1]
		c2 := s[2]
		c3 := s[3]

		start := 0
		if c0 == 'o' || c0 == 'O' {
			if (c1 == 'r' || c1 == 'R') && unicode.IsSpace(rune(c2)) {
				start = 2
			}
		} else if c0 == 'A' || c0 == 'a' {
			if (c1 == 'N' || c1 == 'n') && (c2 == 'D' || c2 == 'd') && unicode.IsSpace(rune(c3)) {
				start = 3
			}
		}

		c0 = s[len(s)-1]
		c1 = s[len(s)-2]
		c2 = s[len(s)-3]
		c3 = s[len(s)-4]

		end := 0
		if c0 == 'D' || c0 == 'd' {
			if (c1 == 'N' || c1 == 'n') && (c2 == 'A' || c2 == 'a') && unicode.IsSpace(rune(c3)) {
				end = 3
			}
		} else if c0 == 'R' || c0 == 'r' {
			if (c1 == 'O' || c1 == 'o') && unicode.IsSpace(rune(c2)) {
				end = 2
			}
		}

		if start != 0 || end != 0 {
			printer.sb.Reset()
			printer.sb.WriteString(full[:whereStart])
			printer.sb.WriteString(s[start : len(s)-end])
		}
	}
}

type setExpression struct {
	expressions expressionArray
}

func (set *setExpression) String() string {
	var sb strings.Builder
	sb.WriteString("<set>")
	for idx := range set.expressions {
		sb.WriteString(set.expressions[idx].String())
	}
	sb.WriteString("</set>")
	return sb.String()
}

func (set *setExpression) writeTo(printer *sqlPrinter) {
	old := printer.sb.String()
	printer.sb.WriteString(" SET ")
	oldLen := printer.sb.Len()
	set.expressions.writeTo(printer)
	if printer.err != nil {
		return
	}

	if printer.sb.Len() == oldLen {
		printer.sb.Reset()
		printer.sb.WriteString(old)
	} else {
		s := strings.TrimSpace(printer.sb.String())
		if len(s) < 1 {
			return
		}

		c := s[len(s)-1]
		if c == ',' {
			printer.sb.Reset()
			printer.sb.WriteString(s[:len(s)-1])
		}
	}
}

type expressionArray []sqlExpression

func (expressions expressionArray) String() string {
	var sb strings.Builder
	for idx := range expressions {
		sb.WriteString(expressions[idx].String())
	}
	return sb.String()
}

func (expressions expressionArray) writeTo(printer *sqlPrinter) {
	for idx := range expressions {
		expressions[idx].writeTo(printer)

		if printer.err != nil {
			return
		}
	}
}

func (expressions expressionArray) GenerateSQL(ctx *Context) (string, []interface{}, error) {
	printer := &sqlPrinter{
		ctx: ctx,
	}
	expressions.writeTo(printer)
	return printer.sb.String(), printer.params, printer.err
}

type printExpression struct {
	prefix string
	suffix string
	fmt    string
	value  string

	inStr string
}

func (expr printExpression) String() string {
	if expr.fmt == "" {
		return expr.prefix + `<print value="` + expr.value + `" />` + expr.suffix
	}
	var inStr string
	if expr.inStr != "" {
		 inStr = ` inStr="`+expr.inStr+`"`
	}
	return expr.prefix + `<print fmt="` + expr.fmt + `" value="` + expr.value + `"`+inStr+` />` + expr.suffix
}

func (expr printExpression) writeTo(printer *sqlPrinter) {
	value, err := printer.ctx.Get(expr.value)
	if err != nil {
		printer.err = errors.New("search '" + expr.value + "' fail, " + err.Error())
	} else if value == nil {
		printer.err = errors.New("'" + expr.value + "' isnot found")
	} else if expr.fmt == "" {
		inStr := strings.ToLower(expr.inStr) == "true"
		printer.sb.WriteString(expr.prefix)
		err := isValidPrintValue(value, inStr)
		if err != nil {
			printer.err = err
			return
		}
		printer.sb.WriteString(fmt.Sprint(value))
		printer.sb.WriteString(expr.suffix)
	} else {
		inStr := strings.ToLower(expr.inStr) == "true"
		printer.sb.WriteString(expr.prefix)
		err := isValidPrintValue(value, inStr)
		if err != nil {
			printer.err = err
			return
		}
		printer.sb.WriteString(fmt.Sprintf(expr.fmt, value))
		printer.sb.WriteString(expr.suffix)
	}
}

func isValidPrintValue(value interface{}, inStr bool) error {
	switch v := value.(type) {
	case string:
		_, err := isValidPrintString(v, inStr)
		return err
	case int, int8, int16, int32, int64,
		uint, uint8, uint16, uint32, uint64, 
		float32, float64,
		bool:
		return nil
	default:
		return errors.New("print value is invalid type - "+fmt.Sprintf("%T", value))
	}
}

var ErrInvalidPrintValue = errors.New("print value is invalid")

func isValidPrintString(value string, inStr bool) (string, error) {
	for _, r := range value {
		switch r {
		case '\'', '"':
			return "", ErrInvalidPrintValue
		default:
			if !inStr {
				if  unicode.IsSpace(r) {
					return "", ErrInvalidPrintValue
				}
				switch r {
				case '=',
					'>',
					'<',
					'!',
					'`',
					'~',
					'@',
					'#',
					'$',
					'%',
					'^',
					'&',
					'*',
					'-',
					'+',
					'{',
					'}',
					'(',
					')',
					'[',
					']',
					'/',
					'\'',
					'\\',
					'|',
					'?',
					':',
					';',
					',',
					'.':
					return "", ErrInvalidPrintValue
				}
			}
		}
	}

	return value, nil
}

type likeExpression struct {
	prefix string
	suffix string
	value  string

	isPrefix string
	isSuffix string
}

func (expr likeExpression) String() string {
	var s string
	if expr.isPrefix != "" {
		 s = ` isPrefix="`+expr.isPrefix+`"`
	}
	if expr.isSuffix != "" {
		 s += ` isSuffix="`+expr.isSuffix+`"`
	}
	return expr.prefix + `<like value="` + expr.value + `"`+s+` />` + expr.suffix
}

func (expr likeExpression) writeTo(printer *sqlPrinter) {
	value, err := printer.ctx.Get(expr.value)
	if err != nil {
		printer.err = errors.New("search '" + expr.value + "' fail, " + err.Error())
	} else if value == nil {
		printer.err = errors.New("'" + expr.value + "' isnot found")
	} else {
		s := fmt.Sprint(value)
		if s == "" {
			printer.err = errors.New("like param '" + expr.value + "' is empty")
			return
		}

		// s 是不是为空的， 之前有地方用 <none> 代替空绕开限限制了
		// 现在我让它出错， 以便改正
		if s == "<none>" {
			printer.err = errors.New("like param '" + expr.value + "' is <none>")
			return
		}

		printer.sb.WriteString(expr.prefix)
		if strings.HasPrefix(s, "%") || strings.HasSuffix(s, "%") {
			printer.addPlaceholderAndParam(s)
		} else {
			if strings.ToLower(expr.isPrefix) == "true" {
				printer.addPlaceholderAndParam(s+"%")
			} else if strings.ToLower(expr.isSuffix) == "true" {
				printer.addPlaceholderAndParam("%" + s)
			} else {
				printer.addPlaceholderAndParam("%" + s + "%")
			}
		}
		printer.sb.WriteString(expr.suffix)
	}
}

type limitExpression struct {
	offset string
	limit  string
}

func (expr limitExpression) String() string {
	return `<pagination offset="` + expr.offset + `" limit="` + expr.limit + `" />`
}

func (expr limitExpression) writeTo(printer *sqlPrinter) {
	o, _ := printer.ctx.Get(expr.offset)
	offset := int64With(o, 0)
	o, _ = printer.ctx.Get(expr.limit)
	limit := int64With(o, 0)

	s := printer.ctx.Dialect.Limit(offset, limit)
	printer.sb.WriteString(s)
}

type pageExpression struct {
	page string
	size string
}

func (expr pageExpression) String() string {
	return `<pagination page="` + expr.page + `" size="` + expr.size + `" />`
}

func (expr pageExpression) writeTo(printer *sqlPrinter) {
	o, _ := printer.ctx.Get(expr.page)
	page := int64With(o, 0)
	o, _ = printer.ctx.Get(expr.size)
	size := int64With(o, 0)

	if size <= 0 {
		size = 20
	}
	offset := int64(0)
	if page > 1 {
		offset = (page - 1) * size
	}

	s := printer.ctx.Dialect.Limit(offset, size)
	printer.sb.WriteString(s)
}

type orderByExpression struct {
	sort string

	prefix string
	direction string
}

func (expr orderByExpression) String() string {
	if expr.sort == "" {
		return "<order_by />"
	}
	return `<order_by prefix="`+expr.prefix+`" sort="` + expr.sort + `" direction="`+expr.direction+`"/>`
}

func (expr orderByExpression) writeTo(printer *sqlPrinter) {
	var o interface{}
	if expr.sort == "" {
		o, _ = printer.ctx.Get("sort")
	} else {
		o, _ = printer.ctx.Get(expr.sort)
	}
	if o == nil {
		return
	}

	sortStr := fmt.Sprint(o)
	if sortStr == "" {
		return
	}

	printer.sb.WriteString(" ORDER BY ")
	for idx, s := range strings.Fields(sortStr) {
		if idx > 0 {
			printer.sb.WriteString(", ")
		}

		printer.sb.WriteString(expr.prefix)
		if strings.HasPrefix(s, "+") {
			printer.sb.WriteString(strings.TrimPrefix(s, "+"))
			printer.sb.WriteString(" ASC")
		} else if strings.HasPrefix(s, "-") {
			printer.sb.WriteString(strings.TrimPrefix(s, "-"))
			printer.sb.WriteString(" DESC")
		} else {
			printer.sb.WriteString(s)
		}
		printer.sb.WriteString(" ")
		printer.sb.WriteString(expr.direction)
	}
}

type trimExpression struct {
	expressions    expressionArray
	prefixoverride []string
	prefix         sqlExpression
	suffixoverride []string
	suffix         sqlExpression
}

func (expr trimExpression) String() string {
	var sb strings.Builder
	sb.WriteString("<trim ")
	if len(expr.prefixoverride) != 0 {
		sb.WriteString("prefixOverrides=\"")
		sb.WriteString(strings.Join(expr.prefixoverride, " | "))
		sb.WriteString("\"")
	}
	if expr.prefix != nil {
		sb.WriteString("prefix=\"")
		sb.WriteString(expr.prefix.String())
		sb.WriteString("\"")
	}

	if len(expr.suffixoverride) != 0 {
		sb.WriteString("suffixOverrides=\"")
		sb.WriteString(strings.Join(expr.suffixoverride, " | "))
		sb.WriteString("\"")
	}
	if expr.suffix != nil {
		sb.WriteString("suffix=\"")
		sb.WriteString(expr.suffix.String())
		sb.WriteString("\"")
	}
	sb.WriteString(">")
	for idx := range expr.expressions {
		sb.WriteString(expr.expressions[idx].String())
	}
	sb.WriteString("</trim>")
	return sb.String()
}

func splitTrimStrings(s string) []string {
	var ss []string
	for _, prefix := range strings.Split(s, "|") {
		if a := strings.TrimSpace(prefix); a != "" {
			ss = append(ss, strings.ToLower(a))
		}
	}
	return ss
}

func (expr trimExpression) hasPrefix(s string) (bool, string) {
	if len(expr.prefixoverride) > 0 {
		s = strings.ToLower(s)
		for _, prefix := range expr.prefixoverride {
			if strings.HasPrefix(s, prefix) {
				return true, prefix
			}
		}
	}
	return false, ""
}

func (expr trimExpression) hasSuffix(s string) (bool, string) {
	if len(expr.suffixoverride) > 0 {
		s = strings.ToLower(s)

		for _, suffix := range expr.suffixoverride {
			if strings.HasSuffix(s, suffix) {
				return true, suffix
			}
		}
	}
	return false, ""
}

func (expr trimExpression) writeTo(printer *sqlPrinter) {
	newPrinter := &sqlPrinter{
		ctx:    printer.ctx,
		params: printer.params,
	}

	expr.expressions.writeTo(newPrinter)
	if newPrinter.err != nil {
		printer.err = newPrinter.err
		return
	}

	s := newPrinter.sb.String()
	if strings.TrimSpace(s) == "" {
		return
	}

	hasPrefix, prefix := expr.hasPrefix(s)
	hasSuffix, suffix := expr.hasSuffix(s)
	if hasPrefix {
		s = strings.TrimPrefix(s, prefix)
	}
	if hasSuffix {
		s = strings.TrimSuffix(s, suffix)
	}
	if strings.TrimSpace(s) == "" {
		return
	}

	hasParamsInPrefix := false

	if expr.prefix != nil {
		old := len(printer.params)
		expr.prefix.writeTo(printer)
		hasParamsInPrefix = old != len(printer.params)
	}

	if hasParamsInPrefix {
		newPrinter = printer.Clone()
		newPrinter.sb.Reset()
		expr.expressions.writeTo(newPrinter)
		if newPrinter.err != nil {
			printer.err = newPrinter.err
			return
		}

		s = newPrinter.sb.String()

		if hasPrefix, prefix = expr.hasPrefix(s); hasPrefix {
			s = strings.TrimPrefix(s, prefix)
		}
		if hasSuffix, suffix = expr.hasSuffix(s); hasSuffix {
			s = strings.TrimSuffix(s, suffix)
		}
	}

	printer.sb.WriteString(s)
	printer.params = newPrinter.params

	if expr.suffix != nil {
		expr.suffix.writeTo(printer)
	}
}

type valueRangeExpression struct {
	field string
	value string

	prefix sqlExpression
	suffix sqlExpression
}

func (expr valueRangeExpression) String() string {
	return `<value-range field="` + expr.field + `" value="` + expr.value + `" />`
}

func (expr valueRangeExpression) writeTo(printer *sqlPrinter) {
	o, _ := printer.ctx.Get(expr.value)
	if o == nil {
		printer.err = errors.New("argument '" + expr.value + "' is missing")
		return
	}

	start, end, err := toRange(o)
	if err != nil {
		printer.err = errors.New("argument '" + expr.value + "' is invalid value: " + err.Error())
		return
	}

	if start == nil {
		if end != nil {
			if expr.prefix != nil {
				expr.prefix.writeTo(printer)
			}
			printer.sb.WriteString(" ")
			printer.sb.WriteString(expr.field)
			printer.sb.WriteString(" <= ")
			printer.addPlaceholderAndParam(end)
			if expr.suffix != nil {
				expr.suffix.writeTo(printer)
			}
		}
		return
	}

	if end == nil {
		if expr.prefix != nil {
			expr.prefix.writeTo(printer)
		}
		printer.sb.WriteString(" ")
		printer.sb.WriteString(expr.field)
		printer.sb.WriteString(" >= ")
		printer.addPlaceholderAndParam(start)
		if expr.suffix != nil {
			expr.suffix.writeTo(printer)
		}
		return
	}

	if expr.prefix != nil {
		expr.prefix.writeTo(printer)
	}

	printer.sb.WriteString(" ")
	printer.sb.WriteString(expr.field)
	printer.sb.WriteString(" BETWEEN ")
	printer.addPlaceholderAndParam(start)
	printer.sb.WriteString(" AND ")
	printer.addPlaceholderAndParam(end)

	if expr.suffix != nil {
		expr.suffix.writeTo(printer)
	}
}

func toRange(o interface{}) (start interface{}, end interface{}, err error) {
	switch v := o.(type) {
	case []time.Time:
		switch len(v) {
		case 0:
			return nil, nil, nil
		case 1:
			if v[0].IsZero() {
				return nil, nil, nil
			}
			return v[0], nil, nil
		case 2:
			if v[0].IsZero() {
				if v[1].IsZero() {
					return nil, nil, nil
				}
				return nil, v[1], nil
			}
			if v[1].IsZero() {
				return v[0], nil, nil
			}
			return v[0], v[1], nil
		default:
			return nil, nil, errors.New("size isnot match")
		}
	case []int:
		switch len(v) {
		case 0:
			return nil, nil, nil
		case 1:
			return v[0], nil, nil
		case 2:
			return v[0], v[1], nil
		default:
			return nil, nil, errors.New("size isnot match")
		}
	case []int64:
		switch len(v) {
		case 0:
			return nil, nil, nil
		case 1:
			return v[0], nil, nil
		case 2:
			return v[0], v[1], nil
		default:
			return nil, nil, errors.New("size isnot match")
		}
	case []uint:
		switch len(v) {
		case 0:
			return nil, nil, nil
		case 1:
			return v[0], nil, nil
		case 2:
			return v[0], v[1], nil
		default:
			return nil, nil, errors.New("size isnot match")
		}
	case []uint64:
		switch len(v) {
		case 0:
			return nil, nil, nil
		case 1:
			return v[0], nil, nil
		case 2:
			return v[0], v[1], nil
		default:
			return nil, nil, errors.New("size isnot match")
		}

	case []interface{}:
		switch len(v) {
		case 0:
			return nil, nil, nil
		case 1:
			return v[0], nil, nil
		case 2:
			return v[0], v[1], nil
		default:
			return nil, nil, errors.New("size isnot match")
		}
	}

	rv := reflect.ValueOf(o)
	if !rv.IsValid() {
		return nil, nil, errors.New("type is invalid")
	}
	if rv.Kind() == reflect.Slice {
		length := rv.Len()
		switch length {
		case 0:
			return nil, nil, nil
		case 1:
			return toActualValue(rv.Index(0).Interface()), nil, nil
		case 2:
			return toActualValue(rv.Index(0).Interface()), toActualValue(rv.Index(1).Interface()), nil
		}
		return nil, nil, errors.New("size isnot match")
	}

	if rv.Kind() == reflect.Ptr {
		rv = rv.Elem()
	}

	if rv.Kind() == reflect.Struct {
		startField := rv.FieldByName("Start")
		endField := rv.FieldByName("End")

		if startField.IsValid() && endField.IsValid() {
			return startField.Interface(), endField.Interface(), nil
		}
	}
	return nil, nil, errors.New("unsupport type '" + rv.Type().Name() + "'")
}

func toActualValue(a interface{}) interface{} {
	switch v := a.(type) {
	case sql.NullInt64:
		if v.Valid {
			return v.Int64
		}
		return nil
	case sql.NullString:
		if v.Valid {
			return v.String
		}
		return nil
	case sql.NullBool:
		if v.Valid {
			return v.Bool
		}
		return nil
	case sql.NullTime:
		if v.Valid {
			return v.Time
		}
		return nil
		// case sql.NullByte:
		// 	if v.Valid {
		// 		return v.Byte
		// 	}
		// 	return nil
	}
	return a
}

func int64With(v interface{}, defaultValue int64) int64 {
	if v == nil {
		return defaultValue
	}
	switch value := v.(type) {
	case int:
		return int64(value)
	case int64:
		return value
	case int8:
		return int64(value)
	case int16:
		return int64(value)
	case int32:
		return int64(value)
	case uint:
		return int64(value)
	case uint64:
		return int64(value)
	case uint8:
		return int64(value)
	case uint16:
		return int64(value)
	case uint32:
		return int64(value)
	case float32:
		return int64(value)
	case float64:
		return int64(value)
	default:
		return defaultValue
	}
}
