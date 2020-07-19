module github.com/micro/go-plugins/wrapper/trace/opentracing/v2

go 1.13

require (
	github.com/micro/go-micro/v2 v2.9.1-0.20200716153311-f9bf56239306
	github.com/opentracing/opentracing-go v1.1.0
	github.com/stretchr/testify v1.4.0
)

replace github.com/coreos/etcd => github.com/ozonru/etcd v3.3.20-grpc1.27-origmodule+incompatible
