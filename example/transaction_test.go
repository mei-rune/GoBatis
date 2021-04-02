package example

import (
	"context"
	"fmt"
	"log"
	"os"

	gobatis "github.com/runner-mei/GoBatis"
	"github.com/runner-mei/GoBatis/tests"
)

func ExampleTx() {
	insertUser := User{
		Username: "abc",
		Phone:    "123",
		Status:   1,
	}

	factory, err := gobatis.New(&gobatis.Config{
		Tracer:     gobatis.StdLogger{Logger: log.New(os.Stderr, "", log.Lshortfile)},
		DriverName: tests.TestDrv,
		DataSource: tests.TestConnURL,
		//XMLPaths: []string{"example/test.xml"},
		//ShowSQL: false,
	})
	if err != nil {
		fmt.Println(err)
		return
	}

	defer func() {
		if err = factory.Close(); err != nil {
			fmt.Println(err)
		}
	}()

	switch factory.Dialect() {
	case gobatis.DbTypePostgres:
		_, err = factory.DB().ExecContext(context.Background(), postgres)
	case gobatis.DbTypeMSSql:
		_, err = factory.DB().ExecContext(context.Background(), mssql)
	default:
		_, err = factory.DB().ExecContext(context.Background(), mysql)
	}
	if err != nil {
		fmt.Println(err)
		return
	}

	conn := NewConnection(factory)

	tx, err := conn.Begin()
	if err != nil {
		fmt.Println("begin tx:", err)
		return
	}

	id, err := tx.Users().Insert(&insertUser)
	if err != nil {
		fmt.Println("insert in tx:", err)
		return
	}

	if err = tx.Commit(); err != nil {
		fmt.Println("commit tx:", err)
		return
	}

	_, err = conn.Users().Delete(id)
	if err != nil {
		fmt.Println("delete:", err)
		return
	}
	tx, err = conn.Begin()
	if err != nil {
		fmt.Println("begin tx:", err)
		return
	}

	_, err = tx.Users().Insert(&insertUser)
	if err != nil {
		fmt.Println("insert tx:", err)
		return
	}

	if err = tx.Rollback(); err != nil {
		fmt.Println("rollback tx:", err)
		return
	}

	c, err := conn.Users().Count()
	if err != nil {
		fmt.Println("count", err)
		return
	}
	if c != 0 {
		fmt.Println("want 0 got", c)
	} else {
		fmt.Println("test ok!")
	}

	// Output:
	// test ok!
}