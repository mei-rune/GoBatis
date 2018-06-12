package gobatis

type Multiple struct {
	Names   []string
	Returns []interface{}
}

func (m *Multiple) Set(name string, ret interface{}) {
	m.Names = append(m.Names, name)
	m.Returns = append(m.Returns, ret)
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
