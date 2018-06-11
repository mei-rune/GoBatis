package tests

import (
	"bytes"
	"flag"
	"log"
	"math"
	"net"
	"reflect"
	"testing"
	"time"

	_ "github.com/go-sql-driver/mysql"
	_ "github.com/lib/pq"
	gobatis "github.com/runner-mei/GoBatis"
)

type User struct {
	ID          int64                  `db:"id,pk"`
	Name        string                 `db:"name"`
	Nickname    string                 `db:"nickname"`
	Password    string                 `db:"password"`
	Description string                 `db:"description"`
	Birth       time.Time              `db:"birth"`
	Address     string                 `db:"address"`
	HostIP      net.IP                 `db:"host_ip"`
	HostMAC     net.HardwareAddr       `db:"host_mac"`
	HostIPPtr   *net.IP                `db:"host_ip_ptr"`
	HostMACPtr  *net.HardwareAddr      `db:"host_mac_ptr"`
	Sex         string                 `db:"sex"`
	ContactInfo map[string]interface{} `db:"contact_info"`
	CreateTime  time.Time              `db:"create_time"`
}

func AssertUser(t testing.TB, excepted, actual User) {
	if helper, ok := t.(interface {
		Helper()
	}); ok {
		helper.Helper()
	}
	if excepted.ID != actual.ID {
		t.Error("[ID] excepted is", excepted.ID)
		t.Error("[ID] actual   is", actual.ID)
	}
	if excepted.Name != actual.Name {
		t.Error("[Name] excepted is", excepted.Name)
		t.Error("[Name] actual   is", actual.Name)
	}
	if excepted.Nickname != actual.Nickname {
		t.Error("[Nickname] excepted is", excepted.Nickname)
		t.Error("[Nickname] actual   is", actual.Nickname)
	}
	if excepted.Password != actual.Password {
		t.Error("[Password] excepted is", excepted.Password)
		t.Error("[Password] actual   is", actual.Password)
	}
	if excepted.Description != actual.Description {
		t.Error("[Description] excepted is", excepted.Description)
		t.Error("[Description] actual   is", actual.Description)
	}
	if excepted.Address != actual.Address {
		t.Error("[Address] excepted is", excepted.Address)
		t.Error("[Address] actual   is", actual.Address)
	}

	if len(excepted.HostIP) != 0 || len(actual.HostIP) != 0 {
		if !bytes.Equal(excepted.HostIP, actual.HostIP) {
			t.Error("[HostIP] excepted is", excepted.HostIP)
			t.Error("[HostIP] actual   is", actual.HostIP)
		}
	}
	if len(excepted.HostMAC) != 0 || len(actual.HostMAC) != 0 {
		if !bytes.Equal(excepted.HostMAC, actual.HostMAC) {
			t.Error("[HostMAC] excepted is", excepted.HostMAC)
			t.Error("[HostMAC] actual   is", actual.HostMAC)
		}
	}

	if (excepted.HostIPPtr != nil && len(*excepted.HostIPPtr) != 0) ||
		(actual.HostIPPtr != nil && len(*actual.HostIPPtr) != 0) {
		if !reflect.DeepEqual(excepted.HostIPPtr, actual.HostIPPtr) {
			t.Error("[HostIPPtr] excepted is", excepted.HostIPPtr)
			t.Error("[HostIPPtr] actual   is", actual.HostIPPtr)
		}
	}

	if (excepted.HostMACPtr != nil && len(*excepted.HostMACPtr) != 0) ||
		(actual.HostMACPtr != nil && len(*actual.HostMACPtr) != 0) {
		if !reflect.DeepEqual(excepted.HostMACPtr, actual.HostMACPtr) {
			t.Error("[HostMACPtr] excepted is", excepted.HostMACPtr)
			t.Error("[HostMACPtr] actual   is", actual.HostMACPtr)
		}
	}
	if excepted.Sex != actual.Sex {
		t.Error("[Sex] excepted is", excepted.Sex)
		t.Error("[Sex] actual   is", actual.Sex)
	}
	if len(excepted.ContactInfo) != len(actual.ContactInfo) ||
		!reflect.DeepEqual(excepted.ContactInfo, actual.ContactInfo) {
		t.Error("[ContactInfo] excepted is", excepted.ContactInfo)
		t.Error("[ContactInfo] actual   is", actual.ContactInfo)
	}

	if excepted.Birth.Format("2006-01-02") != actual.Birth.Format("2006-01-02") {
		t.Error("[Birth] excepted is", excepted.Birth.Format("2006-01-02"))
		t.Error("[Birth] actual   is", actual.Birth.Format("2006-01-02"))
	}
	if math.Abs(excepted.CreateTime.Sub(actual.CreateTime).Seconds()) > 2 {
		t.Error("[CreateTime] excepted is", excepted.CreateTime.Format(time.RFC1123))
		t.Error("[CreateTime] actual   is", actual.CreateTime.Format(time.RFC1123))
	}
}

