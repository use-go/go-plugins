package main

import (
	"fmt"

	"github.com/micro/go-micro"
	"github.com/micro/go-micro/registry/mdns"

	// static selector offloads load balancing to k8s services
	// note: requires user to create k8s services
	"github.com/use-go/go-plugins/client/selector/static"

	"github.com/use-go/go-plugins/transport/udp"
)

func main() {

	// create registry and selector
	r := mdns.NewRegistry()
	s := static.NewSelector()
	t := udp.NewTransport()
	// Create a new service. Optionally include some options here.
	service := micro.NewService(
		micro.Name("udp-transport"),
		micro.Version("latest"),
		micro.Metadata(map[string]string{
			"type": "udp-server-demo",
		}),
		micro.Registry(r),
		micro.Selector(s),
		micro.Transport(t), //use UDP transport
	)

	// Init will parse the command line flags. Any flags set will
	// override the above settings. Options defined here will
	// override anything set on the command line.
	service.Init()

	// Run the server
	if err := service.Run(); err != nil {
		fmt.Println(err)
	}
}
