package gobatis

import (
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"strings"
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

		return strings.HasPrefix(args[0].(string), args[1].(string)), nil
	},
	"hasSuffix": func(args ...interface{}) (interface{}, error) {
		if len(args) != 2 {
			return nil, errors.New("hasSuffix args is invalid")
		}

		return strings.HasSuffix(args[0].(string), args[1].(string)), nil
	},
	"trimPrefix": func(args ...interface{}) (interface{}, error) {
		if len(args) != 2 {
			return nil, errors.New("hasSuffix args is invalid")
		}

		return strings.TrimPrefix(args[0].(string), args[1].(string)), nil
	},
	"trimSuffix": func(args ...interface{}) (interface{}, error) {
		if len(args) != 2 {
			return nil, errors.New("hasSuffix args is invalid")
		}

		return strings.TrimSuffix(args[0].(string), args[1].(string)), nil
	},
	"trimSpace": func(args ...interface{}) (interface{}, error) {
		if len(args) != 1 {
			return nil, errors.New("hasSuffix args is invalid")
		}
		return strings.TrimSpace(args[0].(string)), nil
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
	sql := printer.ctx.Dialect.Placeholder().Concat([]string{"", ""}, nil, len(printer.params))
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

type sqlExpression interface {
	String() string
	writeTo(printer *sqlPrinter)
}

func newRawExpression(content string) (sqlExpression, error) {
	fragments, bindParams, err := compileNamedQuery(content)
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
	sql := printer.ctx.Dialect.Placeholder().Concat(stmt.fragments, stmt.bindParams, len(printer.params))
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
	test     *govaluate.EvaluableExpression
	segement sqlExpression
}

func (ifExpr ifExpression) String() string {
	return "<if test=\"" + ifExpr.test.String() + "\">" + ifExpr.segement.String() + "</if>"
}

func (ifExpr ifExpression) writeTo(printer *sqlPrinter) {
	bResult, err := ifExpr.isOK(printer)
	if err != nil {
		printer.err = err
		return
	}

	if bResult {
		ifExpr.segement.writeTo(printer)
	}
}

func (ifExpr ifExpression) isOK(printer *sqlPrinter) (bool, error) {
	result, err := ifExpr.test.Eval(evalParameters{ctx: printer.ctx})
	if err != nil {
		return false, err
	}

	if result == nil {
		return false, errors.New("result of if expression  is nil - " + ifExpr.String())
	}

	bResult, ok := result.(bool)
	if !ok {
		return false, errors.New("result of if expression isnot bool got " + fmt.Sprintf("%T", result) + " - " + ifExpr.String())
	}

	return bResult, nil
}

func newIFExpression(test string, segement sqlExpression) (sqlExpression, error) {
	if test == "" {
		return nil, errors.New("if test is empty")
	}
	if segement == nil {
		return nil, errors.New("if content is empty")
	}
	expr, err := govaluate.NewEvaluableExpressionWithFunctions(test, expFunctions)
	if err != nil {
		return nil, errors.New("expression '" + test + "' is invalid: " + err.Error())
	}
	return ifExpression{test: expr, segement: segement}, nil
}

type choseExpression struct {
	el xmlChoseElement

	when      []ifExpression
	otherwise sqlExpression
}

func (chose *choseExpression) String() string {
	return chose.el.String()
}

func (chose *choseExpression) writeTo(printer *sqlPrinter) {
	for idx := range chose.when {
		bResult, err := chose.when[idx].isOK(printer)
		if err != nil {
			printer.err = err
			return
		}

		if bResult {
			chose.when[idx].segement.writeTo(printer)
			return
		}
	}

	if chose.otherwise != nil {
		chose.otherwise.writeTo(printer)
	}
}

func newChoseExpression(el xmlChoseElement) (sqlExpression, error) {
	var when []ifExpression

	for idx := range el.when {
		s, err := newIFExpression(el.when[idx].test, el.when[idx].content)
		if err != nil {
			return nil, err
		}

		when = append(when, s.(ifExpression))
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

		var start = 0
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

		var end = 0
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
	var printer = &sqlPrinter{
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
}

func (expr printExpression) String() string {
	if expr.fmt == "" {
		return expr.prefix + `<print value="` + expr.value + `" />` + expr.suffix
	}
	return expr.prefix + `<print fmt="` + expr.fmt + `" value="` + expr.value + `" />` + expr.suffix
}

func (expr printExpression) writeTo(printer *sqlPrinter) {
	value, err := printer.ctx.Get(expr.value)
	if err != nil {
		printer.err = errors.New("search '" + expr.value + "' fail, " + err.Error())
	} else if value == nil {
		printer.err = errors.New("'" + expr.value + "' isnot found")
	} else if expr.fmt == "" {
		printer.sb.WriteString(expr.prefix)
		printer.sb.WriteString(fmt.Sprint(value))
		printer.sb.WriteString(expr.suffix)
	} else {
		printer.sb.WriteString(expr.prefix)
		printer.sb.WriteString(fmt.Sprintf(expr.fmt, value))
		printer.sb.WriteString(expr.suffix)
	}
}

type likeExpression struct {
	prefix string
	suffix string
	value  string
}

func (expr likeExpression) String() string {
	return expr.prefix + `<like value="` + expr.value + `" />` + expr.suffix
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

		printer.sb.WriteString(expr.prefix)
		if strings.HasPrefix(s, "%") || strings.HasSuffix(s, "%") {
			printer.addPlaceholderAndParam(s)
		} else {
			printer.addPlaceholderAndParam("%" + s + "%")
		}
		printer.sb.WriteString(expr.suffix)
	}
}

type paginationExpression struct {
	offset string
	limit  string
}

func (expr paginationExpression) String() string {
	return `<pagination offset="` + expr.offset + `" limit="` + expr.limit + `" />`
}

func (expr paginationExpression) writeTo(printer *sqlPrinter) {
	o, _ := printer.ctx.Get(expr.offset)
	offset := int64With(o, 0)
	o, _ = printer.ctx.Get(expr.limit)
	limit := int64With(o, 0)

	s, args := printer.ctx.Dialect.GeneratePagination(offset, limit)
	printer.sb.WriteString(s)
	if len(args) > 0 {
		printer.params = append(printer.params, args...)
	}
}

type orderByExpression struct {
	sort string
}

func (expr orderByExpression) String() string {
	if expr.sort == "" {
		return "<order_by />"
	}
	return `<order_by sort="` + expr.sort + `"/>`
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

	s := fmt.Sprint(o)
	if s == "" {
		return
	}

	printer.sb.WriteString(" ORDER BY ")
	if strings.HasPrefix(s, "+") {
		printer.sb.WriteString(strings.TrimPrefix(s, "+"))
		printer.sb.WriteString(" ASC")
	} else if strings.HasPrefix(s, "-") {
		printer.sb.WriteString(strings.TrimPrefix(s, "-"))
		printer.sb.WriteString(" DESC")
	} else {
		printer.sb.WriteString(s)
	}
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
