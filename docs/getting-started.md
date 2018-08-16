
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

## 3. 然后运行 `go generate ./...` 命令，生成代码文件 user.gobatis.go

    # go generate ./...

````go
// Please don't edit this file!
package example

import (
  "errors"

  gobatis "github.com/runner-mei/GoBatis"
)

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

func (impl *UserDaoImpl) Get(id int64) (*AuthUser, error) {
  var instance = &AuthUser{}

  err := impl.session.SelectOne("UserDao.Get",
    []string{
      "id",
    },
    []interface{}{
      id,
    }).Scan(instance)
  if err != nil {
    return nil, err
  }
  return instance, nil
}


func (impl *UserDaoImpl) Update(id int64, u *AuthUser) (int64, error) {
  return impl.session.Update("UserDao.Update",
    []string{
      "id",
      "u",
    },
    []interface{}{
      id,
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