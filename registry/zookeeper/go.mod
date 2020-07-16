module github.com/micro/go-plugins/registry/zookeeper/v2

go 1.13

require (
	github.com/micro/go-micro/v2 v2.9.1-0.20200716153311-f9bf56239306
	github.com/mitchellh/hashstructure v1.0.0
	github.com/samuel/go-zookeeper v0.0.0-20190923202752-2cc03de413da
)

replace github.com/coreos/etcd => github.com/ozonru/etcd v3.3.20-grpc1.27-origmodule+incompatible
