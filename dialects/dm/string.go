package dm

import (
	"fmt"

	"gitee.com/runner.mei/dm" // 达梦
	"github.com/runner-mei/GoBatis/dialects"
)

func init() {
	dialects.NewDMClob = func() dialects.Clob {
		return &clob{}
	}
	dialects.NewDMBlob = func() dialects.Blob {
		return &blob{}
	}
}

type clob struct {
	Str   string
	Valid bool // Valid is true if Int64 is not NULL
}

// Scan implements the Scanner interface.
func (n *clob) Scan(value interface{}) error {
	if value == nil {
		n.Str, n.Valid = "", false
		return nil
	}
	switch s := value.(type) {
	case []byte:
		if s != nil {
			n.Valid = true
			n.Str = string(s)
		} else {
			n.Str, n.Valid = "", false
		}
		return nil
	case string:
		n.Valid = true
		n.Str = s
		return nil
	case *[]byte:
		if s != nil && *s != nil {
			n.Valid = true
			n.Str = string(*s)
		} else {
			n.Str, n.Valid = "", false
		}
		return nil
	case *string:
		if s == nil {
			n.Str, n.Valid = "", false
		} else {
			n.Valid = true
			n.Str = *s
		}
		return nil
	case *dm.DmClob:
		l, err := s.GetLength()
		if err != nil {
			return err
		}
		if l == 0 {
			n.Valid = true
			return nil
		}
		n.Str, err = s.ReadString(1, int(l))
		if err != nil {
			return err
		}
		n.Valid = true
		return nil
	}
	return fmt.Errorf("unsupported Scan, storing driver.Value type %T into type Clob", value)
}

func (clob *clob) Invalid() dialects.Clob {
	clob.Valid = false
	clob.Str = ""
	return clob
}

func (clob *clob) Set(s string) dialects.Clob {
	clob.Valid = true
	clob.Str = s
	return clob
}

func (clob *clob) IsValid() bool {
	return clob.Valid
}

func (clob *clob) Length() int {
	return len(clob.Str)
}

func (clob *clob) String() string {
	if clob.Valid {
		return clob.Str
	}
	return ""
}

type blob struct {
	bs []byte
}

func (blob *blob) Scan(src interface{}) error {
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
	case *dm.DmBlob:
		if bs == nil {
			blob.bs = nil
			return nil
		}
		l, err := bs.GetLength()
		if err != nil {
			return err
		}
		if l == 0 {
			blob.bs = nil
			return nil
		}
		blob.bs = make([]byte, int(l))
		_, err = bs.Read(blob.bs)
		if err != nil {
			blob.bs = nil
			return err
		}
		return nil
	}
	return fmt.Errorf("unsupported Scan, storing driver.Value type %T into type Blob", src)
}

func (blob *blob) Invalid() dialects.Blob {
	blob.bs = nil
	return blob
}

func (blob *blob) Set(bs []byte) dialects.Blob {
	blob.bs = bs
	return blob
}

func (blob *blob) IsValid() bool {
	return blob.bs != nil
}

func (blob *blob) Length() int {
	return len(blob.bs)
}

func (blob *blob) Bytes() []byte {
	return blob.bs
}
