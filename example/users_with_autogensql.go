//go:generate gobatis users_with_autogensql.go
package example

import (
	"time"

	gobatis "github.com/runner-mei/GoBatis"
)

type Users interface {
	Insert(u *User) (int64, error)

	Update(id int64, u *User) (int64, error)

	DeleteAll() (int64, error)

	Delete(id int64) (int64, error)

	Get(id int64) (*User, error)

	Count() (int64, error)
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
