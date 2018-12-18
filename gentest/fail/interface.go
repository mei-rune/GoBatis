//go:generate gobatis interface.go
package failtest

type TestInterface interface {
	// @default select * from xxx where name = #{name}
	GetByCallback1(name string) func(int64) error
	// @default select * from xxx where name = #{name}
	GetByCallback2(name string) func(int64) int
	// @default select * from xxx where name = #{name}
	GetByCallback3(name string) func(...int64) int
	// @default select * from xxx where name = #{name}
	GetByCallback4(name string) func(int64, int64) int
}
