package core

import (
	"fmt"
	"strings"
	"testing"
)

func TestInt64(t *testing.T) {
	for _, test := range []struct {
		except64, defaultValue int64
		value                  interface{}
	}{
		{except64: 1, defaultValue: 1, value: nil},
		{except64: 1, defaultValue: 0, value: int8(1)},
		{except64: 1, defaultValue: 0, value: int16(1)},
		{except64: 1, defaultValue: 0, value: int32(1)},
		{except64: 1, defaultValue: 0, value: int64(1)},
		{except64: 1, defaultValue: 0, value: int(1)},
		{except64: 1, defaultValue: 0, value: uint8(1)},
		{except64: 1, defaultValue: 0, value: uint16(1)},
		{except64: 1, defaultValue: 0, value: uint32(1)},
		{except64: 1, defaultValue: 0, value: uint64(1)},
		{except64: 1, defaultValue: 0, value: uint(1)},
		{except64: 1, defaultValue: 0, value: float32(1)},
		{except64: 1, defaultValue: 0, value: float64(1)},
	} {
		i64 := int64With(test.value, test.defaultValue)
		if i64 != test.except64 {
			t.Error("want", test.except64, "got", i64)
		}
	}
}

func TestValidPrintValue(t *testing.T) {
	for _, test := range []struct {
		inStr        bool
		value        interface{}
		exceptResult string
	}{
		// {value: nil},
		{value: int8(1)},
		{value: int16(1)},
		{value: int32(1)},
		{value: int64(1)},
		{value: int(1)},
		{value: uint8(1)},
		{value: uint16(1)},
		{value: uint32(1)},
		{value: uint64(1)},
		{value: uint(1)},
		{value: float32(1)},
		{value: float64(1)},
		{value: "abc"},
		{value: "ab_c"},
		{value: "ab_c_"},

		{value: "ab0123456789"},
		{value: "a0123456789"},
		{value: "A0123456789"},
		{value: "ab+c", exceptResult: "invalid"},
		{value: "ab-c", exceptResult: "invalid"},
		{value: "ab+c", inStr: true},
		{value: "ab-c", inStr: true},

		{value: "'abc", exceptResult: "invalid"},
		{value: "\"abc", inStr: true, exceptResult: "invalid"},

		{value: "'abc'"},
		{value: "\"abc\""},
		{value: "'abc'", inStr: true, exceptResult: "invalid"},
		{value: "\"abc\"", inStr: true, exceptResult: "invalid"},

		{value: "aa->>'abc'"},
		{value: "aa->>\"abc\""},
		{value: "aa->>'abc'", inStr: true, exceptResult: "invalid"},
		{value: "aa->>\"abc\"", inStr: true, exceptResult: "invalid"},

		{value: "aa->'abc'"},
		{value: "aa->\"abc\""},
		{value: "aa->'abc'", inStr: true, exceptResult: "invalid"},
		{value: "aa->\"abc\"", inStr: true, exceptResult: "invalid"},

		{value: "aa->>'a-bc'"},
		{value: "aa->>\"a-bc\""},
		{value: "aa->>'a-bc'", inStr: true, exceptResult: "invalid"},
		{value: "aa->>\"a-bc\"", inStr: true, exceptResult: "invalid"},

		{value: "aa->'a-bc'"},
		{value: "aa->\"a-bc\""},
		{value: "aa->'a-bc'", inStr: true, exceptResult: "invalid"},
		{value: "aa->\"a-bc\"", inStr: true, exceptResult: "invalid"},
	} {
		t.Log(test.value)
		fmt.Println("====", test.value)

		err := isValidPrintValue(test.value, test.inStr)
		if test.exceptResult == "" {
			if err != nil {
				t.Error("want ok got", err)
			}
		} else {
			if err == nil {
				t.Error("want error got ok")
			} else if !strings.Contains(err.Error(), test.exceptResult) {
				t.Error("want", test.exceptResult, "got", err)
			}
		}
	}
}

func TestReplaceAndOr(t *testing.T) {
	for _, test := range []struct {
		txt string
		result string
	} {
		{
			txt: "c or b",
			result: "c || b",
		},
		{
			txt: "a or b",
			result: "a || b",
		},
		{
			txt: "a  or  b",
			result: "a  ||  b",
		},
		{
			txt: "a\tor b",
			result: "a\t|| b",
		},
		{
			txt: "a or\tb",
			result: "a ||\tb",
		},
		{
			txt: "a\tor\tb",
			result: "a\t||\tb",
		},
		{
			txt: "a o b",	
			result: "a o b",
		},

		{
			txt: "a \"or\" b",	
			result: "a \"or\" b",
		},

		{
			txt: "a \"or\" or b",	
			result: "a \"or\" || b",
		},



		{
			txt: "c and b",
			result: "c && b",
		},
		{
			txt: "a and b",
			result: "a && b",
		},
		{
			txt: "a  and  b",
			result: "a  &&  b",
		},
		{
			txt: "a\tand b",
			result: "a\t&& b",
		},
		{
			txt: "a and\tb",
			result: "a &&\tb",
		},
		{
			txt: "a\tand\tb",
			result: "a\t&&\tb",
		},

		{
			txt: "a \"and\" b",	
			result: "a \"and\" b",
		},

		{
			txt: "a \"and\" and b",	
			result: "a \"and\" && b",
		},
	} {
		s := replaceAndOr(test.txt)
		if s != test.result {
			t.Error("txt :", test.txt)
			t.Error("want:", test.result)
			t.Error(" got:", s)
		}
	}
}
