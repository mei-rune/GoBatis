package user

import "time"

type Profile struct {
	ID        uint64
	Name      string
	Value     string
	CreatedAt time.Time
	UpdatedAt time.Time
}

type ProfileDao interface {
	Insert(name, value string) (int64, error)

	Remove(name string) error

	Users(name string) ([]User, error)
}
