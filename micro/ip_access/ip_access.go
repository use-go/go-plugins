// Package ip_access is a micro plugin for accessing ip addresses
package ip_access

import (
	"net"
	"net/http"
	"strings"

	"github.com/micro/cli/v2"
	log "github.com/micro/go-micro/v2/logger"
	"github.com/micro/micro/v2/plugin"
)

type access struct {
	cidrs map[string]*net.IPNet
	ips   map[string]bool
}

func (w *access) Flags() []cli.Flag {
	return []cli.Flag{
		&cli.StringFlag{
			Name:    "ip_access",
			Usage:   "Comma separated access of accessed IPs",
			EnvVars: []string{"IP_WHITELIST"},
		},
	}
}

func (w *access) load(ips ...string) {
	for _, ip := range ips {
		parts := strings.Split(ip, "/")

		switch len(parts) {
		// assume just an ip
		case 1:
			w.ips[ip] = true
		case 2:
			// parse cidr
			_, ipnet, err := net.ParseCIDR(ip)
			if err != nil {
				log.Fatalf("[ip_access] failed to parse %v: %v", ip, err)
			}
			w.cidrs[ipnet.String()] = ipnet
		default:
			log.Fatalf("[ip_access] failed to parse %v", ip)
		}
	}

}

func (w *access) match(ip string) bool {
	// make ip
	nip := net.ParseIP(ip)

	// check ips
	if ok := w.ips[nip.String()]; ok {
		return true
	}

	// check cidrs
	for _, cidr := range w.cidrs {
		if cidr.Contains(nip) {
			return true
		}
	}

	// no match
	return false
}

func (w *access) Commands() []*cli.Command {
	return nil
}

func (w *access) Handler() plugin.Handler {
	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
			// check remote addr; if we can't parse it passes through
			if ip, _, err := net.SplitHostPort(r.RemoteAddr); err == nil {
				// reject if no match
				if !w.match(ip) {
					http.Error(rw, "forbidden", 403)
					return
				}
			}

			// serve the request
			h.ServeHTTP(rw, r)
		})
	}
}

func (w *access) Init(ctx *cli.Context) error {
	if access := ctx.String("ip_access"); len(access) > 0 {
		w.load(strings.Split(access, ",")...)
	}
	return nil
}

func (w *access) String() string {
	return "ip_access"
}

func NewPlugin() plugin.Plugin {
	return NewIPAccess()
}

func NewIPAccess(ips ...string) plugin.Plugin {
	// create plugin
	w := &access{
		cidrs: make(map[string]*net.IPNet),
		ips:   make(map[string]bool),
	}

	// load ips
	w.load(ips...)

	return w
}
