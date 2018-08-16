# GoBatis

[![Build Status](https://travis-ci.org/runner-mei/GoBatis.svg?branch=master)](https://travis-ci.org/runner-mei/GoBatis)
[![Semver](http://img.shields.io/SemVer/0.5.1.png)](http://semver.org/spec/v0.5.1.html)
[![Coverage Status](https://coveralls.io/repos/github/runner-mei/GoBatis/badge.svg?branch=master)](https://coveralls.io/github/runner-mei/GoBatis?branch=master)



[文档](https://runner-mei.github.io/GoBatis)

## 简介

GoBatis 是用 golang 编写的 ORM 工具，目前已在生产环境中使用，理论上支持任何数据库 (只测试过 postgresql, mysql, mssql)。

GoBatis 就是对 MyBatis 的简单模仿。但有下列不同
  1. 动态 sql 语句的格式
     我没有时间来实现一个和  mybatis 类似的 if else end 或 for 之类的 xml 执行实现，而是简单地用 go template 包来生成动态sql的生成是使用go和template包，所以sql mapping的格式与MyBatis的不同。
  2. 自动生成 sql 语句
     MyBatis 是不会自动生成 sql 语句的， 我觉得能像大部份的 orm 一样能生成 sql，可以省很多工作

## 待完成的任务
1. 为 sql 语句的 ‘?’ 的支持，如 
    select * from user where id = ?
    当数据库为 postgresql 能自动转成 select * from user where id = $1
2. 增加命名参数的支持， 如 `select * from user where id = :id`
3. 对象继承的实现

## 思路
1. 用户定义对象和接口
2. 在接口的方法上定义 sql
2. 用工具生成接口的实现
3. 创建接口的实例并使用它


## 注意
GoBatis 是基于 [osm](https://github.com/yinshuwei/osm) 的基础上修改来的，goparser 则是在 [light](https://github.com/arstd/light) 的基础上修改来的, reflectx 则从 [sqlx](https://github.com/jmoiron/sqlx) 拷贝过来的