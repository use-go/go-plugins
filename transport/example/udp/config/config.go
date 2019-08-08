// Package config implements go-config with env and k8s configmap
package config

import (
	"github.com/micro/go-micro/config"
	"github.com/micro/go-micro/config/source/env"
	"github.com/micro/go-plugins/config/source/configmap"
)

// NewConfig returns config with env and k8s configmap setup
func NewConfig(opts ...config.Option) config.Config {
	cfg := config.NewConfig()
	cfg.Load(
		env.NewSource(),
		configmap.NewSource(),
	)
	return cfg
}
