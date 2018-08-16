# delete 方法

接口中凡是以  delete, remove 开头或加 @type delete 的方法, 都是对应 delete 语句, 格式如下

deleteXXX(....) (rowsAffected int64, err error)