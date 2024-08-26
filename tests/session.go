package tests

import (
	"context"
	"flag"
	"log"
	"os"
	"strings"
	"testing"

	gobatis "github.com/runner-mei/GoBatis"
	"github.com/runner-mei/GoBatis/dialects"
	_ "github.com/runner-mei/GoBatis/dialects/dm"

	// _ "github.com/SAP/go-hdb/driver"                  // sap hana
	_ "gitee.com/chunanyong/dm"                       // 达梦
	_ "gitee.com/opengauss/openGauss-connector-go-pq" // openGauss
	_ "gitee.com/runner.mei/gokb"                     // 人大金仓
	_ "github.com/microsoft/go-mssqldb"
	_ "github.com/go-sql-driver/mysql"
	_ "github.com/lib/pq"
	_ "github.com/sijms/go-ora/v2" // oracle
	// _ "github.com/ibmdb/go_ibm_db"
)

const (
	MysqlScript = `DROP TABLE IF EXISTS gobatis_users;
		  DROP TABLE IF EXISTS gobatis_usergroups; 
		  DROP TABLE IF EXISTS gobatis_user_and_groups;

		
		 CREATE TABLE gobatis_users (
		  id int(11) NOT NULL AUTO_INCREMENT,
		  name varchar(45) DEFAULT NULL,
		  nickname varchar(45) DEFAULT NULL,
		  password varchar(255) DEFAULT NULL,
		  description varchar(255) DEFAULT NULL COMMENT '自我描述',
		  birth date DEFAULT NULL,
		  address varchar(45) DEFAULT NULL COMMENT '地址',
		  host_ip varchar(50) DEFAULT NULL,
		  host_mac varchar(50) DEFAULT NULL,
		  host_ip_ptr varchar(50) DEFAULT NULL,
		  host_mac_ptr varchar(50) DEFAULT NULL,
		  sex varchar(45) DEFAULT NULL COMMENT '性别',
		  contact_info varchar(1000) DEFAULT NULL COMMENT '联系方式：如qq,msn,网站等 json方式保存{"key","value"}',
		  create_time datetime,
		  field1      int NULL,
		  field2      int NULL,
		  field3      float NULL,
		  field4      float NULL,
		  field5      varchar(50) NULL,
		  field6      datetime NULL,
		  field7      datetime NULL,
		  fieldBool      boolean NULL,
		  fieldBoolP     boolean NULL,
		  PRIMARY KEY (id)
		) ENGINE=InnoDB AUTO_INCREMENT=1 DEFAULT CHARSET=utf8 COMMENT='用户表';

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


		DROP TABLE IF EXISTS gobatis_settings; 
		DROP TABLE IF EXISTS gobatis_list;
    
    CREATE TABLE gobatis_settings (
		  id int(11) NOT NULL AUTO_INCREMENT,
		  name varchar(45) DEFAULT NULL,
		  value varchar(45) DEFAULT NULL,
		  UNIQUE(name),
		  PRIMARY KEY (id)
		);
    
    CREATE TABLE gobatis_list (
		  id int(11) NOT NULL AUTO_INCREMENT,
		  name varchar(45) DEFAULT NULL,
		  UNIQUE(name),
		  PRIMARY KEY (id)
		);


		DROP TABLE IF EXISTS gobatis_testa; 
		DROP TABLE IF EXISTS gobatis_testb;
		DROP TABLE IF EXISTS gobatis_testc;
		DROP TABLE IF EXISTS gobatis_teste1;
		DROP TABLE IF EXISTS gobatis_testf1;
		DROP TABLE IF EXISTS gobatis_teste2;
		DROP TABLE IF EXISTS gobatis_testf2;
		DROP TABLE IF EXISTS gobatis_convert1;
		DROP TABLE IF EXISTS gobatis_convert2;

		CREATE TABLE gobatis_testa (
		  id int(11) NOT NULL AUTO_INCREMENT,
		  field0      boolean NULL,
		  field1      int NULL,
		  field2      int NULL,
		  field3      float NULL,
		  field4      float NULL,
		  field5      varchar(50) NULL,
		  field6      datetime NULL,
		  field7      varchar(50) NULL,
		  field8      varchar(50) NULL,
		  field9      TEXT NULL,

		  PRIMARY KEY (id)
		) ENGINE=InnoDB AUTO_INCREMENT=1 DEFAULT CHARSET=utf8;


		CREATE TABLE gobatis_testb (
		  id int(11) NOT NULL AUTO_INCREMENT,
		  field0      boolean NOT NULL,
		  field1      int NOT NULL,
		  field2      int NOT NULL,
		  field3      float NOT NULL,
		  field4      float NOT NULL,
		  field5      varchar(50) NOT NULL,
		  field6      datetime NOT NULL,
		  field7      varchar(50) NOT NULL,
		  field8      varchar(50) NOT NULL,
		  field9      TEXT NULL,
		  PRIMARY KEY (id)
		) ENGINE=InnoDB AUTO_INCREMENT=1 DEFAULT CHARSET=utf8;


		CREATE TABLE gobatis_testc (
		  id         int(11) NOT NULL AUTO_INCREMENT,
		  field0     varchar(500),
		  PRIMARY KEY (id)
		) ENGINE=InnoDB AUTO_INCREMENT=1 DEFAULT CHARSET=utf8;


		CREATE TABLE gobatis_teste1 (
		  id         int(11) NOT NULL AUTO_INCREMENT,
		  field0     varchar(500),
		  PRIMARY KEY (id)
		) ENGINE=InnoDB AUTO_INCREMENT=1 DEFAULT CHARSET=utf8;


		CREATE TABLE gobatis_teste2 (
		  id         int(11) NOT NULL AUTO_INCREMENT,
		  field0     varchar(500) NOT NULL,
		  PRIMARY KEY (id)
		) ENGINE=InnoDB AUTO_INCREMENT=1 DEFAULT CHARSET=utf8;

		CREATE TABLE gobatis_testf1 (
		  id         int(11) NOT NULL AUTO_INCREMENT,
		  field0     varchar(500),
		  PRIMARY KEY (id)
		) ENGINE=InnoDB AUTO_INCREMENT=1 DEFAULT CHARSET=utf8;

		CREATE TABLE gobatis_testf2 (
		  id         int(11) NOT NULL AUTO_INCREMENT,
		  field0     varchar(500) NOT NULL,
		  PRIMARY KEY (id)
		) ENGINE=InnoDB AUTO_INCREMENT=1 DEFAULT CHARSET=utf8;


		CREATE TABLE gobatis_convert1 (
		  id         int(11) NOT NULL AUTO_INCREMENT,
		  field0     int,
		  PRIMARY KEY (id)
		) ENGINE=InnoDB AUTO_INCREMENT=1 DEFAULT CHARSET=utf8;

		CREATE TABLE gobatis_convert2 (
		  id         int(11) NOT NULL AUTO_INCREMENT,
		  field0     varchar(500),
		  PRIMARY KEY (id)
		) ENGINE=InnoDB AUTO_INCREMENT=1 DEFAULT CHARSET=utf8;

		

		DROP TABLE IF EXISTS computers;

		DROP TABLE IF EXISTS keyboards;
		CREATE TABLE keyboards ( id int(11) NOT NULL  auto_increment PRIMARY KEY, description VARCHAR(56)) ENGINE=InnoDB DEFAULT CHARSET=utf8;

		DROP TABLE IF EXISTS motherboards;
		CREATE TABLE motherboards ( id int(11) NOT NULL  auto_increment PRIMARY KEY, description VARCHAR(56)) ENGINE=InnoDB DEFAULT CHARSET=utf8;

		DROP TABLE IF EXISTS mouses;
		CREATE TABLE mouses (
		  id          int(11) NOT NULL AUTO_INCREMENT,
		  field1      boolean,
		  field2      int,
		  field3      float,
		  field4      varchar(50),
		  field5      varchar(50),
		  field6      varchar(50),
		  PRIMARY KEY (id)
		);

		CREATE TABLE computers ( id int(11) NOT NULL  auto_increment PRIMARY KEY, description VARCHAR(56), mother_id int(11), key_id int(11), mouse_id int(11)) ENGINE=InnoDB DEFAULT CHARSET=utf8;
	`

	// DROP TABLE IF EXISTS people;
	// CREATE TABLE people (id  int(11) NOT NULL auto_increment PRIMARY KEY, name VARCHAR(56) NOT NULL, last_name VARCHAR(56), dob DATE, graduation_date DATE, created_at DATETIME, updated_at DATETIME) ENGINE=InnoDB DEFAULT CHARSET=utf8;

	// DROP TABLE IF EXISTS accounts;
	// CREATE TABLE accounts (id  int(11) NOT NULL auto_increment PRIMARY KEY, account VARCHAR(56), description VARCHAR(56), amount DECIMAL(10,2), total DECIMAL(10,2)) ENGINE=InnoDB DEFAULT CHARSET=utf8;

	// DROP TABLE IF EXISTS temperatures;
	// CREATE TABLE temperatures (id  int(11) NOT NULL  auto_increment PRIMARY KEY, temp SMALLINT) ENGINE=InnoDB DEFAULT CHARSET=utf8;

	// DROP TABLE IF EXISTS shard1_temperatures;
	// CREATE TABLE shard1_temperatures (id  int(11) NOT NULL  auto_increment PRIMARY KEY, temp SMALLINT) ENGINE=InnoDB DEFAULT CHARSET=utf8;

	// DROP TABLE IF EXISTS shard2_temperatures;
	// CREATE TABLE shard2_temperatures (id  int(11) NOT NULL  auto_increment PRIMARY KEY, temp SMALLINT) ENGINE=InnoDB DEFAULT CHARSET=utf8;

	// DROP TABLE IF EXISTS salaries;
	// CREATE TABLE salaries (id  int(11) NOT NULL  auto_increment PRIMARY KEY, salary DECIMAL(7, 2)) ENGINE=InnoDB DEFAULT CHARSET=utf8;

	// DROP TABLE IF EXISTS users;
	// CREATE TABLE users (id  int(11) NOT NULL  auto_increment PRIMARY KEY, first_name VARCHAR(56), last_name VARCHAR(56), email VARCHAR(56)) ENGINE=InnoDB DEFAULT CHARSET=utf8;

	// DROP TABLE IF EXISTS shard1_users;
	// CREATE TABLE shard1_users (id  int(11) NOT NULL  auto_increment PRIMARY KEY, first_name VARCHAR(56), last_name VARCHAR(56), email VARCHAR(56)) ENGINE=InnoDB DEFAULT CHARSET=utf8;

	// DROP TABLE IF EXISTS addresses;
	// CREATE TABLE addresses (id  int(11) NOT NULL  auto_increment PRIMARY KEY, address1 VARCHAR(56), address2 VARCHAR(56), city VARCHAR(56), state VARCHAR(56), zip VARCHAR(56), user_id int(11)) ENGINE=InnoDB DEFAULT CHARSET=utf8;

	// DROP TABLE IF EXISTS shard1_addresses;
	// CREATE TABLE shard1_addresses (id  int(11) NOT NULL  auto_increment PRIMARY KEY, address1 VARCHAR(56), address2 VARCHAR(56), city VARCHAR(56), state VARCHAR(56), zip VARCHAR(56), user_id int(11)) ENGINE=InnoDB DEFAULT CHARSET=utf8;

	// DROP TABLE IF EXISTS rooms;
	// CREATE TABLE rooms (id  int(11) NOT NULL  auto_increment PRIMARY KEY, name VARCHAR(56), address_id int(11)) ENGINE=InnoDB DEFAULT CHARSET=utf8;

	// DROP TABLE IF EXISTS legacy_universities;
	// CREATE TABLE legacy_universities (id  int(11) NOT NULL  auto_increment PRIMARY KEY, univ_name VARCHAR(56), address1 VARCHAR(56), address2 VARCHAR(56), city VARCHAR(56), state VARCHAR(56), zip VARCHAR(56)) ENGINE=InnoDB DEFAULT CHARSET=utf8;

	// DROP TABLE IF EXISTS libraries;
	// CREATE TABLE libraries (id  int(11) NOT NULL  auto_increment PRIMARY KEY, address VARCHAR(56), city VARCHAR(56), state VARCHAR(56)) ENGINE=InnoDB DEFAULT CHARSET=utf8;

	// DROP TABLE IF EXISTS books;
	// CREATE TABLE books (id  int(11) NOT NULL  auto_increment PRIMARY KEY, title VARCHAR(56), author VARCHAR(56), isbn VARCHAR(56), lib_id int(11)) ENGINE=InnoDB DEFAULT CHARSET=utf8;

	// DROP TABLE IF EXISTS readers;
	// CREATE TABLE readers (id  int(11) NOT NULL  auto_increment PRIMARY KEY, first_name VARCHAR(56), last_name VARCHAR(56), book_id int(11)) ENGINE=InnoDB DEFAULT CHARSET=utf8;

	// DROP TABLE IF EXISTS animals;
	// CREATE TABLE animals (animal_id  int(11) NOT NULL  auto_increment PRIMARY KEY, animal_name VARCHAR(56)) ENGINE=InnoDB DEFAULT CHARSET=utf8;

	// DROP TABLE IF EXISTS patients;
	// CREATE TABLE patients (id  int(11) NOT NULL  auto_increment PRIMARY KEY, first_name VARCHAR(56), last_name VARCHAR(56)) ENGINE=InnoDB DEFAULT CHARSET=utf8;

	// DROP TABLE IF EXISTS shard1_patients;
	// CREATE TABLE shard1_patients (id  int(11) NOT NULL  auto_increment PRIMARY KEY, first_name VARCHAR(56), last_name VARCHAR(56)) ENGINE=InnoDB DEFAULT CHARSET=utf8;

	// DROP TABLE IF EXISTS prescriptions;
	// CREATE TABLE prescriptions (id  int(11) NOT NULL  auto_increment PRIMARY KEY, name VARCHAR(56), patient_id int(11)) ENGINE=InnoDB DEFAULT CHARSET=utf8;

	// DROP TABLE IF EXISTS doctors;
	// CREATE TABLE doctors (id  int(11) NOT NULL  auto_increment PRIMARY KEY, first_name VARCHAR(56), last_name VARCHAR(56), discipline varchar(56)) ENGINE=InnoDB DEFAULT CHARSET=utf8;

	// DROP TABLE IF EXISTS shard1_doctors;
	// CREATE TABLE shard1_doctors (id  int(11) NOT NULL  auto_increment PRIMARY KEY, first_name VARCHAR(56), last_name VARCHAR(56), discipline varchar(56)) ENGINE=InnoDB DEFAULT CHARSET=utf8;

	// DROP TABLE IF EXISTS doctors_patients;
	// CREATE TABLE doctors_patients (id  int(11) NOT NULL  auto_increment PRIMARY KEY, doctor_id int(11), patient_id int(11)) ENGINE=InnoDB DEFAULT CHARSET=utf8;

	// DROP TABLE IF EXISTS students;
	// CREATE TABLE students (id int(11) NOT NULL  auto_increment PRIMARY KEY, first_name VARCHAR(56), last_name VARCHAR(56), dob DATE, enrollment_date DATETIME) ENGINE=InnoDB DEFAULT CHARSET=utf8;

	// DROP TABLE IF EXISTS courses;
	// CREATE TABLE courses (id  int(11) NOT NULL  auto_increment PRIMARY KEY, course_name VARCHAR(56)) ENGINE=InnoDB DEFAULT CHARSET=utf8;

	// DROP TABLE IF EXISTS registrations;
	// CREATE TABLE registrations (id  int(11) NOT NULL  auto_increment PRIMARY KEY, astudent_id int(11), acourse_id int(11)) ENGINE=InnoDB DEFAULT CHARSET=utf8;

	// DROP TABLE IF EXISTS items;
	// CREATE TABLE items (id  int(11) NOT NULL  auto_increment PRIMARY KEY, item_number int(11), item_description VARCHAR(56), lock_version int(11)) ENGINE=InnoDB DEFAULT CHARSET=utf8;

	// DROP TABLE IF EXISTS articles;
	// CREATE TABLE articles (id  int(11) NOT NULL  auto_increment PRIMARY KEY, title VARCHAR(56), content TEXT) ENGINE=InnoDB DEFAULT CHARSET=utf8;

	// DROP TABLE IF EXISTS shard1_articles;
	// CREATE TABLE shard1_articles (id  int(11) NOT NULL  auto_increment PRIMARY KEY, title VARCHAR(56), content TEXT) ENGINE=InnoDB DEFAULT CHARSET=utf8;

	// DROP TABLE IF EXISTS posts;
	// CREATE TABLE posts (id  int(11) NOT NULL  auto_increment PRIMARY KEY, title VARCHAR(56), post TEXT) ENGINE=InnoDB DEFAULT CHARSET=utf8;

	// DROP TABLE IF EXISTS shard1_posts;
	// CREATE TABLE shard1_posts (id  int(11) NOT NULL  auto_increment PRIMARY KEY, title VARCHAR(56), post TEXT) ENGINE=InnoDB DEFAULT CHARSET=utf8;

	// DROP TABLE IF EXISTS comments;
	// CREATE TABLE comments (id  int(11) NOT NULL  auto_increment PRIMARY KEY, author VARCHAR(56), content TEXT, parent_id int(11), parent_type VARCHAR(256)) ENGINE=InnoDB DEFAULT CHARSET=utf8;

	// DROP TABLE IF EXISTS shard1_comments;
	// CREATE TABLE shard1_comments (id  int(11) NOT NULL  auto_increment PRIMARY KEY, author VARCHAR(56), content TEXT, parent_id int(11), parent_type VARCHAR(256)) ENGINE=InnoDB DEFAULT CHARSET=utf8;

	// DROP TABLE IF EXISTS tags;
	// CREATE TABLE tags (id  int(11) NOT NULL  auto_increment PRIMARY KEY, content TEXT, parent_id int(11), parent_type VARCHAR(256)) ENGINE=InnoDB DEFAULT CHARSET=utf8;

	// DROP TABLE IF EXISTS fruits;
	// CREATE TABLE fruits (id  int(11) NOT NULL  auto_increment PRIMARY KEY, fruit_name VARCHAR(56), category VARCHAR(56), created_at DATETIME, updated_at DATETIME) ENGINE=InnoDB DEFAULT CHARSET=utf8;

	// DROP TABLE IF EXISTS vegetables;
	// CREATE TABLE vegetables (id  int(11) NOT NULL  auto_increment PRIMARY KEY, vegetable_name VARCHAR(56), category VARCHAR(56), created_at DATETIME, updated_at DATETIME) ENGINE=InnoDB DEFAULT CHARSET=utf8;

	// DROP TABLE IF EXISTS plants;
	// CREATE TABLE plants (id  int(11) NOT NULL  auto_increment PRIMARY KEY, plant_name VARCHAR(56), category VARCHAR(56), created_at DATETIME, updated_at DATETIME) ENGINE=InnoDB DEFAULT CHARSET=utf8;

	// DROP TABLE IF EXISTS pages;
	// CREATE TABLE pages ( id int(11) NOT NULL  auto_increment PRIMARY KEY, description VARCHAR(56), word_count int(11) ) ENGINE=InnoDB DEFAULT CHARSET=utf8;

	// DROP TABLE IF EXISTS watermelons;
	// CREATE TABLE watermelons ( id int(11) NOT NULL  auto_increment PRIMARY KEY, melon_type VARCHAR(56), record_version INT, created_at DATETIME, updated_at DATETIME) ENGINE=InnoDB DEFAULT CHARSET=utf8;

	// DROP TABLE IF EXISTS schools;
	// CREATE TABLE schools ( id int(11) NOT NULL  auto_increment PRIMARY KEY, school_name VARCHAR(56), school_type VARCHAR(56), email VARCHAR(56), created_at DATETIME, updated_at DATETIME) ENGINE=InnoDB DEFAULT CHARSET=utf8;

	// DROP TABLE IF EXISTS programmers;
	// CREATE TABLE programmers ( id int(11) NOT NULL  auto_increment PRIMARY KEY, first_name VARCHAR(56), last_name VARCHAR(56), email VARCHAR(56), created_at DATETIME, updated_at DATETIME) ENGINE=InnoDB DEFAULT CHARSET=utf8;

	// DROP TABLE IF EXISTS projects;
	// CREATE TABLE projects ( id int(11) NOT NULL  auto_increment PRIMARY KEY, project_name VARCHAR(56), created_at DATETIME, updated_at DATETIME) ENGINE=InnoDB DEFAULT CHARSET=utf8;

	// DROP TABLE IF EXISTS programmers_projects;
	// CREATE TABLE programmers_projects ( id int(11) NOT NULL  auto_increment PRIMARY KEY, duration_weeks int(3), project_id int(11), programmer_id int(11), created_at DATETIME, updated_at DATETIME) ENGINE=InnoDB DEFAULT CHARSET=utf8;

	// DROP TABLE IF EXISTS ingredients;
	// CREATE TABLE ingredients (ingredient_id  int(11) NOT NULL  auto_increment PRIMARY KEY, ingredient_name VARCHAR(56)) ENGINE=InnoDB DEFAULT CHARSET=utf8;

	// DROP TABLE IF EXISTS recipes;
	// CREATE TABLE recipes (recipe_id  int(11) NOT NULL  auto_increment PRIMARY KEY, recipe_name VARCHAR(56)) ENGINE=InnoDB DEFAULT CHARSET=utf8;

	// DROP TABLE IF EXISTS ingredients_recipes;
	// CREATE TABLE ingredients_recipes (the_id  int(11) NOT NULL  auto_increment PRIMARY KEY, recipe_id int(11), ingredient_id int(11)) ENGINE=InnoDB DEFAULT CHARSET=utf8;

	// DROP TABLE IF EXISTS vehicles;
	// CREATE TABLE vehicles (id  int(11) NOT NULL  auto_increment PRIMARY KEY, name VARCHAR(56)) ENGINE=InnoDB DEFAULT CHARSET=utf8;

	// DROP TABLE IF EXISTS mammals;
	// CREATE TABLE mammals (id  int(11) NOT NULL  auto_increment PRIMARY KEY, name VARCHAR(56)) ENGINE=InnoDB DEFAULT CHARSET=utf8;

	// DROP TABLE IF EXISTS classifications;
	// CREATE TABLE classifications (id  int(11) NOT NULL  auto_increment PRIMARY KEY, name VARCHAR(56), parent_id int(11), parent_type VARCHAR(56)) ENGINE=InnoDB DEFAULT CHARSET=utf8;

	// DROP TABLE IF EXISTS sub_classifications;
	// CREATE TABLE sub_classifications (id  int(11) NOT NULL  auto_increment PRIMARY KEY, name VARCHAR(56), classification_id int(11)) ENGINE=InnoDB DEFAULT CHARSET=utf8;

	// DROP TABLE IF EXISTS content_groups;
	// create table content_groups ( id  int(11) NOT NULL  auto_increment PRIMARY KEY, group_name INT(11) ) ENGINE=InnoDB DEFAULT CHARSET=utf8;

	// DROP TABLE IF EXISTS cakes;
	// CREATE TABLE cakes (id int(11) NOT NULL auto_increment PRIMARY KEY, name VARCHAR(56) NOT NULL) ENGINE=InnoDB DEFAULT CHARSET=utf8;

	// DROP TABLE IF EXISTS swords;
	// CREATE TABLE swords (id int(11) NOT NULL auto_increment PRIMARY KEY, name VARCHAR(56) NOT NULL) ENGINE=InnoDB DEFAULT CHARSET=utf8;

	// DROP TABLE IF EXISTS meals;
	// CREATE TABLE meals (id int(11) NOT NULL auto_increment PRIMARY KEY, name VARCHAR(56) NOT NULL) ENGINE=InnoDB DEFAULT CHARSET=utf8;

	// DROP TABLE IF EXISTS Member;
	// CREATE TABLE Member (id int(11) NOT NULL auto_increment PRIMARY KEY, name VARCHAR(56) NOT NULL) ENGINE=InnoDB DEFAULT CHARSET=utf8;

	// DROP TABLE IF EXISTS nodes;
	// CREATE TABLE nodes (id int(11) NOT NULL auto_increment PRIMARY KEY, name VARCHAR(56) NOT NULL, parent_id int(11)) ENGINE=InnoDB DEFAULT CHARSET=utf8;

	// DROP TABLE IF EXISTS images;
	// CREATE TABLE images (id int(11) NOT NULL auto_increment PRIMARY KEY, name VARCHAR(56) NOT NULL, content BLOB) ENGINE=InnoDB DEFAULT CHARSET=utf8;

	// DROP TABLE IF EXISTS apples;
	// CREATE TABLE apples (id int(11) NOT NULL PRIMARY KEY, apple_type VARCHAR(56) NOT NULL ) ENGINE=InnoDB DEFAULT CHARSET=utf8;

	// DROP TABLE IF EXISTS alarms;
	// CREATE TABLE alarms (id int(11) NOT NULL auto_increment PRIMARY KEY, alarm_time TIME NOT NULL) ENGINE=InnoDB DEFAULT CHARSET=utf8;

	// DROP TABLE IF EXISTS developers;
	// CREATE TABLE developers (first_name VARCHAR(56) NOT NULL, last_name VARCHAR(56) NOT NULL, email VARCHAR(56) NOT NULL, address VARCHAR(56), CONSTRAINT developers_uq UNIQUE (first_name, last_name, email)) ENGINE=InnoDB DEFAULT CHARSET=utf8;

	// # interval is a reserved word in MySQL
	// DROP TABLE IF EXISTS `interval`;
	// CREATE TABLE `interval` (id int(11) NOT NULL auto_increment PRIMARY KEY, `begin` INT, `end` INT);

	// DROP TABLE IF EXISTS boxes;
	// CREATE TABLE boxes (id  int(11) NOT NULL auto_increment PRIMARY KEY, color VARCHAR(56) NOT NULL, fruit_id int(11), created_at DATETIME, updated_at DATETIME) ENGINE=InnoDB DEFAULT CHARSET=utf8;

	// DROP TABLE IF EXISTS passengers;
	// CREATE TABLE passengers (id  int(11) NOT NULL auto_increment PRIMARY KEY, user_id INT(11) NOT NULL, vehicle VARCHAR(10), transportation_mode VARCHAR(10), created_at DATETIME, updated_at DATETIME) ENGINE=InnoDB DEFAULT CHARSET=utf8;

	// DROP TABLE IF EXISTS teams;
	// CREATE TABLE teams (team_id int(11) NOT NULL auto_increment PRIMARY KEY, name VARCHAR(56) NOT NULL);

	// DROP TABLE IF EXISTS players;
	// CREATE TABLE players (id int(11) NOT NULL auto_increment PRIMARY KEY, first_name VARCHAR(56) NOT NULL, last_name VARCHAR(56) NOT NULL, team_id int(11));

	// DROP TABLE IF EXISTS bands;
	// CREATE TABLE bands (band_id int(11) NOT NULL auto_increment PRIMARY KEY, name VARCHAR(56) NOT NULL);

	// DROP TABLE IF EXISTS genres;
	// CREATE TABLE genres (genre_id int(11) NOT NULL auto_increment PRIMARY KEY, name VARCHAR(56) NOT NULL);

	// DROP TABLE IF EXISTS musicians;
	// CREATE TABLE musicians (musician_id int(11) NOT NULL auto_increment PRIMARY KEY, first_name VARCHAR(56) NOT NULL, last_name VARCHAR(56) NOT NULL);

	// DROP TABLE IF EXISTS bands_genres;
	// CREATE TABLE bands_genres (the_id  int(11) NOT NULL  auto_increment PRIMARY KEY, band_id int(11), genre_id int(11));

	// DROP TABLE IF EXISTS bands_musicians;
	// CREATE TABLE bands_musicians (the_id  int(11) NOT NULL  auto_increment PRIMARY KEY, band_id int(11), musician_id int(11));

	// DROP TABLE IF EXISTS employees;
	// CREATE TABLE employees (
	//   id  int(11) NOT NULL auto_increment PRIMARY KEY,
	//   first_name VARCHAR(56) NOT NULL,
	//   last_name VARCHAR(56),
	//   position  VARCHAR(56),
	//   active int(2),
	//   department VARCHAR(56),
	//   created_at DATETIME,
	//   updated_at DATETIME) ENGINE=InnoDB DEFAULT CHARSET=utf8;

	MssqlScript = `
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
		  field7      datetimeoffset NULL,
		  fieldBool      bit NULL,
		  fieldBoolP     bit NULL,

		  create_time datetimeoffset
		);

		CREATE TABLE gobatis_usergroups (
		  id int IDENTITY NOT NULL PRIMARY KEY,
		  name varchar(45) DEFAULT NULL
		);

		CREATE TABLE gobatis_user_and_groups (
		  user_id int NOT NULL,
		  group_id int NOT NULL,
		  PRIMARY KEY (user_id,group_id)
		);


		IF OBJECT_ID('dbo.gobatis_settings', 'U') IS NOT NULL 
		DROP TABLE gobatis_settings;
		IF OBJECT_ID('dbo.gobatis_list', 'U') IS NOT NULL 
		DROP TABLE gobatis_list;
    
    CREATE TABLE gobatis_settings (
		  id int IDENTITY NOT NULL PRIMARY KEY,
		  name varchar(45) DEFAULT NULL,
		  value varchar(45) DEFAULT NULL,
		  UNIQUE(name)
		);
    
    CREATE TABLE gobatis_list (
		  id int IDENTITY NOT NULL PRIMARY KEY,
		  name varchar(45) DEFAULT NULL,
		  UNIQUE(name)
		);


		IF OBJECT_ID('dbo.gobatis_testa', 'U') IS NOT NULL 
		DROP TABLE gobatis_testa;
		IF OBJECT_ID('dbo.gobatis_testb', 'U') IS NOT NULL 
		DROP TABLE gobatis_testb;
		IF OBJECT_ID('dbo.gobatis_testc', 'U') IS NOT NULL 
		DROP TABLE gobatis_testc;
		IF OBJECT_ID('dbo.gobatis_teste1', 'U') IS NOT NULL 
		DROP TABLE gobatis_teste1;
		IF OBJECT_ID('dbo.gobatis_teste2', 'U') IS NOT NULL 
		DROP TABLE gobatis_teste2;
		IF OBJECT_ID('dbo.gobatis_testf1', 'U') IS NOT NULL 
		DROP TABLE gobatis_testf1;
		IF OBJECT_ID('dbo.gobatis_testf2', 'U') IS NOT NULL 
		DROP TABLE gobatis_testf2;

		CREATE TABLE gobatis_testa (
		  id int IDENTITY NOT NULL PRIMARY KEY,
		  field0      bit NULL,
		  field1      int NULL,
		  field2      int NULL,
		  field3      float NULL,
		  field4      float NULL,
		  field5      varchar(50) NULL,
		  field6      datetimeoffset NULL,
		  field7      varchar(50) NULL,
		  field8      varchar(50) NULL,
		  field9      TEXT NULL
		) ;


		CREATE TABLE gobatis_testb (
		  id int IDENTITY NOT NULL PRIMARY KEY,
		  field0      bit NOT NULL,
		  field1      int NOT NULL,
		  field2      int NOT NULL,
		  field3      float NOT NULL,
		  field4      float NOT NULL,
		  field5      varchar(50) NOT NULL,
		  field6      datetimeoffset NOT NULL,
		  field7      varchar(50) NOT NULL,
		  field8      varchar(50) NOT NULL,
		  field9      TEXT NULL
		) ;


		CREATE TABLE gobatis_testc (
		  id int IDENTITY NOT NULL PRIMARY KEY,
		  field0     varchar(500)
		) ;


		CREATE TABLE gobatis_teste1 (
		  id int IDENTITY NOT NULL PRIMARY KEY,
		  field0     varchar(500)
		) ;


		CREATE TABLE gobatis_teste2 (
		  id int IDENTITY NOT NULL PRIMARY KEY,
		  field0     varchar(500) NOT NULL
		) ;

		CREATE TABLE gobatis_testf1 (
		  id int IDENTITY NOT NULL PRIMARY KEY,
		  field0     varchar(500)
		) ;

		CREATE TABLE gobatis_testf2 (
		  id int IDENTITY NOT NULL PRIMARY KEY,
		  field0     varchar(500) NOT NULL
		);


		IF object_id('dbo.gobatis_convert1') IS NOT NULL
		BEGIN
		    DROP TABLE [dbo].[gobatis_convert1]
		END
		CREATE TABLE gobatis_convert1 (
		  id int IDENTITY NOT NULL PRIMARY KEY,
		  field0     int
		);

		IF object_id('dbo.gobatis_convert2') IS NOT NULL
		BEGIN
		    DROP TABLE [dbo].[gobatis_convert2]
		END
		CREATE TABLE gobatis_convert2 (
		  id int IDENTITY NOT NULL PRIMARY KEY,
		  field0     varchar(500)
		);


		IF object_id('dbo.computers') IS NOT NULL
		BEGIN
		    DROP TABLE [dbo].[computers]
		END
		CREATE TABLE computers ( id INT IDENTITY PRIMARY KEY, description VARCHAR(56), mother_id INT, key_id INT, mouse_id INT);

		IF object_id('dbo.keyboards') IS NOT NULL
		BEGIN
		    DROP TABLE [dbo].[keyboards]
		END
		CREATE TABLE keyboards ( id INT IDENTITY PRIMARY KEY, description VARCHAR(56));

		IF object_id('dbo.motherboards') IS NOT NULL
		BEGIN
		    DROP TABLE [dbo].[motherboards]
		END
		CREATE TABLE motherboards ( id INT IDENTITY PRIMARY KEY, description VARCHAR(56));

		IF object_id('dbo.mouses') IS NOT NULL
		BEGIN
		    DROP TABLE [dbo].[mouses]
		END
		CREATE TABLE mouses (
		  id int IDENTITY NOT NULL PRIMARY KEY,
		  field1      bit,
		  field2      int,
		  field3      float,
		  field4      varchar(50),
		  field5      varchar(50),
		  field6      varchar(50)
		);
	`

	// -- noinspection SqlNoDataSourceInspectionForFile
	// IF object_id('dbo.people') IS NOT NULL
	// BEGIN
	//     DROP TABLE [dbo].[people]
	// END
	// CREATE TABLE people (id INT IDENTITY PRIMARY KEY, name VARCHAR(56) NOT NULL, last_name VARCHAR(56), dob DATE, graduation_date DATE, created_at DATETIME2, updated_at DATETIME2);

	// IF object_id('dbo.accounts') IS NOT NULL
	// BEGIN
	//     DROP TABLE [dbo].[accounts]
	// END
	// CREATE TABLE accounts (id  INT IDENTITY PRIMARY KEY, account VARCHAR(56), description VARCHAR(56), amount DECIMAL(10,2), total DECIMAL(10,2));

	// IF object_id('dbo.temperatures') IS NOT NULL
	// BEGIN
	//     DROP TABLE [dbo].[temperatures]
	// END
	// CREATE TABLE temperatures (id  INT IDENTITY PRIMARY KEY, temp SMALLINT);

	// IF object_id('dbo.shard1_temperatures') IS NOT NULL
	//     BEGIN
	//         DROP TABLE [dbo].[shard1_temperatures]
	//     END
	// CREATE TABLE shard1_temperatures (id  INT IDENTITY PRIMARY KEY, temp SMALLINT);

	// IF object_id('dbo.shard2_temperatures') IS NOT NULL
	//     BEGIN
	//         DROP TABLE [dbo].[shard2_temperatures]
	//     END
	// CREATE TABLE shard2_temperatures (id  INT IDENTITY PRIMARY KEY, temp SMALLINT);

	// IF object_id('dbo.salaries') IS NOT NULL
	// BEGIN
	//     DROP TABLE [dbo].[salaries]
	// END
	// CREATE TABLE salaries (id  INT IDENTITY PRIMARY KEY, salary DECIMAL(7, 2));

	// IF object_id('dbo.users') IS NOT NULL
	// BEGIN
	//     DROP TABLE [dbo].[users]
	// END
	// CREATE TABLE users (id  INT IDENTITY PRIMARY KEY, first_name VARCHAR(56), last_name VARCHAR(56), email VARCHAR(56));

	// IF object_id('dbo.shard1_users') IS NOT NULL
	//     BEGIN
	//         DROP TABLE [dbo].[shard1_users]
	//     END
	// CREATE TABLE shard1_users (id  INT IDENTITY PRIMARY KEY, first_name VARCHAR(56), last_name VARCHAR(56), email VARCHAR(56));

	// IF object_id('dbo.addresses') IS NOT NULL
	// BEGIN
	//     DROP TABLE [dbo].[addresses]
	// END
	// CREATE TABLE addresses (id  INT IDENTITY PRIMARY KEY, address1 VARCHAR(56), address2 VARCHAR(56), city VARCHAR(56), state VARCHAR(56), zip VARCHAR(56), user_id INT);

	// IF object_id('dbo.shard1_addresses') IS NOT NULL
	//     BEGIN
	//         DROP TABLE [dbo].[shard1_addresses]
	//     END
	// CREATE TABLE shard1_addresses (id  INT IDENTITY PRIMARY KEY, address1 VARCHAR(56), address2 VARCHAR(56), city VARCHAR(56), state VARCHAR(56), zip VARCHAR(56), user_id INT);

	// IF object_id('dbo.rooms') IS NOT NULL
	// BEGIN
	//     DROP TABLE [dbo].[rooms]
	// END
	// CREATE TABLE rooms (id  INT IDENTITY PRIMARY KEY, name VARCHAR(56), address_id INT);

	// IF object_id('dbo.legacy_universities') IS NOT NULL
	// BEGIN
	//     DROP TABLE [dbo].[legacy_universities]
	// END
	// CREATE TABLE legacy_universities (id  INT IDENTITY PRIMARY KEY, univ_name VARCHAR(56), address1 VARCHAR(56), address2 VARCHAR(56), city VARCHAR(56), state VARCHAR(56), zip VARCHAR(56));

	// IF object_id('dbo.libraries') IS NOT NULL
	// BEGIN
	//     DROP TABLE [dbo].[libraries]
	// END
	// CREATE TABLE libraries (id  INT IDENTITY PRIMARY KEY, address VARCHAR(56), city VARCHAR(56), state VARCHAR(56));

	// IF object_id('dbo.books') IS NOT NULL
	// BEGIN
	//     DROP TABLE [dbo].[books]
	// END
	// CREATE TABLE books (id  INT IDENTITY PRIMARY KEY, title VARCHAR(56), author VARCHAR(56), isbn VARCHAR(56), lib_id INT);

	// IF object_id('dbo.readers') IS NOT NULL
	// BEGIN
	//     DROP TABLE [dbo].[readers]
	// END
	// CREATE TABLE readers (id  INT IDENTITY PRIMARY KEY, first_name VARCHAR(56), last_name VARCHAR(56), book_id INT);

	// IF object_id('dbo.animals') IS NOT NULL
	// BEGIN
	//     DROP TABLE [dbo].[animals]
	// END
	// CREATE TABLE animals (animal_id  INT IDENTITY PRIMARY KEY, animal_name VARCHAR(56));

	// IF object_id('dbo.patients') IS NOT NULL
	// BEGIN
	//     DROP TABLE [dbo].[patients]
	// END
	// CREATE TABLE patients (id  INT IDENTITY PRIMARY KEY, first_name VARCHAR(56), last_name VARCHAR(56));

	// IF object_id('dbo.shard1_patients') IS NOT NULL
	//     BEGIN
	//         DROP TABLE [dbo].[shard1_patients]
	//     END
	// CREATE TABLE shard1_patients (id  INT IDENTITY PRIMARY KEY, first_name VARCHAR(56), last_name VARCHAR(56));

	// IF object_id('dbo.prescriptions') IS NOT NULL
	// BEGIN
	//     DROP TABLE [dbo].[prescriptions]
	// END
	// CREATE TABLE prescriptions (id  INT IDENTITY PRIMARY KEY, name VARCHAR(56), patient_id INT);

	// IF object_id('dbo.doctors') IS NOT NULL
	// BEGIN
	//     DROP TABLE [dbo].[doctors]
	// END
	// CREATE TABLE doctors (id  INT IDENTITY PRIMARY KEY, first_name VARCHAR(56), last_name VARCHAR(56), discipline varchar(56));

	// IF object_id('dbo.shard1_doctors') IS NOT NULL
	//     BEGIN
	//         DROP TABLE [dbo].[shard1_doctors]
	//     END
	// CREATE TABLE shard1_doctors (id  INT IDENTITY PRIMARY KEY, first_name VARCHAR(56), last_name VARCHAR(56), discipline varchar(56));

	// IF object_id('dbo.doctors_patients') IS NOT NULL
	// BEGIN
	//     DROP TABLE [dbo].[doctors_patients]
	// END
	// CREATE TABLE doctors_patients (id  INT IDENTITY PRIMARY KEY, doctor_id INT, patient_id INT);

	// IF object_id('dbo.students') IS NOT NULL
	// BEGIN
	//     DROP TABLE [dbo].[students]
	// END
	// CREATE TABLE students (id INT IDENTITY PRIMARY KEY, first_name VARCHAR(56), last_name VARCHAR(56), dob DATE, enrollment_date DATETIME2);

	// IF object_id('dbo.courses') IS NOT NULL
	// BEGIN
	//     DROP TABLE [dbo].[courses]
	// END
	// CREATE TABLE courses (id  INT IDENTITY PRIMARY KEY, course_name VARCHAR(56));

	// IF object_id('dbo.registrations') IS NOT NULL
	// BEGIN
	//     DROP TABLE [dbo].[registrations]
	// END
	// CREATE TABLE registrations (id  INT IDENTITY PRIMARY KEY, astudent_id INT, acourse_id INT);

	// IF object_id('dbo.items') IS NOT NULL
	// BEGIN
	//     DROP TABLE [dbo].[items]
	// END
	// CREATE TABLE items (id  INT IDENTITY PRIMARY KEY, item_number INT, item_description VARCHAR(56), lock_version INT);

	// IF object_id('dbo.articles') IS NOT NULL
	// BEGIN
	//     DROP TABLE [dbo].[articles]
	// END
	// CREATE TABLE articles (id  INT IDENTITY PRIMARY KEY, title VARCHAR(56), content TEXT);

	// IF object_id('dbo.shard1_articles') IS NOT NULL
	//     BEGIN
	//         DROP TABLE [dbo].[shard1_articles]
	//     END
	// CREATE TABLE shard1_articles (id  INT IDENTITY PRIMARY KEY, title VARCHAR(56), content TEXT);

	// IF object_id('dbo.posts') IS NOT NULL
	// BEGIN
	//     DROP TABLE [dbo].[posts]
	// END
	// CREATE TABLE posts (id  INT IDENTITY PRIMARY KEY, title VARCHAR(56), post TEXT);

	// IF object_id('dbo.shard1_posts') IS NOT NULL
	//     BEGIN
	//         DROP TABLE [dbo].[shard1_posts]
	//     END
	// CREATE TABLE shard1_posts (id  INT IDENTITY PRIMARY KEY, title VARCHAR(56), post TEXT);

	// IF object_id('dbo.comments') IS NOT NULL
	// BEGIN
	//     DROP TABLE [dbo].[comments]
	// END
	// CREATE TABLE comments (id  INT IDENTITY PRIMARY KEY, author VARCHAR(56), content TEXT, parent_id INT, parent_type VARCHAR(256));

	// IF object_id('dbo.shard1_comments') IS NOT NULL
	//     BEGIN
	//         DROP TABLE [dbo].[shard1_comments]
	//     END
	// CREATE TABLE shard1_comments (id  INT IDENTITY PRIMARY KEY, author VARCHAR(56), content TEXT, parent_id INT, parent_type VARCHAR(256));

	// IF object_id('dbo.tags') IS NOT NULL
	// BEGIN
	//     DROP TABLE [dbo].[tags]
	// END
	// CREATE TABLE tags (id  INT IDENTITY PRIMARY KEY, content TEXT, parent_id INT, parent_type VARCHAR(256));

	// IF object_id('dbo.fruits') IS NOT NULL
	// BEGIN
	//     DROP TABLE [dbo].[fruits]
	// END
	// CREATE TABLE fruits (id  INT IDENTITY PRIMARY KEY, fruit_name VARCHAR(56), category VARCHAR(56), created_at DATETIME2, updated_at DATETIME2);

	// IF object_id('dbo.vegetables') IS NOT NULL
	// BEGIN
	//     DROP TABLE [dbo].[vegetables]
	// END
	// CREATE TABLE vegetables (id  INT IDENTITY PRIMARY KEY, vegetable_name VARCHAR(56), category VARCHAR(56), created_at DATETIME2, updated_at DATETIME2);

	// IF object_id('dbo.plants') IS NOT NULL
	// BEGIN
	//     DROP TABLE [dbo].[plants]
	// END
	// CREATE TABLE plants (id  INT IDENTITY PRIMARY KEY, plant_name VARCHAR(56), category VARCHAR(56), created_at DATETIME2, updated_at DATETIME2);

	// IF object_id('dbo.pages') IS NOT NULL
	// BEGIN
	//     DROP TABLE [dbo].[pages]
	// END
	// CREATE TABLE pages ( id INT IDENTITY PRIMARY KEY, description VARCHAR(56), word_count INT);

	// IF object_id('dbo.watermelons') IS NOT NULL
	// BEGIN
	//     DROP TABLE [dbo].[watermelons]
	// END
	// CREATE TABLE watermelons ( id INT IDENTITY PRIMARY KEY, melon_type VARCHAR(56), record_version INT, created_at DATETIME2, updated_at DATETIME2);

	// IF object_id('dbo.schools') IS NOT NULL
	// BEGIN
	//     DROP TABLE [dbo].[schools]
	// END
	// CREATE TABLE schools ( id INT IDENTITY PRIMARY KEY, school_name VARCHAR(56), school_type VARCHAR(56), email VARCHAR(56), created_at DATETIME2, updated_at DATETIME2);

	// IF object_id('dbo.dual') IS NOT NULL
	// BEGIN
	//     DROP TABLE [dbo].[dual]
	// END
	// CREATE TABLE dual ( id INT IDENTITY PRIMARY KEY, next_val BIGINT);
	// INSERT INTO dual (next_val) VALUES (0);

	// IF object_id('dbo.programmers') IS NOT NULL
	// BEGIN
	//     DROP TABLE [dbo].[programmers]
	// END
	// CREATE TABLE programmers ( id INT IDENTITY PRIMARY KEY, first_name VARCHAR(56), last_name VARCHAR(56), email VARCHAR(56), created_at DATETIME2, updated_at DATETIME2);

	// IF object_id('dbo.projects') IS NOT NULL
	// BEGIN
	//     DROP TABLE [dbo].[projects]
	// END
	// CREATE TABLE projects ( id INT IDENTITY PRIMARY KEY, project_name VARCHAR(56), created_at DATETIME2, updated_at DATETIME2);

	// IF object_id('dbo.programmers_projects') IS NOT NULL
	// BEGIN
	//     DROP TABLE [dbo].[programmers_projects]
	// END
	// CREATE TABLE programmers_projects ( id INT IDENTITY PRIMARY KEY, duration_weeks INT, project_id INT, programmer_id INT, created_at DATETIME2, updated_at DATETIME2);

	// IF object_id('dbo.ingredients') IS NOT NULL
	// BEGIN
	//     DROP TABLE [dbo].[ingredients]
	// END
	// CREATE TABLE ingredients (ingredient_id  INT IDENTITY PRIMARY KEY, ingredient_name VARCHAR(56));

	// IF object_id('dbo.recipes') IS NOT NULL
	// BEGIN
	//     DROP TABLE [dbo].[recipes]
	// END
	// CREATE TABLE recipes (recipe_id  INT IDENTITY PRIMARY KEY, recipe_name VARCHAR(56));

	// IF object_id('dbo.ingredients_recipes') IS NOT NULL
	// BEGIN
	//     DROP TABLE [dbo].[ingredients_recipes]
	// END
	// CREATE TABLE ingredients_recipes (the_id  INT IDENTITY PRIMARY KEY, recipe_id INT, ingredient_id INT);

	// IF object_id('dbo.vehicles') IS NOT NULL
	// BEGIN
	//     DROP TABLE [dbo].[vehicles]
	// END
	// CREATE TABLE vehicles (id  INT IDENTITY PRIMARY KEY, name VARCHAR(56));

	// IF object_id('dbo.mammals') IS NOT NULL
	// BEGIN
	//     DROP TABLE [dbo].[mammals]
	// END
	// CREATE TABLE mammals (id  INT IDENTITY PRIMARY KEY, name VARCHAR(56));

	// IF object_id('dbo.classifications') IS NOT NULL
	// BEGIN
	//     DROP TABLE [dbo].[classifications]
	// END
	// CREATE TABLE classifications (id  INT IDENTITY PRIMARY KEY, name VARCHAR(56), parent_id INT, parent_type VARCHAR(56));

	// IF object_id('dbo.sub_classifications') IS NOT NULL
	// BEGIN
	//     DROP TABLE [dbo].[sub_classifications]
	// END
	// CREATE TABLE sub_classifications (id  INT IDENTITY PRIMARY KEY, name VARCHAR(56), classification_id INT);

	// IF object_id('dbo.content_groups') IS NOT NULL
	// BEGIN
	//     DROP TABLE [dbo].[content_groups]
	// END
	// create table content_groups ( id  INT IDENTITY PRIMARY KEY, group_name INT);

	// IF object_id('dbo.cakes') IS NOT NULL
	// BEGIN
	//     DROP TABLE [dbo].[cakes]
	// END
	// CREATE TABLE cakes (id INT IDENTITY PRIMARY KEY, name VARCHAR(56) NOT NULL);

	// IF object_id('dbo.swords') IS NOT NULL
	// BEGIN
	//     DROP TABLE [dbo].[swords]
	// END
	// CREATE TABLE swords (id INT IDENTITY PRIMARY KEY, name VARCHAR(56) NOT NULL);

	// IF object_id('dbo.meals') IS NOT NULL
	// BEGIN
	//     DROP TABLE [dbo].[meals]
	// END
	// CREATE TABLE meals (id INT IDENTITY PRIMARY KEY, name VARCHAR(56) NOT NULL);

	// IF object_id('dbo.Member') IS NOT NULL
	// BEGIN
	//     DROP TABLE [dbo].[Member]
	// END
	// CREATE TABLE Member (id INT IDENTITY PRIMARY KEY, name VARCHAR(56) NOT NULL);

	// IF object_id('dbo.nodes') IS NOT NULL
	// BEGIN
	//     DROP TABLE [dbo].[nodes]
	// END
	// CREATE TABLE nodes (id INT IDENTITY PRIMARY KEY, name VARCHAR(56) NOT NULL, parent_id INT);

	// IF object_id('dbo.images') IS NOT NULL
	// BEGIN
	//     DROP TABLE [dbo].[images]
	// END
	// CREATE TABLE images (id INT IDENTITY PRIMARY KEY, name VARCHAR(56) NOT NULL, content VARBINARY(MAX));

	// IF object_id('dbo.apples') IS NOT NULL
	// BEGIN
	//     DROP TABLE [dbo].[apples]
	// END
	// CREATE TABLE apples (id INT PRIMARY KEY, apple_type VARCHAR(56) NOT NULL);

	// IF object_id('dbo.alarms') IS NOT NULL
	// BEGIN
	//     DROP TABLE [dbo].[alarms]
	// END
	// CREATE TABLE alarms (id INT IDENTITY PRIMARY KEY, alarm_time TIME NOT NULL);

	// IF object_id('dbo.developers') IS NOT NULL
	// BEGIN
	//     DROP TABLE [dbo].[developers]
	// END
	// CREATE TABLE developers (first_name VARCHAR(56) NOT NULL, last_name VARCHAR(56) NOT NULL, email VARCHAR(56) NOT NULL, address VARCHAR(56), CONSTRAINT developers_uq UNIQUE (first_name, last_name, email));

	// IF object_id('dbo.boxes') IS NOT NULL
	// BEGIN
	//     DROP TABLE [dbo].[boxes]
	// END
	// CREATE TABLE boxes (id INT IDENTITY PRIMARY KEY, color VARCHAR(56) NOT NULL, fruit_id INT);

	// IF object_id('dbo.passengers') IS NOT NULL
	// BEGIN
	//     DROP TABLE [dbo].[passengers]
	// END
	// CREATE TABLE passengers (id INT IDENTITY PRIMARY KEY, user_id INT NOT NULL, vehicle VARCHAR(10), transportation_mode VARCHAR(10));

	// IF object_id('dbo.teams') IS NOT NULL
	//     BEGIN
	//         DROP TABLE [dbo].[teams]
	//     END
	// CREATE TABLE teams (team_id INT IDENTITY PRIMARY KEY, name VARCHAR(56) NOT NULL);

	// IF object_id('dbo.players') IS NOT NULL
	//     BEGIN
	//         DROP TABLE [dbo].[players]
	//     END
	// CREATE TABLE players (id INT IDENTITY PRIMARY KEY, first_name VARCHAR(56) NOT NULL, last_name VARCHAR(56) NOT NULL, team_id INT);

	// IF object_id('dbo.bands') IS NOT NULL
	// BEGIN
	//     DROP TABLE [dbo].[bands]
	// END
	// CREATE TABLE bands (band_id INT IDENTITY PRIMARY KEY, name VARCHAR(56) NOT NULL);

	// IF object_id('dbo.genres') IS NOT NULL
	//   BEGIN
	//     DROP TABLE [dbo].[genres]
	//   END
	// CREATE TABLE genres (genre_id INT IDENTITY PRIMARY KEY, name VARCHAR(56) NOT NULL);

	// IF object_id('dbo.musicians') IS NOT NULL
	//   BEGIN
	//     DROP TABLE [dbo].[musicians]
	//   END
	// CREATE TABLE musicians (musician_id INT IDENTITY PRIMARY KEY, first_name VARCHAR(56) NOT NULL, last_name VARCHAR(56) NOT NULL);

	// IF object_id('dbo.bands_genres') IS NOT NULL
	//   BEGIN
	//     DROP TABLE [dbo].[bands_genres]
	//   END
	// CREATE TABLE bands_genres (the_id INT IDENTITY PRIMARY KEY, band_id INT, genre_id INT);

	// IF object_id('dbo.bands_musicians') IS NOT NULL
	//   BEGIN
	//     DROP TABLE [dbo].bands_musicians
	//   END
	// CREATE TABLE bands_musicians (the_id INT IDENTITY PRIMARY KEY, band_id INT, musician_id INT);

	// IF object_id('dbo.employees') IS NOT NULL
	//     BEGIN
	//         DROP TABLE [dbo].[employees]
	//     END
	// CREATE TABLE employees ( id INT IDENTITY PRIMARY KEY, first_name VARCHAR(56), last_name VARCHAR(56), position VARCHAR(56), active INT, department VARCHAR(56), created_at DATETIME2, updated_at DATETIME2);

	PostgresqlScript = `
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
		  field6      timestamp with time zone NULL,
		  field7      timestamp with time zone NULL,
		  fieldBool      boolean NULL,
		  fieldBoolP     boolean NULL,
		  PRIMARY KEY (id)
		);
    
		CREATE TABLE IF NOT EXISTS gobatis_usergroups (
		  id bigserial NOT NULL,
		  name varchar(45) DEFAULT NULL,
		  PRIMARY KEY (id),
		  UNIQUE(name)
		);
    
		CREATE TABLE IF NOT EXISTS gobatis_user_and_groups (
		  user_id int NOT NULL,
		  group_id int NOT NULL,
		  PRIMARY KEY (user_id,group_id)
		);

		DROP TABLE IF EXISTS gobatis_settings; 
		DROP TABLE IF EXISTS gobatis_list;
    
    CREATE TABLE IF NOT EXISTS gobatis_settings (
		  id bigserial NOT NULL,
		  name varchar(45) DEFAULT NULL,
		  value varchar(45) DEFAULT NULL,
		  PRIMARY KEY (id),
		  UNIQUE(name)
		);
    
    CREATE TABLE IF NOT EXISTS gobatis_list (
		  id bigserial NOT NULL,
		  name varchar(45) DEFAULT NULL,
		  PRIMARY KEY (id),
		  UNIQUE(name)
		);


		DROP TABLE IF EXISTS gobatis_testa; 
		DROP TABLE IF EXISTS gobatis_testb;
		DROP TABLE IF EXISTS gobatis_testc;
		DROP TABLE IF EXISTS gobatis_teste1;
		DROP TABLE IF EXISTS gobatis_teste2;
		DROP TABLE IF EXISTS gobatis_testf1;
		DROP TABLE IF EXISTS gobatis_testf2;


		CREATE TABLE IF NOT EXISTS gobatis_testa (
		  id          bigserial NOT NULL,
		  field0      boolean NULL,
		  field1      int NULL,
		  field2      int NULL,
		  field3      float NULL,
		  field4      float NULL,
		  field5      varchar(50) NULL,
		  field6      timestamp with time zone NULL,
		  field7      varchar(50) NULL,
		  field8      varchar(50) NULL,
		  field9      TEXT NULL,
		  PRIMARY KEY (id)
		);


		CREATE TABLE IF NOT EXISTS gobatis_testb (
		  id          bigserial NOT NULL,
		  field0      boolean NOT NULL,
		  field1      int NOT NULL,
		  field2      int NOT NULL,
		  field3      float NOT NULL,
		  field4      float NOT NULL,
		  field5      varchar(50) NOT NULL,
		  field6      timestamp with time zone NOT NULL,
		  field7      varchar(50) NOT NULL,
		  field8      varchar(50) NOT NULL,
		  field9      TEXT NULL,
		  PRIMARY KEY (id)
		);


		CREATE TABLE IF NOT EXISTS gobatis_testc (
		  id          bigserial NOT NULL,
		  field0      varchar(500) NULL,
		  PRIMARY KEY (id)
		) ;


		CREATE TABLE IF NOT EXISTS gobatis_teste1 (
		  id          bigserial NOT NULL,
		  field0      integer[] NULL,
		  PRIMARY KEY (id)
		) ;


		CREATE TABLE IF NOT EXISTS gobatis_teste2 (
		  id          bigserial NOT NULL,
		  field0      integer[] NOT NULL,
		  PRIMARY KEY (id)
		) ;

		CREATE TABLE IF NOT EXISTS gobatis_testf1 (
		  id          bigserial NOT NULL,
		  field0      varchar(500) NULL,
		  PRIMARY KEY (id)
		) ;


		CREATE TABLE IF NOT EXISTS gobatis_testf2 (
		  id          bigserial NOT NULL,
		  field0      varchar(500) NOT NULL,
		  PRIMARY KEY (id)
		) ;

		DROP TABLE IF EXISTS gobatis_convert1;
		CREATE TABLE IF NOT EXISTS gobatis_convert1 (
		  id          bigserial NOT NULL,
		  field0     int,
		  PRIMARY KEY (id)
		);

		DROP TABLE IF EXISTS gobatis_convert2;
		CREATE TABLE IF NOT EXISTS gobatis_convert2 (
		  id          bigserial NOT NULL,
		  field0     varchar(500),
		  PRIMARY KEY (id)
		);

		DROP TABLE IF EXISTS computers;
		CREATE TABLE IF NOT EXISTS computers ( id serial PRIMARY KEY, description VARCHAR(56), mother_id INT, key_id INT, mouse_id INT);

		DROP TABLE IF EXISTS keyboards;
		CREATE TABLE IF NOT EXISTS keyboards ( id serial PRIMARY KEY, description VARCHAR(56));

		DROP TABLE IF EXISTS motherboards;
		CREATE TABLE IF NOT EXISTS motherboards ( id serial PRIMARY KEY, description VARCHAR(56));

		DROP TABLE IF EXISTS mouses;
		CREATE TABLE IF NOT EXISTS mouses (
		  id          bigserial NOT NULL,
		  field1      boolean,
		  field2      int,
		  field3      float,
		  field4      varchar(50),
		  field5      varchar(50),
		  field6      varchar(50),
		  PRIMARY KEY (id)
		);
	`

	// -- noinspection SqlDialectInspectionForFile
	// DROP TABLE IF EXISTS people;
	// CREATE TABLE people (id serial PRIMARY KEY, name VARCHAR(56) NOT NULL, last_name VARCHAR(56), dob DATE, graduation_date DATE, created_at TIMESTAMP, updated_at TIMESTAMP);

	// DROP TABLE IF EXISTS accounts;
	// CREATE TABLE accounts (id  serial PRIMARY KEY, account VARCHAR(56), description VARCHAR(56), amount DECIMAL(10,2), total DECIMAL(10,2));

	// DROP TABLE IF EXISTS temperatures;
	// CREATE TABLE temperatures (id  serial PRIMARY KEY, temp SMALLINT);

	// DROP TABLE IF EXISTS shard1_temperatures;
	// CREATE TABLE shard1_temperatures (id  serial PRIMARY KEY, temp SMALLINT);

	// DROP TABLE IF EXISTS shard2_temperatures;
	// CREATE TABLE shard2_temperatures (id  serial PRIMARY KEY, temp SMALLINT);

	// DROP TABLE IF EXISTS salaries;
	// CREATE TABLE salaries (id  serial PRIMARY KEY, salary DECIMAL(7, 2));

	// DROP TABLE IF EXISTS users;
	// CREATE TABLE users (id  serial PRIMARY KEY, first_name VARCHAR(56), last_name VARCHAR(56), email VARCHAR(56));

	// DROP TABLE IF EXISTS shard1_users;
	// CREATE TABLE shard1_users (id  serial PRIMARY KEY, first_name VARCHAR(56), last_name VARCHAR(56), email VARCHAR(56));

	// DROP TABLE IF EXISTS addresses;
	// CREATE TABLE addresses (id  serial PRIMARY KEY, address1 VARCHAR(56), address2 VARCHAR(56), city VARCHAR(56), state VARCHAR(56), zip VARCHAR(56), user_id INT);

	// DROP TABLE IF EXISTS shard1_addresses;
	// CREATE TABLE shard1_addresses (id  serial PRIMARY KEY, address1 VARCHAR(56), address2 VARCHAR(56), city VARCHAR(56), state VARCHAR(56), zip VARCHAR(56), user_id INT);

	// DROP TABLE IF EXISTS rooms;
	// CREATE TABLE rooms (id  serial PRIMARY KEY, name VARCHAR(56), address_id INT);

	// DROP TABLE IF EXISTS legacy_universities;
	// CREATE TABLE legacy_universities (id  serial PRIMARY KEY, univ_name VARCHAR(56), address1 VARCHAR(56), address2 VARCHAR(56), city VARCHAR(56), state VARCHAR(56), zip VARCHAR(56));

	// DROP TABLE IF EXISTS libraries;
	// CREATE TABLE libraries (id  serial PRIMARY KEY, address VARCHAR(56), city VARCHAR(56), state VARCHAR(56));

	// DROP TABLE IF EXISTS books;
	// CREATE TABLE books (id  serial PRIMARY KEY, title VARCHAR(56), author VARCHAR(56), isbn VARCHAR(56), lib_id INT);

	// DROP TABLE IF EXISTS readers;
	// CREATE TABLE readers (id  serial PRIMARY KEY, first_name VARCHAR(56), last_name VARCHAR(56), book_id INT);

	// DROP TABLE IF EXISTS animals;
	// CREATE TABLE animals (animal_id  serial PRIMARY KEY, animal_name VARCHAR(56));

	// DROP TABLE IF EXISTS patients;
	// CREATE TABLE patients (id  serial PRIMARY KEY, first_name VARCHAR(56), last_name VARCHAR(56));

	// DROP TABLE IF EXISTS shard1_patients;
	// CREATE TABLE shard1_patients (id  serial PRIMARY KEY, first_name VARCHAR(56), last_name VARCHAR(56));

	// DROP TABLE IF EXISTS prescriptions;
	// CREATE TABLE prescriptions (id  serial PRIMARY KEY, name VARCHAR(56), patient_id INT);

	// DROP TABLE IF EXISTS doctors;
	// CREATE TABLE doctors (id  serial PRIMARY KEY, first_name VARCHAR(56), last_name VARCHAR(56), discipline varchar(56));

	// DROP TABLE IF EXISTS shard1_doctors;
	// CREATE TABLE shard1_doctors (id  serial PRIMARY KEY, first_name VARCHAR(56), last_name VARCHAR(56), discipline varchar(56));

	// DROP TABLE IF EXISTS doctors_patients;
	// CREATE TABLE doctors_patients (id  serial PRIMARY KEY, doctor_id INT, patient_id INT);

	// DROP TABLE IF EXISTS students;
	// CREATE TABLE students (id serial PRIMARY KEY, first_name VARCHAR(56), last_name VARCHAR(56), dob DATE, enrollment_date TIMESTAMP);

	// DROP TABLE IF EXISTS courses;
	// CREATE TABLE courses (id  serial PRIMARY KEY, course_name VARCHAR(56));

	// DROP TABLE IF EXISTS registrations;
	// CREATE TABLE registrations (id  serial PRIMARY KEY, astudent_id INT, acourse_id INT);

	// DROP TABLE IF EXISTS items;
	// CREATE TABLE items (id  serial PRIMARY KEY, item_number INT, item_description VARCHAR(56), lock_version INT);

	// DROP TABLE IF EXISTS articles;
	// CREATE TABLE articles (id  serial PRIMARY KEY, title VARCHAR(56), content TEXT);

	// DROP TABLE IF EXISTS shard1_articles;
	// CREATE TABLE shard1_articles (id  serial PRIMARY KEY, title VARCHAR(56), content TEXT);

	// DROP TABLE IF EXISTS posts;
	// CREATE TABLE posts (id  serial PRIMARY KEY, title VARCHAR(56), post TEXT);

	// DROP TABLE IF EXISTS shard1_posts;
	// CREATE TABLE shard1_posts (id  serial PRIMARY KEY, title VARCHAR(56), post TEXT);

	// DROP TABLE IF EXISTS comments;
	// CREATE TABLE comments (id  serial PRIMARY KEY, author VARCHAR(56), content TEXT, parent_id INT, parent_type VARCHAR(256));

	// DROP TABLE IF EXISTS shard1_comments;
	// CREATE TABLE shard1_comments (id  serial PRIMARY KEY, author VARCHAR(56), content TEXT, parent_id INT, parent_type VARCHAR(256));

	// DROP TABLE IF EXISTS tags;
	// CREATE TABLE tags (id  serial PRIMARY KEY, content TEXT, parent_id INT, parent_type VARCHAR(256));

	// DROP TABLE IF EXISTS fruits;
	// CREATE TABLE fruits (id  serial PRIMARY KEY, fruit_name VARCHAR(56), category VARCHAR(56), created_at TIMESTAMP, updated_at TIMESTAMP);

	// DROP TABLE IF EXISTS vegetables;
	// CREATE TABLE vegetables (id  serial PRIMARY KEY, vegetable_name VARCHAR(56), category VARCHAR(56), created_at TIMESTAMP, updated_at TIMESTAMP);

	// DROP TABLE IF EXISTS plants;
	// CREATE TABLE plants (id  serial PRIMARY KEY, plant_name VARCHAR(56), category VARCHAR(56), created_at TIMESTAMP, updated_at TIMESTAMP);

	// DROP TABLE IF EXISTS pages;
	// CREATE TABLE pages ( id serial PRIMARY KEY, description VARCHAR(56), word_count INT);

	// DROP TABLE IF EXISTS watermelons;
	// CREATE TABLE watermelons ( id serial PRIMARY KEY, melon_type VARCHAR(56), record_version INT, created_at TIMESTAMP, updated_at TIMESTAMP);

	// DROP TABLE IF EXISTS schools;
	// CREATE TABLE schools ( id serial PRIMARY KEY, school_name VARCHAR(56), school_type VARCHAR(56), email VARCHAR(56), created_at TIMESTAMP, updated_at TIMESTAMP);

	// DROP TABLE IF EXISTS dual;
	// CREATE TABLE dual ( id serial PRIMARY KEY, next_val BIGINT);
	// INSERT INTO dual (next_val) VALUES (0);

	// DROP TABLE IF EXISTS programmers;
	// CREATE TABLE programmers ( id serial PRIMARY KEY, first_name VARCHAR(56), last_name VARCHAR(56), email VARCHAR(56), created_at TIMESTAMP, updated_at TIMESTAMP);

	// DROP TABLE IF EXISTS projects;
	// CREATE TABLE projects ( id serial PRIMARY KEY, project_name VARCHAR(56), created_at TIMESTAMP, updated_at TIMESTAMP);

	// DROP TABLE IF EXISTS programmers_projects;
	// CREATE TABLE programmers_projects ( id serial PRIMARY KEY, duration_weeks INT, project_id INT, programmer_id INT, created_at TIMESTAMP, updated_at TIMESTAMP);

	// DROP TABLE IF EXISTS ingredients;
	// CREATE TABLE ingredients (ingredient_id  serial PRIMARY KEY, ingredient_name VARCHAR(56));

	// DROP TABLE IF EXISTS recipes;
	// CREATE TABLE recipes (recipe_id  serial PRIMARY KEY, recipe_name VARCHAR(56));

	// DROP TABLE IF EXISTS ingredients_recipes;
	// CREATE TABLE ingredients_recipes (the_id  serial PRIMARY KEY, recipe_id INT, ingredient_id INT);

	// DROP TABLE IF EXISTS vehicles;
	// CREATE TABLE vehicles (id  serial PRIMARY KEY, name VARCHAR(56));

	// DROP TABLE IF EXISTS mammals;
	// CREATE TABLE mammals (id  serial PRIMARY KEY, name VARCHAR(56));

	// DROP TABLE IF EXISTS classifications;
	// CREATE TABLE classifications (id  serial PRIMARY KEY, name VARCHAR(56), parent_id INT, parent_type VARCHAR(56));

	// DROP TABLE IF EXISTS sub_classifications;
	// CREATE TABLE sub_classifications (id  serial PRIMARY KEY, name VARCHAR(56), classification_id INT);

	// DROP TABLE IF EXISTS content_groups;
	// create table content_groups ( id  serial PRIMARY KEY, group_name INT);

	// DROP TABLE IF EXISTS cakes;
	// CREATE TABLE cakes (id serial PRIMARY KEY, name VARCHAR(56) NOT NULL);

	// DROP TABLE IF EXISTS swords;
	// CREATE TABLE swords (id serial PRIMARY KEY, name VARCHAR(56) NOT NULL);

	// DROP TABLE IF EXISTS meals;
	// CREATE TABLE meals (id serial PRIMARY KEY, name VARCHAR(56) NOT NULL);

	// DROP TABLE IF EXISTS Member;
	// CREATE TABLE Member (id serial PRIMARY KEY, name VARCHAR(56) NOT NULL);

	// DROP TABLE IF EXISTS nodes;
	// CREATE TABLE nodes (id serial PRIMARY KEY, name VARCHAR(56) NOT NULL, parent_id INT);

	// DROP TABLE IF EXISTS images;
	// CREATE TABLE images (id serial PRIMARY KEY, name VARCHAR(56) NOT NULL, content BYTEA);

	// DROP TABLE IF EXISTS apples;
	// CREATE TABLE apples (id serial PRIMARY KEY, apple_type VARCHAR(56) NOT NULL );

	// DROP TABLE IF EXISTS alarms;
	// CREATE TABLE alarms (id serial PRIMARY KEY, alarm_time TIME NOT NULL);

	// DROP TABLE IF EXISTS developers;
	// CREATE TABLE developers (first_name VARCHAR(56) NOT NULL, last_name VARCHAR(56) NOT NULL, email VARCHAR(56) NOT NULL, address VARCHAR(56), CONSTRAINT developers_uq UNIQUE (first_name, last_name, email));

	// DROP TABLE IF EXISTS "Wild Animals";
	// CREATE TABLE "Wild Animals" (id serial PRIMARY KEY, "Name" VARCHAR(56) NOT NULL);

	// DROP TABLE IF EXISTS boxes;
	// CREATE TABLE boxes (id serial PRIMARY KEY, "color" VARCHAR(56) NOT NULL, fruit_id INT);

	// DROP TABLE IF EXISTS passengers;
	// CREATE TABLE passengers (id serial PRIMARY KEY, "vehicle" VARCHAR(56),transportation_mode VARCHAR(56), user_id INT);

	// DROP TABLE IF EXISTS teams;
	// CREATE TABLE teams (team_id serial PRIMARY KEY, name VARCHAR(56) NOT NULL);

	// DROP TABLE IF EXISTS players;
	// CREATE TABLE players (id serial PRIMARY KEY, first_name VARCHAR(56) NOT NULL, last_name VARCHAR(56) NOT NULL, team_id INT);

	// DROP TABLE IF EXISTS bands;
	// CREATE TABLE bands (band_id serial PRIMARY KEY, name VARCHAR(56) NOT NULL);

	// DROP TABLE IF EXISTS genres;
	// CREATE TABLE genres (genre_id serial PRIMARY KEY, name VARCHAR(56) NOT NULL);

	// DROP TABLE IF EXISTS musicians;
	// CREATE TABLE musicians (musician_id serial PRIMARY KEY, first_name VARCHAR(56) NOT NULL, last_name VARCHAR(56) NOT NULL);

	// DROP TABLE IF EXISTS bands_genres;
	// CREATE TABLE bands_genres (the_id  serial PRIMARY KEY, band_id INT, genre_id INT);

	// DROP TABLE IF EXISTS bands_musicians;
	// CREATE TABLE bands_musicians (the_id  serial PRIMARY KEY, band_id INT, musician_id INT);

	// DROP TABLE IF EXISTS employees;
	// CREATE TABLE employees ( id serial PRIMARY KEY, first_name VARCHAR(56) NOT NULL, last_name VARCHAR(56), position VARCHAR(56), active INT, department VARCHAR(56), created_at TIMESTAMP, updated_at TIMESTAMP);

	DMScript = `DROP TABLE  IF EXISTS gobatis_user_and_groups;
		DROP TABLE  IF EXISTS gobatis_users;
		DROP TABLE  IF EXISTS gobatis_usergroups;

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
		  field6      timestamp with time zone NULL,
		  field7      timestamp with time zone NULL,
		  fieldBool      bit NULL,
		  fieldBoolP     bit NULL,

		  create_time timestamp with time zone
		);

		CREATE TABLE gobatis_usergroups (
		  id int IDENTITY NOT NULL PRIMARY KEY,
		  name varchar(45) DEFAULT NULL
		);

		CREATE TABLE gobatis_user_and_groups (
		  user_id int NOT NULL,
		  group_id int NOT NULL,
		  PRIMARY KEY (user_id,group_id)
		);


		DROP TABLE IF EXISTS gobatis_settings;
		DROP TABLE IF EXISTS gobatis_list;
    
    CREATE TABLE gobatis_settings (
		  id int IDENTITY NOT NULL PRIMARY KEY,
		  name varchar(45) DEFAULT NULL,
		  value varchar(45) DEFAULT NULL,
		  UNIQUE(name)
		);
    
    CREATE TABLE gobatis_list (
		  id int IDENTITY NOT NULL PRIMARY KEY,
		  name varchar(45) DEFAULT NULL,
		  UNIQUE(name)
		);


		DROP TABLE IF EXISTS gobatis_testa;
		DROP TABLE IF EXISTS gobatis_testb;
		DROP TABLE IF EXISTS gobatis_testc;
		DROP TABLE IF EXISTS gobatis_teste1;
		DROP TABLE IF EXISTS gobatis_teste2;
		DROP TABLE IF EXISTS gobatis_testf1;
		DROP TABLE IF EXISTS gobatis_testf2;

		CREATE TABLE gobatis_testa (
		  id int IDENTITY NOT NULL PRIMARY KEY,
		  field0      bit NULL,
		  field1      int NULL,
		  field2      int NULL,
		  field3      float NULL,
		  field4      float NULL,
		  field5      varchar(50) NULL,
		  field6      timestamp with time zone NULL,
		  field7      varchar(50) NULL,
		  field8      varchar(50) NULL,
		  field9      TEXT NULL
		) ;


		CREATE TABLE gobatis_testb (
		  id int IDENTITY NOT NULL PRIMARY KEY,
		  field0      bit NOT NULL,
		  field1      int NOT NULL,
		  field2      int NOT NULL,
		  field3      float NOT NULL,
		  field4      float NOT NULL,
		  field5      varchar(50) NOT NULL,
		  field6      timestamp with time zone NOT NULL,
		  field7      varchar(50) NOT NULL,
		  field8      varchar(50) NOT NULL,
		  field9      TEXT NULL
		) ;


		CREATE TABLE gobatis_testc (
		  id int IDENTITY NOT NULL PRIMARY KEY,
		  field0     varchar(500)
		) ;


		CREATE TABLE gobatis_teste1 (
		  id int IDENTITY NOT NULL PRIMARY KEY,
		  field0     varchar(500)
		) ;


		CREATE TABLE gobatis_teste2 (
		  id int IDENTITY NOT NULL PRIMARY KEY,
		  field0     varchar(500) NOT NULL
		) ;

		CREATE TABLE gobatis_testf1 (
		  id int IDENTITY NOT NULL PRIMARY KEY,
		  field0     varchar(500)
		) ;

		CREATE TABLE gobatis_testf2 (
		  id int IDENTITY NOT NULL PRIMARY KEY,
		  field0     varchar(500) NOT NULL
		);


		DROP TABLE IF EXISTS gobatis_convert1;
		CREATE TABLE gobatis_convert1 (
		  id int IDENTITY NOT NULL PRIMARY KEY,
		  field0     int
		);

		DROP TABLE IF EXISTS gobatis_convert2;
		CREATE TABLE gobatis_convert2 (
		  id int IDENTITY NOT NULL PRIMARY KEY,
		  field0     varchar(500)
		);


		DROP TABLE IF EXISTS computers;
		CREATE TABLE computers ( id INT IDENTITY PRIMARY KEY, description VARCHAR(56), mother_id INT, key_id INT, mouse_id INT);

		DROP TABLE IF EXISTS keyboards;
		CREATE TABLE keyboards ( id INT IDENTITY PRIMARY KEY, description VARCHAR(56));

		DROP TABLE IF EXISTS motherboards;
		CREATE TABLE motherboards ( id INT IDENTITY PRIMARY KEY, description VARCHAR(56));

		DROP TABLE IF EXISTS mouses;
		CREATE TABLE mouses (
		  id int IDENTITY NOT NULL PRIMARY KEY,
		  field1      bit,
		  field2      int,
		  field3      float,
		  field4      varchar(50),
		  field5      varchar(50),
		  field6      varchar(50)
		);
	`

	Db2Script = `DROP TABLE  IF EXISTS gobatis_user_and_groups;
		DROP TABLE  IF EXISTS gobatis_users;
		DROP TABLE  IF EXISTS gobatis_usergroups;

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
		  field6      timestamp with time zone NULL,
		  field7      timestamp with time zone NULL,
		  fieldBool      bit NULL,
		  fieldBoolP     bit NULL,

		  create_time timestamp with time zone
		);

		CREATE TABLE gobatis_usergroups (
		  id int IDENTITY NOT NULL PRIMARY KEY,
		  name varchar(45) DEFAULT NULL
		);

		CREATE TABLE gobatis_user_and_groups (
		  user_id int NOT NULL,
		  group_id int NOT NULL,
		  PRIMARY KEY (user_id,group_id)
		);


		DROP TABLE IF EXISTS gobatis_settings;
		DROP TABLE IF EXISTS gobatis_list;
    
    CREATE TABLE gobatis_settings (
		  id int IDENTITY NOT NULL PRIMARY KEY,
		  name varchar(45) DEFAULT NULL,
		  value varchar(45) DEFAULT NULL,
		  UNIQUE(name)
		);
    
    CREATE TABLE gobatis_list (
		  id int IDENTITY NOT NULL PRIMARY KEY,
		  name varchar(45) DEFAULT NULL,
		  UNIQUE(name)
		);


		DROP TABLE IF EXISTS gobatis_testa;
		DROP TABLE IF EXISTS gobatis_testb;
		DROP TABLE IF EXISTS gobatis_testc;
		DROP TABLE IF EXISTS gobatis_teste1;
		DROP TABLE IF EXISTS gobatis_teste2;
		DROP TABLE IF EXISTS gobatis_testf1;
		DROP TABLE IF EXISTS gobatis_testf2;

		CREATE TABLE gobatis_testa (
		  id int IDENTITY NOT NULL PRIMARY KEY,
		  field0      bit NULL,
		  field1      int NULL,
		  field2      int NULL,
		  field3      float NULL,
		  field4      float NULL,
		  field5      varchar(50) NULL,
		  field6      timestamp with time zone NULL,
		  field7      varchar(50) NULL,
		  field8      varchar(50) NULL,
		  field9      TEXT NULL
		) ;


		CREATE TABLE gobatis_testb (
		  id int IDENTITY NOT NULL PRIMARY KEY,
		  field0      bit NOT NULL,
		  field1      int NOT NULL,
		  field2      int NOT NULL,
		  field3      float NOT NULL,
		  field4      float NOT NULL,
		  field5      varchar(50) NOT NULL,
		  field6      timestamp with time zone NOT NULL,
		  field7      varchar(50) NOT NULL,
		  field8      varchar(50) NOT NULL,
		  field9      TEXT NULL
		) ;


		CREATE TABLE gobatis_testc (
		  id int IDENTITY NOT NULL PRIMARY KEY,
		  field0     varchar(500)
		) ;


		CREATE TABLE gobatis_teste1 (
		  id int IDENTITY NOT NULL PRIMARY KEY,
		  field0     varchar(500)
		) ;


		CREATE TABLE gobatis_teste2 (
		  id int IDENTITY NOT NULL PRIMARY KEY,
		  field0     varchar(500) NOT NULL
		) ;

		CREATE TABLE gobatis_testf1 (
		  id int IDENTITY NOT NULL PRIMARY KEY,
		  field0     varchar(500)
		) ;

		CREATE TABLE gobatis_testf2 (
		  id int IDENTITY NOT NULL PRIMARY KEY,
		  field0     varchar(500) NOT NULL
		);


		DROP TABLE IF EXISTS gobatis_convert1;
		CREATE TABLE gobatis_convert1 (
		  id int IDENTITY NOT NULL PRIMARY KEY,
		  field0     int
		);

		DROP TABLE IF EXISTS gobatis_convert2;
		CREATE TABLE gobatis_convert2 (
		  id int IDENTITY NOT NULL PRIMARY KEY,
		  field0     varchar(500)
		);


		DROP TABLE IF EXISTS computers;
		CREATE TABLE computers ( id INT IDENTITY PRIMARY KEY, description VARCHAR(56), mother_id INT, key_id INT, mouse_id INT);

		DROP TABLE IF EXISTS keyboards;
		CREATE TABLE keyboards ( id INT IDENTITY PRIMARY KEY, description VARCHAR(56));

		DROP TABLE IF EXISTS motherboards;
		CREATE TABLE motherboards ( id INT IDENTITY PRIMARY KEY, description VARCHAR(56));

		DROP TABLE IF EXISTS mouses;
		CREATE TABLE mouses (
		  id int IDENTITY NOT NULL PRIMARY KEY,
		  field1      bit,
		  field2      int,
		  field3      float,
		  field4      varchar(50),
		  field5      varchar(50),
		  field6      varchar(50)
		);
	`

	OracleScript = `
		BEGIN EXECUTE IMMEDIATE 'DROP TABLE gobatis_user_and_groups'; EXCEPTION WHEN OTHERS THEN NULL; END;
		BEGIN EXECUTE IMMEDIATE 'DROP TABLE gobatis_users'; EXCEPTION WHEN OTHERS THEN NULL; END;
		BEGIN EXECUTE IMMEDIATE 'DROP TABLE gobatis_usergroups'; EXCEPTION WHEN OTHERS THEN NULL; END;

		CREATE TABLE gobatis_users (
		  id NUMBER GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
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
		  field6      TIMESTAMP(9) WITH LOCAL TIME ZONE NULL,
		  field7      TIMESTAMP(9) WITH LOCAL TIME ZONE NULL,
		  fieldBool   Char(1) NOT NULL,
		  fieldBoolP  Char(1) NOT NULL,

		  create_time TIMESTAMP(9) WITH LOCAL TIME ZONE
		);

		CREATE TABLE gobatis_usergroups (
		  id NUMBER GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
		  name varchar(45) DEFAULT NULL
		);

		CREATE TABLE gobatis_user_and_groups (
		  user_id int NOT NULL,
		  group_id int NOT NULL,
		  PRIMARY KEY (user_id,group_id)
		);

		BEGIN EXECUTE IMMEDIATE 'DROP TABLE gobatis_settings'; EXCEPTION WHEN OTHERS THEN NULL; END;
		BEGIN EXECUTE IMMEDIATE 'DROP TABLE gobatis_list'; EXCEPTION WHEN OTHERS THEN NULL; END;
		
    
    CREATE TABLE gobatis_settings (
		  id NUMBER GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
		  name varchar(45) DEFAULT NULL,
		  value varchar(45) DEFAULT NULL,
		  UNIQUE(name)
		);
    
    CREATE TABLE gobatis_list (
		  id NUMBER GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
		  name varchar(45) DEFAULT NULL,
		  UNIQUE(name)
		);

		BEGIN EXECUTE IMMEDIATE 'DROP TABLE gobatis_testa'; EXCEPTION WHEN OTHERS THEN NULL;END;
		BEGIN EXECUTE IMMEDIATE 'DROP TABLE gobatis_testb'; EXCEPTION WHEN OTHERS THEN NULL;END;
		BEGIN EXECUTE IMMEDIATE 'DROP TABLE gobatis_testc'; EXCEPTION WHEN OTHERS THEN NULL;END;
		BEGIN EXECUTE IMMEDIATE 'DROP TABLE gobatis_teste1'; EXCEPTION WHEN OTHERS THEN NULL;END;
		BEGIN EXECUTE IMMEDIATE 'DROP TABLE gobatis_teste2'; EXCEPTION WHEN OTHERS THEN NULL;END;
		BEGIN EXECUTE IMMEDIATE 'DROP TABLE gobatis_testf1'; EXCEPTION WHEN OTHERS THEN NULL;END;
		BEGIN EXECUTE IMMEDIATE 'DROP TABLE gobatis_testf2'; EXCEPTION WHEN OTHERS THEN NULL;END;


		CREATE TABLE gobatis_testa (
		  id NUMBER GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
		  field0      Char(1) NULL,
		  field1      int NULL,
		  field2      int NULL,
		  field3      float NULL,
		  field4      float NULL,
		  field5      varchar(50) NULL,
		  field6      TIMESTAMP(9) WITH LOCAL TIME ZONE NULL,
		  field7      varchar(50) NULL,
		  field8      varchar(50) NULL,
		  field9      clob NULL
		) ;


		CREATE TABLE gobatis_testb (
		  id NUMBER GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
		  field0      Char(1) NOT NULL,
		  field1      int NOT NULL,
		  field2      int NOT NULL,
		  field3      float NOT NULL,
		  field4      float NOT NULL,
		  field5      varchar(50) NOT NULL,
		  field6      TIMESTAMP(9) WITH LOCAL TIME ZONE NOT NULL,
		  field7      varchar(50) NOT NULL,
		  field8      varchar(50) NOT NULL,
		  field9      clob NULL
		) ;


		CREATE TABLE gobatis_testc (
		  id NUMBER GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
		  field0     varchar(500)
		) ;


		CREATE TABLE gobatis_teste1 (
		  id NUMBER GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
		  field0     varchar(500)
		) ;


		CREATE TABLE gobatis_teste2 (
		  id NUMBER GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
		  field0     varchar(500) NOT NULL
		) ;

		CREATE TABLE gobatis_testf1 (
		  id NUMBER GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
		  field0     varchar(500)
		) ;

		CREATE TABLE gobatis_testf2 (
		  id NUMBER GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
		  field0     varchar(500) NOT NULL
		);

		BEGIN EXECUTE IMMEDIATE 'DROP TABLE gobatis_convert1'; EXCEPTION WHEN OTHERS THEN NULL;END;
		CREATE TABLE gobatis_convert1 (
		  id NUMBER GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
		  field0     int
		);

		BEGIN EXECUTE IMMEDIATE 'DROP TABLE gobatis_convert2'; EXCEPTION WHEN OTHERS THEN NULL;END;
		CREATE TABLE gobatis_convert2 (
		  id NUMBER GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
		  field0     varchar(500)
		);


		BEGIN EXECUTE IMMEDIATE 'DROP TABLE computers'; EXCEPTION WHEN OTHERS THEN NULL;END;
		CREATE TABLE computers ( id NUMBER GENERATED ALWAYS AS IDENTITY PRIMARY KEY, description VARCHAR(56), mother_id INT, key_id INT, mouse_id INT);

		BEGIN EXECUTE IMMEDIATE 'DROP TABLE keyboards'; EXCEPTION WHEN OTHERS THEN NULL;END;
		CREATE TABLE keyboards ( id NUMBER GENERATED ALWAYS AS IDENTITY PRIMARY KEY, description VARCHAR(56));

		BEGIN EXECUTE IMMEDIATE 'DROP TABLE motherboards'; EXCEPTION WHEN OTHERS THEN NULL;END;
		CREATE TABLE motherboards ( id NUMBER GENERATED ALWAYS AS IDENTITY PRIMARY KEY, description VARCHAR(56));

		BEGIN EXECUTE IMMEDIATE 'DROP TABLE mouses'; EXCEPTION WHEN OTHERS THEN NULL;END;
		CREATE TABLE mouses (
		  id NUMBER GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
		  field1      Char(1),
		  field2      int,
		  field3      float,
		  field4      varchar(50),
		  field5      varchar(50),
		  field6      varchar(50)
		);
	`
)

