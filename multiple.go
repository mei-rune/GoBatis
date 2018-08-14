package gobatis

import "errors"

type Multiple struct {
	Names   []string
	Returns []interface{}
}

func (m *Multiple) Set(name string, ret interface{}) {
	m.Names = append(m.Names, name)
	m.Returns = append(m.Returns, ret)
}

func (m *Multiple) Scan(dialect Dialect, mapper *Mapper, r colScanner, isUnsafe bool) error {
	return errors.New("not implemented")
}

func NewMultiple() *Multiple {
	return &Multiple{}
}

type MultipleArray struct {
	Names   []string
	Returns []interface{}
}

func (m *MultipleArray) Set(name string, ret interface{}) {
	m.Names = append(m.Names, name)
	m.Returns = append(m.Returns, ret)
}

func NewMultipleArray() *MultipleArray {
	return &MultipleArray{}
}

func (m *MultipleArray) Scan(dialect Dialect, mapper *Mapper, r rowsi, isUnsafe bool) error {
	return errors.New("not implemented")
}
