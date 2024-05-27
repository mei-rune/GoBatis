module github.com/runner-mei/GoBatis

require (
	gitee.com/opengauss/openGauss-connector-go-pq v1.0.2
	gitee.com/runner.mei/gokb v0.0.0-20211216112635-582e81b3d7ad
	github.com/Knetic/govaluate v3.0.1-0.20171022003610-9aa49832a739+incompatible
	github.com/aryann/difflib v0.0.0-20170710044230-e206f873d14a
	github.com/denisenkom/go-mssqldb v0.11.0
	github.com/go-sql-driver/mysql v1.6.0
	github.com/grsmv/inflect v0.0.0-20140723132642-a28d3de3b3ad
	github.com/lib/pq v1.10.4
	github.com/mattn/goveralls v0.0.11
	github.com/sijms/go-ora/v2 v2.2.17
	golang.org/x/tools v0.9.1
)

require github.com/PaesslerAG/gval v1.2.3-0.20240523111506-121093f3c9a6

require (
	github.com/antlr4-go/antlr/v4 v4.13.0 // indirect
	github.com/golang-sql/civil v0.0.0-20190719163853-cb61b32ac6fe // indirect
	github.com/golang/snappy v0.0.4 // indirect
	github.com/shopspring/decimal v1.4.0 // indirect
	github.com/stoewer/go-strcase v1.2.0 // indirect
	github.com/stretchr/testify v1.8.3 // indirect
	golang.org/x/crypto v0.9.0 // indirect
	golang.org/x/exp v0.0.0-20230515195305-f3d0a9c9a5cc // indirect
	golang.org/x/sys v0.8.0 // indirect
	golang.org/x/xerrors v0.0.0-20220907171357-04be3eba64a2 // indirect
	google.golang.org/genproto/googleapis/api v0.0.0-20230803162519-f966b187b2e5 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20230803162519-f966b187b2e5 // indirect
	google.golang.org/protobuf v1.31.0 // indirect
)

require (
	gitee.com/chunanyong/dm v1.8.15-0.20240130091939-38fab3047677
	github.com/alexbrainman/odbc v0.0.0-20211220213544-9c9a2e61c5e2
	github.com/google/cel-go v0.20.1
	github.com/ibmdb/go_ibm_db v0.4.1
	golang.org/x/mod v0.10.0
	golang.org/x/text v0.14.0 // indirect
)

replace github.com/PaesslerAG/gval v1.2.3-0.20240523111506-121093f3c9a6 => github.com/mei-rune/gval v0.0.0-20240527144442-8679586f671f

go 1.17
