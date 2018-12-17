//go:generate gobatis interface.go
package gentest

type TestInterface interface {

	// @default insert into xxx (name)  values (#{name})
	Insert(name string) (int64, error)

	// @default insert into xxx (name)  values (#{name})
	Update(id int64, name string) (int64, error)

	// @default select * from xxx where name = #{name}
	Query(name string) (int64, error)

	// @default delete from xxx where name = #{name}
	Delete(name string) (int64, error)

	// @default select * from xxx where name = #{name}
	GetByCallback1(name string) func(a *int64) error

	// @default select * from xxx where name = #{name}
	GetByCallback2(name string) func(*int64) error
}
