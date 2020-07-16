module github.com/micro/go-plugins/broker/kafka/v2

go 1.13

require (
	github.com/Shopify/sarama v1.25.0
	github.com/google/uuid v1.1.1
	github.com/micro/go-micro/v2 v2.9.1-0.20200716123506-3627e47f04eb
)

replace github.com/coreos/etcd => github.com/ozonru/etcd v3.3.20-grpc1.27-origmodule+incompatible