var (
	PostgreSQLUrl     = "host=127.0.0.1 user=golang password=123456 dbname=golang sslmode=disable"
	PostgreSQLOdbcUrl = "DSN=gobatis_test;uid=golang;pwd=123456" // + ";database=xxx"
	MySQLUrl          = os.Getenv("mysql_username") + ":" + os.Getenv("mysql_password") + "@tcp(192.168.1.2:3306)/golang?autocommit=true&parseTime=true&multiStatements=true"
	MsSqlUrl          = "sqlserver://golang:123456@127.0.0.1?database=golang&connection+timeout=30"
	DMSqlUrl          = "dm://" + os.Getenv("dm_username") + ":" + os.Getenv("dm_password") + "@" + os.Getenv("dm_host") // + "?noConvertToHex=true"
	DmOdbcUrl         = "DSN=" + os.Getenv("dm_odbc_name") + ";uid=" + os.Getenv("dm_odbc_username") + ";pwd=" + os.Getenv("dm_odbc_password") // + ";database=xxx"
	Db2Url            = "HOSTNAME=127.0.0.1;DATABASE=golangtest;PORT=5000;UID=golangtest;PWD=golangtest"
	OracleUrl         = "oracle://" + os.Getenv("oracle_username") + ":" + os.Getenv("oracle_password") + "@" + os.Getenv("oracle_host") + "/" + os.Getenv("oracle_service")
)

