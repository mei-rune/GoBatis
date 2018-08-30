# delete 方法

## 格式

接口中凡是以 delete, remove, clear 开头或加 @type delete 的方法, 都是对应 delete 语句, 格式如下


````go
deleteXXX(....) (rowsAffected int64, err error)
deleteXXX(ctx context.Context, ....) (rowsAffected int64, err error)
````
或

````go
deleteXXX(....) (err error)
deleteXXX(ctx context.Context, ....) (err error)
````
  

## 输入参数
   方法可以有 0 到多个参数，每个参数（除 context.Context 外）都作为 sql 语句中引用的参数
   其中 context.Context 参数会传给  sql.DB 的  ExecContext 方法

## 返回值，必须为一个或两个

    返回值  rowsAffected 为删除数据后，删除的行数
    返回值  err 为执行中如果出错时返回的错误

## 例子

如更新一个用户, 如下

````go
  // @default DELETE FROM auth_users
  DeleteAll() (int64, error)
````

````go
  // @default DELETE FROM auth_users WHERE id=#{id}
  DeleteByID(id int64) (int64, error)
````

