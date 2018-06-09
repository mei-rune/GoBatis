//go:generate gobatis users.go
package example

type Users interface {
	Insert(u *AuthUser) (int64, error)

	Update(id int64, u *AuthUser) (int64, error)

	DeleteAll() (int64, error)

	Delete(id int64) (int64, error)

	Get(id int64) (*AuthUser, error)

	Count() (int64, error)

	GetName(id int64) (string, error)
}
