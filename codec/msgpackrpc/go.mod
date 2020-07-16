module github.com/micro/go-plugins/codec/msgpackrpc/v2

go 1.13

require (
	github.com/micro/go-micro/v2 v2.9.1-0.20200716123506-3627e47f04eb
	github.com/philhofer/fwd v1.0.0 // indirect
	github.com/tinylib/msgp v1.1.1
)

replace github.com/coreos/etcd => github.com/ozonru/etcd v3.3.20-grpc1.27-origmodule+incompatible
