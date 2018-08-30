
# 用法

## 1. 安装 `gobatis` 工具.

    `go get -u -v github.com/runner-mei/GoBatis/cmd/gobatis`


## 2. 定义一个接口，并用方法上的注释定义 SQL 和变量, 然后在源码中加上 `//go:generate gobatis user.go`.

````go
//go:generate gobatis user.go
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

type UserDao interface {
  // @postgres insert into auth_users(username, phone, address, status, birth_day, created_at, updated_at)
  // values (#{username},#{phone},#{address},#{status},#{birth_day},CURRENT_TIMESTAMP, CURRENT_TIMESTAMP) returning id
  //
  // @default insert into auth_users(username, phone, address, status, birth_day, created_at, updated_at)
  // values (#{username},#{phone},#{address},#{status},#{birth_day},CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
  Insert(u *AuthUser) (int64, error)
}

````

## 3. 然后运行 `go generate ./...` 命令，生成代码文件 user.gobatis.go

    # go generate ./...

````go
// Please don't edit this file!
package example

import (
  "errors"

  gobatis "github.com/runner-mei/GoBatis"
)

func init() {
  gobatis.Init(func(ctx *gobatis.InitContext) error {
    { //// UserDao.Insert
      if _, exists := ctx.Statements["UserDao.Insert"]; !exists {
        sqlStr := "insert into auth_users(username, phone, address, status, birth_day, created_at, updated_at)\r\n values (#{username},#{phone},#{address},#{status},#{birth_day},CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)"
        switch ctx.Dialect {
        case gobatis.ToDbType("mssql"):
          sqlStr = "insert into auth_users(username, phone, address, status, birth_day, created_at, updated_at)\r\n output inserted.id\r\n values (#{username},#{phone},#{address},#{status},#{birth_day},CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)"
        case gobatis.ToDbType("postgres"):
          sqlStr = "insert into auth_users(username, phone, address, status, birth_day, created_at, updated_at)\r\n values (#{username},#{phone},#{address},#{status},#{birth_day},CURRENT_TIMESTAMP, CURRENT_TIMESTAMP) returning id"
        }
        stmt, err := gobatis.NewMapppedStatement(ctx, "UserDao.Insert",
          gobatis.StatementTypeInsert,
          gobatis.ResultStruct,
          sqlStr)
        if err != nil {
          return err
        }
        ctx.Statements["UserDao.Insert"] = stmt
      }
    }
  })
}

func NewUserDao(ref *gobatis.Reference) UserDao {
  return &UserDaoImpl{session: ref}
}

type UserDaoImpl struct {
  session *gobatis.Reference
}

func (impl *UserDaoImpl) Insert(u *AuthUser) (int64, error) {
  return impl.session.Insert("UserDao.Insert",
    []string{
      "u",
    },
    []interface{}{
      u,
    })
}

...

````

## 4. 创建接口的实例

````go
  factory, err := gobatis.New(&gobatis.Config{DriverName: tests.TestDrv,
    DataSource: tests.TestConnURL,
    //XMLPaths: []string{"example/test.xml"},
    })
    
  ref := factory.Reference()
  userDao := NewUserDao(&ref)
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