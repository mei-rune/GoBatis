package osm

import (
	"database/sql"
	"log"
	"reflect"
	"testing"
	"time"

	// _ "github.com/go-sql-driver/mysql"
	_ "github.com/lib/pq"
)

type User struct {
	ID           int64     `json:"id"`
	Email        string    `json:"email"`
	Mobile       string    `json:"mobile"`
	Nickname     string    `json:"nickname"`
	Password     string    `json:"password"`
	HeadImageURL string    `json:"head_image_url"`
	Description  string    `json:"description"`
	Name         string    `json:"name"`
	Birth        time.Time `json:"birth"`
	Province     string    `json:"province"`
	City         string    `json:"city"`
	Company      string    `json:"company"`
	Address      string    `json:"address"`
	Sex          string    `json:"sex"`
	ContactInfo  string    `json:"contact_info"`
	CreateTime   time.Time `json:"create_time"`
}

const (
	mysql = "CREATE TABLE `osm_users` (" +
		"  `id` int(11) NOT NULL AUTO_INCREMENT," +
		"  `email` varchar(255) DEFAULT NULL," +
		"  `mobile` varchar(45) DEFAULT NULL," +
		"  `nickname` varchar(45) DEFAULT NULL," +
		"  `password` varchar(255) DEFAULT NULL," +
		"  `head_image_url` varchar(255) DEFAULT NULL," +
		"  `description` varchar(255) DEFAULT NULL COMMENT '自我描述'," +
		"  `name` varchar(45) DEFAULT NULL," +
		"  `birth` date DEFAULT NULL," +
		"  `province` varchar(45) DEFAULT NULL COMMENT '省'," +
		"  `city` varchar(45) DEFAULT NULL COMMENT '市'," +
		"  `company` varchar(45) DEFAULT NULL COMMENT '公司'," +
		"  `address` varchar(45) DEFAULT NULL COMMENT '地址'," +
		"  `sex` varchar(45) DEFAULT NULL COMMENT '性别'," +
		"  `contact_info` varchar(1000) DEFAULT NULL COMMENT '联系方式：如qq,msn,网站等 json方式保存{\"key\",\"value\"}'," +
		"  `create_time` datetime," +
		"  PRIMARY KEY (`id`)" +
		") ENGINE=InnoDB AUTO_INCREMENT=1 DEFAULT CHARSET=utf8 COMMENT='用户表';"

	postgresql = `
DROP TABLE IF EXISTS osm_users;

CREATE TABLE IF NOT EXISTS osm_users
(
  id bigserial NOT NULL,
  email character varying(255),
  mobile character varying(45),
  nickname character varying(45),
  password character varying(255),
  head_image_url character varying(255),
  description character varying(255), -- 自我描述
  name character varying(45),
  birth timestamp with time zone,
  province character varying(45), -- 省
  city character varying(45), -- 市
  company character varying(45), -- 公司
  address character varying(45), -- 地址
  sex character varying(45), -- 性别
  contact_info character varying(1000), -- 联系方式：如qq,msn,网站等 json方式保存{"key","value"}
  create_time timestamp with time zone,
  CONSTRAINT osm_users_pkey PRIMARY KEY (id)
);`
)

