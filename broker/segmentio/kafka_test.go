package segmentio

import (
	"os"
	"strings"
	"testing"

	"github.com/micro/go-micro/v2/broker"
	segjson "github.com/micro/go-plugins/codec/segmentio/v2"
)

var (
	bm = &broker.Message{
		Header: map[string]string{"hkey": "hval"},
		Body:   []byte("body"),
	}
)

func TestPublish(t *testing.T) {
	if tr := os.Getenv("TRAVIS"); len(tr) > 0 {
		t.Skip()
	}

	var addrs []string
	if addr := os.Getenv("BROKER_ADDRS"); len(addr) == 0 {
		addrs = []string{"127.0.0.1:9092"}
	} else {
		addrs = strings.Split(addr, ",")
	}

	b := NewBroker(broker.Codec(segjson.Marshaler{}), broker.Addrs(addrs...))
	if err := b.Connect(); err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := b.Disconnect(); err != nil {
			t.Fatal(err)
		}
	}()

	done := make(chan bool, 1)
	fn := func(msg broker.Event) error {
		done <- true
		return msg.Ack()
	}

	sub, err := b.Subscribe("test_topic", fn, broker.Queue("test"))
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := sub.Unsubscribe(); err != nil {
			t.Fatal(err)
		}
	}()

	if err := b.Publish("test_topic", bm); err != nil {
		t.Fatal(err)
	}
	<-done
}

func BenchmarkSegmentioPublish(b *testing.B) {
	if tr := os.Getenv("TRAVIS"); len(tr) > 0 {
		b.Skip()
	}

	var addrs []string
	if addr := os.Getenv("BROKER_ADDRS"); len(addr) == 0 {
		addrs = []string{"127.0.0.1:9092"}
	} else {
		addrs = strings.Split(addr, ",")
	}

	brk := NewBroker(broker.Codec(segjson.Marshaler{}), broker.Addrs(addrs...))
	if err := brk.Connect(); err != nil {
		b.Fatal(err)
	}
	defer func() {
		if err := brk.Disconnect(); err != nil {
			b.Fatal(err)
		}
	}()

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		if err := brk.Publish("test_topic", bm); err != nil {
			b.Fatal(err)
		}
	}

}

func BenchmarkSegmentioSubscribe(b *testing.B) {
	if tr := os.Getenv("TRAVIS"); len(tr) > 0 {
		b.Skip()
	}

	var addrs []string
	if addr := os.Getenv("BROKER_ADDRS"); len(addr) == 0 {
		addrs = []string{"127.0.0.1:9092"}
	} else {
		addrs = strings.Split(addr, ",")
	}

	brk := NewBroker(broker.Codec(segjson.Marshaler{}), broker.Addrs(addrs...))
	if err := brk.Connect(); err != nil {
		b.Fatal(err)
	}
	defer func() {
		if err := brk.Disconnect(); err != nil {
			b.Fatal(err)
		}
	}()

	cnt := 0
	done := make(chan struct{})
	fn := func(msg broker.Event) error {
		if cnt == 0 {
			b.ResetTimer()
		}
		cnt++
		if cnt == b.N {
			close(done)
		}
		return msg.Ack()
	}

	for i := 0; i < b.N; i++ {
		if err := brk.Publish("test_topic", bm); err != nil {
			b.Fatal(err)
		}
	}

	sub, err := brk.Subscribe("test_topic", fn, broker.Queue("test"))
	if err != nil {
		b.Fatal(err)
	}
	defer func() {
		if err := sub.Unsubscribe(); err != nil {
			b.Fatal(err)
		}
	}()
	<-done
}

/*
func BenchmarkSaramaPublish(b *testing.B) {
	if tr := os.Getenv("TRAVIS"); len(tr) > 0 {
		b.Skip()
	}
	brk := sarama.NewBroker(broker.Addrs("127.0.0.1:9092"))
	if err := brk.Connect(); err != nil {
		b.Fatal(err)
	}
	defer func() {
		if err := brk.Disconnect(); err != nil {
			b.Fatal(err)
		}
	}()

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		if err := brk.Publish("test_topic", bm); err != nil {
			b.Fatal(err)
		}
	}

}


func BenchmarkSaramaSubscribe(b *testing.B) {
	if tr := os.Getenv("TRAVIS"); len(tr) > 0 {
		b.Skip()
	}
	brk := sarama.NewBroker(broker.Addrs("127.0.0.1:9092"))
	if err := brk.Connect(); err != nil {
		b.Fatal(err)
	}
	defer func() {
		if err := brk.Disconnect(); err != nil {
			b.Fatal(err)
		}
	}()

	cnt := 0
	done := make(chan struct{})
	fn := func(msg broker.Event) error {
		if cnt == 0 {
			b.ResetTimer()
		}

		cnt++
		if cnt == 10000 {
			close(done)
		}
		return msg.Ack()
	}

	for i := 0; i < 10000; i++ {
		if err := brk.Publish("test_topic", bm); err != nil {
			b.Fatal(err)
		}
	}

	sub, err := brk.Subscribe("test_topic", fn, broker.Queue("test"))
	if err != nil {
		b.Fatal(err)
	}
	defer func() {
		if err := sub.Unsubscribe(); err != nil {
			b.Fatal(err)
		}
	}()

	<-done
}
*/