const (
	mysql = "DROP TABLE IF EXISTS `gobatis_users`;" +

		" CREATE TABLE `gobatis_users` (" +
		"  `id` int(11) NOT NULL AUTO_INCREMENT," +
		"  `name` varchar(45) DEFAULT NULL," +
		"  `nickname` varchar(45) DEFAULT NULL," +
		"  `password` varchar(255) DEFAULT NULL," +
		"  `description` varchar(255) DEFAULT NULL COMMENT '自我描述'," +
		"  `birth` date DEFAULT NULL," +
		"  `address` varchar(45) DEFAULT NULL COMMENT '地址'," +
		"  `host_ip` varchar(50) DEFAULT NULL," +
		"  `host_mac` varchar(50) DEFAULT NULL," +
		"  `host_ip_ptr` varchar(50) DEFAULT NULL," +
		"  `host_mac_ptr` varchar(50) DEFAULT NULL," +
		"  `sex` varchar(45) DEFAULT NULL COMMENT '性别'," +
		"  `contact_info` varchar(1000) DEFAULT NULL COMMENT '联系方式：如qq,msn,网站等 json方式保存{\"key\",\"value\"}'," +
		"  `create_time` datetime," +
		"  PRIMARY KEY (`id`)" +
		") ENGINE=InnoDB AUTO_INCREMENT=1 DEFAULT CHARSET=utf8 COMMENT='用户表';"

	mssql = `IF OBJECT_ID('dbo.gobatis_users', 'U') IS NOT NULL 
		DROP TABLE gobatis_users;

		CREATE TABLE gobatis_users (
		  id int IDENTITY NOT NULL PRIMARY KEY,
		  name varchar(45) DEFAULT NULL,
		  nickname varchar(45) DEFAULT NULL,
		  password varchar(255) DEFAULT NULL,
		  description varchar(255) DEFAULT NULL,
		  birth date DEFAULT NULL,
		  address varchar(45) DEFAULT NULL,
		  host_ip varchar(50) DEFAULT NULL,
		  host_mac varchar(50) DEFAULT NULL,
			host_ip_ptr varchar(50) DEFAULT NULL,
			host_mac_ptr varchar(50) DEFAULT NULL,
		  sex varchar(45) DEFAULT NULL,
		  contact_info varchar(1000) DEFAULT NULL,
		  create_time datetimeoffset
		);`

	postgresql = `
DROP TABLE IF EXISTS gobatis_users;

CREATE TABLE IF NOT EXISTS gobatis_users
(
  id bigserial NOT NULL,
  name character varying(45),
  nickname character varying(45),
  password character varying(255),
  description character varying(255), -- 自我描述
  birth timestamp with time zone,
  address character varying(45), -- 地址
	host_ip varchar(50) DEFAULT NULL,
	host_mac varchar(50) DEFAULT NULL,
	host_ip_ptr varchar(50) DEFAULT NULL,
	host_mac_ptr varchar(50) DEFAULT NULL,
  sex character varying(45), -- 性别
  contact_info character varying(1000), -- 联系方式：如qq,msn,网站等 json方式保存{"key","value"}
  create_time timestamp with time zone,
  CONSTRAINT gobatis_users_pkey PRIMARY KEY (id)
);`
)

var (
	TestDrv     string
	TestConnURL string
)

func init() {
	flag.StringVar(&TestDrv, "dbDrv", "postgres", "")
	flag.StringVar(&TestConnURL, "dbURL", "host=127.0.0.1 user=golang password=123456 dbname=golang sslmode=disable", "")
	//flag.StringVar(&TestConnURL, "dbURL", "golang:123456@tcp(localhost:3306)/golang?autocommit=true&parseTime=true&multiStatements=true", "")
	//flag.StringVar(&TestConnURL, "dbURL", "sqlserver://golang:123456@127.0.0.1?database=golang&connection+timeout=30", "")
}

func Run(t testing.TB, cb func(t testing.TB, factory *gobatis.SessionFactory)) {
	log.SetFlags(log.Ldate | log.Lshortfile)

	o, err := gobatis.New(&gobatis.Config{DriverName: TestDrv,
		DataSource: TestConnURL,
		XMLPaths: []string{"tests",
			"../tests",
			"../../tests"},
		MaxIdleConns: 2,
		MaxOpenConns: 2,
		ShowSQL:      true})
	if err != nil {
		t.Error(err)
		return
	}

	defer func() {
		if err = o.Close(); err != nil {
			t.Error(err)
		}
	}()

	switch o.DbType() {
	case gobatis.DbTypePostgres:
		_, err = o.DB().Exec(postgresql)
	case gobatis.DbTypeMSSql:
		_, err = o.DB().Exec(mssql)
	default:
		_, err = o.DB().Exec(mysql)
	}

	if err != nil {
		t.Error(err)
		return
	}

	cb(t, o)
}
