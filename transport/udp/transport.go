package udp

import (
	"bufio"
	"context"
	"crypto/tls"
	"encoding/gob"
	"net"

	"github.com/micro/go-micro/transport"
	maddr "github.com/micro/go-micro/util/addr"
	mnet "github.com/micro/go-micro/util/net"
	mls "github.com/micro/go-micro/util/tls"
)

func (u *udpTransport) dialContext(ctx context.Context, addr string) (nc net.Conn, err error) {
	netAddr, err := net.ResolveUDPAddr("udp", addr)

}

func (u *udpTransport) Dial(addr string, opts ...transport.DialOption) (transport.Client, error) {
	dopts := transport.DialOptions{
		Timeout: transport.DefaultDialTimeout,
	}

	for _, opt := range opts {
		opt(&dopts)
	}

	ctx, _ := context.WithTimeout(context.Background(), dopts.Timeout)
	c, err := u.dialContext(ctx, addr)
	if err != nil {
		return nil, err
	}

	if u.opts.Secure || u.opts.TLSConfig != nil {
		config := u.opts.TLSConfig
		if config == nil {
			config = &tls.Config{
				InsecureSkipVerify: true,
			}
		}
		c = tls.Client(c, config)
	}

	encBuf := bufio.NewWriter(c)

	return &udpClient{
		dialOpts: dopts,
		conn:     c,
		encBuf:   encBuf,
		enc:      gob.NewEncoder(encBuf),
		dec:      gob.NewDecoder(c),
		timeout:  u.opts.Timeout,
	}, nil
}

func (u *udpTransport) invokeListen(laddr string) (net.Listener, error) {

}

func (u *udpTransport) Listen(addr string, opts ...transport.ListenOption) (transport.Listener, error) {
	var options transport.ListenOptions
	for _, o := range opts {
		o(&options)
	}
	var l net.Listener
	var err error

	if u.opts.Secure || u.opts.TLSConfig != nil {
		config := u.opts.TLSConfig

		fn := func(addr string) (net.Listener, error) {
			if config == nil {
				hosts := []string{addr}

				// check if its a valid host:port
				if host, _, err := net.SplitHostPort(addr); err == nil {
					if len(host) == 0 {
						hosts = maddr.IPs()
					} else {
						hosts = []string{host}
					}
				}

				// generate a certificate
				cert, err := mls.Certificate(hosts...)
				if err != nil {
					return nil, err
				}
				config = &tls.Config{Certificates: []tls.Certificate{cert}}
			}
			l, err := u.invokeListen(addr)
			if err != nil {
				return nil, err
			}
			return tls.NewListener(l, config), nil
		}

		l, err = mnet.Listen(addr, fn)
	} else {
		l, err = mnet.Listen(addr, u.invokeListen)
	}

	if err != nil {
		return nil, err
	}

	return &udpListener{
		timeout:  u.opts.Timeout,
		listener: l,
		opts:     options,
	}, nil
}

func (u *udpTransport) Init(opts ...transport.Option) error {
	for _, o := range opts {
		o(&u.opts)
	}
	return nil
}

func (u *udpTransport) Options() transport.Options {
	return u.opts
}

func (u *udpTransport) String() string {
	return "udp"
}
