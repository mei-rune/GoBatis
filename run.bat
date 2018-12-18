pushd cmd\gobatis
go install
@if %errorlevel% equ 1 goto :eof
popd
go generate ./...
@if %errorlevel% equ 1 goto :eof
del gentest\fail\interface.gobatis.go
go test -v ./...
@if %errorlevel% equ 0 goto :eof
echo test fail...