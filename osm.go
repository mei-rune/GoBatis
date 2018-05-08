package osm

// osm (Object Sql Mapping)是用go编写的ORM工具，目前很简单，只能算是半成品，只支持mysql(因为我目前的项目是mysql,所以其他数据库没有测试过)。
//
// 以前是使用MyBatis开发java服务端，它的sql mapping很灵活，把sql独立出来，程序通过输入与输出来完成所有的数据库操作。
//
// osm就是对MyBatis的简单模仿。当然动态sql的生成是使用go和template包，所以sql mapping的格式与MyBatis的不同。sql xml 格式如下：
//  <?xml version="1.0" encoding="utf-8"?>
//  <osm>
//   <select id="selectUsers" result="structs">
//     SELECT id,email
//     FROM user
//     {{if ne .Email ""}} where email=#{Email} {{end}}
//     order by id
//   </select>
//  </osm>
//

import (
	"bytes"
	"database/sql"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"reflect"
	"strings"

	"upper.io/db.v3/lib/reflectx"
)

const (
	dbTypeMysql    = 0
	dbTypePostgres = 1
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

type osmBase struct {
	dbType     int
	db         dbRunner
	sqlMappers map[string]*sqlMapper
	mapper     *reflectx.Mapper
	isUnsafe   bool
}

// Osm 对象，通过Struct、Map、Array、value等对象以及Sql Map来操作数据库。可以开启事务。
type Osm struct {
	osmBase
}

// Tx 与Osm对象一样，不过是在事务中进行操作
type Tx struct {
	osmBase
}

type Config struct {
	MaxIdleConns int
	MaxOpenConns int
	IsUnsafe     bool
	TagPrefix    string
}

// New 创建一个新的Osm，这个过程会打开数据库连接。
//
//driverName是数据库驱动名称如"mysql".
//dataSource是数据库连接信息如"root:root@/51jczj?charset=utf8".
//xmlPaths是sql xml的路径如[]string{"test.xml"}.
//params是数据连接的参数，可以是0个1个或2个数字，第一个表示MaxIdleConns，第二个表示MaxOpenConns.
//
//如：
//  o, err := osm.New("mysql", "root:root@/51jczj?charset=utf8", []string{"test.xml"})
func New(driverName, dataSource string, xmlPaths []string, cfg *Config) (*Osm, error) {
	if logger == nil {
		logger = log.New(os.Stdout, "[osm] ", log.Flags())
	}

	db, err := sql.Open(driverName, dataSource)
	if err != nil {
		if db != nil {
			db.Close()
		}
		return nil, fmt.Errorf("create osm error : %s", err.Error())
	}

	if cfg != nil {
		if cfg.MaxIdleConns > 0 {
			db.SetMaxIdleConns(cfg.MaxIdleConns)
		}
		if cfg.MaxOpenConns > 0 {
			db.SetMaxOpenConns(cfg.MaxOpenConns)
		}
	}

	osm := new(Osm)
	switch driverName {
	case "postgres":
		osm.dbType = dbTypePostgres
	default:
		osm.dbType = dbTypeMysql
	}
	osm.db = db
	osm.sqlMappers = make(map[string]*sqlMapper)
	var tagPrefix string
	if cfg != nil {
		osm.isUnsafe = cfg.IsUnsafe
		tagPrefix = cfg.TagPrefix
	}
	osm.mapper = createMapper(tagPrefix, nil)

	osmXMLPaths := []string{}
	for _, xmlPath := range xmlPaths {
		pathInfo, err := os.Stat(xmlPath)
		if err != nil {
			if os.IsNotExist(err) {
				continue
			}
			return nil, err
		}

		if !pathInfo.IsDir() {
			osmXMLPaths = append(osmXMLPaths, xmlPath)
			continue
		}

		fs, err := ioutil.ReadDir(xmlPath)
		if err != nil {
			return nil, err
		}

		for _, fileInfo := range fs {
			if fileName := fileInfo.Name(); strings.ToLower(filepath.Ext(fileName)) == ".xml" {
				osmXMLPaths = append(osmXMLPaths, filepath.Join(xmlPath, fileName))
			}
		}
	}

	for _, osmXMLPath := range osmXMLPaths {
		sqlMappers, err := readMappers(osmXMLPath)
		if err != nil {
			return nil, err
		}

		for _, sm := range sqlMappers {
			osm.sqlMappers[sm.id] = sm
		}
	}
	return osm, nil
}

// Begin 打开事务
//
//如：
//  tx, err := o.Begin()
func (o *Osm) Begin() (tx *Tx, err error) {
	tx = new(Tx)
	tx.sqlMappers = o.sqlMappers
	tx.dbType = o.dbType
	tx.mapper = o.mapper
	tx.isUnsafe = o.isUnsafe

	if o.db == nil {
		err = fmt.Errorf("db no opened")
	} else {
		sqlDb, ok := o.db.(*sql.DB)
		if ok {
			tx.db, err = sqlDb.Begin()
		} else {
			err = fmt.Errorf("db no opened")
		}
	}

	return
}

// Close 与数据库断开连接，释放连接资源
//
//如：
//  err := o.Close()
func (o *Osm) Close() (err error) {
	if o.db == nil {
		err = fmt.Errorf("db no opened")
	} else {
		sqlDb, ok := o.db.(*sql.DB)
		if ok {
			err = sqlDb.Close()
			o.db = nil
		} else {
			err = fmt.Errorf("db no opened")
		}
	}
	return
}

// Commit 提交事务
//
//如：
//  err := tx.Commit()
func (o *Tx) Commit() error {
	if o.db == nil {
		return fmt.Errorf("tx no runing")
	}
	sqlTx, ok := o.db.(*sql.Tx)
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
	if o.db == nil {
		return fmt.Errorf("tx no runing")
	}
	sqlTx, ok := o.db.(*sql.Tx)
	if ok {
		return sqlTx.Rollback()
	}
	return fmt.Errorf("tx no runing")
}

// Delete 执行删除sql
//
//xml
//  <osm>
//  ...
//    <delete id="deleteUser">DELETE FROM user where id = #{Id};</delete>
//  ...
//  </osm>
//代码
//  user := User{Id: 3}
//  count,err := o.Delete("deleteUser", user)
//删除id为3的用户数据
func (o *osmBase) Delete(id string, params ...interface{}) (int64, error) {
	sql, sqlParams, _, err := o.readSQLParams(id, typeDelete, params...)
	if err != nil {
		return 0, err
	}

	if ShowSQL {
		logger.Printf(`id:"%s", sql:"%s", params:"%+v"`, id, sql, sqlParams)
	}

	stmt, err := o.db.Prepare(sql)
	if err != nil {
		return 0, err
	}
	result, err := stmt.Exec(sqlParams...)
	if err != nil {
		return 0, err
	}
	defer stmt.Close()
	return result.RowsAffected()
}

// Update 执行更新sql
//
//xml
//  <osm>
//  ...
//    <update id="updateUserEmail">UPDATE user SET email=#{Email} where id = #{Id};</update>
//  ...
//  </osm>
//代码
//  user := User{Id: 3, Email: "test@foxmail.com"}
//  count,err := o.Update("updateUserEmail", user)
//将id为3的用户email更新为"test@foxmail.com"
func (o *osmBase) Update(id string, params ...interface{}) (int64, error) {
	sql, sqlParams, _, err := o.readSQLParams(id, typeUpdate, params...)
	if err != nil {
		return 0, err
	}

	if ShowSQL {
		logger.Printf(`id:"%s", sql:"%s", params:"%+v"`, id, sql, sqlParams)
	}

	stmt, err := o.db.Prepare(sql)
	if err != nil {
		return 0, err
	}
	result, err := stmt.Exec(sqlParams...)
	if err != nil {
		return 0, err
	}
	defer stmt.Close()
	return result.RowsAffected()
}

// UpdateMulti 批量执行更新sql
//
//xml
//  <osm>
//  ...
//    <update id="updateUserEmail">
//       UPDATE user SET email=#{Email} where id = #{Id};
//       UPDATE user SET email=#{Email} where id = #{Id2};
//    </update>
//  ...
//  </osm>
//代码
//  user := User{Id: 3, Id2: 4, Email: "test@foxmail.com"}
//  err := o.UpdateMulti("updateUserEmail", user)
//将id为3和4的用户email更新为"test@foxmail.com"
func (o *osmBase) UpdateMulti(id string, params ...interface{}) error {
	sql, sqlParams, _, err := o.readSQLParams(id, typeUpdate, params...)
	if err != nil {
		return err
	}

	if ShowSQL {
		logger.Printf(`id:"%s", sql:"%s", params:"%+v"`, id, sql, sqlParams)
	}

	_, err = o.db.Exec(sql, sqlParams...)
	return err
}

// Insert 执行添加sql
//
//xml
//  <osm>
//  ...
//    <insert id="insertUser">INSERT INTO user(email) VALUES(#{Email});</insert>
//  ...
//  </osm>
//代码
//  user := User{Email: "test@foxmail.com"}
//  insertId,count,err := o.Insert("insertUser", user)
//添加一个用户数据，email为"test@foxmail.com"
func (o *osmBase) Insert(id string, params ...interface{}) (int64, int64, error) {
	sql, sqlParams, _, err := o.readSQLParams(id, typeInsert, params...)
	if err != nil {
		return 0, 0, err
	}

	if ShowSQL {
		logger.Printf(`id:"%s", sql:"%s", params:"%+v"`, id, sql, sqlParams)
	}

	stmt, err := o.db.Prepare(sql)
	if err != nil {
		return 0, 0, err
	}

	result, err := stmt.Exec(sqlParams...)
	if err != nil {
		return 0, 0, err
	}
	defer stmt.Close()

	var insertID int64
	if o.dbType == dbTypeMysql {
		insertID, err = result.LastInsertId()
		if err != nil {
			logger.Println(err)
		}
	}

	count, err := result.RowsAffected()
	return insertID, count, err
}

//执行查询sql, 返回单行数据
//
//xml
//  <select id="searchArchives">
//   <![CDATA[
//   SELECT id,email,create_time FROM user WHERE id=#{Id};
//   ]]>
//  </select>
func (o *osmBase) SelectOne(id string, params ...interface{}) Result {
	sql, sqlParams, _, err := o.readSQLParams(id, typeSelect, params...)

	return Result{o: o,
		id:        id,
		sql:       sql,
		sqlParams: sqlParams,
		err:       err}
}

type Result struct {
	o         *osmBase
	id        string
	sql       string
	sqlParams []interface{}
	err       error
}

func (result Result) Scan(value interface{}) error {
	if result.err != nil {
		return result.err
	}

	if ShowSQL {
		logger.Printf(`id:"%s", sql:"%s", params:"%+v"`, result.id, result.sql, result.sqlParams)
	}

	rows, err := result.o.db.Query(result.sql, result.sqlParams...)
	if err != nil {
		return err
	}
	defer rows.Close()

	if !rows.Next() {
		if err := rows.Err(); err != nil {
			return err
		}
		return sql.ErrNoRows
	}

	err = scanAny(result.o.mapper, rows, value, false, result.o.isUnsafe)
	if err != nil {
		return err
	}
	if err := rows.Close(); err != nil {
		return err
	}
	return nil
}

//执行查询sql, 返回多行数据
//
//xml
//  <select id="searchUsers">
//   <![CDATA[
//   SELECT id,email,create_time FROM user WHERE create_time >= #{create_time};
//   ]]>
//  </select>
func (o *osmBase) Select(id string, params ...interface{}) *Results {
	sql, sqlParams, _, err := o.readSQLParams(id, typeSelect, params...)

	return &Results{o: o,
		id:        id,
		sql:       sql,
		sqlParams: sqlParams,
		err:       err}
}

type Results struct {
	o         *osmBase
	id        string
	sql       string
	sqlParams []interface{}
	rows      *sql.Rows
	err       error
}

func (results *Results) Close() error {
	if results.rows != nil {
		return results.rows.Close()
	}
	return nil
}

func (results *Results) Err() error {
	return results.err
}

func (results *Results) Next() bool {
	if results.err != nil {
		return false
	}

	if results.rows == nil {
		if ShowSQL {
			logger.Printf(`id:"%s", sql:"%s", params:"%+v"`, results.id, results.sql, results.sqlParams)
		}

		results.rows, results.err = results.o.db.Query(results.sql, results.sqlParams...)
		if results.err != nil {
			return false
		}
	}

	return results.rows.Next()
}

func (results *Results) Scan(value interface{}) error {
	if results.err != nil {
		return results.err
	}

	if results.rows == nil {
		return errors.New("please invoke Next()")
	}
	return scanAny(results.o.mapper, results.rows, value, false, results.o.isUnsafe)
}

func (results *Results) ScanSlice(value interface{}) error {
	if results.err != nil {
		return results.err
	}

	if results.rows != nil {
		return errors.New("please not invoke Next()")
	}

	if ShowSQL {
		logger.Printf(`id:"%s", sql:"%s", params:"%+v"`, results.id, results.sql, results.sqlParams)
	}

	rows, err := results.o.db.Query(results.sql, results.sqlParams...)
	if err != nil {
		return err
	}
	defer rows.Close()

	err = scanAll(results.o.mapper, rows, value, false, results.o.isUnsafe)
	if err != nil {
		return err
	}

	return rows.Close()
}

func (o *osmBase) readSQLParams(id string, sqlType statementType, params ...interface{}) (sql string, sqlParams []interface{}, rType resultType, err error) {
	sqlParams = make([]interface{}, 0)
	sm, ok := o.sqlMappers[id]
	err = nil

	if !ok {
		err = fmt.Errorf("sql '%s' error : id not found ", id)
		return
	}
	rType = sm.result

	if sm.sqlType != sqlType {
		err = fmt.Errorf("sql '%s' error : Select type Error", id)
		return
	}

	var param interface{}
	paramsSize := len(params)
	if paramsSize == 0 {
		sql = sm.sql
		return
	}

	if paramsSize == 1 {
		param = params[0]
	} else {
		param = params
	}

	//sql start
	var sqlBuilder strings.Builder
	paramNames := []string{}
	var buf bytes.Buffer

	err = sm.sqlTemplate.Execute(&buf, param)
	if err != nil {
		logger.Println(err)
	}
	sqlOrg := buf.String()

	sqlTemp := sqlOrg
	errorIndex := 0
	signIndex := 1
	for strings.Contains(sqlTemp, "#{") {
		si := strings.Index(sqlTemp, "#{")
		sqlBuilder.WriteString(sqlTemp[0:si])
		sqlTemp = sqlTemp[si+2:]
		errorIndex += si + 2

		ei := strings.Index(sqlTemp, "}")
		if ei < 0 {
			err = errors.New(markSQLError(sqlOrg, errorIndex))
			return
		}

		if o.dbType == dbTypePostgres {
			sqlBuilder.WriteString(fmt.Sprintf("$%d", signIndex))
			signIndex++
		} else {
			sqlBuilder.WriteString("?")
		}
		paramNames = append(paramNames, sqlTemp[0:ei])
		sqlTemp = sqlTemp[ei+1:]
		errorIndex += ei + 1
	}
	sqlBuilder.WriteString(sqlTemp)
	//sql end

	sql = sqlBuilder.String()

	v := reflect.ValueOf(param)

	kind := v.Kind()
	switch {
	case kind == reflect.Array || kind == reflect.Slice:
		for i := 0; i < v.Len() && i < len(paramNames); i++ {
			vv := v.Index(i)
			sqlParams = append(sqlParams, vv.Interface())
		}
	case kind == reflect.Map:
		for _, paramName := range paramNames {
			vv := v.MapIndex(reflect.ValueOf(paramName))
			if !vv.IsValid() {
				sqlParams = append(sqlParams, nil)
				log.Printf("[warn] sql '%s' error : '%s' no exist", sm.id, paramName)
				continue
			}

			sqlParams = append(sqlParams, vv.Interface())
		}
	case kind == reflect.Struct:
		for _, paramName := range paramNames {
			vv := v.FieldByName(paramName)
			if !vv.IsValid() {
				sqlParams = append(sqlParams, nil)
				log.Printf("[warn] sql '%s' error : '%s' no exist", sm.id, paramName)
				continue
			}

			sqlParams = append(sqlParams, vv.Interface())
		}
	case kind == reflect.Bool ||
		kind == reflect.Int ||
		kind == reflect.Int8 ||
		kind == reflect.Int16 ||
		kind == reflect.Int32 ||
		kind == reflect.Int64 ||
		kind == reflect.Uint ||
		kind == reflect.Uint8 ||
		kind == reflect.Uint16 ||
		kind == reflect.Uint32 ||
		kind == reflect.Uint64 ||
		kind == reflect.Uintptr ||
		kind == reflect.Float32 ||
		kind == reflect.Float64 ||
		kind == reflect.Complex64 ||
		kind == reflect.Complex128 ||
		kind == reflect.String:
		sqlParams = append(sqlParams, param)
	default:
		// err = fmt.Errorf(`id:"%s", sql:"%s", argType:%T, argValue=%#v`, id, sql, v, v)
		// return
	}

	return
}
