package dialects

import (
	"database/sql"
	"fmt"
)

type Clob interface {
	sql.Scanner
	Invalid() Clob
	Set(s string) Clob

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

func (clob *defaultClob) Invalid() Clob {
	clob.str.Valid = false
	clob.str.String = ""
	return clob
}

func (clob *defaultClob) Set(s string) Clob {
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

type Blob interface {
	sql.Scanner
	Invalid() Blob
	Set(s []byte) Blob

	IsValid() bool
	Length() int
	Bytes() []byte
}

func newBlob() Blob {
	return &defaultBlob{}
}

type defaultBlob struct {
	bs []byte
}

func (blob *defaultBlob) Scan(src interface{}) error {
	if src == nil {
		blob.bs = nil
		return nil
	}

	switch bs := src.(type) {
	case []byte:
		blob.bs = bs
		return nil
	case string:
		blob.bs = []byte(bs)
		return nil
	case *[]byte:
		blob.bs = *bs
	case *string:
		if bs == nil {
			blob.bs = nil
		} else {
			blob.bs = []byte(*bs)
		}
		return nil
	}
	return fmt.Errorf("unsupported Scan, storing driver.Value type %T into type Blob", src)
}

func (blob *defaultBlob) Invalid() Blob {
	blob.bs = nil
	return blob
}

func (blob *defaultBlob) Set(bs []byte) Blob {
	blob.bs = bs
	return blob
}

func (blob *defaultBlob) IsValid() bool {
	return blob.bs != nil
}

func (blob *defaultBlob) Length() int {
	return len(blob.bs)
}

func (blob *defaultBlob) Bytes() []byte {
	return blob.bs
}
