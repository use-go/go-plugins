// Package access is a micro plugin for accessing service requests
package access

import (
	"net/http"
	"strings"

	"github.com/micro/cli/v2"
	"github.com/micro/go-micro/v2/client"
	"github.com/micro/micro/v2/plugin"
)

type access struct {
	services map[string]bool
}

func (w *access) Flags() []cli.Flag {
	return []cli.Flag{
		&cli.StringFlag{
			Name:   "rpc_allow",
			Usage:  "Comma separated access of accessed services for RPC calls",
			EnvVars: []string{"RPC_ALLOW"},
		},
	}
}

func (w *access) Commands() []*cli.Command {
	return nil
}

func (w *access) Handler() plugin.Handler {
	return func(h http.Handler) http.Handler {
		return h
	}
}

func (w *access) Init(ctx *cli.Context) error {
	if access := ctx.String("rpc_allow"); len(access) > 0 {
		client.DefaultClient = newClient(strings.Split(access, ",")...)
	}
	return nil
}

func (w *access) String() string {
	return "access"
}

func NewPlugin() plugin.Plugin {
	return NewRPCAllow()
}

func NewRPCAllow(services ...string) plugin.Plugin {
	list := make(map[string]bool)

	for _, service := range services {
		list[service] = true
	}

	return &access{
		services: list,
	}
}
