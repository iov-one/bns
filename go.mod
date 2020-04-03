module github.com/iov-one/bns

go 1.13

replace (
	github.com/etcd-io/bbolt => go.etcd.io/bbolt v1.3.3
	github.com/swaggo/http-swagger => github.com/orkunkl/http-swagger v1.0.0
)

require (
	github.com/alecthomas/template v0.0.0-20190718012654-fb15b899a751
	github.com/etcd-io/bbolt v1.3.4 // indirect
	github.com/iov-one/weave v0.21.4
	github.com/swaggo/http-swagger v0.0.0-00010101000000-000000000000
	github.com/swaggo/swag v1.6.5
	github.com/tendermint/tendermint v0.31.9
	go.etcd.io/bbolt v1.3.4 // indirect
)
