//go:generate gobatis user.go
package gentest

import (
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

	// @postgres select username FROM auth_users WHERE id=$1
	// @default select username FROM auth_users WHERE id=?
	GetName(id int64) (string, error)

	// @default select username FROM auth_users
	GetNames() ([]string, error)

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
	//            where user_id = #{id} and auth_roles.id = auth_users_and_roles.role_id)
	Roles(id int64) ([]Role, error)

	// @reference UserProfiles.Insert
	InsertProfile(profile *UserProfile) (int64, error)

	// @reference UserProfiles.DeleteByUserID
	DeleteProfile(userID int64) (int64, error)

	// @reference UserProfiles.List
	Profiles(userID int64) ([]*UserProfile, error)
}

type UserProfile struct {
	TableName gobatis.TableName `db:"user_profiles"`
	ID        int64             `db:"id,autoincr"`
	UserID    int64             `db:"user_id"`
	Key       string            `db:"name"`
	Value     string            `db:"value"`
	CreatedAt time.Time         `db:"created_at"`
	UpdatedAt time.Time         `db:"updated_at"`
}

type UserProfiles interface {
	Insert(u *UserProfile) (int64, error)

	Update(id int64, u *UserProfile) (int64, error)

	DeleteByUserID(userID int64) (int64, error)

	DeleteAll() (int64, error)

	Delete(id int64) (int64, error)

	Get(id int64) (*UserProfile, error)

	List(userID int64) ([]*UserProfile, error)

	Count() (int64, error)

	// @option field_delimiter .
	// @default SELECT p.id as "p.id",
	//                 p.user_id as "p.user_id",
	//                 p.name as "p.name",
	//                 p.value "p.value",
	//                 p.created_at as "p.created_at",
	//                 p.updated_at as "p.updated_at",
	//                 u.id as "u.id",
	//                 u.username as "u.username"
	//          FROM user_profiles as p LEFT JOIN auth_users as u On p.user_id = u.id
	//          WHERE p.id = #{id}
	FindByID1(id int64) (p *UserProfile, u *User, err error)

	// @default SELECT p.id as p_id,
	//                 p.user_id as p_user_id,
	//                 p.name as p_name,
	//                 p.value p_value,
	//                 p.created_at as p_created_at,
	//                 p.updated_at as p_updated_at,
	//                 u.id as u_id,
	//                 u.username as u_username
	//          FROM user_profiles as p LEFT JOIN auth_users as u On p.user_id = u.id
	//          WHERE p.id = #{id}
	FindByID2(id int64) (p UserProfile, u User, err error)

	// @option default_return_name p
	// @default SELECT p.id,
	//                 p.user_id,
	//                 p.name,
	//                 p.value,
	//                 p.created_at,
	//                 p.updated_at,
	//                 u.id as userid,
	//                 u.username as username
	//          FROM user_profiles as p LEFT JOIN auth_users as u On p.user_id = u.id
	//          WHERE p.id = #{id}
	FindByID3(id int64) (p UserProfile, userid int64, username string, err error)

	// @default SELECT p.id as p_id,
	//                 p.user_id as p_user_id,
	//                 p.name as p_name,
	//                 p.value p_value,
	//                 p.created_at as p_created_at,
	//                 p.updated_at as p_updated_at,
	//                 u.id as userid,
	//                 u.username as username
	//          FROM user_profiles as p LEFT JOIN auth_users as u On p.user_id = u.id
	//          WHERE p.id = #{id}
	FindByID4(id int64) (p *UserProfile, userid *int64, username *string, err error)

	// @default SELECT p.id as p_id,
	//                 p.user_id as p_user_id,
	//                 p.name as p_name,
	//                 p.value p_value,
	//                 p.created_at as p_created_at,
	//                 p.updated_at as p_updated_at,
	//                 u.id as u_id,
	//                 u.username as u_username
	//          FROM user_profiles as p LEFT JOIN auth_users as u On p.user_id = u.id
	//          WHERE p.user_id = #{userID}
	ListByUserID1(userID int64) (p []*UserProfile, u []*User, err error)

	// @option field_delimiter .
	// @default SELECT p.id as "p.id",
	//                 p.user_id as "p.user_id",
	//                 p.name as "p.name",
	//                 p.value "p.value",
	//                 p.created_at as "p.created_at",
	//                 p.updated_at as "p.updated_at",
	//                 u.id as "u.id",
	//                 u.username as "u.username"
	//          FROM user_profiles as p LEFT JOIN auth_users as u On p.user_id = u.id
	//          WHERE p.user_id = #{userID}
	ListByUserID2(userID int64) (p []UserProfile, u []User, err error)

	// @option default_return_name p
	// @default SELECT p.id,
	//                 p.user_id,
	//                 p.name,
	//                 p.value,
	//                 p.created_at,
	//                 p.updated_at,
	//                 u.id as userids,
	//                 u.username as usernames
	//          FROM user_profiles as p LEFT JOIN auth_users as u On p.user_id = u.id
	//          WHERE p.user_id = #{userID}
	ListByUserID3(userID int64) (p []UserProfile, userids []int64, usernames []string, err error)

	// @option field_delimiter .
	// @default SELECT p.id as "p.id",
	//                 p.user_id as "p.user_id",
	//                 p.name as "p.name",
	//                 p.value "p.value",
	//                 p.created_at as "p.created_at",
	//                 p.updated_at as "p.updated_at",
	//                 u.id as userids,
	//                 u.username as usernames
	//          FROM user_profiles as p LEFT JOIN auth_users as u On p.user_id = u.id
	//          WHERE p.user_id = #{userID}
	ListByUserID4(userID int64) (p []*UserProfile, userids []*int64, usernames []*string, err error)
}
