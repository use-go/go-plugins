module github.com/micro/go-plugins/registry/gossip/v2

go 1.13

require (
	github.com/golang/protobuf v1.4.0
	github.com/google/uuid v1.1.1
	github.com/hashicorp/memberlist v0.1.5
	github.com/micro/go-micro/v2 v2.9.1-0.20200716153311-f9bf56239306
	github.com/mitchellh/hashstructure v1.0.0
)

replace github.com/coreos/etcd => github.com/ozonru/etcd v3.3.20-grpc1.27-origmodule+incompatible
