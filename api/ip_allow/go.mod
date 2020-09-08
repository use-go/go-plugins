module github.com/micro/go-plugins/api/ip_whitelist/v2

go 1.13

require (
	github.com/micro/cli/v2 v2.1.2
	github.com/micro/go-micro/v2 v2.9.1-0.20200723075038-fbdf1f2c1c4c
	github.com/micro/micro/v2 v2.9.2-0.20200728090142-c7f7e4a71077
)

replace github.com/coreos/etcd => github.com/ozonru/etcd v3.3.20-grpc1.27-origmodule+incompatible
