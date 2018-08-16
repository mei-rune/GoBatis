package gobatis

import (
	"errors"
	"reflect"
	"strings"

	"github.com/runner-mei/GoBatis/reflectx"
)

func NewMultipleArray() *MultipleArray {
	return &MultipleArray{}
}

type Multiple struct {
	Names   []string
	Returns []interface{}

	delimiter  string
	columns    []string
	positions  []int
	fields     []string
	traversals []*FieldInfo
}

func (m *Multiple) SetDelimiter(delimiter string) {
	m.delimiter = delimiter
}

func (m *Multiple) Set(name string, ret interface{}) {
	m.Names = append(m.Names, name)
	m.Returns = append(m.Returns, ret)
}

func (m *Multiple) setColumns(columns []string) (err error) {
	m.columns = columns
	m.positions, m.fields, err = indexColumns(columns, m.Names, m.delimiter)
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

	if err = m.ensureTraversals(mapper); err != nil {
		return err
	}

	return m.scan(dialect, mapper, r, isUnsafe)
}

func (m *Multiple) ensureTraversals(mapper *Mapper) error {
	if len(m.traversals) != 0 {
		return nil
	}

	var traversals = make([]*FieldInfo, len(m.columns))
	for idx := range m.columns {
		vp := m.Returns[m.positions[idx]]
		t := reflectx.Deref(reflect.TypeOf(vp))
		if isScannable(mapper, t) {
			continue
		}

		tm := mapper.TypeMap(t)
		fi, _ := tm.Names[m.fields[idx]]
		if fi == nil {
			var sb strings.Builder
			for key := range tm.Names {
				sb.WriteString(key)
				sb.WriteString(",")
			}
			return errors.New("field '" + m.fields[idx] + "' isnot found in the " + t.Name() + "(" + sb.String() + ")")
		}
		traversals[idx] = fi
	}
	m.traversals = traversals
	return nil
}

func (m *Multiple) scan(dialect Dialect, mapper *Mapper, r colScanner, isUnsafe bool) error {
	values := make([]interface{}, len(m.traversals))
	for idx, traversal := range m.traversals {
		vp := m.Returns[m.positions[idx]]
		if traversal == nil {
			values[idx] = vp
			continue
		}

		fieldName := m.fields[m.positions[idx]]

		v := reflect.Indirect(reflect.ValueOf(vp))
		fvalue, err := traversal.RValue(dialect, fieldName, v)
		if err != nil {
			return err
		}
		values[idx] = fvalue
	}

	if err := r.Scan(values...); err != nil {
		return errors.New("Scan into multiple(" + strings.Join(m.columns, ",") + ") error : " + err.Error())
	}
	return nil
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

func (m *MultipleArray) SetDelimiter(delimiter string) {
	m.multiple.SetDelimiter(delimiter)
}

func (m *MultipleArray) Set(name string, newRows func(int) interface{}) {
	m.Names = append(m.Names, name)
	m.NewRows = append(m.NewRows, newRows)
}

func (m *MultipleArray) setColumns(columns []string) error {
	m.multiple.Names = m.Names
	return m.multiple.setColumns(columns)
}

func (m *MultipleArray) Next(mapper *Mapper) (*Multiple, error) {
	for idx := range m.NewRows {
		m.multiple.Returns = append(m.multiple.Returns, m.NewRows[idx](m.Index))
	}

	if err := m.multiple.ensureTraversals(mapper); err != nil {
		return nil, err
	}

	m.Index++
	return &m.multiple, nil
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
		multiple, err := m.Next(mapper)
		if err != nil {
			return err
		}

		err = multiple.Scan(dialect, mapper, r, isUnsafe)
		if err != nil {
			return err
		}
	}
	return r.Err()
}

func indexColumns(columns, names []string, delimiter string) ([]int, []string, error) {
	if delimiter == "" {
		delimiter = "_"
	}

	results := make([]int, len(columns))
	fields := make([]string, len(columns))
	for idx, column := range columns {

		foundIndex := -1
		for nameIdx, name := range names {
			if name == column {
				foundIndex = nameIdx
				break
			}
		}

		if foundIndex >= 0 {
			fields[idx] = column
			results[idx] = foundIndex
			continue
		}

		position := strings.Index(column, delimiter)
		if position < 0 {
			return nil, nil, errors.New("column '" + strings.Join(columns, ",") + "' isnot exists in the names - " + strings.Join(names, ","))
		}

		tagName := column[:position]
		for nameIdx, name := range names {
			if name == tagName {
				foundIndex = nameIdx
				break
			}
		}

		if foundIndex < 0 {
			return nil, nil, errors.New("column '" + strings.Join(columns, ",") + "' isnot exists in the names - " + strings.Join(names, ","))
		}

		results[idx] = foundIndex
		fields[idx] = column[position+len(delimiter):]
	}
	return results, fields, nil
}
