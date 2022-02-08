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

func newClob(target *string) Clob {
	return &DefaultClob{Target: target}
}

type DefaultClob struct {
	Str sql.NullString
	Target *string
}

func (clob *DefaultClob) Scan(src interface{}) error {	
	err := clob.Str.Scan(src)
	if err != nil {
		return err
	}
	if clob.Target != nil {
		if clob.Str.Valid {
			*clob.Target = clob.Str.String
		} else {
			*clob.Target = ""
		}
	}
	return nil
}

func (clob *DefaultClob) Invalid() Clob {
	clob.Str.Valid = false
	clob.Str.String = ""

	if clob.Target != nil {
		*clob.Target = ""
	}
	return clob
}

func (clob *DefaultClob) Set(s string) Clob {
	clob.Str.Valid = true
	clob.Str.String = s

	if clob.Target != nil {
		*clob.Target = s
	}
	return clob
}

func (clob *DefaultClob) IsValid() bool {
	return clob.Str.Valid
}

func (clob *DefaultClob) Length() int {
	return len(clob.Str.String)
}

func (clob *DefaultClob) String() string {
	if clob.Str.Valid {
		return clob.Str.String
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

func newBlob(target *[]byte) Blob {
	return &DefaultBlob{Target: target}
}

type DefaultBlob struct {
	Target *[]byte
}

func (blob *DefaultBlob) Scan(src interface{}) error {
	if src == nil {
		blob.Invalid()
		return nil
	}

	switch bs := src.(type) {
	case []byte:
		blob.Set(bs)
		return nil
	case string:
		blob.Set([]byte(bs))
		return nil
	case *[]byte:
		if bs != nil {
			blob.Set(*bs)
		} else {
			blob.Set(nil)
		}
	case *string:
		if bs != nil {
			blob.Set([]byte(*bs))
		} else {
			blob.Set(nil)
		}
		return nil
	}
	return fmt.Errorf("unsupported Scan, storing driver.Value type %T into type Blob", src)
}

func (blob *DefaultBlob) Invalid() Blob {
	if blob.Target != nil {
		*blob.Target = nil
	}
	return blob
}

func (blob *DefaultBlob) Set(bs []byte) Blob {
	if blob.Target == nil {
		if bs != nil {
			blob.Target = &bs
		} else {
			// 不用重新赋值
		}
	} else {
		*blob.Target = bs
	}
	return blob
}

func (blob *DefaultBlob) IsValid() bool {
	return blob.Target != nil && *blob.Target != nil
}

func (blob *DefaultBlob) Length() int {
	if blob.Target == nil {
		return 0
	}
	return len(*blob.Target)
}

func (blob *DefaultBlob) Bytes() []byte {
	if blob.Target == nil {
		return nil
	}
	return *blob.Target
}
