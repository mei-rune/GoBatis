# 接口定义规范

定义接口时， 对接口中的方法是有一些要求的，不然代码生成工具也无法正确地生成代码, 和 mybatis 一致有 4 种 sql 语句，不管哪一种语句，它对参数都不无限制的， 它只对返回参数有限制， 具体如下

#### insert 方法
凡是以  insert, create，upsert 开头或加 @type insert 的方法, 都是对应 insert 语句, 格式如下

insertXXX(....) (lastInsertID int64, err error)


#### update 方法
凡是以  update 开头或加 @type update 的方法, 都是对应 update 语句, 格式如下

updateXXX(....) (rowsAffected int64, err error)


#### delete 方法
凡是以  delete, remove 开头或加 @type delete 的方法, 都是对应 delete 语句, 格式如下

deleteXXX(....) (rowsAffected int64, err error)


#### query 方法
凡是以  query, get, find, all, list 开头或加 @type select 的方法, 都是对应 select 语句, 格式如下

````go
queryXXX(....) (result XXXX, err error)
queryXXX(....) (result *XXXX, err error)
queryXXX(....) (results []XXXX, err error)
queryXXX(....) (results []*XXXX, err error)
queryXXX(....) (results map[int64]*XXXX, err error)
queryXXX(....) (results map[int64]XXXX, err error)
````