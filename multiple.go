package gobatis

import (
	"database/sql"
	"errors"
	"reflect"
	"strings"

	"github.com/runner-mei/GoBatis/reflectx"
)

func NewMultipleArray() *MultipleArray {
	return &MultipleArray{}
}

type committer struct {
	commitFunc func(bool)
	canCommits []func() bool
}

func (ct *committer) reset() {
	ct.commitFunc = nil
	if len(ct.canCommits) > 0 {
		ct.canCommits = ct.canCommits[:0]
	}
}

func (ct *committer) commit() bool {
	if ct.commitFunc == nil {
		return false
	}
	if len(ct.canCommits) == 0 {
		ct.commitFunc(false)
		return false
	}
	isCommit := false
	for _, can := range ct.canCommits {
		if can() {
			isCommit = true
			break
		}
	}

	ct.commitFunc(isCommit)
	return isCommit
}

func (ct *committer) estimateWith(value interface{}) interface{} {
	if ct.commitFunc == nil {
		return value
	}

	switch v := value.(type) {
	case *sql.NullInt64:
		ct.canCommits = append(ct.canCommits, func() bool {
			return v.Valid
		})
		return value
	case *sql.NullFloat64:
		ct.canCommits = append(ct.canCommits, func() bool {
			return v.Valid
		})
		return value
	case *sql.NullBool:
		ct.canCommits = append(ct.canCommits, func() bool {
			return v.Valid
		})
		return value
	case *sql.NullString:
		ct.canCommits = append(ct.canCommits, func() bool {
			return v.Valid
		})
		return value
	case *Nullable:
		ct.canCommits = append(ct.canCommits, func() bool {
			return v.Valid
		})
		return value
	case *emptyScanner:
		ct.canCommits = append(ct.canCommits, func() bool {
			return false
		})
		return value
	case *sScanner:
		ct.canCommits = append(ct.canCommits, func() bool {
			return v.Valid
		})
		return value
	case *scanner:
		ct.canCommits = append(ct.canCommits, func() bool {
			return v.Valid
		})
		return value
	default:
		nullable := &Nullable{Value: value}
		ct.canCommits = append(ct.canCommits, func() bool {
			return nullable.Valid
		})
		return nullable
	}
}

type Multiple struct {
	Names   []string
	Returns []interface{}
	commits []committer

	defaultReturnName string
	delimiter         string
	columns           []string     // sql 执行后的列名， 下面三个变曙与本数组长度一致
	positions         []int        // columns 中的列在 Returns 中的位置
	fields            []string     // columns 中的列中 Returns 中的值为结构时，表示字段名 ，否则为空
	traversals        []*FieldInfo // columns 中的列中 Returns 中的值为结构时，表示字段的指针 ，否则为 nil
}

func (m *Multiple) SetDelimiter(delimiter string) {
	m.delimiter = delimiter
}

func (m *Multiple) SetDefaultReturnName(returnName string) {
	m.defaultReturnName = returnName
}

func (m *Multiple) Set(name string, ret interface{}) {
	m.Names = append(m.Names, name)
	m.Returns = append(m.Returns, ret)
	m.commits = append(m.commits, committer{})
}

func (m *Multiple) setColumns(columns []string) (err error) {
	defaultField := -1
	if m.defaultReturnName != "" {
		for idx, name := range m.Names {
			if m.defaultReturnName == name {
				defaultField = idx
				break
			}
		}
	}

	m.columns = columns
	m.positions, m.fields, err = indexColumns(columns, m.Names, defaultField, m.delimiter)
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
	rValues := make([]reflect.Value, len(m.Returns))
	isPtrs := make([]bool, len(m.Returns))

	for idx, traversal := range m.traversals {
		valueIndex := m.positions[idx]
		if traversal == nil {
			vp := m.Returns[valueIndex]
			if _, ok := vp.(sql.Scanner); ok {
				values[idx] = m.commits[valueIndex].estimateWith(vp)
				continue
			}

			nullable := &Nullable{Name: m.columns[idx], Value: vp}
			values[idx] = m.commits[valueIndex].estimateWith(nullable)
			continue
		}

		rv := rValues[valueIndex]
		isPtr := isPtrs[valueIndex]
		if !rv.IsValid() {
			vp := m.Returns[valueIndex]
			rv = reflect.ValueOf(vp)
			if rv.Kind() == reflect.Ptr {
				isPtr = true
				rv = rv.Elem()
			}

			rValues[valueIndex] = rv
			isPtrs[valueIndex] = isPtr
		}

		fvalue, err := traversal.LValue(dialect, m.columns[idx], rv)
		if err != nil {
			return err
		}
		if !isPtr {
			values[idx] = fvalue
			continue
		}

		values[idx] = m.commits[valueIndex].estimateWith(fvalue)
	}

	if err := r.Scan(values...); err != nil {
		return errors.New("Scan into multiple(" + strings.Join(m.columns, ",") + ") error : " + err.Error())
	}

	for _, commit := range m.commits {
		commit.commit()
	}
	return nil
}

