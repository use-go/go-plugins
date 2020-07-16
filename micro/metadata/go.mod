module github.com/micro/go-plugins/micro/metadata/v2

go 1.13

require (
	github.com/micro/cli/v2 v2.1.2
	github.com/micro/go-micro/v2 v2.9.1-0.20200716123506-3627e47f04eb
	github.com/micro/micro/v2 v2.9.2-0.20200716115650-3579db7b7955
)

replace github.com/coreos/etcd => github.com/ozonru/etcd v3.3.20-grpc1.27-origmodule+incompatible
