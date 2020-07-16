module github.com/micro/go-plugins/config/encoder/cue/v2

go 1.13

require (
	cuelang.org/go v0.0.15
	github.com/ghodss/yaml v1.0.0
	github.com/micro/go-micro/v2 v2.9.1-0.20200716123506-3627e47f04eb
	github.com/stretchr/testify v1.4.0
)

replace github.com/coreos/etcd => github.com/ozonru/etcd v3.3.20-grpc1.27-origmodule+incompatible
