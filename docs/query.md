# query 方法


## 格式

接口中凡是以  query, get, find, all, list 开头或加 @type select 的方法, 都是对应 select 语句, 格式如下

````go
queryXXX(....) (result XXXX, err error)
queryXXX(....) (result *XXXX, err error)
queryXXX(....) (results []XXXX, err error)
queryXXX(....) (results []*XXXX, err error)
queryXXX(....) (results map[int64]*XXXX, err error)
queryXXX(....) (results map[int64]XXXX, err error)
````



### 输入参数
   方法可以有 0 到多个参数，每个参数都作为 sql 语句中引用的参数

### 返回值，必须为两个或两个以上

    不管返回值有多少个，最后一个必须为  error 类型，它是执行中如果出错时返回的错误




## 形式1

它的返回值必须为两个， 最后一个必须是 error 类型：

第一个返回值可以是下面几种情况

 1. slice  类型

        这种情况会认为 sql 语句会返回多条记录

 2. map  类型

        这种情况有一个特例，除特例一般会认为 sql 语句会返回多条记录, 
        如果 map 的 value 是一个 struct 时，则将返回的列映射到结构，然后将结构中有 tag 为 pk 的 字段的值作为 map 有 key.
        如果 map 的 value 不是一个 struct 时， 则认为 sql 语句会返回两列， 第一列为 key ,  第一列为 value

        ** 特例： 类型为 map[string]interface{} 时，认为  sql 语句会返回一个记录, 然后将返回的列名为 key, 列值为 value **

 3. 非 slice 或 map 类型

        这种情况会认为 sql 语句会返回一条记录


### 例子


````go

  // @default select * FROM auth_users WHERE id=#{id}
  Get(id int64) (*User, error)


  // @default select username FROM auth_users WHERE id=#{id}
  GetName(id int64) (string, error)

  // @default select username FROM auth_users
  GetNames() ([]string, error)

  // @default select id, username FROM auth_users
  GetIDNames() (map[int64]string, error)

  // @default select * FROM auth_users WHERE id=#{id}
  GetMap(id int64) (map[string]interface{}, error)

  // @default select count(*) from auth_users
  Count() (int64, error)

  // @mssql select * from auth_users ORDER BY username OFFSET #{offset} ROWS FETCH NEXT #{size}  ROWS ONLY
  // @mysql select * from auth_users limit #{offset}, #{size}
  // @default select * from auth_users offset #{offset} limit  #{size}
  List1(offset, size int) (users []*User, err error)

  // @mssql select * from auth_users ORDER BY username OFFSET #{offset} ROWS FETCH NEXT #{size}  ROWS ONLY
  // @mysql select * from auth_users limit #{offset}, #{size}
  // @default select * from auth_users offset #{offset} limit  #{size}
  List2(offset, size int) (users map[int64]*User, err error)

  // @default select username from auth_users where id = #{id}
  GetNameByID(id int64) (string, error)
````



## 形式2


它的返回变量必须为两个以上， 最后一个必须是 error 类型

这时一般会要求返回变量是有名字的， 然后按返回变量的名字来映射返回字段， 它一个作如下处理
字

* 以返回的列名与方法的返回变量名进行匹配，如果找到则该列作为这个返回变量的值
* 否则将返回的列名以分隔符（默认为 _ ）将它们拆成 前缀 和 字段名 两个字符串
* 以 前缀 与方法的返回变量名进行匹配，如果找到则该列作为这个返回变量的值
* 否则看有没有指定缺省返回变量(以选项 default_return_name 指定)，有则将该列作为这个缺省返回变量的值
* 否则报错

#### 例如

````go
  // @default SELECT p.id as p_id,
  //                 p.user_id as p_user_id,
  //                 p.name as p_name,
  //                 p.value p_value,
  //                 p.created_at as p_created_at,
  //                 p.updated_at as p_updated_at,
  //                 u.id as u_id,
  //                 u.username as u_username
  //          FROM user_profiles as p LEFT JOIN auth_users as u On p.user_id = u.id
  //          WHERE p.id = #{id}
  FindByID1(id int64) (p *UserProfile, u *User, err error)
````

 sql 语句返回  p_id,p_user_id, p_name, p_value, p_created_at, p_updated_at, u_id, u_username 这些列， 
 程序会以 _ 字符将它们拆成 前缀和字段名 两个字符串 ，如下

  前缀 | 字段名
  ---- | ------
   p | id
   p | user_id
   p | name
   p | value
   p | created_at
   p | updated_at
   u | id
   u | username

  然后按方法上的返回变量的名字将它们映射到不同的结构中：

  前缀 p 和方法的 p 返回变量匹配，则 p_id,p_user_id, p_name, p_value, p_created_at, p_updated_at 将去除 p_ 前缀后按字段名映射到 UserProfile 结构中

  前缀 u 和方法的 u 返回变量匹配，则 u_id,u_username 将去除 u_ 前缀后按字段名映射到 User 结构中

#### 再如

````go
  // @option default_return_name p
  // @default SELECT p.id,
  //                 p.user_id,
  //                 p.name,
  //                 p.value,
  //                 p.created_at,
  //                 p.updated_at,
  //                 u.id as userid,
  //                 u.username as username
  //          FROM user_profiles as p LEFT JOIN auth_users as u On p.user_id = u.id
  //          WHERE p.id = #{id}
  FindByID3(id int64) (p UserProfile, userid int64, username string, err error)
