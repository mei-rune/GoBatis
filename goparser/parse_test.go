package goparser

import "testing"

func TestParse(t *testing.T) {
	ifs, e := Parse("../example/user.go")
	if e != nil {
		t.Error(e)
		return
	}
	t.Log(ifs)

	for idx := range ifs.Interfaces {
		t.Log(ifs.Interfaces[idx].Name)

		for _, m := range ifs.Interfaces[idx].Methods {
			t.Log(*m)
		}
	}
}
