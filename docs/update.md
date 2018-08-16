# update 方法

接口中凡是以  update 开头或加 @type update 的方法, 都是对应 update 语句, 格式如下

updateXXX(....) (rowsAffected int64, err error)