func NewMultiple() *Multiple {
	return &Multiple{}
}

type MultipleArray struct {
	Names   []string
	NewRows []func(int) (interface{}, func(bool))

	Index int

	multiple Multiple
}

func (m *MultipleArray) SetDefaultReturnName(returnName string) {
	m.multiple.SetDefaultReturnName(returnName)
}

func (m *MultipleArray) SetDelimiter(delimiter string) {
	m.multiple.SetDelimiter(delimiter)
}

func (m *MultipleArray) Set(name string, newRows func(int) (interface{}, func(bool))) {
	m.Names = append(m.Names, name)
	m.NewRows = append(m.NewRows, newRows)
}

func (m *MultipleArray) setColumns(columns []string) error {
	m.multiple.Names = m.Names
	return m.multiple.setColumns(columns)
}

func (m *MultipleArray) Next(mapper *Mapper) (*Multiple, error) {
	if len(m.multiple.Returns) != len(m.NewRows) {
		m.multiple.Returns = make([]interface{}, len(m.NewRows))
		m.multiple.commits = make([]committer, len(m.NewRows))
	}

	for idx := range m.NewRows {
		m.multiple.commits[idx].reset()
		m.multiple.Returns[idx], m.multiple.commits[idx].commitFunc = m.NewRows[idx](m.Index)
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

const (
	alreadyExistsBasic  = 1
	alreadyExistsStruct = 2
)

func indexColumns(columns, names []string, defaultField int, delimiter string) ([]int, []string, error) {
	if delimiter == "" {
		delimiter = "_"
	}

	results := make([]int, len(columns))
	fields := make([]string, len(columns))

	alreadyExists := make([]int, len(names))
	for idx, column := range columns {

		foundIndex := -1
		for nameIdx, name := range names {
			if name == column {
				foundIndex = nameIdx
				break
			}
		}

		if foundIndex >= 0 {
			if alreadyExists[foundIndex] != 0 {
				var oldColumn string
				for i := 0; i < idx; i++ {
					if results[i] == foundIndex {
						oldColumn = columns[i]
						break
					}
				}
				return nil, nil, errors.New("column '" + column +
					"' is duplicated with '" + oldColumn + "' in the names - " + strings.Join(names, ","))
			}

			fields[idx] = column
			results[idx] = foundIndex
			alreadyExists[foundIndex] = alreadyExistsBasic
			continue
		}

		position := strings.Index(column, delimiter)
		if position < 0 {
			if defaultField >= 0 {
				fields[idx] = column
				results[idx] = defaultField
				alreadyExists[defaultField] = alreadyExistsStruct
				continue
			}

			return nil, nil, errors.New("column '" + strings.Join(columns, ",") +
				"' isnot exists in the names - " + strings.Join(names, ","))
		}

		tagName := column[:position]
		for nameIdx, name := range names {
			if name == tagName {
				foundIndex = nameIdx
				break
			}
		}

		if foundIndex < 0 {

			if defaultField >= 0 {
				fields[idx] = column
				results[idx] = defaultField
				alreadyExists[defaultField] = alreadyExistsStruct
				continue
			}

			return nil, nil, errors.New("column '" + column + "' isnot exists in the names - " + strings.Join(names, ","))
		}

		if alreadyExists[foundIndex] == alreadyExistsBasic {
			var oldColumn string
			for i := 0; i < idx; i++ {
				if results[i] == foundIndex {
					oldColumn = columns[i]
					break
				}
			}

			return nil, nil, errors.New("column '" + column +
				"' is duplicated with '" + oldColumn + "' in the names - " + strings.Join(names, ","))
		}

		fields[idx] = column[position+len(delimiter):]
		results[idx] = foundIndex
		alreadyExists[foundIndex] = alreadyExistsStruct
	}
	return results, fields, nil
}
