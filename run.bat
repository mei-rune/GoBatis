pushd cmd\gobatis
go build
popd
pushd example
..\cmd\gobatis\gobatis.exe user.go
go test -v
popd