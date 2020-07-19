module github.com/micro/go-plugins/config/source/grpc/v2

go 1.13

require (
	github.com/golang/protobuf v1.4.0
	github.com/micro/go-micro/v2 v2.9.1-0.20200716153311-f9bf56239306
	golang.org/x/net v0.0.0-20200520182314-0ba52f642ac2
	google.golang.org/grpc v1.26.0
)

replace github.com/coreos/etcd => github.com/ozonru/etcd v3.3.20-grpc1.27-origmodule+incompatible
