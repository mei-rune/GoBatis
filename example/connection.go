package example

import gobatis "github.com/runner-mei/GoBatis"

func NewConnection(factory *gobatis.SessionFactory) *Connection {
	return &Connection{
		SessionFactory: factory,
		Base:           Base{Reference: factory.Reference()},
	}
}

type Connection struct {
	*gobatis.SessionFactory
	Base
}

func (o *Connection) Begin() (*Tx, error) {
	tx, err := o.SessionFactory.Begin()
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

type Base struct {
	gobatis.Reference
}

func (base *Base) Users() AuthUserDao {
	return NewAuthUserDao(&base.Reference)
}

func (base *Base) Roles() AuthRoleDao {
	return NewAuthRoleDao(&base.Reference)
}