func TestOsm(t *testing.T) {
	log.SetFlags(log.Ldate | log.Lshortfile)

	ShowSQL = true

	o, err := New("postgres", "host=127.0.0.1 user=golang password=123456 dbname=golang sslmode=disable",
		[]string{"example/test.xml"}, nil)
	// o, err := osm.New("mysql", "root:123456@tcp(127.0.0.1:3306)/test?charset=utf8", []string{"test.xml"})
	if err != nil {
		t.Error(err)
		return
	}

	defer func() {
		if err = o.Close(); err != nil {
			t.Error(err)
		}
	}()

	if _, err = o.db.Exec(postgresql); err != nil {
		t.Error(err)
		return
	}

	insertUser := User{
		Email:        "test@foxmail.com",
		Mobile:       "13113113113",
		Nickname:     "haha",
		HeadImageURL: "www.baidu.com",
		Password:     "password",
		Description:  "地球人",
		Name:         "张三",
		Province:     "上海",
		City:         "上海",
		Company:      "电信",
		Address:      "沪南路1155号",
		Sex:          "女",
		ContactInfo:  `{"QQ":"8888888"}`,
		Birth:        time.Now(),
		CreateTime:   time.Now(),
	}

	user := User{
		Email: "test@foxmail.com",
	}

	t.Run("selectUsers", func(t *testing.T) {
		_, _, err := o.Insert("insertUser", insertUser)
		if err != nil {
			t.Error(err)
		}
		_, _, err = o.Insert("insertUser", insertUser)
		if err != nil {
			t.Error(err)
		}

		var users []User
		err = o.Select("selectUsers", user).ScanSlice(&users)
		if err != nil {
			t.Error(err)
			return
		}

		if len(users) != 2 {
			t.Error("excepted size is", 2)
			t.Error("actual size   is", len(users))
		} else {

			insertUser2 := insertUser
			insertUser2.Birth = insertUser2.Birth.UTC()
			insertUser2.CreateTime = insertUser2.CreateTime.UTC()

			for _, u := range users {

				insertUser2.ID = u.ID
				u.Birth = u.Birth.UTC()
				u.CreateTime = u.CreateTime.UTC()

				if !reflect.DeepEqual(insertUser2, u) {
					t.Error("excepted is", insertUser2)
					t.Error("actual   is", u)
				}
			}
		}
	})

	t.Run("selectUser", func(t *testing.T) {
		_, _, err := o.Insert("insertUser", insertUser)
		if err != nil {
			t.Error(err)
		}

		u := User{Email: insertUser.Email}
		err = o.SelectOne("selectUser", u).Scan(&u)
		if err != nil {
			t.Error(err)
			return
		}

		insertUser.ID = u.ID
		insertUser.Birth = insertUser.Birth.UTC()
		insertUser.CreateTime = insertUser.CreateTime.UTC()
		u.Birth = u.Birth.UTC()
		u.CreateTime = u.CreateTime.UTC()
		if !reflect.DeepEqual(insertUser, u) {
			t.Error("excepted is", insertUser)
			t.Error("actual   is", u)
		}
	})

	t.Run("updateUser", func(t *testing.T) {
		_, _, err := o.Insert("insertUser", insertUser)
		if err != nil {
			t.Error(err)
		}

		u := User{Email: insertUser.Email}
		err = o.SelectOne("selectUser", u).Scan(&u)
		if err != nil {
			t.Error(err)
			return
		}

		updateUser := insertUser
		updateUser.ID = u.ID
		updateUser.Email = "test@foxmail.com"
		updateUser.Birth = time.Now()
		updateUser.HeadImageURL = "www.qq.com"
		updateUser.CreateTime = time.Now()
		_, err = o.Update("updateUser", updateUser)
		if err != nil {
			t.Error(err)
		}

		updateUser.Birth = updateUser.Birth.UTC()
		updateUser.CreateTime = updateUser.CreateTime.UTC()

		err = o.SelectOne("selectUser", u).Scan(&u)
		if err != nil {
			t.Error(err)
			return
		}
		u.Birth = u.Birth.UTC()
		u.CreateTime = u.CreateTime.UTC()
		if !reflect.DeepEqual(updateUser, u) {
			t.Error("excepted is", updateUser)
			t.Error("actual   is", u)
		}
	})

	t.Run("deleteUser", func(t *testing.T) {
		if _, err = o.db.Exec(`DELETE FROM osm_users`); err != nil {
			t.Error(err)
			return
		}

		_, _, err := o.Insert("insertUser", insertUser)
		if err != nil {
			t.Error(err)
		}

		u := User{Email: insertUser.Email}
		err = o.SelectOne("selectUser", u).Scan(&u)
		if err != nil {
			t.Error(err)
			return
		}

		deleteUser := User{ID: u.ID}
		_, err = o.Delete("deleteUser", deleteUser)
		if err != nil {
			t.Error(err)
		}

		err = o.SelectOne("selectUser", u).Scan(&u)
		if err == nil {
			t.Error("DELETE fail")
			return
		}

		if err != sql.ErrNoRows {
			t.Error(err)
		}
	})
}
