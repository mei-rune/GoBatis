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
)

type sqlPrinter struct {
	sb     strings.Builder
	ctx    *Context
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
	panic("这个不能作为 sql 或 sql 片段")
}

func toElse(s SqlExpression) (elseExpression, bool) {
	expr, ok := s.(elseExpression)
	return expr, ok
}

type SqlExpression interface {
	String() string
	writeTo(printer *sqlPrinter)
}

func newRawExpression(content string) (SqlExpression, error) {
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
	// isPrint bool
	ctx *Context
}

func (eval evalParameters) Get(name string) (interface{}, error) {
	value, err := eval.ctx.Get(name)
	// if eval.isPrint {
	// 	fmt.Println("===================", name, value, err)
	// }

	if err == nil {
		return value, nil
	}
	if err == ErrNotFound {
		return nil, nil
	}
	return nil, err
}

type TestGetter interface {
	Get(name string) (interface{}, error)
}

type Testable interface {
	Test(TestGetter) (bool, error)
	String() string
}

type ifExpression struct {
	test                            Testable
	trueExpression, falseExpression SqlExpression
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
	bResult, err := ifExpr.test.Test(evalParameters{ctx: printer.ctx})
	if err != nil {
		printer.err = errors.New("execute `" + ifExpr.test.String() + "` fail, " + err.Error())
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

func newIFExpression(test string, segements []SqlExpression) (SqlExpression, error) {
	if test == "" {
		return nil, errors.New("if test is empty")
	}
	if len(segements) == 0 {
		return nil, errors.New("if content is empty")
	}

	expr, err := ParseEvaluableExpression(test)
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

	var trueExprs, falseExprs []SqlExpression
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
	otherwise SqlExpression
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
		bResult, err := chose.when[idx].test.Test(evalParameters{ctx: printer.ctx})
		if err != nil {
			printer.err = errors.New("execute `" + chose.when[idx].test.String() + "` fail, " + err.Error())
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
	test       Testable
	expression SqlExpression
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

func newChoseExpression(el xmlChoseElement) (SqlExpression, error) {
	var when []whenExpression

	for idx := range el.when {
		if el.when[idx].test == "" {
			return nil, errors.New("when test is empty")
		}
		if el.when[idx].content == nil {
			return nil, errors.New("when content is empty")
		}

		expr, err := ParseEvaluableExpression(el.when[idx].test)
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
	segements []SqlExpression
}

func (foreach *forEachExpression) String() string {
	return foreach.el.String()
}

func (foreach *forEachExpression) execOne(printer *sqlPrinter, key, value interface{}) {
	// oldParams := printer.params
	// oldErr := printer.err
	oldFinder := printer.ctx.finder

	// newPrinter := printer.Clone()
	// ctx := *printer.ctx
	// newPrinter.ctx = &ctx

	printer.ctx.finder = &kvFinder{
		Parameters:  oldFinder,
		mapper:      printer.ctx.Mapper,
		paramNames:  []string{foreach.el.item, foreach.el.index},
		paramValues: []interface{}{value, key},
	}

	for idx := range foreach.segements {
		foreach.segements[idx].writeTo(printer)
		if printer.err != nil {
			break
		}
	}

	printer.ctx.finder = oldFinder
	// printer.sb.WriteString(newPrinter.sb.String())
	// printer.params = newPrinter.params
	// printer.err = newPrinter.err
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
	case []string:
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

	case map[string]string:
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

func newForEachExpression(el xmlForEachElement) (SqlExpression, error) {
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

type expressionArray []SqlExpression

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
		inStr = ` inStr="` + expr.inStr + `"`
	}
	return expr.prefix + `<print fmt="` + expr.fmt + `" value="` + expr.value + `"` + inStr + ` />` + expr.suffix
}

func (expr printExpression) writeTo(printer *sqlPrinter) {
	value, err := printer.ctx.Get(expr.value)
	if err != nil {
		printer.err = errors.New("search '" + expr.value + "' fail, " + err.Error())
	} else if value == nil {
		printer.err = errors.New("'" + expr.value + "' isnot found")
	} else {
		printer.sb.WriteString(expr.prefix)

		inStr := strings.ToLower(expr.inStr) == "true"
		var s string
		if expr.fmt != "" {
			s = fmt.Sprintf(expr.fmt, value)
			value = s
		} else {
			s = fmt.Sprint(value)
		}
		err := isValidPrintValue(value, inStr)
		if err != nil {
			printer.err = err
			return
		}
		printer.sb.WriteString(s)

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
		return errors.New("print value is invalid type - " + fmt.Sprintf("%T", value))
	}
}

var ErrInvalidPrintValue = errors.New("print value is invalid")

func isValidPrintString(value string, inStr bool) (string, error) {
	runes := []rune(value)
	if !inStr {
		return value, validSqlFieldString(runes)
	}

	for _, r := range runes {
		switch r {
		case '\'', '"':
			return "", ErrInvalidPrintValue
		default:
			if !inStr {
				if unicode.IsSpace(r) {
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

func skipBy(runes []rune, fn func(r rune) bool) []rune {
	for idx := range runes {
		if !fn(runes[idx]) {
			return runes[idx:]
		}
	}
	return nil
}
func isAlphabet(r rune) bool {
	return r >= 'a' && r <= 'z' ||
		r >= 'A' && r <= 'Z' || r == '_'
}
func isAlphabetOrDigit(r rune) bool {
	return r >= 'a' && r <= 'z' ||
		r >= 'A' && r <= 'Z' || r == '_' ||
		r >= '0' && r <= '9'
}
func isNotExceptedRune(excepted rune) func(r rune) bool {
	return func(r rune) bool {
		return r != excepted
	}
}
func validSqlFieldString(runes []rune) error {
	runes = skipBy(runes, isAlphabet)
	if len(runes) == 0 {
		return nil
	}
	runes = skipBy(runes, isAlphabetOrDigit)
	if len(runes) == 0 {
		return nil
	}

	for {
		c := runes[0]
		switch c {
		case '.':
			return validSqlFieldString(runes[1:])
		case '-':
			if len(runes) <= 2 {
				return ErrInvalidPrintValue
			}

			if runes[1] == '>' {
				runes = runes[2:]
				if runes[0] == '>' {
					runes = runes[1:]
					if len(runes) == 0 {
						return ErrInvalidPrintValue
					}
				}
				return validSqlFieldString(runes)
			}
			return ErrInvalidPrintValue
		case '\'', '"':
			runes = runes[1:]
			runes = skipBy(runes, isNotExceptedRune(c))
			if len(runes) == 0 {
				return ErrInvalidPrintValue
			}

			if runes[0] != c {
				return ErrInvalidPrintValue
			}
			runes = runes[1:]
			if len(runes) == 0 {
				return nil
			}

			// 加上下面几句意味着 aaa'a'aaa 也是合法的
			runes = skipBy(runes, isAlphabet)
			if len(runes) == 0 {
				return nil
			}
			runes = skipBy(runes, isAlphabetOrDigit)
			if len(runes) == 0 {
				return nil
			}
		default:
			return ErrInvalidPrintValue
		}
	}
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
		s = ` isPrefix="` + expr.isPrefix + `"`
	}
	if expr.isSuffix != "" {
		s += ` isSuffix="` + expr.isSuffix + `"`
	}
	return expr.prefix + `<like value="` + expr.value + `"` + s + ` />` + expr.suffix
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
			appended := false
			if strings.ToLower(expr.isPrefix) == "true" {
				printer.addPlaceholderAndParam(s + "%")
				appended = true
			}
			if strings.ToLower(expr.isSuffix) == "true" {
				printer.addPlaceholderAndParam("%" + s)
				appended = true
			}
			if !appended {
				printer.addPlaceholderAndParam("%" + s + "%")
			}
		}
		printer.sb.WriteString(expr.suffix)
	}
}

type limitExpression struct {
	offset string
	offsetValid bool
	offsetInt int64
	limit  string
	limitValid bool
	limitInt int64
}

func (expr limitExpression) String() string {
	return `<pagination offset="` + expr.offset + `" limit="` + expr.limit + `" />`
}

func (expr limitExpression) writeTo(printer *sqlPrinter) {
	var offset = expr.offsetInt
	if !expr.offsetValid {
		o, _ := printer.ctx.Get(expr.offset)
		offset = int64With(o, 0)
	}

	var limit = expr.limitInt
	if !expr.limitValid {
		o, _ := printer.ctx.Get(expr.limit)
		limit = int64With(o, 0)
	}

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

	prefix    string
	direction string
}

func (expr orderByExpression) String() string {
	if expr.sort == "" {
		return "<order_by />"
	}
	return `<order_by prefix="` + expr.prefix + `" sort="` + expr.sort + `" direction="` + expr.direction + `"/>`
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
			s = strings.TrimPrefix(s, "+")
			s = strings.TrimSuffix(s, ",")
			if err := checkOrderBy(s); err != nil {
					printer.err = errors.New("order by '" + s + "' is invalid value, " + err.Error())
					return
			}
			printer.sb.WriteString(s)
			printer.sb.WriteString(" ASC")
		} else if strings.HasPrefix(s, "-") {
			s = strings.TrimPrefix(s, "-")
			s = strings.TrimSuffix(s, ",")
			if err := checkOrderBy(s); err != nil {
					printer.err = errors.New("order by '" + s + "' is invalid value, " + err.Error())
					return
			}
			printer.sb.WriteString(s)
			printer.sb.WriteString(" DESC")
		} else {
			s = strings.TrimSuffix(s, ",")
			if err := checkOrderBy(s); err != nil {
					printer.err = errors.New("order by '" + s + "' is invalid value, " + err.Error())
					return
			}
			printer.sb.WriteString(s)
		}
		if expr.direction != "" {
			printer.sb.WriteString(" ")
			if err := checkOrderByDirection(s); err != nil {
				printer.err = errors.New("order by '" + s + "' is invalid value, " + err.Error())
				return
			}
			printer.sb.WriteString(expr.direction)
		}
	}
}

func checkOrderBy(s string) error {
	if strings.Contains(s, ";") {
		return errors.New("invalid field name")
	}
	return nil
}

func checkOrderByDirection(s string) error {
	s = strings.ToLower(s)
	if s != "asc" && s != "desc" {
		return errors.New("invalid direction")
	}
	return nil
}

type trimExpression struct {
	expressions    expressionArray
	prefixoverride []string
	prefix         SqlExpression
	suffixoverride []string
	suffix         SqlExpression
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

		idx := strings.IndexFunc(s, func(r rune) bool {
			return !unicode.IsSpace(r)
		})

		for _, prefix := range expr.prefixoverride {
			if strings.HasPrefix(s[idx:], prefix) {
				return true, s[:idx] + prefix
			}
		}
	}
	return false, ""
}

func (expr trimExpression) hasSuffix(s string) (bool, string) {
	if len(expr.suffixoverride) > 0 {
		s = strings.ToLower(s)

		idx := strings.LastIndexFunc(s, func(r rune) bool {
			return !unicode.IsSpace(r)
		})

		for _, suffix := range expr.suffixoverride {
			if strings.HasSuffix(s[:idx+1], suffix) {
				return true, suffix + s[idx+1:]
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

	prefix SqlExpression
	suffix SqlExpression
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
		typ := rv.Type()
		startField := rv.FieldByName("Start")
		startType, _ := typ.FieldByName("Start")
		endField := rv.FieldByName("End")
		endType, _ := typ.FieldByName("End")

		if startField.IsValid() && endField.IsValid() {
			return toValueFromReflectValue(startType, startField),
				toValueFromReflectValue(endType, endField), nil
		}
	}
	return nil, nil, errors.New("unsupport type '" + rv.Type().Name() + "'")
}

func toValueFromReflectValue(t reflect.StructField, a reflect.Value) interface{} {
	if t.Type.Kind() == reflect.Ptr && a.IsNil() {
		return nil
	}

	jsonTag := t.Tag.Get("json")
	if !strings.Contains(jsonTag, ",omitempty") {
		return a.Interface()
	}

	if a.IsZero() {
		return nil
	}

	value := a.Interface()
	if tt, ok := value.(interface {
		IsZero() bool
	}); ok && tt.IsZero() {
		return nil
	}
	return value
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

func newIncludeExpression(refid string, propertyArray [][2]string,
	findSqlFragment func(string) (SqlExpression, error)) (SqlExpression, error) {

	var findFragment func(Parameters) (SqlExpression, error)
	if (strings.HasPrefix(refid, "${") || strings.HasPrefix(refid, "#{")) && strings.HasSuffix(refid, "}") {
		id := strings.TrimPrefix(refid, "${")
		id = strings.TrimPrefix(id, "#{")
		id = strings.TrimSuffix(id, "}")
		findFragment = func(parameters Parameters) (SqlExpression, error) {
			value, err := parameters.Get(id)
			if err != nil {
				return nil, errors.New("element include.refid '" + id + "' value not found, " + err.Error())
			}

			refExpr, err := findSqlFragment(fmt.Sprint(value))
			if err != nil {
				return nil, errors.New("element include.refid '" + id + "' invalid, " + err.Error())
			}
			return refExpr, nil
		}
	} else {
		refExpr, err := findSqlFragment(refid)
		if err != nil {
			return nil, errors.New("element include.refid '" + refid + "' invalid, " + err.Error())
		}
		findFragment = func(Parameters) (SqlExpression, error) {
			return refExpr, nil
		}
	}
	return includeExpression{
		refid:         refid,
		findFragment:  findFragment,
		propertyArray: propertyArray,
	}, nil
}

type includeExpression struct {
	refid         string
	findFragment  func(Parameters) (SqlExpression, error)
	propertyArray [][2]string
}

func (expr includeExpression) String() string {
	if len(expr.propertyArray) > 0 {
		var sb strings.Builder
		sb.WriteString(`<include refid="`)
		sb.WriteString(expr.refid)
		sb.WriteString(`">`)

		for _, a := range expr.propertyArray {
			sb.WriteString(`<property name="`)
			sb.WriteString(a[0])
			sb.WriteString(`" value="`)
			sb.WriteString(a[1])
			sb.WriteString(`" />`)
		}

		sb.WriteString(`</include>`)
		return sb.String()
	}
	return `<include refid="` + expr.refid + `" />`
}

func (expr includeExpression) writeTo(printer *sqlPrinter) {
	// oldParams := printer.params
	// oldErr := printer.err
	oldFinder := printer.ctx.finder

	refExpr, err := expr.findFragment(oldFinder)
	if err != nil {
		printer.err = err
		return
	}

	if len(expr.propertyArray) == 0 {
		refExpr.writeTo(printer)
		return
	}

	// newPrinter := printer.Clone()
	// ctx := *printer.ctx
	// newPrinter.ctx = &ctx

	var values = map[string]func(name string) (interface{}, error){}

	for _, a := range expr.propertyArray {
		value := a[1]
		if value == "true" {
			values[a[0]] = func(name string) (interface{}, error) {
				return true, nil
			}
		} else if value == "false" {
			values[a[0]] = func(name string) (interface{}, error) {
				return false, nil
			}
		} else if (strings.HasPrefix(value, "${") || strings.HasPrefix(value, "#{")) && strings.HasSuffix(value, "}") {
			value = strings.TrimPrefix(value, "${")
			value = strings.TrimPrefix(value, "#{")
			value = strings.TrimSuffix(value, "}")
			values[a[0]] = func(name string) (interface{}, error) {
				return oldFinder.Get(value)
			}
		} else {
			if i64, err := strconv.ParseInt(value, 10, 64); err == nil {
				values[a[0]] = func(name string) (interface{}, error) {
					return i64, nil
				}
			} else if u64, err := strconv.ParseUint(value, 10, 64); err == nil {
				values[a[0]] = func(name string) (interface{}, error) {
					return u64, nil
				}
			} else if f64, err := strconv.ParseFloat(value, 64); err == nil {
				values[a[0]] = func(name string) (interface{}, error) {
					return f64, nil
				}
			} else {
				values[a[0]] = func(name string) (interface{}, error) {
					return value, nil
				}
			}
		}
	}

	printer.ctx.finder = nestParameters{
		Parameters: oldFinder,
		values:     values,
	}

	refExpr.writeTo(printer)

	printer.ctx.finder = oldFinder
	// printer.sb.WriteString(newPrinter.sb.String())
	// printer.params = newPrinter.params
	// printer.err = newPrinter.err
}

type nestParameters struct {
	Parameters

	values map[string]func(name string) (interface{}, error)
}

func (s nestParameters) RValue(dialect Dialect, param *Param) (interface{}, error) {
	get, ok := s.values[param.Name]
	if ok {
		value, err := get(param.Name)
		if err != nil {
			return nil, err
		}
		return toSQLType(dialect, param, value)
	}
	return s.Parameters.RValue(dialect, param)
}

func (s nestParameters) Get(name string) (interface{}, error) {
	get, ok := s.values[name]
	if ok {
		return get(name)
	}
	return s.Parameters.Get(name)
}
