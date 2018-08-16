# 结构和接口定义规范


## 结构对象的定义

GoBatis 支持将一个struct映射为数据库中对应的一张表。映射规则如下：

我们在field对应的Tag中对Column的一些属性进行定义，定义的方法基本和我们写SQL定义表结构类似，比如：

````go
type User struct {
  TableName gobatis.TableName `db:"auth_users"`
  ID        int64             `db:"id,autoincr"`
  Username  string            `db:"username"`
  Phone     string            `db:"phone"`
  Address   *string           `db:"address"`
  Status    Status            `db:"status"`
  BirthDay  *time.Time        `db:"birth_day"`
  CreatedAt time.Time         `db:"created_at"`
  UpdatedAt time.Time         `db:"updated_at"`
}
````

表名可以用一个名为 TableName  的字段的 tag 值来提供， TableName  的字段的类型可以是一个 struct{} 类型， 也支持 TableName() 方法获取表名， 如下

````go
func (u User) TableName() string {
  return "auth_users"
}
````
或

````go
func (u *User) TableName() string {
  return "auth_users"
}
````


各个字段的具体的Tag规则如下，另Tag中的关键字均不区分大小写，但字段名根据不同的数据库是区分大小写：


字段 | 说明
--- | ----
| name | 当前field对应的字段的名称，可选，如不写，则自动根据field名字和转换规则命名，如与其它关键字冲突，请使用单引号括起来。 |
| pk | 是否是Primary Key，|
| autoincr  | 是否是自增 |
| [not ]null 或 notnull  | 是否可以为空 |
| - | 这个Field将不进行字段映射 |
| json | 表示内容将先转成Json格式，然后存储到数据库中，数据库中的字段类型可以为Text或者二进制 |



## 接口的定义

定义接口时， 对接口中的方法是有一些要求的，不然代码生成工具也无法正确地生成代码, 和 mybatis 一致有 4 种 sql 语句，不管哪一种语句，它对参数都不无限制的， 它只对返回参数有限制， 具体如下

#### insert 方法
凡是以  insert, create，upsert 开头或加 @type insert 的方法, 都是对应 insert 语句, 格式如下

insertXXX(....) (lastInsertID int64, err error)


#### update 方法
凡是以  update 开头或加 @type update 的方法, 都是对应 update 语句, 格式如下

updateXXX(....) (rowsAffected int64, err error)


#### delete 方法
凡是以  delete, remove，clear 开头或加 @type delete 的方法, 都是对应 delete 语句, 格式如下

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

#### 方法引用
有时一个接口的方法可以引用另一个接口的方法，是很有用的， 可以如下
````go
  // @reference RoleDao.ListByUserID
  Roles(id int64) ([]Role, error)
````

#### 例子

````go
type Users interface {
  Insert(u *User) (int64, error)

  Update(id int64, u *User) (int64, error)

  DeleteAll() (int64, error)

  Delete(id int64) (int64, error)

  Get(id int64) (*User, error)

  Count() (int64, error)
}

````