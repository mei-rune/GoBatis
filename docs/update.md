# update 方法

## 格式

接口中凡是以  update 开头或加 @type update 的方法, 都是对应 update 语句, 格式如下

````go
updateXXX(....) (rowsAffected int64, err error)
````
或

````go
updateXXX(....) (err error)

````


## 输入参数
   方法可以有 0 到多个参数，每个参数都作为 sql 语句中引用的参数

## 返回值，必须为一个或两个

    返回值  rowsAffected 为更新数据后，影响的行数
    返回值  err 执行中如果出错时返回的错误

## 例子

如更新一个用户, 如下

````go
  // @default UPDATE auth_users
  //     SET username=#{u.username},
  //          phone=#{u.phone},
  //          address=#{u.address},
  //          status=#{u.status},
  //          birth_day=#{u.birth_day},
  //          updated_at=CURRENT_TIMESTAMP
  //      WHERE id=#{id}
  Update(id int64, u *User) (int64, error)
````

````go
  // @default UPDATE auth_users
  //      SET username=#{username},
  //          updated_at=CURRENT_TIMESTAMP
  //      WHERE id=#{id}
  UpdateName(id int64, username string) (int64, error)
````

