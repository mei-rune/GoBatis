package sqlite

import (
	"strings"

	_ "github.com/glebarez/go-sqlite"
	"github.com/runner-mei/GoBatis/dialects"
)
func init() {
  dialects.SetHandleError(dialects.Sqlite.Name(), handleError)
}

func handleError(e error) error {
  if e == nil {
    return nil
  }

  
  // if se, ok := e.(*sqlite.Error); ok {

  // }

  if strings.Contains(e.Error(), "no such table:") {
  	return dialects.ErrTableNotExists{
				Err:       e,
				Tablename: "xxx",
			}
  }

  return e
}
