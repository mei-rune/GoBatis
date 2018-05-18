package goparser

func MethodSignature(m *Method) string {
	name := m.Itf.Name
	if name[0] == 'I' {
		name = name[1:]
	}
	return m.Name + "(" + m.Params.String() + ")(" + m.Results.String() + ")"
}

type Method struct {
	Itf     *Interface `json:"-"`
	Name    string
	Doc     []string
	Params  *Params
	Results *Results
}

func NewMethod(itf *Interface, name string, doc []string) *Method {
	return &Method{Itf: itf, Name: name, Doc: doc}
}
