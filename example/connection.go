package example

import gobatis "github.com/runner-mei/GoBatis"

type Base struct {
	gobatis.Reference
}

func (base *Base) Users() UserDao {
	return NewUserDao(&base.Reference, base.UserProfiles())
}

func (base *Base) Roles() RoleDao {
	return NewRoleDao(&base.Reference)
}

func (base *Base) UserProfiles() UserProfiles {
	return NewUserProfiles(&base.Reference)
}

type Connection struct {
	*gobatis.Session
	Base
}

func (o *Connection) Begin() (*Tx, error) {
	tx, err := o.Session.Begin()
	if err != nil {
		return nil, err
	}
	return &Tx{Tx: tx, Base: Base{Reference: tx.Reference()}}, nil
}

type Tx struct {
	*gobatis.Tx
	Base
}

func (tx *Tx) Commit() error {
	return tx.Tx.Commit()
}

func (tx *Tx) Rollback() error {
	return tx.Tx.Rollback()
}

func NewConnection(factory *gobatis.Session) *Connection {
	return &Connection{
		Session: factory,
		Base:    Base{Reference: factory.Reference()},
	}
}
