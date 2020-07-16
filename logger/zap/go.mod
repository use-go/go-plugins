module github.com/micro/go-plugins/logger/zap/v2

go 1.13

require (
	github.com/micro/go-micro/v2 v2.9.1-0.20200716123506-3627e47f04eb
	go.uber.org/zap v1.13.0
)

replace github.com/coreos/etcd => github.com/ozonru/etcd v3.3.20-grpc1.27-origmodule+incompatible
