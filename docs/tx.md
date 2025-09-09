# 事务的支持


事务的支持比较简单，有三种方式


## 1. 定义一个 WithDB 

  你可以在接口中定义一个 WithDB(gobatis.DBRunner) XXXDao 方法，让它返回一个新接口实例

````go

  type UserDao interface  {
    WithDB(gobatis.DBRunner) UserDao
  }
  // ...
  userDao := NewUserDao(sessionRef)
  // ...
  db, _ := sql.Open("mysql", "test:@tcp(localhost:3306)/golang?autocommit=true&parseTime=true&multiStatements=true")

  // ...
  tx := db.Begin()
  defer tx.Rollback()
  txDao := userDao.WithDB(tx)
  txDao.Update(1, u1)
  txDao.Update(2, u2)
  tx.Commit()
  
````
完整例子请看 https://github.com/runner-mei/GoBatis/blob/master/example/transaction_test.go 中的 ExampleUserDao_WithDB

## 2. 通过 context 参数传递

你可以在 接口的方法上增加一个 context.Context 参数，然后用 gobatis.WithTx 传递事务

````go

  type UserDao interface  {
    Update(ctx context.Context, id int64, user User) error
  }
  // ...
  userDao := NewUserDao(sessionRef)
  // ...
  db, _ := sql.Open("mysql", "test:@tcp(localhost:3306)/golang?autocommit=true&parseTime=true&multiStatements=true")


  // ...
  tx := db.Begin()
  defer tx.Rollback()
  ctx := WithTx(context.Background(), tx)
  userDao.Update(ctx, 1, u1)
  userDao.Update(ctx, 2, u2)
  tx.Commit()


````

完整例子请看 https://github.com/runner-mei/GoBatis/blob/master/example/transaction_test.go 中的 ExampleUserDao_InsertWithContext

## 3. 另一种方法

````go

  // 下面的 Connection， Tx 和 NewConnection 请看
  // https://github.com/runner-mei/GoBatis/blob/master/example/connection.go

  conn := NewConnection(factory)

  tx := conn.Begin()
  defer tx.Rollback()
  tx.Users().Update(ctx, 1, u1)
  tx.Users().Update(ctx, 2, u2)
  tx.Commit()
  
````
完整例子请看 https://github.com/runner-mei/GoBatis/blob/master/example/transaction_test.go 中的 ExampleTx_Commit
