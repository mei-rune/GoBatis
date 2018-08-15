package gobatis

import (
	"errors"
	"strings"
)

func NewMultipleArray() *MultipleArray {
	return &MultipleArray{}
}

type Multiple struct {
	Names   []string
	Returns []interface{}

	positions []int
	fields    []string
}

func (m *Multiple) Set(name string, ret interface{}) {
	m.Names = append(m.Names, name)
	m.Returns = append(m.Returns, ret)
}

func (m *Multiple) setColumns(columns []string) (err error) {
	m.positions, m.fields, err = indexColumns(columns, m.Names)
	return err
}

func (m *Multiple) Scan(dialect Dialect, mapper *Mapper, r colScanner, isUnsafe bool) error {
	columns, err := r.Columns()
	if err != nil {
		return err
	}
	if err = m.setColumns(columns); err != nil {
		return err
	}

	return m.scan(dialect, mapper, r, isUnsafe)
}

func (m *Multiple) scan(dialect Dialect, mapper *Mapper, r colScanner, isUnsafe bool) error {
	return errors.New("not implemented")
}

func NewMultiple() *Multiple {
	return &Multiple{}
}

type MultipleArray struct {
	Names   []string
	NewRows []func(int) interface{}
	Index   int

	multiple Multiple
}

func (m *MultipleArray) Set(name string, newRows func(int) interface{}) {
	m.Names = append(m.Names, name)
	m.NewRows = append(m.NewRows, newRows)
}

func (m *MultipleArray) setColumns(columns []string) error {
	m.multiple.Names = m.Names
	return m.multiple.setColumns(columns)
}

func (m *MultipleArray) Next() *Multiple {
	for idx := range m.NewRows {
		m.multiple.Returns = append(m.multiple.Returns, m.NewRows[idx](m.Index))
	}
	m.Index++
	return &m.multiple
}

func (m *MultipleArray) Scan(dialect Dialect, mapper *Mapper, r rowsi, isUnsafe bool) error {
	columns, err := r.Columns()
	if err != nil {
		return err
	}
	if err = m.setColumns(columns); err != nil {
		return err
	}

	for r.Next() {
		multiple := m.Next()
		err = multiple.Scan(dialect, mapper, r, isUnsafe)
		if err != nil {
			return err
		}
	}
	return r.Err()
}

func indexColumns(columns, names []string) ([]int, []string, error) {
	results := make([]int, len(columns))
	fields := make([]string, len(columns))
	for idx, column := range columns {
		tagName := column
		position := strings.IndexByte(column, '.')
		if position >= 0 {
			tagName = column[:position]
			fields[idx] = column[position+1:]
		}

		foundIndex := -1
		for nameIdx, name := range names {
			if name == tagName {
				foundIndex = nameIdx
				break
			}
		}
		if foundIndex < 0 {
			return nil, fields, errors.New("column '" + column + "' isnot exists in the names - " + strings.Join(names, ","))
		}

		results[idx] = foundIndex
	}
	return results, fields, nil
}
