package core

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
)

// Session 对象，通过Struct、Map、Array、value等对象以及Sql Map来操作数据库。可以开启事务。
type Session struct {
	base
}

func (sess *Session) WithDB(db DBRunner) DbSession {
	newSess := &Session{}
	*newSess = *sess
	newSess.conn = *sess.conn.CloneWith(db)
	return newSess
}

func (sess *Session) SetDB(db DBRunner) {
	sess.conn.SetDB(db)
}

func (sess *Session) InTx(ctx context.Context, optionalDB DBRunner, failIfInTx bool, cb func(ctx context.Context, tx *Tx) error) error {
	if optionalDB != nil {
		return InTx(ctx, optionalDB, failIfInTx, func(ctx context.Context, tx DBRunner) error {
			stx, err := sess.WithTx(tx)
			if err != nil {
				return err
			}
			return cb(ctx, stx)
		})
	}
	return InTx(ctx, sess.conn.db, failIfInTx, func(ctx context.Context, tx DBRunner) error {
		stx, err := sess.WithTx(tx)
		if err != nil {
			return err
		}

		return cb(ctx, stx)
	})
}

// Begin 打开事务
//
// 如：
//
//	tx, err := o.Begin()
func (sess *Session) Begin(nativeTx ...DBRunner) (tx *Tx, err error) {
	tx = new(Tx)
	tx.base = sess.base

	var native DBRunner
	if len(nativeTx) > 0 {
		_, ok := nativeTx[0].(*sql.Tx)
		if !ok {
			return nil, errTx{method: "open", inner: errors.New("argument tx invalid")}
		}

		native = nativeTx[0]
	}

	if native == nil {
		if sess.conn.db == nil {
			return nil, errTx{method: "open", inner: errors.New("db no opened")}
		}

		sqlDb, ok := sess.conn.db.(*sql.DB)
		if !ok {
			return nil, errTx{method: "open", inner: errors.New("db no *sql.DB")}
		}

		native, err = sqlDb.Begin()
		if err != nil {
			return nil, errTx{method: "begin", inner: err}
		}
	}

	tx.conn.db = native
	return tx, err
}

// Close 与数据库断开连接，释放连接资源
//
// 如：
//
//	err := o.Close()
func (o *Session) Close() error {
	return o.conn.Close()
}

// Tx 与 Session 对象一样，不过是在事务中进行操作
type Tx struct {
	base
}

func (o *Tx) withDB(db DBRunner) *Tx {
	newSess := &Tx{}
	*newSess = *o
	newSess.conn = *o.conn.CloneWith(db)
	return newSess
}

func (o *Tx) WithDB(db DBRunner) DbSession {
	return o.withDB(db)
}

func (o *Tx) SetDB(db DBRunner) {
	o.conn.SetDB(db)
}

func (o *Tx) InTx(ctx context.Context, optionalDB DBRunner, failIfInTx bool, cb func(ctx context.Context, tx *Tx) error) error {
	if optionalDB != nil {
		return InTx(ctx, optionalDB, failIfInTx, func(ctx context.Context, tx DBRunner) error {
			stx := o.withDB(tx)
			return cb(ctx, stx)
		})
	}
	if failIfInTx {
		return ErrAlreadyTx
	}
	return cb(ctx, o)
}

// Commit 提交事务
//
// 如：
//
//	err := tx.Commit()
func (o *Tx) Commit() error {
	if o.conn.db == nil {
		return errTx{method: "commit", inner: errors.New("tx no running")}
	}
	sqlTx, ok := o.conn.db.(*sql.Tx)
	if ok {
		err := sqlTx.Commit()
		if err != nil {
			return errTx{method: "commit", inner: err}
		}
		return nil
	}
	return errTx{method: "commit", inner: errors.New("tx no running")}
}

// Rollback 事务回滚
//
// 如：
//
//	err := tx.Rollback()
func (o *Tx) Rollback() error {
	if o.conn.db == nil {
		return errTx{method: "rollback", inner: errors.New("tx no running")}
	}
	sqlTx, ok := o.conn.db.(*sql.Tx)
	if ok {
		err := sqlTx.Rollback()
		if err != nil && err != sql.ErrTxDone {
			return errTx{method: "rollback", inner: err}
		}
		return nil
	}
	return errTx{method: "rollback", inner: errors.New("tx no running")}
}

type base struct {
	conn connection
}

func (sess *base) SqlStatements() [][2]string {
	return sess.conn.SqlStatements()
}

func (sess *base) ToXML() (map[string]*xmlConfig, error) {
	return sess.conn.ToXML()
}

func (sess *base) ToXMLFiles(dir string) error {
	return sess.conn.ToXMLFiles(dir)
}

func (sess *base) DB() DBRunner {
	return sess.conn.db
}

func (sess *base) Tracer() Tracer {
	return sess.conn.tracer
}

func (sess *base) WithTx(nativeTx DBRunner) (*Tx, error) {
	if nativeTx == nil {
		return nil, errTx{method: "withTx", inner: errors.New("argument tx missing")}
	}
	_, ok := nativeTx.(*sql.Tx)
	if !ok {
		return nil, errTx{method: "withTx", inner: errors.New("argument tx invalid")}
	}

	tx := new(Tx)
	tx.base = *sess
	tx.conn.db = nativeTx
	return tx, nil
}

