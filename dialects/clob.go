package dialects

import "database/sql"

type Clob interface {
	sql.Scanner
	SetZero() Clob
	SetString(s string) Clob

	IsValid() bool
	Length() int
	String() string
}

func newClob() Clob {
	return &defaultClob{}
}

type defaultClob struct {
	str sql.NullString
}

func (clob *defaultClob) Scan(src interface{}) error {
	return clob.str.Scan(src)
}

func (clob *defaultClob) SetZero() Clob {
	clob.str.Valid = false
	clob.str.String = ""
	return clob
}

func (clob *defaultClob) SetString(s string) Clob {
	clob.str.Valid = true
	clob.str.String = s
	return clob
}

func (clob *defaultClob) IsValid() bool {
	return clob.str.Valid
}

func (clob *defaultClob) Length() int {
	return len(clob.str.String)
}

func (clob *defaultClob) String() string {
	if clob.str.Valid {
		return clob.str.String
	}
	return ""
}
