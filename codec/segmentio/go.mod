module github.com/micro/go-plugins/codec/segmentio/v2

go 1.13

require (
	github.com/golang/protobuf v1.4.0
	github.com/micro/go-micro/v2 v2.9.1-0.20200716153311-f9bf56239306
	github.com/oxtoacart/bpool v0.0.0-20190530202638-03653db5a59c
	github.com/segmentio/encoding v0.1.10
)

replace github.com/coreos/etcd => github.com/ozonru/etcd v3.3.20-grpc1.27-origmodule+incompatible
