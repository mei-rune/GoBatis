# SQL 自动生成

和 MyBatis 不同，我觉得能像大部份的 orm 一样能生成 sql 的话，可以省很多工作.


生成规则如下
1. 生成 sql 不是很智能，只能生成一些针对一张表的简单操作，复杂的查询生成不了，请自已写 sql.
2. 接口中的每个方法都是对应 sql 的 insert, update, delete 和 select 四种中的一个，请见[SQL 配置](sql_config.md)，所以我们也只会生成这四种语句
3. 表名是根据配置中的 record_type 来获取的，或返回值的类型来获取的。


## Insert 语句的生成

生成 sql 时， 如果 字段的 tag 中有 autoincr，<- 或 deleted 时将跳过这个字段不处理


### 形式1

````go 
 Insert(x *XXX) (error)
 // or
 Insert(x *XXX) (int64, error)
```

它将和其它 orm 一样， 将能自动生成为如下 sql

````sql
 INSERT INTO xxx(field1, field2, ...) VALUES(#{x.Field1}, #{x.Field2}, ...)
````

### 形式2

````go 
 Insert(field1 type1, field2 type2, ...) (error)
 // or
 Insert(field1 type1, field2 type2, ...) (int64, error)
```

它将和其它 orm 一样， 将能自动生成为如下 sql

````sql
 INSERT INTO xxx(field1, field2, ...) VALUES(#{field1}, #{field2}, ...)
````

## Upddate 语句的生成

生成 sql 时， 如果 字段的 tag 中有 autoincr，<-, pk, created 或 deleted 时将跳过这个字段不处理

### 形式1

````go 
 Update(xField xtype, x *XXX) (error)
 // or
 Update(xField xtype, x *XXX) (int64, error)
```

它将和其它 orm 一样， 将能自动生成为如下 sql

````sql
 UPDATE xxx SET field1 = #{x.Field1}, field2 = #{x.Field2}, ... WHERE xField = #{xField} 
````

注意如果有 updated_at 字段（或 tag 中有 updated 的字段），则会自动加上 updated_at = now() 或 CURRENT_TIMESTAMP

### 形式2

````go 
 Update(xField xtype, field1 type1, field2 type2, ...) (error)
 // or
 Update(xField xtype, field1 type1, field2 type2, ...) (int64, error)
```

它将和其它 orm 一样， 将能自动生成为如下 sql

````sql
 UPDATE xxx SET field1 = #{field1}, field2 = #{field2}, ... WHERE xField = #{xField} 
````


注意如果有 updated_at 字段（或 tag 中有 updated 的字段），那么不管参数中有没有 updated_at 参数都会自动加上 updated_at = now() 或 CURRENT_TIMESTAMP

## Delete 语句的生成

````go 
 Delete(field1 type1, field2 type2, ...) (error)
 // or
 Delete(field1 type1, field2 type2, ...) (int64, error)
```

它将和其它 orm 一样， 将能自动生成为如下 sql

````sql
 DELETE FROM xxx WHERE field1 = #{field1} AND field2 = #{field2} AND ...
````

注意， WHERE 子句更详细的说明请见 Select 语句的生成


### 软删除的支持

如果结构中有一个在 tag 中有 deleted 的字段, 那么它支持软删除，也就是可以不真实删除， 只是加上一个删除标记， 如


````sql
 UPDATE xxx SET deleted = now() WHERE field1 = #{field1} AND field2 = #{field2} AND ...
````

想真实删除可如下定义方法

````go 
 Delete(force bool, field1 type1, field2 type2, ...) (error)
 // or
 Delete(force bool, field1 type1, field2 type2, ...) (int64, error)
```

生成 SQL 如下

````sql
 <if test="force">DELETE FROM xxx WHERE field1 = #{field1} AND field2 = #{field2} AND ...</if>
 <if test="!force">UPDATE xxx SET deleted = now() WHERE field1 = #{field1} AND field2 = #{field2} AND ...</if>
````

## Count 语句的生成

````go 
 Count(xField xtype, field1 type1, field2 type2, ...) (int64, error)
```

它将和其它 orm 一样， 将能自动生成为如下 sql

````sql
 SELECT count(*) FROM xxx WHERE field1 = #{field1} AND field2 = #{field2} AND ...
````

注意， WHERE 子句更详细的说明请见 Select 语句的生成


## SELECT 语句的生成

````go 
 Query(field1 type1, field2 type2, ...) (xxx, error)
```

它将和其它 orm 一样， 将能自动生成为如下 sql

````sql
 SELECT * FROM xxx WHERE field1 = #{field1} AND field2 = #{field2} AND ...
````

### where 子句

一般来说系统会将方法参数作 ‘=’ 处理， 如 field1 = #{field1}, 但有一些例外处理以增强系统

#### sql.NullXXXX 参数

当参数的类型为 sql.NullXXX 或类似的类型(如 pq.NullTime, gopkg.in/guregu/null.Int64)时, 会生成如下表达式

````
<if test="x.Vaild"> xfield = #{x}</if>
````


#### 参数名为 xFieldLike 形式

参数名为 xFieldLike 形式时（删除后面的 Like 后的 "xField" 字符为字段名）, 会生成如下表达式

````
   xfield Like #{xFieldLike}
````

#### 参数所对应的字段类型为 slice 或 array 时（[]byte 除外）

````
   #{xField} = ANY (xfield)
````


#### 软删除的处理

结构支持软删除时， 会自动加上 “deleted IS NULL”, 当然如果有下面的 isDeleted 参数则不会加上
参数中可加上一个名为 isDeleted 且类型为 bool 或 sql.NullBool 的参数，那么它会生成下面的表达式

````
 <if test="isDeleted"> deleted IS NOT NULL </if><if test="!isDeleted"> deleted IS NULL </if>
````

如果 isDeleted 的类型为  sql.NullBool 时
`````
 <if test="isDeleted.Vaild"> <if test="isDeleted.Bool"> deleted IS NOT NULL </if><if test="!isDeleted.Bool"> deleted IS NULL </if></if>
````

### OFFSET 和 LIMIT 子句

如果参数中有  offset 和  limit 参数， 那么会生成

````
<if test="offset &gt 0"> OFFSET #{offset} </if>
<if test="limit &gt 0"> LIMIT #{limit} </if> 
````