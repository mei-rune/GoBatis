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

		func (a A) f1() {}

		type A struct {
			A int
			B string
			C bool
			D time.Time
			E IntType
		}
		func (a A) f2() {}
		`))

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

	f1 := aClass.MethodByName("f1")
	if f1 == nil {
		t.Error("f1 not found")
	}
	f2 := aClass.MethodByName("f2")
	if f2 == nil {
		t.Error("f2 not found")
	}
	afield := aClass.Struct.FieldByName("A")
	if afield == nil {
		t.Error("A not found")
		return
	}
	if ctx.IsStructType(file, afield.Expr) {
		t.Error("A type isnot struct?")
	}
	if !ctx.IsNumericType(file, afield.Expr, true) {
		t.Error("A type isnot numeric?")
	}

	bfield := aClass.Struct.FieldByName("B")
	if bfield == nil {
		t.Error("B not found")
		return
	}
	if ctx.IsStructType(file, bfield.Expr) {
		t.Error("B type isnot struct?")
	}
	if ctx.IsNumericType(file, bfield.Expr, true) {
		t.Error("B type isnot numeric?")
	}

	cfield := aClass.Struct.FieldByName("C")
	if cfield == nil {
		t.Error("C not found")
		return
	}
	if ctx.IsStructType(file, cfield.Expr) {
		t.Error("C type isnot struct?")
	}
	if ctx.IsNumericType(file, cfield.Expr, true) {
		t.Error("C type isnot numeric?")
	}
	// if !ctx.IsStringType(file, bfield.Expr) {
	// 	t.Error("B type isnot numeric?")
	// }

	dfield := aClass.Struct.FieldByName("D")
	if dfield == nil {
		t.Error("D not found")
		return
	}
	if !ctx.IsStructType(file, dfield.Expr) {
		t.Error("D type isnot struct?")
	}
	if ctx.IsBasicType(file, dfield.Expr, true) {
		t.Error("D type isnot basic type?")
	}

	efield := aClass.Struct.FieldByName("E")
	if efield == nil {
		t.Error("E not found")
		return
	}
	if ctx.IsStructType(file, efield.Expr) {
		t.Error("E type isnot struct?")
	}
	if !ctx.IsNumericType(file, efield.Expr, true) {
		t.Error("E type isnot numeric?")
	}
	if ctx.IsBasicType(file, efield.Expr, false) {
		t.Error("E type is basic type, no check underlying?")
	}
	if !ctx.IsBasicType(file, efield.Expr, true) {
		t.Error("E type isnot basic type?")
	}
}
