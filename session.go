package gobatis

// gobatis (Object Sql Mapping)是用go编写的ORM工具，目前很简单，只能算是半成品，只支持mysql(因为我目前的项目是mysql,所以其他数据库没有测试过)。
//
// 以前是使用MyBatis开发java服务端，它的sql mapping很灵活，把sql独立出来，程序通过输入与输出来完成所有的数据库操作。
//
// osm就是对MyBatis的简单模仿。当然动态sql的生成是使用go和template包，所以sql mapping的格式与MyBatis的不同。sql xml 格式如下：
//  <?xml version="1.0" encoding="utf-8"?>
//  <gobatis>
//   <select id="selectUsers" result="structs">
//     SELECT id,email
//     FROM user
//     {{if ne .Email ""}} where email=#{Email} {{end}}
//     order by id
//   </select>
//  </gobatis>
//

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
)

// SessionFactory 对象，通过Struct、Map、Array、value等对象以及Sql Map来操作数据库。可以开启事务。
type SessionFactory struct {
	Session
}

func (sess *SessionFactory) WithDB(db DBRunner) *SessionFactory {
	newSess := &SessionFactory{}
	newSess.base = *sess.base.WithDB(db)
	return newSess
}

func (sess *SessionFactory) SetDB(db DBRunner) {
	sess.base.SetDB(db)
}

func (o *SessionFactory) DB() DBRunner {
	return o.base.DB()
}

// Begin 打开事务
//
//如：
//  tx, err := o.Begin()
func (o *SessionFactory) Begin(nativeTx ...DBRunner) (tx *Tx, err error) {
	tx = new(Tx)
	tx.Session = o.Session

	var native DBRunner
	if len(nativeTx) > 0 {
		native = nativeTx[0]
	}

	if native == nil {
		if o.base.db == nil {
			return nil, errors.New("db no opened")
		}

		sqlDb, ok := o.base.db.(*sql.DB)
		if !ok {
			return nil, errors.New("db no *sql.DB")
		}

		native, err = sqlDb.Begin()
	}

	tx.base.db = native
	return tx, err
}

// WithTx 打开事务
func (o *SessionFactory) WithTx(nativeTx DBRunner) *Tx {
	tx := new(Tx)
	tx.Session = o.Session
	tx.base.db = nativeTx
	return tx
}

// Close 与数据库断开连接，释放连接资源
//
//如：
//  err := o.Close()
func (o *SessionFactory) Close() (err error) {
	if o.base.db == nil {
		err = fmt.Errorf("db no opened")
	} else {
		sqlDb, ok := o.base.db.(*sql.DB)
		if ok {
			err = sqlDb.Close()
			o.base.db = nil
		} else {
			err = fmt.Errorf("db no opened")
		}
	}
	return
}

// Tx 与Osm对象一样，不过是在事务中进行操作
type Tx struct {
	Session
}

// Commit 提交事务
//
//如：
//  err := tx.Commit()
func (o *Tx) Commit() error {
	if o.base.db == nil {
		return fmt.Errorf("tx no runing")
	}
	sqlTx, ok := o.base.db.(*sql.Tx)
	if ok {
		return sqlTx.Commit()
	}
	return fmt.Errorf("tx no runing")
}

// Rollback 事务回滚
//
//如：
//  err := tx.Rollback()
func (o *Tx) Rollback() error {
	if o.base.db == nil {
		return fmt.Errorf("tx no runing")
	}
	sqlTx, ok := o.base.db.(*sql.Tx)
	if ok {
		return sqlTx.Rollback()
	}
	return fmt.Errorf("tx no runing")
}

type Session struct {
	base Connection
}

func (sess *Session) DB() DBRunner {
	return sess.base.db
}

func (sess *Session) Dialect() Dialect {
	return sess.base.Dialect()
}

func (sess *Session) Reference() Reference {
	return Reference{&sess.base}
}

// Delete 执行删除sql
//
//xml
//    <delete id="deleteUser">DELETE FROM user where id = #{Id};</delete>
//代码
//  user := User{Id: 3}
//  count,err := o.Delete("deleteUser", user)
//删除id为3的用户数据
func (sess *Session) Delete(id string, params ...interface{}) (int64, error) {
	return sess.base.Delete(context.Background(), id, nil, params)
}

// Update 执行更新sql
//
//xml
//    <update id="updateUserEmail">UPDATE user SET email=#{Email} where id = #{Id};</update>
//代码
//  user := User{Id: 3, Email: "test@foxmail.com"}
//  count,err := o.Update("updateUserEmail", user)
//将id为3的用户email更新为"test@foxmail.com"
func (sess *Session) Update(id string, params ...interface{}) (int64, error) {
	return sess.base.Update(context.Background(), id, nil, params)
}

// Insert 执行添加sql
//
//xml
//  <insert id="insertUser">INSERT INTO user(email) VALUES(#{Email});</insert>
//代码
//  user := User{Email: "test@foxmail.com"}
//  insertId,count,err := o.Insert("insertUser", user)
//添加一个用户数据，email为"test@foxmail.com"
func (sess *Session) Insert(id string, params ...interface{}) (int64, error) {
	return sess.base.Insert(context.Background(), id, nil, params)
}

// SelectOne 执行查询sql, 返回单行数据
//
//xml
//  <select id="searchArchives">
//   <![CDATA[
//   SELECT id,email,create_time FROM user WHERE id=#{Id};
//   ]]>
//  </select>
func (sess *Session) SelectOne(id string, params ...interface{}) Result {
	return sess.base.SelectOne(context.Background(), id, nil, params)
}

// Select 执行查询sql, 返回多行数据
//
//xml
//  <select id="searchUsers">
//   <![CDATA[
//   SELECT id,email,create_time FROM user WHERE create_time >= #{create_time};
//   ]]>
//  </select>
func (sess *Session) Select(id string, params ...interface{}) *Results {
	return sess.base.Select(context.Background(), id, nil, params)
}

// New 创建一个新的Osm，这个过程会打开数据库连接。
//
// cfg 是数据连接的参数，可以是0个1个或2个数字，第一个表示MaxIdleConns，第二个表示MaxOpenConns.
//
// 如：
//  o, err := gobatis.New(&gobatis.Config{DriverName: "mysql",
//         DataSource: "root:root@/51jczj?charset=utf8",
//         XMLPaths: []string{"test.xml"}})
func New(cfg *Config) (*SessionFactory, error) {
	conn, err := newConnection(cfg)
	if err != nil {
		return nil, err
	}

	return &SessionFactory{Session: Session{base: *conn}}, nil
}
