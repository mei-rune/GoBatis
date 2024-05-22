pushd cmd\gobatis
go install
@if %errorlevel% equ 1 goto :eof
popd


del tests\*.gobatis.go
del gentest\*.gobatis.go
del example\*.gobatis.go
del example_xml\*.gobatis.go
go generate ./...
@if %errorlevel% equ 1 goto :eof
del gentest\fail\interface.gobatis.go

go test -v   ./core
@if %errorlevel% NEQ 0 goto test_error
go test -v   .
@if %errorlevel% NEQ 0 goto test_error
go test -v   ./cmd/gobatis/goparser2
@if %errorlevel% NEQ 0 goto test_error
go test -v   ./cmd/gobatis/goparser2/astutil
@if %errorlevel% NEQ 0 goto test_error
go test -v   ./example
@if %errorlevel% NEQ 0 goto test_error
go test -v   ./example_xml
@if %errorlevel% NEQ 0 goto test_error
go test -v   ./convert
@if %errorlevel% NEQ 0 goto test_error
go test -v   ./dialects
@if %errorlevel% NEQ 0 goto test_error
go test -v   ./reflectx
@if %errorlevel% NEQ 0 goto test_error
go test -v   ./tests
@if %errorlevel% NEQ 0 goto test_error

go test -v -tags gval  ./core
@if %errorlevel% NEQ 0 goto test_error
go test -v -tags gval   .
@if %errorlevel% NEQ 0 goto test_error
go test -v -tags gval   ./cmd/gobatis/goparser2
@if %errorlevel% NEQ 0 goto test_error
go test -v -tags gval   ./cmd/gobatis/goparser2/astutil
@if %errorlevel% NEQ 0 goto test_error
go test -v -tags gval   ./example
@if %errorlevel% NEQ 0 goto test_error
go test -v -tags gval   ./example_xml
@if %errorlevel% NEQ 0 goto test_error
go test -v -tags gval   ./convert
@if %errorlevel% NEQ 0 goto test_error
go test -v -tags gval   ./dialects
@if %errorlevel% NEQ 0 goto test_error
go test -v -tags gval   ./reflectx
@if %errorlevel% NEQ 0 goto test_error
go test -v -tags gval   ./tests
@if %errorlevel% NEQ 0 goto test_error


:test_ok
@echo test ok
@goto :eof


:test_error
@echo test fail...
@goto :eof

