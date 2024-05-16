# GoBatis

[![GoDoc](https://godoc.org/github.com/runner-mei/GoBatis?status.svg)](https://godoc.org/github.com/runner-mei/GoBatis)
[![Travis Build Status](https://travis-ci.org/runner-mei/GoBatis.svg?branch=master)](https://travis-ci.org/runner-mei/GoBatis)
[![GitHub Actions](https://github.com/runner-mei/GoBatis/actions/workflows/test.yml/badge.svg)](https://github.com/runner-mei/GoBatis/actions)
![GitHub tag (latest SemVer)](https://img.shields.io/github/tag/runner-mei/GoBatis.svg)
[![Coverage Status](https://coveralls.io/repos/github/runner-mei/GoBatis/badge.svg?branch=master)](https://coveralls.io/github/runner-mei/GoBatis?branch=master)
[![Appveyor Build status](https://ci.appveyor.com/api/projects/status/hmg1mecib5j46r55?svg=true)](https://ci.appveyor.com/project/runner-mei/gobatis)


## [中文文档](https://runner-mei.github.io/GoBatis)

## Introduction

An easy ORM tool for Golang, support MyBatis-Like XML template SQL

## 待完成的任务
1. 对象继承的实现
2. 延迟加载或加密字段(或特殊处理)的实现
     有泛型了，可以尝试下 
     ````go
     type Lazy[T any] struct {
        value T
        session SqlSession
        sqlstr string
     }
     func (l *Lazy[T]) Read() T {
         session.Query()
     }

     // 加密字段(或特殊处理)
     type Passworder struct {
        value string
     }
     func (p *Passworder) Scan(interface{}) error {
         xxxxx
     }
     func (p *Passworder) Value() driver.Value {
         xxxxx
     }

     
     type Record struct {
     TableName struct{}    `db:records`
     Blob   Lazy[[]byte]   `db:"blob"`
     Password Passworder   `db:"password"`
     }

     ``````

3. 返回大量数据记录时用泛型来改进
   ````go
   type Results[T any] struct  {}
   func (rs *Results) Next() bool {}
   func (rs *Results) Read(value *T) error {}
   ````

## 已知 bug

1. 当 sql 中含有 xml 标签时 <code>&lt; </code> 号需要转义为 <code>&amp;lt; </code>，而不含 xml 标签时<code>&amp;lt; </code> 又不转义为 <code>&lt; </code>, 这很不一致。
   最近我改进了一个像 mybatis 一样用 gt, gte，lt 和 lte 代替 >,>=, < 和 <=, 如

     a > 8 写成 a gt 8
   
     a >= 8 写成 a gte 8

     a < 8 写成 a lt 8

     a <= 8 写成 a lte 8

3. 达梦数据库实现 upsert 时无法返回 insert id (达梦数据库的问题)。


## 和 MyBatis 的区别

GoBatis 就是对 MyBatis 的简单模仿。 但有下列不同

  1. 动态 sql 语句的格式

     我实现一个和  mybatis 类似的 if, chose, foreach, trim, set 和 where 之类的 xml 基本实现，同时也支持 go template 来生成 sql。

     1.1 另外我不支持 ${xxx}, 但是我提供了一个更安全的 <print fmt="%s" value="b" inStr="true" /> 来替换它

     当有  inStr="true" 时我会检查 value 的值中是不是有 引号之类的字符，防止 sql 注入
     当    inStr="false" 时我会检查 value 的值中是不是有 and 或  or 之类的逻辑表达式，防止 sql 注入

     1.2 我为 if 标签 增加了 else 支持， 用法为 <if> xx <else /> xxx </if>

  3. 自动生成 sql 语句

     MyBatis 是不会自动生成 sql 语句的， 我觉得能像大部份的 orm 一样能生成 sql 的话，可以省很多工作
     请见 [SQL 自动生成](https://runner-mei.github.io/GoBatis/#/sql_genrate)


## 基本思路
1. 用户定义结构和接口
2. 在接口的方法上定义 sql （可以在 xml 中或方法的注释中）
3. 用工具生成接口的实现
4. 创建接口的实例并使用它

## Roadmap
1. 升级 go1.14 后 goparser 特别慢，准备用 goparser2 替换
2. 将 xml 相关代码移到 xml 子目录
3. 将 sql 生成工具 builder 相关代码移到 sql 子目录

## Usage

注意， gobatis 也支持 xml, 请见 example_xml 目录

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
        case gobatis.NewDialect("mssql"):
          sqlStr = "insert into auth_users(username, phone, address, status, birth_day, created_at, updated_at)\r\n output inserted.id\r\n values (#{username},#{phone},#{address},#{status},#{birth_day},CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)"
        case gobatis.NewDialect("postgres"):
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

func NewUserDao(session gobatis.SqlSession) UserDao {
  return &UserDaoImpl{session: session}
}

type UserDaoImpl struct {
  session gobatis.SqlSession
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
    DataSource: tests.GetTestConnURL(),
    // XMLPaths: []string{"example/test.xml"},
    })
    
  userDao := NewUserDao(factory.SessionReference())
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


## 注意
GoBatis 是基于 [osm](https://github.com/yinshuwei/osm) 的基础上修改来的，goparser 则是在 [light](https://github.com/arstd/light) 的基础上修改来的, reflectx 则从 [sqlx](https://github.com/jmoiron/sqlx) 拷贝过来的
