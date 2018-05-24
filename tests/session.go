package tests

import (
	"flag"
	"log"
	"testing"
	"time"

	"github.com/runner-mei/gobatis"
)

type User struct {
	ID          int64     `json:"id"`
	Name        string    `json:"name"`
	Nickname    string    `json:"nickname"`
	Password    string    `json:"password"`
	Description string    `json:"description"`
	Birth       time.Time `json:"birth"`
	Location    string    `json:"location"`
	Company     string    `json:"company"`
	Address     string    `json:"address"`
	Sex         string    `json:"sex"`
	ContactInfo string    `json:"contact_info"`
	CreateTime  time.Time `json:"create_time"`
}

const (
	mysql = "CREATE TABLE `gobatis_users` (" +
		"  `id` int(11) NOT NULL AUTO_INCREMENT," +
		"  `name` varchar(45) DEFAULT NULL," +
		"  `nickname` varchar(45) DEFAULT NULL," +
		"  `password` varchar(255) DEFAULT NULL," +
		"  `description` varchar(255) DEFAULT NULL COMMENT '自我描述'," +
		"  `birth` date DEFAULT NULL," +
		"  `address` varchar(45) DEFAULT NULL COMMENT '地址'," +
		"  `sex` varchar(45) DEFAULT NULL COMMENT '性别'," +
		"  `contact_info` varchar(1000) DEFAULT NULL COMMENT '联系方式：如qq,msn,网站等 json方式保存{\"key\",\"value\"}'," +
		"  `create_time` datetime," +
		"  PRIMARY KEY (`id`)" +
		") ENGINE=InnoDB AUTO_INCREMENT=1 DEFAULT CHARSET=utf8 COMMENT='用户表';"

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
  sex character varying(45), -- 性别
  contact_info character varying(1000), -- 联系方式：如qq,msn,网站等 json方式保存{"key","value"}
  create_time timestamp with time zone,
  CONSTRAINT gobatis_users_pkey PRIMARY KEY (id)
);`
)

var (
	testDrv     string
	testConnURL string
)

func init() {
	flag.StringVar(&testDrv, "drv", "postgres", "")
	flag.StringVar(&testConnURL, "connURL", "host=127.0.0.1 user=golang password=123456 dbname=golang sslmode=disable", "")
}

func Run(t testing.TB, cb func(t testing.TB, factory *gobatis.SessionFactory)) {
	log.SetFlags(log.Ldate | log.Lshortfile)

	gobatis.ShowSQL = true

	o, err := gobatis.New(testDrv, testConnURL,
		[]string{"example/test.xml",
			"../example/test.xml",
			"../../example/test.xml"}, nil)
	if err != nil {
		t.Error(err)
		return
	}

	defer func() {
		if err = o.Close(); err != nil {
			t.Error(err)
		}
	}()

	switch testDrv {
	case "postgres":
		_, err = o.DB().Exec(postgresql)
	default:
		_, err = o.DB().Exec(mysql)
	}

	if err != nil {
		t.Error(err)
		return
	}

	cb(t, o)
}
