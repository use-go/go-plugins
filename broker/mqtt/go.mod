module github.com/micro/go-plugins/broker/mqtt/v2

go 1.13

require (
	github.com/eclipse/paho.mqtt.golang v1.2.0
	github.com/micro/go-micro/v2 v2.9.1-0.20200716153311-f9bf56239306
)

replace github.com/coreos/etcd => github.com/ozonru/etcd v3.3.20-grpc1.27-origmodule+incompatible
