# 插入数据


## 格式
接口中凡是以  insert, create，upsert 开头或加 @type insert 的方法, 都是对应 insert 语句, 格式如下

````go
insertXXX(....) (lastInsertID int64, err error)
insertXXX(ctx context.Context, ....) (lastInsertID int64, err error)
````
或

````go
insertXXX(....) (err error)
insertXXX(ctx context.Context, ....) (err error)

````


## 输入参数
   方法可以有 0 到多个参数，每个参数（除 context.Context 外）都作为 sql 语句中引用的参数
   其中 context.Context 参数会传给  sql.DB 的  ExecContext 方法

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


## selectKey

### 格式

````go
// @selectKey <dialect1,dialect2,...> <sql>
````

`@selectKey` 用于在 INSERT 或 UPSERT 语句执行之后，再执行一条额外的 SQL 查询来获取生成的 ID。它**仅适用于 insert 类型的操作**（即方法名以 insert、create、upsert 开头，或标注了 `@type insert` 的方法）。

- `<dialect>` 指定该 selectKey SQL 适用的数据库方言，多个方言用逗号分隔。**不能**使用 `default` 作为方言名称。
- `<sql>` 为 selectKey 要执行的 SQL 语句，可以引用方法参数中的字段（如 `#{name}`、`#{user.id}`）。

### 工作原理

生成代码会根据当前数据库的 `DatabaseID()` 匹配 `@selectKey` 中指定的方言，如果匹配成功，则将 selectKey 的 SQL 追加到主 INSERT 语句之后（用 `;\r\n` 分隔），并将结果类型标记为 `ResultSelectKey`。

在执行阶段，GoBatis 检测到 `ResultSelectKey` 后，会使用 `QueryRowContext` 代替 `ExecContext` 执行拼接后的 SQL，并从查询结果中扫描出 int64 类型的 ID 返回。

### 使用场景

1. **Oracle 数据库获取序列值**——Oracle 不支持 `RETURNING` 子句（或需要特定版本），通常使用序列来生成主键，可通过 selectKey 查询序列的当前值：

    ````go
    // @selectKey oracle select "users_seq".currval from dual
    Insert(u *User) (int64, error)
    ````

2. **自定义主键生成规则**——当主键不是自增列，而是由业务规则或触发器生成时，可在 INSERT 后通过唯一条件查询回该 ID：

    ````go
    // @selectKey oracle select id from gobatis_settings where name = #{name}
    UpsertSetting1(s *Setting) (int64, error)
    ````

3. **UPSERT 场景**——在使用 `INSERT … ON DUPLICATE KEY UPDATE` 或 `MERGE` 实现的 upsert 操作中，如果插入时发生了冲突并执行了更新，可以通过 selectKey 查询现有记录的 ID：

    ````go
    // @selectKey oracle select id from gobatis_settings where name = #{name}
    UpsertSetting1(s *Setting) (int64, error)
    ````

### 完整示例

````go
// @selectKey oracle  select id from <tablename /> where name = #{name}
// @default  insert into auth_users(username, phone, address) values (#{username}, #{phone}, #{address})
Insert(u *User) (int64, error)
````

当数据库为 Oracle 时，生成的 SQL 为：

```sql
insert into auth_users(username, phone, address) values (?, ?, ?);
select id from auth_users where name = ?
```

GoBatis 执行此语句后，从第二句查询的结果中扫描出 ID 并返回。

