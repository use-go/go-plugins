package memcached

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"net"
	"strings"
	"time"

	mc "github.com/bradfitz/gomemcache/memcache"
	"github.com/micro/go-micro/v2/store"
)

type mkv struct {
	options store.Options
	Server  *mc.ServerList
	Client  *mc.Client
}

func (m *mkv) Init(...store.Option) error {
	return nil
}

func (m *mkv) Read(key string, opts ...store.ReadOption) ([]*store.Record, error) {
	// TODO: implement read options
	records := make([]*store.Record, 0, 1)

	keyval, err := m.Client.Get(key)
	if err != nil && err == mc.ErrCacheMiss {
		return nil, store.ErrNotFound
	} else if err != nil {
		return nil, err
	}

	if keyval == nil {
		return nil, store.ErrNotFound
	}

	records = append(records, &store.Record{
		Key:    keyval.Key,
		Value:  keyval.Value,
		Expiry: time.Second * time.Duration(keyval.Expiration),
	})

	return records, nil
}

func (m *mkv) Delete(key string) error {
	return m.Client.Delete(key)
}

func (m *mkv) Write(record *store.Record) error {
	return m.Client.Set(&mc.Item{
		Key:        record.Key,
		Value:      record.Value,
		Expiration: int32(record.Expiry.Seconds()),
	})
}

func (m *mkv) List() ([]*store.Record, error) {
	// stats
	// cachedump
	// get keys

	var keys []string

	//store := make(map[string]string)
	if err := m.Server.Each(func(c net.Addr) error {
		cc, err := net.Dial("tcp", c.String())
		if err != nil {
			return err
		}
		defer cc.Close()

		b := bufio.NewReadWriter(bufio.NewReader(cc), bufio.NewWriter(cc))

		// get records
		if _, err := fmt.Fprintf(b, "stats records\r\n"); err != nil {
			return err
		}

		b.Flush()

		v, err := b.ReadSlice('\n')
		if err != nil {
			return err
		}

		parts := bytes.Split(v, []byte("\n"))
		if len(parts) < 1 {
			return nil
		}
		vals := strings.Split(string(parts[0]), ":")
		records := vals[1]

		// drain
		for {
			buf, err := b.ReadSlice('\n')
			if err == io.EOF {
				break
			}
			if err != nil {
				return err
			}
			if strings.HasPrefix(string(buf), "END") {
				break
			}
		}

		b.Writer.Reset(cc)
		b.Reader.Reset(cc)

		if _, err := fmt.Fprintf(b, "lru_crawler metadump %s\r\n", records); err != nil {
			return err
		}
		b.Flush()

		for {
			v, err := b.ReadString('\n')
			if err == io.EOF {
				break
			}
			if err != nil {
				return err
			}
			if strings.HasPrefix(v, "END") {
				break
			}
			key := strings.Split(v, " ")[0]
			keys = append(keys, strings.TrimPrefix(key, "key="))
		}

		return nil
	}); err != nil {
		return nil, err
	}

	var vals []*store.Record

	// concurrent op
	ch := make(chan []*store.Record, len(keys))

	for _, k := range keys {
		go func(key string) {
			i, _ := m.Read(key)
			ch <- i
		}(k)
	}

	for i := 0; i < len(keys); i++ {
		records := <-ch

		if records == nil {
			continue
		}

		vals = append(vals, records...)
	}

	close(ch)

	return vals, nil
}

func (m *mkv) String() string {
	return "memcached"
}

func NewStore(opts ...store.Option) store.Store {
	var options store.Options
	for _, o := range opts {
		o(&options)
	}

	nodes := options.Nodes

	if len(nodes) == 0 {
		nodes = []string{"127.0.0.1:11211"}
	}

	ss := new(mc.ServerList)
	ss.SetServers(nodes...)

	return &mkv{
		options: options,
		Server:  ss,
		Client:  mc.New(nodes...),
	}
}
