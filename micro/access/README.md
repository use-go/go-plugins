# Access Plugin

The access plugin is a plugin for micro which accesss the services that can be used via the /rpc HTTP endpoint.

## Usage

Register the plugin before building Micro

```
package main

import (
	"github.com/micro/micro/plugin"
	"github.com/micro/go-plugins/micro/access"
)

func init() {
	plugin.Register(access.NewRPCAccess())
}
```

It can then be applied on the command line like so.

```
micro --rpc_access go.micro.srv.greeter,go.micro.srv.example api
```

### Scoped to API

If you like to only apply the plugin for a specific component you can register it with that specifically. 
For example, below you'll see the plugin registered with the API.

```
package main

import (
	"github.com/micro/micro/api"
	"github.com/micro/go-plugins/micro/access"
)

func init() {
	api.Register(access.NewRPCAccess())
}
```

Here's what the help displays when you do that.

```
$ go run main.go link.go api --help
NAME:
   main api - Run the micro API

USAGE:
   main api [command options] [arguments...]

OPTIONS:
   --rpc_access 	Comma separated access of accessed services for RPC calls [$MICRO_RPC_WHITELIST]
```

In this case the usage would be

```
micro api --rpc_access go.micro.srv.greeter
```
