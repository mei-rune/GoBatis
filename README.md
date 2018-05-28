# GoBatis

[![Build Status](https://travis-ci.org/runner-mei/GoBatis.svg?branch=master)](https://travis-ci.org/runner-mei/GoBatis)
[![Semver](http://img.shields.io/SemVer/0.5.1.png)](http://semver.org/spec/v0.5.1.html)
[![Coverage Status](https://coveralls.io/repos/github/runner-mei/GoBatis/badge.svg?branch=master)](https://coveralls.io/github/runner-mei/GoBatis?branch=master)

GoBatis 是用 golang 编写的 ORM 工具，目前已在生产环境中使用，理论上支持任何数据库 (只测试过 postgresql)。

GoBatis 就是对 MyBatis 的简单模仿。当然动态sql的生成是使用go和template包，所以sql mapping的格式与MyBatis的不同。

GoBatis 是基于 [osm](https://github.com/yinshuwei/osm) 的基础上修改来的，goparser 则是在 [light](https://github.com/arstd/light) 的基础上修改来的, reflectx 则从 [sqlx](https://github.com/jmoiron/sqlx) 拷贝过来的


### 思路
1. 用户定义对象和接口
2. 用工具生成接口的实现
3. 创建接口的实例并使用它



### 用法

1. 安装 `gobatis` 工具.

    `go get -u -v github.com/runner-mei/GoBatis/cmd/gobatis`


2. 定义一个接口，并用方法上的注释定义 SQL 和变量, 然后在源码中加上 `//go:generate gobatis`.

````go
//go:generate gobatis
package example

import (
	"time"
)

type AuthUser struct {
	ID        int64      `json:"id"`
	Username  string     `json:"username"`
	Phone     string     `json:"phone"`
	Address   *string    `json:"address"`
	Status    uint8      `json:"status"`
	BirthDay  *time.Time `json:"birth_day"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
}

type AuthUserDao interface {
	// @postgres insert into auth_users(username, phone, address, status, birth_day, created_at, updated_at)
	// values (#{username},#{phone},#{address},#{status},#{birth_day},CURRENT_TIMESTAMP, CURRENT_TIMESTAMP) returning id
	//
	// @default insert into auth_users(username, phone, address, status, birth_day, created_at, updated_at)
	// values (#{username},#{phone},#{address},#{status},#{birth_day},CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
	Insert(u *AuthUser) (int64, error)

	// @mysql insert into auth_users(username, phone, address, status, birth_day, created_at, updated_at)
	// values (?,?,?,?,?,CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
	// on duplicate key update
	//   username=values(username), phone=values(phone), address=values(address),
	//   status=values(status), birth_day=values(birth_day), updated_at=CURRENT_TIMESTAMP
	//
	// @postgres insert into auth_users(username, phone, address, status, birth_day, created_at, updated_at)
	// values (?,?,?,?,?,CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
	// on duplicate key update
	//   username=values(username), phone=values(phone), address=values(address),
	//   status=values(status), birth_day=values(birth_day), updated_at=CURRENT_TIMESTAMP
	Upsert(u *AuthUser) (int64, error)

	// @default UPDATE auth_users
	// SET username=#{u.username},
	//     phone=#{u.phone},
	//     address=#{u.address},
	//     status=#{u.status},
	//     birth_day=#{u.birth_day},
	//     updated_at=CURRENT_TIMESTAMP
	// WHERE id=#{id}
	Update(id int64, u *AuthUser) (int64, error)

	// @default UPDATE auth_users
	// SET username=#{username},
	//     updated_at=CURRENT_TIMESTAMP
	// WHERE id=#{id}
	UpdateName(id int64, username string) (int64, error)

	// @postgres DELETE FROM auth_users
	// @default DELETE FROM auth_users
	DeleteAll() (int64, error)

	// @postgres DELETE FROM auth_users WHERE id=$1
	// @default DELETE FROM auth_users WHERE id=?
	Delete(id int64) (int64, error)

	// @postgres select * FROM auth_users WHERE id=$1
	// @default select * FROM auth_users WHERE id=?
	Get(id int64) (*AuthUser, error)

	// @default select count(*) from auth_users
	Count() (int64, error)

	// @default select * from auth_users limit #{offset}, #{size}
	List(offset, size int) ([]*AuthUser, error)

	// @default select username from auth_users where id = #{id}
    GetNameByID(id int64) (string, error)
}

````

2. 然后运行 `go generate ./...` 命令，生成代码文件 user.gobatis.go

	# go generate ./...

````go
// Please don't edit this file!
package example

import (
	"errors"

	gobatis "github.com/runner-mei/GoBatis"
)

func NewAuthUserDao(ref *gobatis.Reference) AuthUserDao {
	return &AuthUserDaoImpl{session: ref}
}

type AuthUserDaoImpl struct {
	session *gobatis.Reference
}

func (impl *AuthUserDaoImpl) Insert(u *AuthUser) (int64, error) {
	return impl.session.Insert("AuthUserDao.Insert",
		[]string{
			"u",
		},
		[]interface{}{
			u,
		})
}
 ...

````

3. 创建接口的实例

````go

	factory, err := gobatis.New(&gobatis.Config{DriverName: tests.TestDrv,
		DataSource: tests.TestConnURL,
		//XMLPaths: []string{"example/test.xml"},
    })
    
	ref := factory.Reference()
	userDao := NewAuthUserDao(&ref)
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

	_, err = userDao.Delete(id)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println("delete success!")
````

更详细的例子请见 example/example_test.go

