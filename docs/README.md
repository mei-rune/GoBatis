# GoBatis

## 简介

GoBatis 是用 golang 编写的 ORM 工具，目前已在生产环境中使用，理论上支持任何数据库 (只测试过 postgresql, mysql, mssql)。


## 基本思路
1. 用户定义结构和接口
2. 在接口的方法上定义 sql （可以是 xml 或 方法的注释中）
3. 用工具生成接口的实现
4. 创建接口的实例并使用它
5. 仅可能地桵据方法的参数和返回值来生成 sql，请见 [SQL 自动生成](sql_genrate.md)


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

### 1. 动态 sql 语句的格式

     我实现一个和  mybatis 类似的 if, choose, foreach, trim, set 和 where 之类的 xml 基本实现，同时也支持 go template 来生成 sql。

#### 1.1 另外我不支持 ${xxx}, 但是我提供了一个更安全的 <print fmt="%s" value="b" inStr="true" /> 来替换它

     当  inStr="true" 时我会检查 value 的值中是不是有 引号之类的字符，防止 sql 注入
     
     当    inStr="false" 时我会检查 value 的值中是不是有 and 或  or 之类的逻辑表达式，防止 sql 注入

#### 1.2 我为 if 标签 增加了 else 支持， 用法为 <if> xx <else /> xxx </if>

### 2. 自动生成 sql 语句

MyBatis 是不会自动生成 sql 语句的， 我觉得能像大部份的 orm 一样能生成 sql 的话，可以省很多工作
     请见 [SQL 自动生成](https://mei-rune.github.io/GoBatis/#/sql_genrate)



## 待完成的任务

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


## 注意
GoBatis 是基于 [osm](https://github.com/yinshuwei/osm) 的基础上修改来的，goparser 则是在 [light](https://github.com/arstd/light) 的基础上修改来的, reflectx 则从 [sqlx](https://github.com/jmoiron/sqlx) 拷贝过来的