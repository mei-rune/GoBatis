package goparser

import (
	"go/types"
	"strings"
)

type Param struct {
	Name string
	Type types.Type
}

type Params struct {
	Method *Method      `json:"-"`
	Tuple  *types.Tuple `json:"-"`
	List   []Param
}

func NewParams(method *Method, tuple *types.Tuple) *Params {
	ps := &Params{
		Method: method,
		Tuple:  tuple,
		List:   make([]Param, tuple.Len()),
	}

	for i := 0; i < tuple.Len(); i++ {
		v := tuple.At(i)
		ps.List[i] = Param{
			Name: v.Name(),
			Type: v.Type(),
		}
	}

	return ps
}

func (ps *Params) String() string {
	var ss []string
	for idx := range ps.List {
		ss = append(ss, ps.Tuple.At(idx).String())
	}
	return strings.Join(ss, ", ")
}

func (t *Params) Len() int {
	return len(t.List)
}

func (t *Params) ByName(name string) *Param {
	name = strings.Trim(name, "`")
	if name == "" {
		panic("name must not blank")
	}

	for idx := range t.List {
		if t.List[idx].Name == name {
			return &t.List[idx]
		}
	}
	return nil
}
