
package user

import (
	"time"
	role "github.com/runner-mei/GoBatis/goparser/rr"
	g "github.com/runner-mei/GoBatis/goparser/group"
)

type Status uint8

type User struct {
	ID        uint64     `json:"id"`
	Username  string     `json:"username"`
	Phone     string     `json:"phone"`
	Address   *string    `json:"address"`
	Status    Status     `json:"status"`
	BirthDay  *time.Time `json:"birth_day"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
}

type UserDao interface {
	// insert ignore into users(`username`, phone, address, status, birth_day, created, updated)
	// values (?,?,?,?,?,CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
	Insert(u *User) (int64, error)

	// select id, username, phone, address, status, birth_day, created, updated
	// FROM users WHERE id=?
	Get(id uint64) (*User, error)

	// select count(1)
	// from users
	Count() (int64, error)

	Ping()

	RemoveAll() (err error)

	List(offset int, size int) ([]*User, error)

	List2(offset int, size int) ([]User, error)

	List3(offset int, size int) (users []User, err error)

	ListAll() (map[int]*User, error)

	UpdateByID(id int, user map[string]interface{}) error

	Roles(id int) ([]role.Role, error)

	Groups(id int) ([]g.Group, error)

	GroupsWithID(id int) (map[int64]g.Group, error)

	Prefiles(id int) ([]Profile, error)
}