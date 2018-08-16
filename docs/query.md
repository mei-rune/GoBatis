# query 方法

接口中凡是以  query, get, find, all, list 开头或加 @type select 的方法, 都是对应 select 语句, 格式如下

````go
queryXXX(....) (result XXXX, err error)
queryXXX(....) (result *XXXX, err error)
queryXXX(....) (results []XXXX, err error)
queryXXX(....) (results []*XXXX, err error)
queryXXX(....) (results map[int64]*XXXX, err error)
queryXXX(....) (results map[int64]XXXX, err error)
````