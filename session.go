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
	"database/sql"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
)

var (
	logger *log.Logger

	// ShowSQL 显示执行的sql，用于调试，使用logger打印
	ShowSQL = false
)

type dbRunner interface {
	Prepare(query string) (*sql.Stmt, error)
	Exec(query string, args ...interface{}) (sql.Result, error)
	Query(query string, args ...interface{}) (*sql.Rows, error)
	QueryRow(query string, args ...interface{}) *sql.Row
}

// SessionFactory 对象，通过Struct、Map、Array、value等对象以及Sql Map来操作数据库。可以开启事务。
type SessionFactory struct {
	Session
}

type Config struct {
	DriverName string
	DataSource string
	XMLPaths   []string

	MaxIdleConns int
	MaxOpenConns int
	IsUnsafe     bool
	TagPrefix    string
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
	if logger == nil {
		logger = log.New(os.Stdout, "[gobatis] ", log.Flags())
	}

	db, err := sql.Open(cfg.DriverName, cfg.DataSource)
	if err != nil {
		if db != nil {
			db.Close()
		}
		return nil, fmt.Errorf("create gobatis error : %s", err.Error())
	}

	if cfg != nil {
		if cfg.MaxIdleConns > 0 {
			db.SetMaxIdleConns(cfg.MaxIdleConns)
		}
		if cfg.MaxOpenConns > 0 {
			db.SetMaxOpenConns(cfg.MaxOpenConns)
		}
	}

	base := Connection{
		dbType: ToDbType(cfg.DriverName),
	}
	if base.dbType == DbTypeNone {
		base.dbType = DbTypePostgres
	}
	base.db = db
	base.sqlStatements = make(map[string]*MappedStatement)
	var tagPrefix string
	if cfg != nil {
		base.isUnsafe = cfg.IsUnsafe
		tagPrefix = cfg.TagPrefix
	}
	base.mapper = CreateMapper(tagPrefix, nil)

	dbName := strings.ToLower(ToDbName(base.DbType()))
	xmlPaths := []string{}
	for _, xmlPath := range cfg.XMLPaths {
		pathInfo, err := os.Stat(xmlPath)
		if err != nil {
			if os.IsNotExist(err) {
				continue
			}
			return nil, err
		}

		if !pathInfo.IsDir() {
			xmlPaths = append(xmlPaths, xmlPath)
			continue
		}

		fs, err := ioutil.ReadDir(xmlPath)
		if err != nil {
			return nil, err
		}

		for _, fileInfo := range fs {
			if !fileInfo.IsDir() {
				if fileName := fileInfo.Name(); strings.ToLower(filepath.Ext(fileName)) == ".xml" {
					xmlPaths = append(xmlPaths, filepath.Join(xmlPath, fileName))
				}
				continue
			}

			if dbName != strings.ToLower(fileInfo.Name()) {
				continue
			}

			dialectDirs, err := ioutil.ReadDir(filepath.Join(xmlPath, fileInfo.Name()))
			if err != nil {
				return nil, err
			}

			for _, dialectInfo := range dialectDirs {
				if fileName := dialectInfo.Name(); strings.ToLower(filepath.Ext(fileName)) == ".xml" {
					xmlPaths = append(xmlPaths, filepath.Join(xmlPath, fileInfo.Name(), fileName))
				}
			}
		}
	}

	for _, xmlPath := range xmlPaths {
		log.Println("load xml -", xmlPath)
		statements, err := readMappedStatements(xmlPath)
		if err != nil {
			return nil, err
		}

		for _, sm := range statements {
			base.sqlStatements[sm.id] = sm
		}
	}

	if err := runInit(&InitContext{DbType: base.dbType, Statements: base.sqlStatements}); err != nil {
		return nil, err
	}
	return &SessionFactory{Session: Session{base: base}}, nil
}

func (o *SessionFactory) DB() dbRunner {
	return o.base.db
}

// Begin 打开事务
//
//如：
//  tx, err := o.Begin()
func (o *SessionFactory) Begin(nativeTx ...*sql.Tx) (tx *Tx, err error) {
	tx = new(Tx)
	tx.Session = o.Session

	var native *sql.Tx
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

func (sess *Session) DB() dbRunner {
	return sess.base.db
}

func (sess *Session) DbType() int {
	return sess.base.dbType
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
	return sess.base.Delete(id, nil, params)
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
	return sess.base.Update(id, nil, params)
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
	return sess.base.Insert(id, nil, params)
}

//执行查询sql, 返回单行数据
//
//xml
//  <select id="searchArchives">
//   <![CDATA[
//   SELECT id,email,create_time FROM user WHERE id=#{Id};
//   ]]>
//  </select>
func (sess *Session) SelectOne(id string, params ...interface{}) Result {
	return sess.base.SelectOne(id, nil, params)
}

//执行查询sql, 返回多行数据
//
//xml
//  <select id="searchUsers">
//   <![CDATA[
//   SELECT id,email,create_time FROM user WHERE create_time >= #{create_time};
//   ]]>
//  </select>
func (sess *Session) Select(id string, params ...interface{}) *Results {
	return sess.base.Select(id, nil, params)
}
