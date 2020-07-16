module github.com/micro/go-plugins/wrapper/ratelimiter/uber/v2

go 1.13

require (
	github.com/micro/go-micro/v2 v2.9.1-0.20200716123506-3627e47f04eb
	go.uber.org/ratelimit v0.1.0
)

replace github.com/coreos/etcd => github.com/ozonru/etcd v3.3.20-grpc1.27-origmodule+incompatible
