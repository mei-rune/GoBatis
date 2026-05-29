pushd cmd\gobatis
go install
@if %errorlevel% equ 1 goto :eof
popd


del tests\*.gobatis.go
del gentest\*.gobatis.go
del example\*.gobatis.go
del example_xml\*.gobatis.go
set PATH=%PATH%;.\cmd\gobatis
go generate ./...
@if %errorlevel% equ 1 goto :eof
del gentest\fail\interface.gobatis.go

rem set gobatis_db_drv=dm
rem set dm_host=192.168.100.2:5236
rem set dm_username=golang
rem set dm_password=Test@123456


rem set gobatis_db_drv=postgres
rem set gobatis_db_url=host=192.168.1.98 user=golang password=Test@123456 dbname=golang sslmode=disable

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

set gobatis_db_drv=oracle
@rem set gobatis_db_url=oracle://golang:Test@123456@192.168.1.51:30211/ORCLPDB1
@rem 密码中 @ 是特殊字符要转义，转义为 %40，但是因为这个在 cmd 中运行 % 也要转义一下
set gobatis_db_url=oracle://golang:Test%%40123456@192.168.1.51:30211/ORCLPDB1


@rem set gobatis_db_drv=oracle
@rem set mariadb_username=root
@rem set mariadb_password=xxx

@rem set tags=-tags gval
@if "%tags%" == "" (
  set tags=-tags gval
)

go test -v  %tags% ./core
@if %errorlevel% NEQ 0 goto test_error
go test -v  %tags% .
@if %errorlevel% NEQ 0 goto test_error
go test -v  %tags% ./cmd/gobatis/goparser2
@if %errorlevel% NEQ 0 goto test_error
go test -v  %tags% ./cmd/gobatis/goparser2/astutil
@if %errorlevel% NEQ 0 goto test_error
go test -v  %tags% ./example
@if %errorlevel% NEQ 0 goto test_error
go test -v  %tags% ./example_xml
@if %errorlevel% NEQ 0 goto test_error
go test -v  %tags% ./convert
@if %errorlevel% NEQ 0 goto test_error
go test -v  %tags% ./dialects
@if %errorlevel% NEQ 0 goto test_error
go test -v  %tags% ./reflectx
@if %errorlevel% NEQ 0 goto test_error
go test -v  %tags% ./tests
@if %errorlevel% NEQ 0 goto test_error

:test_ok
@echo test ok
@goto :eof


:test_error
@echo test fail...
@goto :eof

