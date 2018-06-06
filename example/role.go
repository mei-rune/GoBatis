//go:generate gobatis role.go
package example

import (
	"time"
)

type AuthRole struct {
	ID        int64     `db:"id"`
	Name      string    `db:"name"`
	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
}

type AuthRoleDao interface {
	// @mssql insert into auth_roles(name, created_at, updated_at)
	// output inserted.id
	// values (#{name}, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
	//
	// @postgres insert into auth_roles(name, created_at, updated_at)
	// values (#{name}, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP) returning id
	//
	// @default insert into auth_roles(name, created_at, updated_at)
	// values (#{name}, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
	Insert(name string) (int64, error)

	// @postgres select name FROM auth_roles WHERE id=$1
	// @default select name FROM auth_roles WHERE id=?
	Get(id int64) (string, error)
}
