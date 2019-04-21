package gobatis

import (
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
