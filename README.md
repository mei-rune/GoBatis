# GoBatis

[![Build Status](https://travis-ci.org/runner-mei/GoBatis.svg?branch=master)](https://travis-ci.org/runner-mei/GoBatis)
[![Semver](http://img.shields.io/SemVer/0.9.5.png)](http://semver.org/spec/v0.9.5.html)
[![Coverage Status](https://coveralls.io/repos/github/runner-mei/GoBatis/badge.svg?branch=master)](https://coveralls.io/github/runner-mei/GoBatis?branch=master)



## [文档](https://runner-mei.github.io/GoBatis)

## Introduction

An easy ORM tool for Golang, support MyBatis-Like XML template SQL


## 基本思路
1. 用户定义结构和接口
2. 在接口的方法上定义 sql （可以在 xml 或 方法的注释中）
3. 用工具生成接口的实现
4. 创建接口的实例并使用它


## Usage

1. install `gobatis` tools.

    `go get -u -v github.com/runner-mei/GoBatis/cmd/gobatis`


2. Define a struct, interface and comment methods with SQLs and Variables, then write a directive `//go:generate gobatis user.go`.

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

3. After that, run `go generate ./...` ， user.gobatis.go is generated

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

4. use UserDao

````go
  factory, err := gobatis.New(&gobatis.Config{DriverName: tests.TestDrv,
    DataSource: tests.TestConnURL,
    // XMLPaths: []string{"example/test.xml"},
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

  _, err = userDao.Delete(id)
  if err != nil {
    fmt.Println(err)
    return
  }
  fmt.Println("delete success!")
````
     

## 待完成的任务
1. 为 sql 语句的 ‘?’ 的支持，如 
    select * from user where id = ?
    当数据库为 postgresql 能自动转成 select * from user where id = $1
2. 对象继承的实现
3. 延迟加载的实现


## 注意
GoBatis 是基于 [osm](https://github.com/yinshuwei/osm) 的基础上修改来的，goparser 则是在 [light](https://github.com/arstd/light) 的基础上修改来的, reflectx 则从 [sqlx](https://github.com/jmoiron/sqlx) 拷贝过来的
