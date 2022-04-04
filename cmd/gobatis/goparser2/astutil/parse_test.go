package astutil

import (
	"strings"
	"testing"
)

func TestParse(t *testing.T) {
	ctx := NewContext(nil)

	file, err := Parse(ctx, "a.go", strings.NewReader(`package main
		import (
			"time"
		)

		type IntType int

		type A struct {
			A int
			B string
			C bool
			D time.Time
			E IntType
		}`))

	if err != nil {
		t.Error(err)
		return
	}

	for _, name := range []string{
		"time",
	} {
		found := false
		for idx := range file.Imports {
			if file.Imports[idx].Name != nil {
				if file.Imports[idx].Name.Name == name {
					found = true
					break
				}
			}
			if strings.Trim(file.Imports[idx].Path.Value, "\"") == name {
				found = true
				break
			}
		}
		if !found {
			t.Error(name, "not found")
			return
		}
	}

	aClass := file.GetType("A")
	if aClass == nil {
		t.Error("A not found")
		return
	}
	

	afield := aClass.Struct.FieldByName("A")
	if afield == nil {
		t.Error("A not found")
		return
	}
	if ctx.IsStructType(file, afield.Typ) {
		t.Error("A type isnot struct?")
	}
	if !ctx.IsNumericType(file, afield.Typ) {
		t.Error("A type isnot numeric?")
	}

	bfield := aClass.Struct.FieldByName("B")
	if bfield == nil {
		t.Error("B not found")
		return
	}
	if ctx.IsStructType(file, bfield.Typ) {
		t.Error("B type isnot struct?")
	}
	if ctx.IsNumericType(file, bfield.Typ) {
		t.Error("B type isnot numeric?")
	}


	cfield := aClass.Struct.FieldByName("C")
	if cfield == nil {
		t.Error("C not found")
		return
	}
	if ctx.IsStructType(file, cfield.Typ) {
		t.Error("C type isnot struct?")
	}
	if ctx.IsNumericType(file, cfield.Typ) {
		t.Error("C type isnot numeric?")
	}
	// if !ctx.IsStringType(file, bfield.Typ) {
	// 	t.Error("B type isnot numeric?")
	// }

	dfield := aClass.Struct.FieldByName("D")
	if dfield == nil {
		t.Error("D not found")
		return
	}
	if !ctx.IsStructType(file, dfield.Typ) {
		t.Error("D type isnot struct?")
	}
	if ctx.IsBasicType(file, dfield.Typ) {
		t.Error("D type isnot basic type?")
	}

	efield := aClass.Struct.FieldByName("E")
	if efield == nil {
		t.Error("E not found")
		return
	}
	if ctx.IsStructType(file, efield.Typ) {
		t.Error("E type isnot struct?")
	}
	if !ctx.IsNumericType(file, efield.Typ) {
		t.Error("E type isnot numeric?")
	}
	if !ctx.IsBasicType(file, efield.Typ) {
		t.Error("E type isnot basic type?")
	}
}
