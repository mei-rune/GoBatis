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

