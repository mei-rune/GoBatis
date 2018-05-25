pushd cmd\gobatis
go install
popd
cd example
go generate
popd

go test -v ./...