func init() {
	if ss := strings.Fields(os.Getenv("oracle_username")); len(ss) == 3 && strings.ToLower(ss[1]) == "as" {
		OracleUrl = "oracle://" + ss[0] + ":" + os.Getenv("oracle_password") + "@" + os.Getenv("oracle_host") + "/" + os.Getenv("oracle_service") + "?DBA PRIVILEGE=" + ss[2]
	}
}

var (
	TestDrv     string
	TestConnURL string
)

func init() {
	flag.StringVar(&TestDrv, "dbDrv", "postgres", "")
	flag.StringVar(&TestConnURL, "dbURL", "", "缺省值会根据 dbDrv 的值自动选择，请见 GetTestConnURL()")
	//flag.StringVar(&TestConnURL, "dbURL", "golang:123456@tcp(localhost:3306)/golang?autocommit=true&parseTime=true&multiStatements=true", "")
	//flag.StringVar(&TestConnURL, "dbURL", "sqlserver://golang:123456@127.0.0.1?database=golang&connection+timeout=30", "")
}

func GetTestSQLText(drvName string) string {
	drvName = strings.ToLower(drvName)
retrySwitch:
	switch drvName {
	case "kingbase", "postgres", "":
		return PostgresqlScript
	case "sqlserver", "mssql":
		return MssqlScript
	case "mysql":
		return MysqlScript
	case "oracle":
		return OracleScript
	case "dm":
		return DMScript
	case "go_ibm_db":
		return Db2Script
	default:
		if strings.HasPrefix(drvName, dialects.OdbcPrefix) {
			drvName = strings.TrimPrefix(drvName, dialects.OdbcPrefix)
			goto retrySwitch
		}
		return "******************* no sql script *******************"
	}
}

