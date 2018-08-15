//go:generate gobatis users_with_autogensql.go
package example

import (
	"time"

	gobatis "github.com/runner-mei/GoBatis"
)

type Users interface {
	Insert(u *AuthUser) (int64, error)

	Update(id int64, u *AuthUser) (int64, error)

	DeleteAll() (int64, error)

	Delete(id int64) (int64, error)

	Get(id int64) (*AuthUser, error)

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

	// @default SELECT p.id,
	//                 p.user_id,
	//                 p.name,
	//                 p.value,
	//                 p.created_at,
	//                 p.updated_at,
	//                 u.id,
	//                 u.username
	//          FROM user_profiles as p LEFT JOIN auth_users as u On p.user_id = u.id
	//          WHERE p.id = #{id}
	FindByID(id int64) (p *UserProfile, u *AuthUser, err error)

	// @default SELECT p.id,
	//                 p.user_id,
	//                 p.name,
	//                 p.value,
	//                 p.created_at,
	//                 p.updated_at,
	//                 u.id,
	//                 u.username
	//          FROM user_profiles as p LEFT JOIN auth_users as u On p.user_id = u.id
	//          WHERE p.id = #{id}
	FindByID2(id int64) (p UserProfile, u AuthUser, err error)

	// @default SELECT p.id,
	//                 p.user_id,
	//                 p.name,
	//                 p.value,
	//                 p.created_at,
	//                 p.updated_at,
	//                 u.id,
	//                 u.username
	//          FROM user_profiles as p LEFT JOIN auth_users as u On p.user_id = u.id
	//          WHERE p.user_id = #{userID}
	ListByUserID(userID int64) (p []*UserProfile, u []*AuthUser, err error)

	// @default SELECT p.id,
	//                 p.user_id,
	//                 p.name,
	//                 p.value,
	//                 p.created_at,
	//                 p.updated_at,
	//                 u.id,
	//                 u.username
	//          FROM user_profiles as p LEFT JOIN auth_users as u On p.user_id = u.id
	//          WHERE p.user_id = #{userID}
	ListByUserID2(userID int64) (p []UserProfile, u []AuthUser, err error)
}