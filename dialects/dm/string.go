package dm

import (
	"fmt"
	"gitee.com/runner.mei/dm" // 达梦
	"github.com/runner-mei/GoBatis/dialects"
)

func init() {
	dialects.SetNewDMClob(func(target *string) dialects.Clob {
		return &clob{DefaultClob: dialects.DefaultClob{Target: target}}
	})
	dialects.SetNewDMBlob(func(target *[]byte) dialects.Blob {
		return &blob{DefaultBlob: dialects.DefaultBlob{Target: target}}
	})
}

type clob struct {
	dialects.DefaultClob
}

// Scan implements the Scanner interface.
func (n *clob) Scan(value interface{}) error {
	fmt.Println("=======", fmt.Sprintf("%T %#v", value, value))
	if value == nil {
		n.Invalid()
		return nil
	}

	if dmClob, ok := value.(*dm.DmClob); ok {
		l, err := dmClob.GetLength()
		if err != nil {
			return err
		}
		if l == 0 {
			n.Set("")
			return nil
		}
		str, err := dmClob.ReadString(1, int(l))
		if err != nil {
			return err
		}
		n.Set(str)
		return nil
	}
	return n.DefaultClob.Scan(value)
}

type blob struct {
	dialects.DefaultBlob
}

func (blob *blob) Scan(src interface{}) error {
	if dmBlob, ok := src.(*dm.DmBlob); ok {
		if dmBlob == nil {
			blob.Invalid()
			return nil
		}
		l, err := dmBlob.GetLength()
		if err != nil {
			return err
		}
		if l == 0 {
			blob.Set(nil)
			return nil
		}
		bs := make([]byte, int(l))
		_, err = dmBlob.Read(bs)
		if err != nil {
			return err
		}
		blob.Set(bs)
		return nil
	}
	return blob.DefaultBlob.Scan(src)
}