func (sess *base) DriverName() string {
	return sess.conn.DriverName()
}

func (sess *base) Dialect() Dialect {
	return sess.conn.Dialect()
}

func (sess *base) Reference() Reference {
	return Reference{&sess.conn}
}

func (sess *base) SessionReference() SqlSession {
	return &sess.conn
}

func (sess *base) Mapper() *Mapper {
	return sess.conn.Mapper()
}

// QueryRow 执行SQL, 返回结果
func (sess *base) QueryRow(ctx context.Context, sqlstr string, params []interface{}) SingleRowResult {
	return sess.conn.QueryRow(ctx, sqlstr, params)
}

// Query 执行SQL, 返回结果集
func (sess *base) Query(ctx context.Context, sqlstr string, params []interface{}) *MultRowResult {
	return sess.conn.Query(ctx, sqlstr, params)
}

// Delete 执行删除sql
//
// xml
//
//	<delete id="deleteUser">DELETE FROM user where id = #{Id};</delete>
//
// 代码
//
//	user := User{Id: 3}
//	count,err := o.Delete("deleteUser", user)
//
// 删除id为3的用户数据
func (sess *base) Delete(ctx context.Context, id string, params ...interface{}) (int64, error) {
	return sess.conn.Delete(ctx, id, nil, params)
}

// Update 执行更新sql
//
// xml
//
//	<update id="updateUserEmail">UPDATE user SET email=#{Email} where id = #{Id};</update>
//
// 代码
//
//	user := User{Id: 3, Email: "test@foxmail.com"}
//	count,err := o.Update("updateUserEmail", user)
//
// 将id为3的用户email更新为"test@foxmail.com"
func (sess *base) Update(ctx context.Context, id string, params ...interface{}) (int64, error) {
	return sess.conn.Update(ctx, id, nil, params)
}

// Insert 执行添加sql
//
// xml
//
//	<insert id="insertUser">INSERT INTO user(email) VALUES(#{Email});</insert>
//
// 代码
//
//	user := User{Email: "test@foxmail.com"}
//	insertId,count,err := o.Insert("insertUser", user)
//
// 添加一个用户数据，email为"test@foxmail.com"
func (sess *base) Insert(ctx context.Context, id string, params ...interface{}) (int64, error) {
	return sess.conn.Insert(ctx, id, nil, params)
}

// SelectOne 执行查询sql, 返回单行数据
//
// xml
//
//	<select id="searchArchives">
//	 <![CDATA[
//	 SELECT id,email,create_time FROM user WHERE id=#{Id};
//	 ]]>
//	</select>
func (sess *base) SelectOne(ctx context.Context, id string, params ...interface{}) SingleRowResult {
	return sess.conn.SelectOne(ctx, id, nil, params)
}

// Select 执行查询sql, 返回多行数据
//
// xml
//
//	<select id="searchUsers">
//	 <![CDATA[
//	 SELECT id,email,create_time FROM user WHERE create_time >= #{create_time};
//	 ]]>
//	</select>
func (sess *base) Select(ctx context.Context, id string, params ...interface{}) *MultRowResult {
	return sess.conn.Select(ctx, id, nil, params)
}

// New 创建一个新的 Session，这个过程会打开数据库连接。
//
// cfg 是数据连接的参数
//
// 如：
//
//	o, err := core.New(&core.Config{
//	         DriverName: "mysql",
//	         DataSource: "root:root@/51jczj?charset=utf8",
//	         XMLPaths: []string{"test.xml"},
//	       })
func New(cfg *Config) (*Session, error) {
	conn, err := newConnection(cfg)
	if err != nil {
		return nil, err
	}

	return &Session{base: base{conn: *conn}}, nil
}

type DbSession interface {
	SqlStatements() [][2]string
	ToXML() (map[string]*xmlConfig, error)
	ToXMLFiles(dir string) error
	DB() DBRunner
	Tracer() Tracer
	WithTx(nativeTx DBRunner) (*Tx, error)
	WithDB(nativeTx DBRunner) DbSession
	InTx(ctx context.Context, optionalDB DBRunner, failIfInTx bool, cb func(ctx context.Context, tx *Tx) error) error
	DriverName() string
	Dialect() Dialect
	Reference() Reference
	SessionReference() SqlSession
	Mapper() *Mapper
	QueryRow(ctx context.Context, sqlstr string, params []interface{}) SingleRowResult
	Query(ctx context.Context, sqlstr string, params []interface{}) *MultRowResult
	Delete(ctx context.Context, id string, params ...interface{}) (int64, error)
	Update(ctx context.Context, id string, params ...interface{}) (int64, error)
	Insert(ctx context.Context, id string, params ...interface{}) (int64, error)
	SelectOne(ctx context.Context, id string, params ...interface{}) SingleRowResult
	Select(ctx context.Context, id string, params ...interface{}) *MultRowResult
}

var (
	_ DbSession = &Session{}
	_ DbSession = &Tx{}
)

func InTxFactory(ctx context.Context, sess DbSession, optionalTx DBRunner, failIfInTx bool, cb func(ctx context.Context, tx *Tx) error) error {
	return sess.InTx(ctx, optionalTx, failIfInTx, cb)
}
