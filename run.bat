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
go test -v ./...
@if %errorlevel% equ 0 goto :eof
echo test fail...