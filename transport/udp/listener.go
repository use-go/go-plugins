package udp

import (
	"bufio"
	"encoding/gob"
	"net"
	"time"

	"github.com/micro/go-micro/transport"
	"github.com/micro/go-micro/util/log"
)

func (u *udpListener) Addr() string {
	return u.listener.Addr().String()
}

func (u *udpListener) Close() error {
	return u.listener.Close()
}

//Accept and handle a data package
func (u *udpListener) Accept(fn func(transport.Socket)) error {
	var tempDelay time.Duration

	for {
		c, err := u.listener.Accept()
		//bytesLenth, comefrom, err := u.listenerCon.ReadFromUDP(buf)
		if err != nil {
			if ne, ok := err.(net.Error); ok && ne.Temporary() {
				if tempDelay == 0 {
					tempDelay = 5 * time.Millisecond
				} else {
					tempDelay *= 2
				}
				if max := 1 * time.Second; tempDelay > max {
					tempDelay = max
				}
				log.Logf("udp: Accept error: %v; retrying in %v\n", err, tempDelay)
				time.Sleep(tempDelay)
				continue
			}
			return err
		}

		encBuf := bufio.NewWriter(c)
		sock := &udpSocket{
			timeout: u.timeout,
			conn:    c,
			encBuf:  encBuf,
			enc:     gob.NewEncoder(encBuf),
			dec:     gob.NewDecoder(c),
		}

		go func() {
			// TODO: think of a better error response strategy
			defer func() {
				if r := recover(); r != nil {
					sock.Close()
				}
			}()

			fn(sock)
		}()
	}
}