func GetTestConnURL() string {
	if TestConnURL == "" {
		switch TestDrv {
		case "postgres", "":
			return PostgreSQLUrl
		case "odbc_with_postgres":
			return PostgreSQLOdbcUrl
		case "mysql":
			return MySQLUrl
		case "sqlserver", "mssql":
			return MsSqlUrl
		case "dm":
			return DMSqlUrl
		case "odbc_with_dm":
			return DmOdbcUrl
		case "go_ibm_db":
			return Db2Url
		case "oracle":
			return OracleUrl
		}
	}

	return TestConnURL
}

func Run(t testing.TB, cb func(t testing.TB, factory *gobatis.SessionFactory)) {
	log.SetFlags(log.Ldate | log.Lshortfile)

	o, err := gobatis.New(&gobatis.Config{
		DriverName: TestDrv,
		DataSource: GetTestConnURL(),
		XMLPaths: []string{"tests",
			"../tests",
			"../../tests"},
		MaxIdleConns: 2,
		MaxOpenConns: 2,
		//ShowSQL:      true,
		Tracer:   gobatis.TraceWriter{Output: os.Stderr},
		IsUnsafe: true,
	})
	if err != nil {
		t.Error(err)
		return
	}

	defer func() {
		if err = o.Close(); err != nil {
			t.Error(err)
		}
	}()

	tryCount := 0
	sqltext := GetTestSQLText(o.Dialect().Name())
retry:
	err = gobatis.ExecContext(context.Background(), o.DB(), sqltext)
	if err != nil {
		t.Error(o.Dialect().Name())
		t.Error(GetTestConnURL())

		t.Error("执行 SQL 失败")
		t.Error(sqltext)

		if e, ok := err.(*gobatis.SqlError); ok {
			t.Error("其中 SQL 失败")
			t.Error(e.SQL)
		}
		if strings.Contains(err.Error(), "pg_type_typname_nsp_index") {
			tryCount++
			if tryCount <= 5 {
				goto retry
			}
		}
		// if sqlErr := errors.ToSQLError(err); sqlErr != nil {
		// 	t.Error(sqlErr.SqlStr)
		// }
		t.Error(err)

		return
	}

	cb(t, o)
}
