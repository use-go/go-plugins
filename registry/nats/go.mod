module github.com/micro/go-plugins/registry/nats/v2

go 1.13

require (
	github.com/go-log/log v0.2.0
	github.com/micro/go-micro/v2 v2.9.1-0.20200716123506-3627e47f04eb
	github.com/nats-io/nats.go v1.9.2
)

replace github.com/coreos/etcd => github.com/ozonru/etcd v3.3.20-grpc1.27-origmodule+incompatible
