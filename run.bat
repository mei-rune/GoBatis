pushd cmd\gobatis
go install
@if errorlevel 1 goto failed
popd
go generate ./...
@if errorlevel 1 goto failed
go test -v ./...