//go:generate gobatis embedded.go
package gentest

type Test1 interface {

	// @type insert
	// @default insert into auth_roles(name, created_at, updated_at)
	// values (#{name}, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
	Insert1(name string) (int64, error)
}

type Test2 interface {
	// @type insert
	// @default insert into auth_roles(name, created_at, updated_at)
	// values (#{name}, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
	Insert2(name string) (int64, error)
}

type TestEmbedded interface {
	Test1

	Test2

	// @type insert
	// @default insert into auth_roles(name, created_at, updated_at)
	// values (#{name}, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
	Insert3() (int64, error)
}
