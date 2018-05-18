package goparser

import (
	"go/types"
	"strings"
)

type Result struct {
	Name string
	Type types.Type
}

type Results struct {
	Method *Method      `json:"-"`
	Tuple  *types.Tuple `json:"-"`
	List   []Result
}

func NewResults(method *Method, tuple *types.Tuple) *Results {
	rs := &Results{
		Method: method,
		Tuple:  tuple,
		List:   make([]Result, tuple.Len()),
	}

	for i := 0; i < tuple.Len(); i++ {
		v := tuple.At(i)
		rs.List[i] = Result{
			Name: v.Name(),
			Type: v.Type(),
		}
	}
	return rs
}

func (rs *Results) String() string {
	var ss []string
	for i := 0; i < rs.Tuple.Len(); i++ {
		ss = append(ss, rs.Tuple.At(i).String())
	}
	return strings.Join(ss, ", ")
}

func (rs *Results) Len() int {
	return len(rs.List)
}

func (rs *Results) ByName(name string) *Result {
	name = strings.Trim(name, "`")
	if name == "" {
		panic("name must not blank")
	}

	for idx := range rs.List {
		if rs.List[idx].Name == name {
			return &rs.List[idx]
		}
	}
	return nil
}
