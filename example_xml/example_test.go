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
		DbCompatibility: true,
		DriverName:      tests.TestDrv,
		DataSource:      tests.GetTestConnURL(),
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
	err = gobatis.ExecContext(context.Background(), factory.DB(), sqltxt)
	if err != nil {
		if e, ok := err.(*gobatis.SqlError); ok {
			fmt.Println(e.SQL)
		}
		fmt.Println(err)
		return
	}

	insertUser := tests.User{
		Name:        "abc",
		Nickname:    "nickabc",
		Description: "123",
		Password:    "xxx",
		Address:     "as",
		Sex:         "male",
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

	list, err := userDao.SelectAll("", "all", nil)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println("fetch all user from database!")
	for _, u := range list {
		fmt.Println(u.Nickname)
	}

	listmap, err := userDao.SelectAllForMap("", "all", nil)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println("fetch all user from database!")
	for _, u := range listmap {
		fmt.Println(u["nickname"])
	}

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
	// fetch all user from database!
	// ABC
	// fetch all user from database!
	// ABC
	// delete success!
}