````

 sql 语句返回的多个列，并指定了缺省返回变量(以选项 default_return_name 指定了 返回变量 p), 程序会作以下处理

    字段名    | 处理方式
  ---------- | ----------------------------------------------
  id         | 无法以 _ 拆分，名字也无法与返回变量名匹配，作为缺省返回变量 p 的字段值
  user_id    | 以 _ 拆分为  user 和 id，但前缀 user 无法与返回变量名匹配，作为缺省返回变量 p 的字段值
  name       | 无法以 _ 拆分，名字也无法与返回变量名匹配，作为缺省返回变量 p 的字段值
  value      | 无法以 _ 拆分，名字也无法与返回变量名匹配，作为缺省返回变量 p 的字段值
  created_at | 以 _ 拆分为  created 和 at，但前缀 created 无法与返回变量名匹配，作为缺省返回变量 p 的字段值
  updated_at | 以 _ 拆分为  updated 和 at，但前缀 updated 无法与返回变量名匹配，作为缺省返回变量 p 的字段值
  userid     | 无法以 _ 拆分，但名字可以与返回变量名 userid 匹配， 则作为返回变量 userid 的值
  username   | 无法以 _ 拆分，但名字可以与返回变量名 username 匹配， 则作为返回变量 username 的值


### 例子

````go
  // @option field_delimiter .
  // @default SELECT p.id as "p.id",
  //                 p.user_id as "p.user_id",
  //                 p.name as "p.name",
  //                 p.value "p.value",
  //                 p.created_at as "p.created_at",
  //                 p.updated_at as "p.updated_at",
  //                 u.id as "u.id",
  //                 u.username as "u.username"
  //          FROM user_profiles as p LEFT JOIN auth_users as u On p.user_id = u.id
  //          WHERE p.id = #{id}
  FindByID1(id int64) (p *UserProfile, u *User, err error)

  // @default SELECT p.id as p_id,
  //                 p.user_id as p_user_id,
  //                 p.name as p_name,
  //                 p.value p_value,
  //                 p.created_at as p_created_at,
  //                 p.updated_at as p_updated_at,
  //                 u.id as u_id,
  //                 u.username as u_username
  //          FROM user_profiles as p LEFT JOIN auth_users as u On p.user_id = u.id
  //          WHERE p.id = #{id}
  FindByID2(id int64) (p UserProfile, u User, err error)

  // @option default_return_name p
  // @default SELECT p.id,
  //                 p.user_id,
  //                 p.name,
  //                 p.value,
  //                 p.created_at,
  //                 p.updated_at,
  //                 u.id as userid,
  //                 u.username as username
  //          FROM user_profiles as p LEFT JOIN auth_users as u On p.user_id = u.id
  //          WHERE p.id = #{id}
  FindByID3(id int64) (p UserProfile, userid int64, username string, err error)

  // @default SELECT p.id as p_id,
  //                 p.user_id as p_user_id,
  //                 p.name as p_name,
  //                 p.value p_value,
  //                 p.created_at as p_created_at,
  //                 p.updated_at as p_updated_at,
  //                 u.id as userid,
  //                 u.username as username
  //          FROM user_profiles as p LEFT JOIN auth_users as u On p.user_id = u.id
  //          WHERE p.id = #{id}
  FindByID4(id int64) (p *UserProfile, userid *int64, username *string, err error)

  // @default SELECT p.id as p_id,
  //                 p.user_id as p_user_id,
  //                 p.name as p_name,
  //                 p.value p_value,
  //                 p.created_at as p_created_at,
  //                 p.updated_at as p_updated_at,
  //                 u.id as u_id,
  //                 u.username as u_username
  //          FROM user_profiles as p LEFT JOIN auth_users as u On p.user_id = u.id
  //          WHERE p.user_id = #{userID}
  ListByUserID1(userID int64) (p []*UserProfile, u []*User, err error)

  // @option field_delimiter .
  // @default SELECT p.id as "p.id",
  //                 p.user_id as "p.user_id",
  //                 p.name as "p.name",
  //                 p.value "p.value",
  //                 p.created_at as "p.created_at",
  //                 p.updated_at as "p.updated_at",
  //                 u.id as "u.id",
  //                 u.username as "u.username"
  //          FROM user_profiles as p LEFT JOIN auth_users as u On p.user_id = u.id
  //          WHERE p.user_id = #{userID}
  ListByUserID2(userID int64) (p []UserProfile, u []User, err error)

  // @option default_return_name p
  // @default SELECT p.id,
  //                 p.user_id,
  //                 p.name,
  //                 p.value,
  //                 p.created_at,
  //                 p.updated_at,
  //                 u.id as userids,
  //                 u.username as usernames
  //          FROM user_profiles as p LEFT JOIN auth_users as u On p.user_id = u.id
  //          WHERE p.user_id = #{userID}
  ListByUserID3(userID int64) (p []UserProfile, userids []int64, usernames []string, err error)

  // @option field_delimiter .
  // @default SELECT p.id as "p.id",
  //                 p.user_id as "p.user_id",
  //                 p.name as "p.name",
  //                 p.value "p.value",
  //                 p.created_at as "p.created_at",
  //                 p.updated_at as "p.updated_at",
  //                 u.id as userids,
  //                 u.username as usernames
  //          FROM user_profiles as p LEFT JOIN auth_users as u On p.user_id = u.id
  //          WHERE p.user_id = #{userID}
  ListByUserID4(userID int64) (p []*UserProfile, userids []*int64, usernames []*string, err error)
````