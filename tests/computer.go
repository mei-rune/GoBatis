//go:generate gobatis computer.go
package tests

import (
	"database/sql"
	"net"
)

type Computer struct {
	TableName   struct{} `db:"computers"`
	ID          int64    `db:"id,pk,autoincr"`
	Description string   `db:"description,notnull"`
	KeyID       int64    `db:"key_id,null"`
	MotherID    int64    `db:"mother_id,null"`
	MouseID     int64    `db:"mouse_id,null"`
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

type Mouse struct {
	TableName struct{}               `db:"mouses"`
	ID        int64                  `db:"id,pk,autoincr"`
	Field1    sql.NullBool           `db:"field1"`
	Field2    sql.NullInt64          `db:"field2"`
	Field3    sql.NullFloat64        `db:"field3"`
	Field4    sql.NullString         `db:"field4"`
	Field5    net.IP                 `db:"field5"`
	Field6    map[string]interface{} `db:"field6"`
}

type Computers interface {
	InsertComputer(computer *Computer) (int64, error)
	InsertKeyboard(keyboard *Keyboard) (int64, error)
	InsertMotherboard(motherboard *Motherboard) (int64, error)

	// @default DELETE FROM keyboards WHERE EXISTS(SELECT * FROM computers WHERE computers.id = #{id} AND computers.key_id = keyboards.id);
	//          DELETE FROM mouses WHERE EXISTS(SELECT * FROM computers WHERE computers.id = #{id} AND computers.mouse_id = mouses.id);
	//          DELETE FROM motherboards WHERE EXISTS(SELECT * FROM computers WHERE computers.id = #{id} AND computers.mother_id = motherboards.id);
	//          DELETE FROM computers WHERE id = #{id};
	DeleteByID(id int64) (int64, error)

	// @option default_return_name computer
	// @default SELECT c.id, c.description, c.key_id, c.mother_id,
	//                 m.id as mouse_id, m.field1 as mouse_field1, m.field2 as mouse_field2,
	//                 m.field3 as mouse_field3, m.field4 as mouse_field4,
	//                 m.field5 as mouse_field5, m.field6 as mouse_field6
	//           FROM computers as c LEFT JOIN mouses as m On c.mouse_id = m.id
	//           WHERE c.id = #{id}
	FindByID0(id int64) (computer *Computer, mouse *Mouse, err error)

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
	//           ORDER BY c.id
	QueryAll1() (computers []Computer, keyboards []*Keyboard, err error)

	// @option default_return_name computers
	// @default SELECT c.id, c.description, c.key_id, c.mother_id,
	//                 k.id as keyboards_id, k.description as keyboards_description,
	//                 m.id as motherboards_id, m.description as motherboards_description
	//           FROM (computers as c LEFT JOIN motherboards as m On c.mother_id = m.id)
	//                  LEFT JOIN keyboards as k On c.key_id = k.id
	//           ORDER BY c.id
	QueryAll2() (computers []Computer, keyboards []*Keyboard, motherboards []*Motherboard, err error)

	// @option default_return_name computers
	// @default SELECT c.id, c.description, c.key_id, c.mother_id,
	//                 k.id as keyboards_id, k.description as keyboards_description,
	//                 m.id as motherboards_id, m.description as motherboards_description
	//           FROM (computers as c LEFT JOIN motherboards as m On c.mother_id = m.id)
	//                  LEFT JOIN keyboards as k On c.key_id = k.id
	//           ORDER BY c.id
	QueryAll3() (computers []Computer, keyboards_id []int64, keyboards_description []string, motherboards_id []int64, motherboards_description []string, err error)

	// @option default_return_name computers
	// @default SELECT c.id, c.description, c.key_id, c.mother_id,
	//                 k.id as keyboards_id, k.description as keyboards_description,
	//                 m.id as motherboards_id, m.description as motherboards_description
	//           FROM (computers as c LEFT JOIN motherboards as m On c.mother_id = m.id)
	//                  LEFT JOIN keyboards as k On c.key_id = k.id
	//           ORDER BY c.id
	QueryAllFail1() (computers []Computer, keyboards_id []int64, keyboards_description []int64, motherboards_id []int64, motherboards_description []string, err error)

	// @default SELECT count(*) FROM computers
	CountComputers() (int64, error)

	// @default SELECT count(*) FROM mouses
	CountMouses() (int64, error)

	// @default SELECT count(*) FROM KeyboardS
	CountKeyboards() (int64, error)

	// @default SELECT count(*) FROM motherboards
	CountMotherboards() (int64, error)
}
