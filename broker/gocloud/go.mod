module github.com/micro/go-plugins/broker/gocloud/v2

go 1.13

require (
	github.com/micro/go-micro/v2 v2.9.1-0.20200716123506-3627e47f04eb
	github.com/streadway/amqp v0.0.0-20200108173154-1c71cc93ed71
	gocloud.dev v0.17.0
	gocloud.dev/pubsub/rabbitpubsub v0.17.0
	golang.org/x/oauth2 v0.0.0-20191202225959-858c2ad4c8b6 // indirect
)

replace github.com/coreos/etcd => github.com/ozonru/etcd v3.3.20-grpc1.27-origmodule+incompatible
