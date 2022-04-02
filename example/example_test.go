package example

import (
	"context"
	"fmt"
	"log"
	"os"

	gobatis "github.com/runner-mei/GoBatis"
	"github.com/runner-mei/GoBatis/tests"
)

func ExampleSimple() {
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
			if e, ok := err.(*gobatis.SqlError); ok {
				t.Error(e.SQL)
			}
		fmt.Println(err)
		return
	}

	ref := factory.SessionReference()
	userDao := NewUserDao(ref, NewUserProfiles(ref))
	id, err := userDao.Insert(&insertUser)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println("insert success!")

	u, err := userDao.Get(id)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println("fetch user from database!")
	fmt.Println(u.Username)

	uv, err := userDao.GetNonPtr(id)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println("fetch user from database!")
	fmt.Println(uv.Username)

	username, err := userDao.GetName(id)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println("fetch username from database!")
	fmt.Println(username)

	usernames, err := userDao.GetNames()
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println("fetch usernames from database!")
	fmt.Println(usernames)

	usernameByID, err := userDao.GetIDNames()
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(usernameByID[id])

	_, err = userDao.Delete(id)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println("delete success!")

	tx, err := factory.Begin()
	if err != nil {
		fmt.Println(err)
		return
	}
	txref := factory.Reference()
	userDaoInTx := NewUserDao(&txref, NewUserProfiles(&txref))
	id, err = userDaoInTx.Insert(&insertUser)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println("insert success!")
	if err = tx.Commit(); err != nil {
		fmt.Println(err)
		return
	}

	_, err = userDao.Delete(id)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println("delete success!")

	// Output:
	// insert success!
	// fetch user from database!
	// abc
	// fetch user from database!
	// abc
	// fetch username from database!
	// abc
	// fetch usernames from database!
	// [abc]
	// abc
	// delete success!
	// insert success!
	// delete success!
}
