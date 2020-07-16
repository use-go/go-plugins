module github.com/micro/go-plugins/store/redis/v2

go 1.13

require (
	github.com/go-redis/redis/v7 v7.4.0
	github.com/micro/go-micro/v2 v2.9.1-0.20200716123506-3627e47f04eb
	gopkg.in/yaml.v2 v2.2.7 // indirect
)

replace github.com/coreos/etcd => github.com/ozonru/etcd v3.3.20-grpc1.27-origmodule+incompatible
