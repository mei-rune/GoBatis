package example_xml

import (
	"context"
	"fmt"
	"os"

	gobatis "github.com/runner-mei/GoBatis"
	"github.com/runner-mei/GoBatis/tests"
)

func ExampleSimple() {
	factory, err := gobatis.New(&gobatis.Config{
		DriverName: tests.TestDrv,
		DataSource: tests.GetTestConnURL(),
		XMLPaths: []string{
			// 在不同目录下运行测试试，路径可能会不对，所以多试几次
			"xmlfiles",
			"example_xml/xmlfiles",
			"../example_xml/xmlfiles",
		},
		MaxIdleConns: 2,
		MaxOpenConns: 2,
		// ShowSQL:      false,
		Tracer: gobatis.TraceWriter{Output: os.Stderr},
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

	sqltxt := tests.GetTestSQLText(factory.Dialect().Name())
	_, err = factory.DB().ExecContext(context.Background(), sqltxt)
	
	if err != nil {
		fmt.Println(err)
		return
	}

	insertUser := tests.User{
		Name:        "abc",
		Description: "123",
	}

	ref := factory.SessionReference()
	userDao := NewUsers(ref)
	id, err := userDao.Insert(&insertUser)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println("insert success!")

	u, err := userDao.FindByID(id)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println("fetch user from database!")
	fmt.Println(u.Name)

	u.Nickname = "ABC"
	count, err := userDao.Update(id, u)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println("update user:", count)

	u, err = userDao.FindByID(id)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println("fetch user from database!")
	fmt.Println(u.Nickname)

	_, err = userDao.DeleteByID(id)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println("delete success!")

	// Output:
	// insert success!
	// fetch user from database!
	// abc
	// update user: 1
	// fetch user from database!
	// ABC
	// delete success!
}
