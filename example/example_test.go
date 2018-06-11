package example

import (
	"fmt"

	gobatis "github.com/runner-mei/GoBatis"
	"github.com/runner-mei/GoBatis/tests"
)

func ExampleSimple() {
	insertUser := AuthUser{
		Username: "abc",
		Phone:    "123",
		Status:   1,
	}

	factory, err := gobatis.New(&gobatis.Config{DriverName: tests.TestDrv,
		DataSource: tests.TestConnURL,
		//XMLPaths: []string{"example/test.xml"},
		ShowSQL: false,
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

	switch factory.DbType() {
	case gobatis.DbTypePostgres:
		_, err = factory.DB().Exec(postgres)
	case gobatis.DbTypeMSSql:
		_, err = factory.DB().Exec(mssql)
	default:
		_, err = factory.DB().Exec(mysql)
	}
	if err != nil {
		fmt.Println(err)
		return
	}

	ref := factory.Reference()
	userDao := NewAuthUserDao(&ref, NewUserProfiles(&ref))
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
	userDaoInTx := NewAuthUserDao(&txref, NewUserProfiles(&txref))
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
	// fetch username from database!
	// abc
	// fetch usernames from database!
	// [abc]
	// delete success!
	// insert success!
	// delete success!
}
