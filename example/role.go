//go:generate gobatis role.go
package example

import (
	"time"
)

type Role struct {
	TableName struct{}  `db:"auth_users"`
	ID        int64     `db:"id,autoincr"`
	Name      string    `db:"name"`
	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
}

type RoleDao interface {
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

	// @type select
	// @default select * from auth_users where exists(
	//            select * from auth_users_and_roles
	//            where auth_users_and_roles.role_id = #{id} and auth_users.id = auth_users_and_roles.user_id)
	Users(id int64) ([]User, error)

	// @type insert
	// @default insert into auth_users_and_roles(user_id, role_id)
	// values ((select id from auth_users where username=#{username}), (select id from auth_roles where name=#{rolename}))
	AddUser(username, rolename string) error

	// @type delete
	// @default delete from auth_users_and_roles where exists(
	//              select * from auth_users_and_roles, auth_users, auth_roles
	//              where auth_users.id = auth_users_and_roles.user_id
	//              and auth_roles.id = auth_users_and_roles.role_id
	//              and auth_roles.name = #{rolename}
	//              and auth_users.username = #{username}
	//          )
	// @mysql delete from auth_users_and_roles where
	//     auth_users_and_roles.user_id in (select id from auth_users where username = #{username})
	// AND auth_users_and_roles.role_id in (select id from auth_roles where name = #{rolename})
	RemoveUser(username, rolename string) (e error)
}
