//go:generate gobatis computer.go
package tests

type Computer struct {
	TableName   struct{} `db:"computers"`
	ID          int64    `db:"id,pk,autoincr"`
	Description string   `db:"description,notnull"`
	KeyID       int64    `db:"key_id,null"`
	MotherID    int64    `db:"mother_id,null"`
}

type Keyboard struct {
	TableName   struct{} `db:"keyboards"`
	ID          int64    `db:"id,pk,autoincr"`
	Description string   `db:"description,notnull"`
}

type Motherboard struct {
	TableName   struct{} `db:"motherboards"`
	ID          int64    `db:"id,pk,autoincr"`
	Description string   `db:"description,notnull"`
}

type Computers interface {
	InsertComputer(computer *Computer) (int64, error)
	InsertKeyboard(keyboard *Keyboard) (int64, error)
	InsertMotherboard(motherboard *Motherboard) (int64, error)

	// @option default_return_name computer
	// @default SELECT c.id, c.description, c.key_id, c.mother_id,
	//                 k.id as keyboard_id, k.description as keyboard_description
	//           FROM computers as c LEFT JOIN keyboards as k On c.key_id = k.id
	//           WHERE c.id = #{id}
	FindByID1(id int64) (computer *Computer, keyboard *Keyboard, err error)

	// @option default_return_name computer
	// @default SELECT c.id, c.description, c.key_id, c.mother_id,
	//                 m.id as motherboard_id, m.description as motherboard_description
	//           FROM computers as c LEFT JOIN motherboards as m On c.mother_id = m.id
	//           WHERE c.id = #{id}
	FindByID2(id int64) (computer *Computer, motherboard *Motherboard, err error)

	// @option default_return_name computer
	// @default SELECT c.id, c.description, c.key_id, c.mother_id,
	//                 k.id as keyboard_id, k.description as keyboard_description,
	//                 m.id as motherboard_id, m.description as motherboard_description
	//           FROM (computers as c LEFT JOIN motherboards as m On c.mother_id = m.id)
	//                  LEFT JOIN keyboards as k On c.key_id = k.id
	//           WHERE c.id = #{id}
	FindByID3(id int64) (computer *Computer, keyboard *Keyboard, motherboard *Motherboard, err error)

	// @option default_return_name computers
	// @default SELECT c.id, c.description, c.key_id, c.mother_id,
	//                 k.id as keyboards_id, k.description as keyboards_description
	//           FROM computers as c LEFT JOIN keyboards as k On c.key_id = k.id
	QueryAll1() (computers []Computer, keyboards []*Keyboard, err error)

	// @option default_return_name computers
	// @default SELECT c.id, c.description, c.key_id, c.mother_id,
	//                 k.id as keyboards_id, k.description as keyboards_description,
	//                 m.id as motherboards_id, m.description as motherboards_description
	//           FROM (computers as c LEFT JOIN motherboards as m On c.mother_id = m.id)
	//                  LEFT JOIN keyboards as k On c.key_id = k.id
	QueryAll2() (computers []Computer, keyboards []*Keyboard, motherboards []*Motherboard, err error)

	// @option default_return_name computers
	// @default SELECT c.id, c.description, c.key_id, c.mother_id,
	//                 k.id as keyboards_id, k.description as keyboards_description,
	//                 m.id as motherboards_id, m.description as motherboards_description
	//           FROM (computers as c LEFT JOIN motherboards as m On c.mother_id = m.id)
	//                  LEFT JOIN keyboards as k On c.key_id = k.id
	QueryAll3() (computers []Computer, keyboards_id []int64, keyboards_description []string, motherboards_id []int64, motherboards_description []string, err error)
}
