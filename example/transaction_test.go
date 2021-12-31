package example

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"

	gobatis "github.com/runner-mei/GoBatis"
	"github.com/runner-mei/GoBatis/tests"
)

func ExampleTx1() {
	insertUser := User{
		Username: "abc",
		Phone:    "123",
		Status:   1,
	}

	factory, err := gobatis.New(&gobatis.Config{
		Tracer:     gobatis.StdLogger{Logger: log.New(os.Stderr, "", log.Lshortfile)},
		DriverName: tests.TestDrv,
		DataSource: tests.GetTestConnURL(),
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

	sqltext := GetTestSQL(factory.Dialect().Name())
	err = gobatis.ExecContext(context.Background(), factory.DB(), sqltext)
	if err != nil {
		fmt.Println(factory.Dialect().Name())
		fmt.Println(sqltext)
		fmt.Println(err)
		return
	}

	conn := factory.DB().(*sql.DB)
	ref := factory.SessionReference()
	userDao := NewUserDao(ref, NewUserProfiles(ref))

	tx, err := conn.Begin()
	if err != nil {
		fmt.Println("begin tx:", err)
		return
	}

	ctxWithTx := gobatis.WithTx(context.Background(), tx)
	id, err := userDao.InsertWithContext(ctxWithTx, &insertUser)
	if err != nil {
		fmt.Println("insert in tx:", err)
		return
	}

	if err = tx.Commit(); err != nil {
		fmt.Println("commit tx:", err)
		return
	}

	_, err = userDao.Delete(id)
	if err != nil {
		fmt.Println("delete:", err)
		return
	}
	tx, err = conn.Begin()
	if err != nil {
		fmt.Println("begin tx:", err)
		return
	}
	ctxWithTx = gobatis.WithTx(context.Background(), tx)

	_, err = userDao.InsertWithContext(ctxWithTx, &insertUser)
	if err != nil {
		fmt.Println("insert tx:", err)
		return
	}

	if err = tx.Rollback(); err != nil {
		fmt.Println("rollback tx:", err)
		return
	}

	c, err := userDao.Count()
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

func ExampleTx2() {
	insertUser := User{
		Username: "abc",
		Phone:    "123",
		Status:   1,
	}

	factory, err := gobatis.New(&gobatis.Config{
		Tracer:     gobatis.StdLogger{Logger: log.New(os.Stderr, "", log.Lshortfile)},
		DriverName: tests.TestDrv,
		DataSource: tests.GetTestConnURL(),
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

	sqltext := GetTestSQL(factory.Dialect().Name())
	err = gobatis.ExecContext(context.Background(), factory.DB(), sqltext)
	if err != nil {
		fmt.Println(factory.Dialect().Name())
		fmt.Println(sqltext)
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
