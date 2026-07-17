
SET GOPATH=C:\developing\go\hengwei;C:\developing\go\hengwei\tpt_vendor
SET GO111MODULE=off
SET CGO_ENABLED=1
SET PATH=%PATH%;C:\developing\tools\mingw64\bin;C:\developing\go\hengwei\tpt_vendor\src\gitee.com\shentongdata\go-aci\lib\win64

pushd cmd\gobatis
go install
@if %errorlevel% equ 1 goto :eof
popd

del tests\*.gobatis.go
del gentest\*.gobatis.go
del example\*.gobatis.go
del example_xml\*.gobatis.go
set PATH=%PATH%;%USERPROFILE%\go\bin
go generate ./...
@if %errorlevel% equ 1 goto :eof
del gentest\fail\interface.gobatis.go

rem set gobatis_db_drv=dm
rem set dm_host=192.168.100.2:5236
rem set dm_username=golang
rem set dm_password=Test@123456


rem set gobatis_db_drv=postgres
rem set gobatis_db_url=host=192.168.1.98 user=golang password=Test@123456 dbname=golang sslmode=disable



rem set gobatis_db_drv=oceanbase_mysql
@rem 密码中 @ 是特殊字符要转义，转义为 %40，但是因为这个在 cmd 中运行 % 也要转义一下
rem set gobatis_db_url=golang:12345678@tcp(192.168.1.228:2881)/golang?autocommit=true^&parseTime=true^&multiStatements=true


rem set gobatis_db_drv=oceanbase_mysql
rem @rem 密码中 @ 是特殊字符要转义，转义为 %40，但是因为这个在 cmd 中运行 % 也要转义一下
rem set gobatis_db_url=golang:123456@tcp(192.168.100.2:3306)/golang?autocommit=true^&parseTime=true^&multiStatements=true


rem set gobatis_db_drv=pgx/v5
@rem 密码中 @ 是特殊字符要转义，转义为 %40，但是因为这个在 cmd 中运行 % 也要转义一下
rem set gobatis_db_url=postgres://golang:Test%%40123456@192.168.1.98/golang?sslmode=disable^&client_encoding=UTF8

rem set gobatis_db_drv=kingbase
rem set gobatis_db_url=host=192.168.1.52 port=31432 user=golang password=12345678 dbname=golang sslmode=disable


rem set gobatis_db_drv=mariadb
rem set mariadb_host=192.168.1.50:33306
rem set mariadb_dbname=golang
rem set mariadb_username=golang
rem set mariadb_password=Test@123456

rem set gobatis_db_drv=oracle
rem set oracle_host=192.168.1.51:30211
rem set oracle_service=ORCLCDB
rem set oracle_username=golang
rem set oracle_password=Test@123456

rem set gobatis_db_drv=oracle
rem @rem set gobatis_db_url=oracle://golang:Test@123456@192.168.1.51:30211/ORCLPDB1
rem @rem 密码中 @ 是特殊字符要转义，转义为 %40，但是因为这个在 cmd 中运行 % 也要转义一下
rem set gobatis_db_url=oracle://golang:Test%%40123456@192.168.1.51:30211/ORCLPDB1


rem ALTER USER app_user1 IDENTIFIED BY szoscar55;
rem CREATE DATABASE golang USER sysdba password 'tpt_a5sdfasdf6'
rem CREATE USER golang IDENTIFIED BY tpt_a5sdfasdf6;

set gobatis_db_drv=shengtong_oscar
set gobatis_db_url=golang/tpt_a5sdfasdf6@192.168.1.52:32003/test?dbtext_max_len=100000;fetch_size=100




rem set gobatis_db_drv=oceanbase_mysql
rem @rem 密码中 @ 是特殊字符要转义，转义为 %40，但是因为这个在 cmd 中运行 % 也要转义一下
rem set gobatis_db_url=root@sys:xxxxx@tcp(192.168.1.11:2881)/oceanbase?autocommit=true^&parseTime=true^&multiStatements=true


@rem set gobatis_db_drv=oracle
@rem set mariadb_username=root
@rem set mariadb_password=xxx

@rem set tags=-tags gval
@if "%tags%" == "" (
  set tags=-tags gval
)

set args=-timeout 2h

go test -v %tags% %args% ./core
@if %errorlevel% NEQ 0 goto test_error
go test -v  %tags% %args% .
@if %errorlevel% NEQ 0 goto test_error
go test -v  %tags% %args% ./cmd/gobatis/goparser2
@if %errorlevel% NEQ 0 goto test_error
go test -v  %tags% %args% ./cmd/gobatis/goparser2/astutil
@if %errorlevel% NEQ 0 goto test_error
go test -v  %tags% %args% ./example
@if %errorlevel% NEQ 0 goto test_error
go test -v  %tags% %args% ./example_xml
@if %errorlevel% NEQ 0 goto test_error
go test -v  %tags% %args% ./convert
@if %errorlevel% NEQ 0 goto test_error
go test -v  %tags% %args% ./dialects
@if %errorlevel% NEQ 0 goto test_error
go test -v  %tags% %args% ./reflectx
@if %errorlevel% NEQ 0 goto test_error
go test -v  %tags% %args% ./tests
@if %errorlevel% NEQ 0 goto test_error

:test_ok
@echo test ok
@goto :eof


:test_error
@echo test fail...
@goto :eof

