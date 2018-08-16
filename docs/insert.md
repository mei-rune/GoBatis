# 插入数据

接口中凡是以  insert, create，upsert 开头或加 @type select 的方法, 都是对应 insert 语句, 格式如下

insertXXX(....) (lastInsertID int64, err error)