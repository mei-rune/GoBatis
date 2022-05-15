package dialects

import (
	"strconv"
	"strings"
)

// 这个本项目没有使用，是我公司项目用到了，稍后移走
func NewPrinter(dialect Dialect) *Printer {
	return &Printer{
		Dialect: dialect,
	}
}

// 这个本项目没有使用，是我公司项目用到了，稍后移走
type Checkpoint struct {
	dialect Dialect
	sql     string
	args    []interface{}
	index   int
	err     error

	hasExpr bool
}

// 这个本项目没有使用，是我公司项目用到了，稍后移走
type Printer struct {
	Dialect Dialect
	out     strings.Builder
	args    []interface{}
	index   int
	err     error

	hasExpr   bool
	pgversion int
}

func (p *Printer) IsPg10Compatible() bool {
	return p.pgversion <= 10
}

func (p *Printer) SetPostgresqlVersion(version int) {
	p.pgversion = version
}

func (p *Printer) Reset() {
	p.out.Reset()
	p.args = nil
	p.index = 0
	p.err = nil
	p.hasExpr = false
}

var emptyArgs = []interface{}{}

func (p *Printer) ToSQL() (string, []interface{}, error) {
	if p.err != nil {
		return "", nil, p.err
	}
	args := p.args
	if args == nil {
		args = emptyArgs
	}
	return p.out.String(), args, nil
}

func (p *Printer) HasExpression() bool {
	return p.hasExpr
}

func (p *Printer) SetExpression(hasExpr bool) bool {
	old := p.hasExpr
	p.hasExpr = hasExpr
	return old
}

func (p *Printer) NewExpression(ss ...string) {
	if p.hasExpr {
		p.WriteString(" AND ")
	} else {
		p.hasExpr = true
		p.WriteString(" WHERE ")
	}

	for _, txt := range ss {
		p.WriteString(txt)
	}
}

func (p *Printer) HasError() bool {
	return p.err != nil
}

func (p *Printer) SetError(err error) error {
	e := p.err
	p.err = err
	if e == nil {
		return err
	}
	return e
}

func (p *Printer) Err() error {
	return p.err
}

func (p *Printer) Checkpoint() Checkpoint {
	return Checkpoint{
		dialect: p.Dialect,
		sql:     p.out.String(),
		args:    p.args,
		index:   p.index,
		err:     p.err,
		hasExpr: p.hasExpr,
	}
}

func (p *Printer) RollbackTo(cp Checkpoint) {
	p.out.Reset()

	p.Dialect = cp.dialect
	p.args = cp.args
	p.index = cp.index
	p.err = cp.err
	p.hasExpr = cp.hasExpr
	p.out.WriteString(cp.sql)
}

func (p *Printer) AddParam(value interface{}) error {
	if p.err != nil {
		return p.err
	}

	if p.index < 0 {
		p.err = p.WriteString("?")
		if p.err != nil {
			return p.err
		}
		p.args = append(p.args, value)
		return nil
	}

	err := p.WriteString(p.Dialect.Placeholder().Format(p.index))
	if err != nil {
		return err
	}
	p.index++
	p.args = append(p.args, value)
	return nil
}

func (p *Printer) WriteString(s string, ss ...string) error {
	if p.err != nil {
		return p.err
	}
	_, p.err = p.out.WriteString(s)
	if p.err != nil {
		return p.err
	}
	for _, txt := range ss {
		_, p.err = p.out.WriteString(txt)
		if p.err != nil {
			return p.err
		}
	}
	return p.err
}

func (p *Printer) OrderBy(prefix, orderBy string) error {
	if orderBy == "" {
		return nil
	}

	if strings.HasPrefix(orderBy, "+") {
		orderBy = strings.TrimPrefix(orderBy, "+") + " ASC"
	} else if strings.HasPrefix(orderBy, "-") {
		orderBy = strings.TrimPrefix(orderBy, "-") + " DESC"
	}

	if strings.ContainsRune(orderBy, '$') {
		p.WriteString(" ORDER BY ")
		p.WriteString(strings.Replace(orderBy, "${prefix}", prefix, -1))
	} else if prefix != "" {
		p.WriteString(" ORDER BY ")
		p.WriteString(prefix)
		p.WriteString(".")
		p.WriteString(orderBy)
	} else {
		p.WriteString(" ORDER BY ")
		p.WriteString(orderBy)
	}
	return nil
}

func (p *Printer) GroupBy(prefix, groupBy string) error {
	if groupBy == "" {
		return nil
	}

	if strings.ContainsRune(groupBy, '$') {
		p.WriteString(" GROUP BY ")
		p.WriteString(strings.Replace(groupBy, "${prefix}", prefix, -1))
	} else {
		p.WriteString(" GROUP BY ")
		p.WriteString(prefix)
		p.WriteString(".")
		p.WriteString(groupBy)
	}
	return nil
}

func (p *Printer) Limit(offset, limit int64) error {
	if offset > 0 {
		p.WriteString(" OFFSET ")
		p.WriteString(strconv.FormatInt(offset, 10))
	}
	if limit > 0 {
		p.WriteString(" LIMIT ")
		p.WriteString(strconv.FormatInt(limit, 10))
	}
	return nil
}
