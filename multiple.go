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
	columns           []columnInfo
}

type columnInfo struct {
	columnName string     // sql 执行后的列名
	fieldName  string     // 如果本列在 Returns 中的值为结构时，表示字段名 ，否则为列名
	position   int        // 本列在 Returns 中的位置
	fi         *FieldInfo // columns 中的列中 Returns 中的值为结构时，表示字段的指针 ，否则为 nil
}

func (m *Multiple) SetDelimiter(delimiter string) {
	m.delimiter = delimiter
}

func (m *Multiple) SetDefaultReturnName(returnName string) {
	m.defaultReturnName = returnName
}

func (m *Multiple) Set(name string, ret interface{}, commit ...func(ok bool)) {
	m.Names = append(m.Names, name)
	m.Returns = append(m.Returns, ret)
	if len(commit) > 0 {
		m.commits = append(m.commits, committer{commitFunc: commit[0]})
	} else {
		m.commits = append(m.commits, committer{})
	}
}

func (m *Multiple) setColumns(mapper *Mapper, r colScanner) error {
	if m.columns != nil {
		return nil
	}

	defaultReturn := -1
	if m.defaultReturnName != "" {
		for idx, name := range m.Names {
			if m.defaultReturnName == name {
				defaultReturn = idx
				break
			}
		}
	}

	columnNames, err := r.Columns()
	if err != nil {
		return err
	}

	columns, err := indexColumns(columnNames, m.Names, defaultReturn, m.delimiter)
	if err != nil {
		return err
	}

	for idx := range columns {
		vp := m.Returns[columns[idx].position]
		t := reflectx.Deref(reflect.TypeOf(vp))
		if isScannable(mapper, t) {
			continue
		}

		tm := mapper.TypeMap(t)
		fi, _ := tm.Names[columns[idx].fieldName]
		if fi == nil {
			var sb strings.Builder
			for key := range tm.Names {
				sb.WriteString(key)
				sb.WriteString(",")
			}
			return errors.New("field '" + columns[idx].fieldName + "' isnot found in the " + t.Name() + "(" + sb.String() + ")")
		}
		columns[idx].fi = fi
	}

	m.columns = columns
	return nil
}

func (m *Multiple) Scan(dialect Dialect, mapper *Mapper, r colScanner, isUnsafe bool) error {
	if err := m.setColumns(mapper, r); err != nil {
		return err
	}

	values := make([]interface{}, len(m.columns))
	rValues := make([]reflect.Value, len(m.Returns))
	isPtrs := make([]bool, len(m.Returns))

	for idx := range m.columns {
		valueIndex := m.columns[idx].position
		if m.columns[idx].fi == nil {
			vp := m.Returns[valueIndex]
			if _, ok := vp.(sql.Scanner); ok {
				values[idx] = m.commits[valueIndex].estimateWith(vp)
				continue
			}

			nullable := &Nullable{Name: m.columns[idx].columnName, Value: vp}
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

		fvalue, err := m.columns[idx].fi.LValue(dialect, m.columns[idx].columnName, rv)
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
		var sb strings.Builder
		sb.WriteString("Scan into multiple(")
		for idx := range m.columns {
			if idx != 0 {
				sb.WriteString(",")
			}
			sb.WriteString(m.columns[idx].columnName)
		}
		sb.WriteString(") error : ")
		sb.WriteString(err.Error())
		return errors.New(sb.String())
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
	m.multiple.Names = append(m.multiple.Names, name)
	m.NewRows = append(m.NewRows, newRows)
}

func (m *MultipleArray) Next(mapper *Mapper) *Multiple {
	if len(m.multiple.Returns) != len(m.NewRows) {
		m.multiple.Returns = make([]interface{}, len(m.NewRows))
		m.multiple.commits = make([]committer, len(m.NewRows))
	}

	for idx := range m.NewRows {
		m.multiple.commits[idx].reset()
		m.multiple.Returns[idx], m.multiple.commits[idx].commitFunc = m.NewRows[idx](m.Index)
	}

	m.Index++
	return &m.multiple
}

func (m *MultipleArray) Scan(dialect Dialect, mapper *Mapper, r rowsi, isUnsafe bool) error {
	for r.Next() {
		multiple := m.Next(mapper)

		err := multiple.Scan(dialect, mapper, r, isUnsafe)
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

func indexColumns(columns, names []string, defaultReturn int, delimiter string) ([]columnInfo, error) {
	if delimiter == "" {
		delimiter = "_"
	}

	results := make([]columnInfo, len(columns))
	//fields := make([]string, len(columns))

	alreadyExists := make([]int, len(names))
	for idx, column := range columns {
		foundIndex := -1
		for nameIdx, name := range names {
			if name == column || strings.ToLower(name) == strings.ToLower(column) {
				foundIndex = nameIdx
				break
			}
		}

		if foundIndex >= 0 {
			if alreadyExists[foundIndex] != 0 {
				var oldColumn string
				for i := 0; i < idx; i++ {
					if results[i].position == foundIndex {
						oldColumn = columns[i]
						break
					}
				}
				return nil, errors.New("column '" + column +
					"' is duplicated with '" + oldColumn + "' in the names - " + strings.Join(names, ","))
			}

			results[idx].columnName = column
			results[idx].fieldName = column
			results[idx].position = foundIndex
			alreadyExists[foundIndex] = alreadyExistsBasic
			continue
		}

		position := strings.Index(column, delimiter)
		if position < 0 {
			if defaultReturn >= 0 {
				results[idx].columnName = column
				results[idx].fieldName = column
				results[idx].position = defaultReturn

				alreadyExists[defaultReturn] = alreadyExistsStruct
				continue
			}

			return nil, errors.New("column '" + strings.Join(columns, ",") +
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
			if defaultReturn >= 0 {
				results[idx].columnName = column
				results[idx].fieldName = column
				results[idx].position = defaultReturn
				alreadyExists[defaultReturn] = alreadyExistsStruct
				continue
			}

			return nil, errors.New("column '" + column + "' isnot exists in the names - " + strings.Join(names, ","))
		}

		if alreadyExists[foundIndex] == alreadyExistsBasic {
			var oldColumn string
			for i := 0; i < idx; i++ {
				if results[i].position == foundIndex {
					oldColumn = columns[i]
					break
				}
			}

			return nil, errors.New("column '" + column +
				"' is duplicated with '" + oldColumn + "' in the names - " + strings.Join(names, ","))
		}

		results[idx].columnName = column
		results[idx].fieldName = column[position+len(delimiter):]
		results[idx].position = foundIndex

		alreadyExists[foundIndex] = alreadyExistsStruct
	}
	return results, nil
}
