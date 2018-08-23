package gobatis

import (
	"errors"
	"fmt"
	"reflect"
	"strings"
	"unicode"

	"github.com/Knetic/govaluate"
)

type sqlPrinter struct {
	ctx    *Context
	sb     strings.Builder
	params []interface{}
	err    error
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
	return "<if test=\"" + ifExpr.test.String() + "\">" + ifExpr.String() + "</if>"
}

func (ifExpr ifExpression) writeTo(printer *sqlPrinter) {
	result, err := ifExpr.test.Eval(evalParameters{ctx: printer.ctx})
	if err != nil {
		printer.err = err
		return
	}

	if result == nil {
		printer.err = errors.New("result of if expression  is nil - " + ifExpr.String())
		return
	}

	bResult, ok := result.(bool)
	if !ok {
		printer.err = errors.New("result of if expression isnot bool got " + fmt.Sprintf("%T", result) + " - " + ifExpr.String())
		return
	}
	if bResult {
		ifExpr.segement.writeTo(printer)
	}
}

func newIFExpression(test, content string) (sqlExpression, error) {
	if test == "" {
		return nil, errors.New("if test is empty")
	}
	if content == "" {
		return nil, errors.New("if content is emtpy")
	}
	expr, err := govaluate.NewEvaluableExpression(test)
	if err != nil {
		return nil, err
	}

	segement, err := newRawExpression(content)
	if err != nil {
		return nil, err
	}

	return ifExpression{test: expr, segement: segement}, nil
}

type choseExpression struct {
	el xmlChoseElement

	when      []sqlExpression
	otherwise sqlExpression
}

func (chose *choseExpression) String() string {
	return chose.el.String()
}

func (chose *choseExpression) writeTo(printer *sqlPrinter) {
	oldLen := printer.sb.Len()
	for idx := range chose.when {
		chose.when[idx].writeTo(printer)
		if oldLen != printer.sb.Len() {
			return
		}
	}

	if chose.otherwise != nil {
		chose.otherwise.writeTo(printer)
	}
}

func newChoseExpression(el xmlChoseElement) (sqlExpression, error) {
	var when []sqlExpression
	var otherwise sqlExpression

	for idx := range el.when {
		s, err := newIFExpression(el.when[idx].test, el.when[idx].content)
		if err != nil {
			return nil, err
		}

		when = append(when, s)
	}

	if el.otherwise != "" {
		other, err := newRawExpression(el.otherwise)
		if err != nil {
			return nil, err
		}
		otherwise = other
	}
	return &choseExpression{
		el:        el,
		when:      when,
		otherwise: otherwise,
	}, nil
}

type forEachExpression struct {
	el       xmlForEachElement
	segement sqlExpression
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
	foreach.segement.writeTo(newPrinter)

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
	if el.content == "" {
		return nil, errors.New("conent of foreach is empty")
	}

	if el.collection == "" {
		return nil, errors.New("collection of foreach is empty")
	}

	if el.index == "" && el.item == "" {
		return nil, errors.New("index,item of foreach is empty")
	}

	segement, err := newRawExpression(el.content)
	if err != nil {
		return nil, err
	}

	return &forEachExpression{el: el, segement: segement}, nil
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
	printer.sb.WriteString(" WHERE ")
	oldLen := printer.sb.Len()
	where.expressions.writeTo(printer)
	if printer.sb.Len() == oldLen {
		printer.sb.Reset()
		printer.sb.WriteString(old)
	} else {
		s := strings.TrimSpace(printer.sb.String())
		if len(s) < 4 {
			return
		}

		c := s[len(s)-1]
		b := s[len(s)-2]
		a := s[len(s)-3]
		w := s[len(s)-4]

		if c == 'd' || c == 'D' {
			if (b == 'n' || b == 'N') && (a == 'a' || a == 'A') && unicode.IsSpace(rune(w)) {
				printer.sb.Reset()
				printer.sb.WriteString(s[:len(s)-3])
			}
		} else if c == 'r' || c == 'R' {
			if (b == 'o' || b == 'O') && unicode.IsSpace(rune(a)) {
				printer.sb.Reset()
				printer.sb.WriteString(s[:len(s)-2])
			}
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

func (expressions expressionArray) writeTo(printer *sqlPrinter) {
	for idx := range expressions {
		expressions[idx].writeTo(printer)
	}
}

func (expressions expressionArray) GenerateSQL(ctx *Context) (string, []interface{}, error) {
	var printer = &sqlPrinter{
		ctx: ctx,
	}
	expressions.writeTo(printer)
	return printer.sb.String(), printer.params, printer.err
}
