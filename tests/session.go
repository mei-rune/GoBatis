package tests

import (
	"flag"
	"log"
	"testing"

	_ "github.com/go-sql-driver/mysql"
	_ "github.com/lib/pq"
	gobatis "github.com/runner-mei/GoBatis"
)

const (
	mysql = `DROP TABLE IF EXISTS gobatis_users;
		  DROP TABLE IF EXISTS gobatis_usergroups; 
		  DROP TABLE IF EXISTS gobatis_user_and_groups;

		  ` +
		" CREATE TABLE gobatis_users (" +
		"  id int(11) NOT NULL AUTO_INCREMENT," +
		"  name varchar(45) DEFAULT NULL," +
		"  nickname varchar(45) DEFAULT NULL," +
		"  password varchar(255) DEFAULT NULL," +
		"  description varchar(255) DEFAULT NULL COMMENT '自我描述'," +
		"  birth date DEFAULT NULL," +
		"  address varchar(45) DEFAULT NULL COMMENT '地址'," +
		"  host_ip varchar(50) DEFAULT NULL," +
		"  host_mac varchar(50) DEFAULT NULL," +
		"  host_ip_ptr varchar(50) DEFAULT NULL," +
		"  host_mac_ptr varchar(50) DEFAULT NULL," +
		"  sex varchar(45) DEFAULT NULL COMMENT '性别'," +
		"  contact_inf` varchar(1000) DEFAULT NULL COMMENT '联系方式：如qq,msn,网站等 json方式保存{\"key\",\"value\"}'," +
		"  create_time datetime," +
		"  `field1`      int NULL," +
		"  `field2`      int NULL," +
		"  `field3`      float NULL," +
		"  `field4`      float NULL," +
		"  `field5`      varchar(50) NULL," +
		"  `field6`      datetime NULL," +
		"  PRIMARY KEY (id)" +
		") ENGINE=InnoDB AUTO_INCREMENT=1 DEFAULT CHARSET=utf8 COMMENT='用户表';" + `

		 CREATE TABLE gobatis_usergroups (
		  id int(11) NOT NULL AUTO_INCREMENT,
		  name varchar(45) DEFAULT NULL,
		  PRIMARY KEY (id)
		) ENGINE=InnoDB AUTO_INCREMENT=1 DEFAULT CHARSET=utf8 COMMENT='用户组';

		 CREATE TABLE gobatis_user_and_groups (
		  user_id int(11) NOT NULL,
		  group_id int(11) NOT NULL,
		  PRIMARY KEY (user_id,group_id)
		) ENGINE=InnoDB DEFAULT CHARSET=utf8 COMMENT='用户组';
		`

	mssql = `
		IF OBJECT_ID('dbo.gobatis_user_and_groups', 'U') IS NOT NULL 
		DROP TABLE gobatis_user_and_groups;

		IF OBJECT_ID('dbo.gobatis_users', 'U') IS NOT NULL 
		DROP TABLE gobatis_users;

		IF OBJECT_ID('dbo.gobatis_usergroups', 'U') IS NOT NULL 
		DROP TABLE gobatis_usergroups;

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
		  field1      int NULL,
			field2      int NULL,
			field3      float NULL,
			field4      float NULL,
			field5      varchar(50) NULL,
      field6      datetimeoffset NULL,

		  create_time datetimeoffset
		);

		CREATE TABLE gobatis_usergroups (
		  id int IDENTITY NOT NULL PRIMARY KEY,
		  name varchar(45) DEFAULT NULL,
		  PRIMARY KEY (id)
		);

		 CREATE TABLE gobatis_user_and_groups (
		  user_id int(11) NOT NULL,
		  group_id int(11) NOT NULL,
		  PRIMARY KEY (user_id,group_id)
		);

		`

	postgresql = `
		DROP TABLE IF EXISTS gobatis_user_and_groups;
		DROP TABLE IF EXISTS gobatis_users;
		DROP TABLE IF EXISTS gobatis_usergroups;

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
			field1      int NULL,
			field2      int NULL,
			field3      float NULL,
			field4      float NULL,
			field5      varchar(50) NULL,
		  PRIMARY KEY (id)
		);

		CREATE TABLE gobatis_usergroups (
		  id bigserial NOT NULL,
		  name varchar(45) DEFAULT NULL,
		  PRIMARY KEY (id)
		);

		 CREATE TABLE gobatis_user_and_groups (
		  user_id int(11) NOT NULL,
		  group_id int(11) NOT NULL,
		  PRIMARY KEY (user_id,group_id)
		);
`
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
