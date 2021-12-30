//go:generate gobatis user.go
package example

import (
	"context"
	"time"

	gobatis "github.com/runner-mei/GoBatis"
)

type Status uint8

type User struct {
	TableName gobatis.TableName `db:"auth_users"`
	ID        int64             `db:"id,autoincr"`
	Username  string            `db:"username"`
	Phone     string            `db:"phone"`
	Address   *string           `db:"address"`
	Status    Status            `db:"status"`
	BirthDay  *time.Time        `db:"birth_day"`
	CreatedAt time.Time         `db:"created_at"`
	UpdatedAt time.Time         `db:"updated_at"`
}

type UserDao interface {
	// @mssql insert into auth_users(username, phone, address, status, birth_day, created_at, updated_at)
	// output inserted.id
	// values (#{username},#{phone},#{address},#{status},#{birth_day},CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
	//
	// @postgres insert into auth_users(username, phone, address, status, birth_day, created_at, updated_at)
	// values (#{username},#{phone},#{address},#{status},#{birth_day},CURRENT_TIMESTAMP, CURRENT_TIMESTAMP) returning id
	//
	// @default insert into auth_users(username, phone, address, status, birth_day, created_at, updated_at)
	// values (#{username},#{phone},#{address},#{status},#{birth_day},CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
	Insert(u *User) (int64, error)

	// @mssql insert into auth_users(username, phone, address, status, birth_day, created_at, updated_at)
	// output inserted.id
	// values (#{username},#{phone},#{address},#{status},#{birth_day},CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
	//
	// @postgres insert into auth_users(username, phone, address, status, birth_day, created_at, updated_at)
	// values (#{username},#{phone},#{address},#{status},#{birth_day},CURRENT_TIMESTAMP, CURRENT_TIMESTAMP) returning id
	//
	// @default insert into auth_users(username, phone, address, status, birth_day, created_at, updated_at)
	// values (#{username},#{phone},#{address},#{status},#{birth_day},CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
	InsertWithContext(ctx context.Context, u *User) (int64, error)

	// @mssql MERGE auth_users USING (
	//     VALUES (?,?,?,?,?, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
	// ) AS foo (username, phone, address, status, birth_day, created_at, updated_at)
	// ON auth_users.username = foo.username
	// WHEN MATCHED THEN
	//    UPDATE SET username=foo.username, phone=foo.phone, address=foo.address, status=foo.status, birth_day=foo.birth_day, updated_at=foo.updated_at
	// WHEN NOT MATCHED THEN
	//    INSERT (username, phone, address, status, birth_day, created_at, updated_at)
	//    VALUES (foo.username, foo.phone, foo.address, foo.status, foo.birth_day,  foo.created_at, foo.updated_at);
	//
	// @dm MERGE auth_users USING dual ON auth_users.username = #{username}
	// WHEN MATCHED THEN
	//    UPDATE SET username=#{username}, phone=#{phone}, address=#{address}, status=#{status}, birth_day=#{birth_day}, updated_at=CURRENT_TIMESTAMP
	// WHEN NOT MATCHED THEN
	//    INSERT (username, phone, address, status, birth_day, created_at, updated_at)
	//    VALUES (#{username}, #{phone}, #{address}, #{status}, #{birth_day}, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP);
	//
	// @mysql insert into auth_users(username, phone, address, status, birth_day, created_at, updated_at)
	// values (?,?,?,?,?,CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
	// on duplicate key update
	//   username=values(username), phone=values(phone), address=values(address),
	//   status=values(status), birth_day=values(birth_day), updated_at=CURRENT_TIMESTAMP
	//
	// @postgres insert into auth_users(username, phone, address, status, birth_day, created_at, updated_at)
	// values (?,?,?,?,?,CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
	// on duplicate key update
	//   username=values(username), phone=values(phone), address=values(address),
	//   status=values(status), birth_day=values(birth_day), updated_at=CURRENT_TIMESTAMP
	Upsert(u *User) (int64, error)

	// @default UPDATE auth_users
	// SET username=#{u.username},
	//     phone=#{u.phone},
	//     address=#{u.address},
	//     status=#{u.status},
	//     birth_day=#{u.birth_day},
	//     updated_at=CURRENT_TIMESTAMP
	// WHERE id=#{id}
	Update(id int64, u *User) (int64, error)

	// @default UPDATE auth_users
	// SET username=#{username},
	//     updated_at=CURRENT_TIMESTAMP
	// WHERE id=#{id}
	UpdateName(id int64, username string) (int64, error)

	// @postgres DELETE FROM auth_users
	// @default DELETE FROM auth_users
	DeleteAll() (int64, error)

	// @postgres DELETE FROM auth_users WHERE id=$1
	// @default DELETE FROM auth_users WHERE id=?
	Delete(id int64) (int64, error)

	// @postgres select * FROM auth_users WHERE id=$1
	// @default select * FROM auth_users WHERE id=?
	Get(id int64) (*User, error)

	// @postgres select * FROM auth_users WHERE id=$1
	// @default select * FROM auth_users WHERE id=?
	GetNonPtr(id int64) (User, error)

	// @postgres select username FROM auth_users WHERE id=$1
	// @default select username FROM auth_users WHERE id=?
	GetName(id int64) (string, error)

	// @default select username FROM auth_users
	GetNames() ([]string, error)

	// @default select id, username FROM auth_users
	GetIDNames() (map[int64]string, error)

	// @postgres select * FROM auth_users WHERE id=$1
	// @default select * FROM auth_users WHERE id=?
	GetMap(id int64) (map[string]interface{}, error)

	// @default select count(*) from auth_users
	Count() (int64, error)

	// @mssql select * from auth_users ORDER BY username OFFSET #{offset} ROWS FETCH NEXT #{size}  ROWS ONLY
	// @mysql select * from auth_users limit #{offset}, #{size}
	// @default select * from auth_users offset #{offset} limit  #{size}
	List(offset, size int) (users []*User, err error)

	// @mssql select * from auth_users ORDER BY username OFFSET #{offset} ROWS FETCH NEXT #{size}  ROWS ONLY
	// @mysql select * from auth_users limit #{offset}, #{size}
	// @default select * from auth_users offset #{offset} limit  #{size}
	ListMap(offset, size int) (users map[int64]*User, err error)

	// @default select username from auth_users where id = #{id}
	GetNameByID(id int64) (string, error)

	// @type select
	// @default select * from auth_roles where exists(
	//            select * from auth_users_and_roles
	//            where user_id = #{userid} and auth_roles.id = auth_users_and_roles.role_id)
	Roles(userid int64) ([]Role, error)

	// @reference UserProfiles.Insert
	InsertProfile(profile *UserProfile) (int64, error)

	// @reference UserProfiles.DeleteByUserID
	DeleteProfile(userID int64) (int64, error)

	// @reference UserProfiles.List
	Profiles(userID int64) ([]*UserProfile, error)
}
