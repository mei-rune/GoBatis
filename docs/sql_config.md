
# SQL 的配置

## 1. xml 方式
我们会将接口名称和方法名作为 sql 的标识在 xml 配置中查找对应的 SQL 语句，标识格式如下

    接口名称.方法名

如例子中的 `UserDao.Insert`

## 2. 注释方式

golang 不支持 java 中的 annotation, 所以我们只好将 SQL 放在注释中，我们一般推荐这种方式，它的格式如下：

````
// xxxxx
// @type select
// @option key1 value1
// @option key2 value2
// @mysql 111
// 111
// @posgres  22
// 222
// @default 333
````
解析注释时，将它映射成下面的结构，分割规则为，如果一行以 @ 开头，则以这个位置分割
每一部分以@开始到第一个空格作为 key, 注释的第一部分如果不是以 @ 开头则作为 description
1. description 作为描述用，没有什么实际用途
2. type 为语句的类型，对应 xml 的 select, insert, udpate 和 delete，如果它没有时会按方法名来猜，规则请见 [generator/statement.go](https://github.com/runner-mei/GoBatis/blob/master/generator/statement.go)
3. option 暂时没有什么用，只是作为以后的扩展使用
4. 其它的均作为不同数据库的 sql 方言，key 数据类型
5. 最后将 default 作为缺省数据库 sql

````go
type SQLConfig struct {
  Description   string
  StatementType string
  DefaultSQL    string
  Options       map[string]string
  Dialects      map[string]string
}
````

如例子，我们会生成如下结构


````go
 &SQLConfig{
  Description: "xxxxx",
  StatementType: "select",
  DefaultSQL: "333",
  Options:       map[string]string{
        "key1": "value1",
        "key2": "value2",
    },
  Dialects:      map[string]string{
        "mysql": `111
        111`,
        "posgres":  `22
        222`,
    }
}
````

