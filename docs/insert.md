# 插入数据


## 格式
接口中凡是以  insert, create，upsert 开头或加 @type insert 的方法, 都是对应 insert 语句, 格式如下

````go
insertXXX(....) (lastInsertID int64, err error)
````
或

````go
insertXXX(....) (err error)

````


## 输入参数
   方法可以有 0 到多个参数，每个参数都作为 sql 语句中引用的参数

## 返回值，必须为一个或两个

    返回值  lastInsertID 为插入数据后，生成的 ID, 它是可选的，可以没有
    返回值  err 为插入数据后，如果出错时返回的错误


## 例子

如创建一个对象， 如下

````go
  // @mssql insert into auth_users(username, 
  //          phone, 
  //          address, 
  //          status,
  //          birth_day,
  //          created_at,
  //          updated_at)
  //          output inserted.id
  //        values (#{username},#{phone},
  //          #{address},
  //          #{status},
  //          #{birth_day},
  //          CURRENT_TIMESTAMP,
  //          CURRENT_TIMESTAMP)
  //
  // @postgres insert into auth_users(username,
  //          phone, 
  //          address,
  //          status,
  //          birth_day,
  //          created_at,
  //          updated_at)
  //        values (#{username},
  //          #{phone},
  //          #{address},
  //          #{status},
  //          #{birth_day},
  //          CURRENT_TIMESTAMP,
  //          CURRENT_TIMESTAMP) returning id
  //
  // @default insert into auth_users(username,
  //          phone,
  //          address,
  //          status,
  //          birth_day,
  //          created_at,
  //          updated_at)
  //        values (#{username},
  //          #{phone},
  //          #{address},
  //          #{status},
  //          #{birth_day},
  //          CURRENT_TIMESTAMP,
  //          CURRENT_TIMESTAMP)
  Insert(u *User) (int64, error)
````

它如可以写成如下形式

````go
  // @mssql insert into auth_users(username, 
  //          phone, 
  //          address, 
  //          status,
  //          birth_day,
  //          created_at,
  //          updated_at)
  //          output inserted.id
  //        values (#{username},#{phone},
  //          #{address},
  //          #{status},
  //          #{birth_day},
  //          CURRENT_TIMESTAMP,
  //          CURRENT_TIMESTAMP)
  //
  // @postgres insert into auth_users(username,
  //          phone, 
  //          address,
  //          status,
  //          birth_day,
  //          created_at,
  //          updated_at)
  //        values (#{username},
  //          #{phone},
  //          #{address},
  //          #{status},
  //          #{birth_day},
  //          CURRENT_TIMESTAMP,
  //          CURRENT_TIMESTAMP) returning id
  //
  // @default insert into auth_users(username,
  //          phone,
  //          address,
  //          status,
  //          birth_day,
  //          created_at,
  //          updated_at)
  //        values (#{username},
  //          #{phone},
  //          #{address},
  //          #{status},
  //          #{birth_day},
  //          CURRENT_TIMESTAMP,
  //          CURRENT_TIMESTAMP)
  Insert(username, phone, address string, status int, birth_day time.Time) (int64, error)
````

