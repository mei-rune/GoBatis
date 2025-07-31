# upsert 方法

## 格式

接口中凡是以  upsert 开头或加 @type upsert 的方法, 都是对应 upsert 语句, 格式有两种，如下

格式一

````go
upsertXXX(key1, key2, key3, ....) (id int64, err error)
upsertXXX(key1, key2, key3, ....) error
upsertXXX(ctx context.Context, key1, key2, key3, ....) (id int64, err error)
upsertXXX(ctx context.Context, key1, key2, key3, ....) error
````

## 输入参数

方法可以有 两个 到多个参数，方法参数都是 key, 它们联合起来组成一个 unique key




## 输入参数
   方法可以有 0 到多个参数，每个参数（除 context.Context 外）都作为 sql 语句中引用的参数
   其中 context.Context 参数会传给  sql.DB 的  ExecContext 方法

## 返回值，必须为一个或两个

    返回值  id 为新建时的ID，注意如果记录已存在， 可能返回 0（根据数据库的特性）
    返回值  err 执行中如果出错时返回的错误


格式二

````go
upsertXXX(key1, key2, key3, ..., record RecordType) (id int64, err error)
upsertXXX(key1, key2, key3, ...., record RecordType) error
upsertXXX(ctx context.Context, key1, key2, key3, ..., record RecordType) (id int64, err error)
upsertXXX(ctx context.Context, key1, key2, key3, ..., record RecordType) error
````

## 输入参数

方法可以有 两个 到多个参数，除最后一个参数外都是 key, 它们联合起来组成一个 unique key
最后一个参数是记录，它必须是一个 struct



## 输入参数
   方法可以有 0 到多个参数，每个参数（除 context.Context 外）都作为 sql 语句中引用的参数
   其中 context.Context 参数会传给  sql.DB 的  ExecContext 方法

## 返回值，必须为一个或两个

    返回值  id 为新建时的ID，注意如果记录已存在， 可能返回 0（根据数据库的特性）
    返回值  err 执行中如果出错时返回的错误

