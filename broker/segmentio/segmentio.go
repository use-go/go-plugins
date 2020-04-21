// Package kafka provides a kafka broker using segmentio
package segmentio

import (
	"context"
	"errors"
	"io"
	"sync"

	"github.com/google/uuid"
	"github.com/micro/go-micro/v2/broker"
	"github.com/micro/go-micro/v2/codec/json"
	"github.com/micro/go-micro/v2/config/cmd"
	"github.com/micro/go-micro/v2/logger"
	kafka "github.com/segmentio/kafka-go"
)

type kBroker struct {
	addrs []string

	readerConfig kafka.ReaderConfig
	writerConfig kafka.WriterConfig

	writers map[string]*kafka.Writer

	connected bool
	sync.RWMutex
	opts broker.Options
}

type subscriber struct {
	k         *kBroker
	group     *kafka.ConsumerGroup
	t         string
	opts      broker.SubscribeOptions
	offset    int64
	gen       *kafka.Generation
	partition int
	handler   broker.Handler
	reader    *kafka.Reader
}

type publication struct {
	t      string
	err    error
	m      *broker.Message
	ctx    context.Context
	gen    *kafka.Generation
	reader *kafka.Reader
	km     kafka.Message
	mp     map[string]map[int]int64 // for commit offsets
}

func init() {
	cmd.DefaultBrokers["kafka"] = NewBroker
}

func (p *publication) Topic() string {
	return p.t
}

func (p *publication) Message() *broker.Message {
	return p.m
}

func (p *publication) Ack() error {
	//return p.gen.CommitOffsets(p.mp)
	return p.reader.CommitMessages(p.ctx, p.km)
}

func (p *publication) Error() error {
	return p.err
}

func (s *subscriber) Options() broker.SubscribeOptions {
	return s.opts
}

func (s *subscriber) Topic() string {
	return s.t
}

func (s *subscriber) Unsubscribe() error {
	//return s.group.Close()
	return s.reader.Close()
}

func (k *kBroker) Address() string {
	if len(k.addrs) > 0 {
		return k.addrs[0]
	}
	return "127.0.0.1:9092"
}

func (k *kBroker) Connect() error {
	k.RLock()
	if k.connected {
		k.RUnlock()
		return nil
	}
	k.RUnlock()

	kaddrs := make([]string, 0, len(k.addrs))
	for _, addr := range k.addrs {
		conn, err := kafka.DialContext(k.opts.Context, "tcp", addr)
		if err != nil {
			continue
		}
		if _, err = conn.Brokers(); err != nil {
			conn.Close()
			continue
		}
		kaddrs = append(kaddrs, addr)
		conn.Close()
	}

	if len(kaddrs) == 0 {
		return errors.New("no available brokers")
	}

	k.Lock()
	k.addrs = kaddrs
	k.readerConfig.Brokers = k.addrs
	k.writerConfig.Brokers = k.addrs
	k.connected = true
	k.Unlock()

	return nil
}

func (k *kBroker) Disconnect() error {
	k.RLock()
	if !k.connected {
		k.RUnlock()
		return nil
	}
	k.RUnlock()

	k.Lock()
	defer k.Unlock()
	for _, writer := range k.writers {
		if err := writer.Close(); err != nil {
			return err
		}
	}

	k.connected = false
	return nil
}

func (k *kBroker) Init(opts ...broker.Option) error {
	for _, o := range opts {
		o(&k.opts)
	}
	var cAddrs []string
	for _, addr := range k.opts.Addrs {
		if len(addr) == 0 {
			continue
		}
		cAddrs = append(cAddrs, addr)
	}
	if len(cAddrs) == 0 {
		cAddrs = []string{"127.0.0.1:9092"}
	}
	k.addrs = cAddrs
	return nil
}

func (k *kBroker) Options() broker.Options {
	return k.opts
}

func (k *kBroker) Publish(topic string, msg *broker.Message, opts ...broker.PublishOption) error {
	buf, err := k.opts.Codec.Marshal(msg)
	if err != nil {
		return err
	}

	k.Lock()
	writer, ok := k.writers[topic]
	if !ok {
		cfg := k.writerConfig
		cfg.Topic = topic
		if err = cfg.Validate(); err != nil {
			k.Unlock()
			return err
		}
		writer = kafka.NewWriter(cfg)
		k.writers[topic] = writer
	}
	k.Unlock()

	return writer.WriteMessages(k.opts.Context, kafka.Message{Value: buf})
}

func (k *kBroker) Subscribe(topic string, handler broker.Handler, opts ...broker.SubscribeOption) (broker.Subscriber, error) {
	opt := broker.SubscribeOptions{
		AutoAck: true,
		Queue:   uuid.New().String(),
	}
	for _, o := range opts {
		o(&opt)
	}

	cfg := k.readerConfig
	cfg.Topic = topic
	cfg.GroupID = opt.Queue
	cfg.WatchPartitionChanges = true
	cfg.MaxAttempts = 1
	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	reader := kafka.NewReader(cfg)
	sub := &subscriber{opts: opt, t: topic, reader: reader}

	go func() {
		for {
			select {
			case <-k.opts.Context.Done():
				return
			default:
				msg, err := reader.FetchMessage(k.opts.Context)
				if err != nil && err == io.EOF {
					return
				} else if err != nil {
					logger.Errorf("[kafka] subscribe error: %v", err)
				}
				p := &publication{t: topic, ctx: k.opts.Context, reader: reader, km: msg}

				var m broker.Message
				eh := k.opts.ErrorHandler
				p.m = &m
				if err := k.opts.Codec.Unmarshal(msg.Value, &m); err != nil {
					p.err = err
					p.m.Body = msg.Value
					if eh != nil {
						eh(p)
					} else {
						if logger.V(logger.ErrorLevel, logger.DefaultLogger) {
							logger.Errorf("[kafka]: failed to unmarshal: %v", err)
						}
					}
					continue
				}
				err = handler(p)
				if err == nil && opt.AutoAck {
					if err = p.Ack(); err != nil {
						logger.Errorf("[kafka]: unable to commit msg: %v", err)
					}
				} else if err != nil {
					p.err = err
					if eh != nil {
						eh(p)
					} else {
						if logger.V(logger.ErrorLevel, logger.DefaultLogger) {
							logger.Errorf("[kafka]: subscriber error: %v", err)
						}
					}
				}

			}
		}
	}()

	return sub, nil
}

/*
func (k *kBroker) Subscribe(topic string, handler broker.Handler, opts ...broker.SubscribeOption) (broker.Subscriber, error) {
	opt := broker.SubscribeOptions{
		AutoAck: true,
		Queue:   uuid.New().String(),
	}
	for _, o := range opts {
		o(&opt)
	}

	gcfg := kafka.ConsumerGroupConfig{
		ID:                    opt.Queue,
		WatchPartitionChanges: true,
		Brokers:               k.readerConfig.Brokers,
		Topics:                []string{topic},
		GroupBalancers:        []kafka.GroupBalancer{kafka.RangeGroupBalancer{}},
	}
	if err := gcfg.Validate(); err != nil {
		return nil, err
	}

	sub := &subscriber{k: k, opts: opt, t: topic, handler: handler}

	chErr := make(chan error)

	go func() {
		for {
			logger.Info("new consumer group")
			group, err := kafka.NewConsumerGroup(gcfg)
			if err != nil {
				chErr <- err
				time.Sleep(1 * time.Second)
				continue
			}
			logger.Info("group next")
			gen, err := group.Next(k.opts.Context)
			if err == kafka.ErrGroupClosed {
				chErr <- nil
				time.Sleep(1 * time.Second)
				continue
			} else if err != nil {
				chErr <- err
				time.Sleep(1 * time.Second)
				continue
			}
			chErr <- nil
			logger.Info("gen assign")
			assignments := gen.Assignments[topic]
			for _, assignment := range assignments {
				partition, offset := assignment.ID, assignment.Offset
				sub.offset = offset
				sub.partition = partition
				sub.gen = gen
				logger.Infof("gen start part %v off %v", partition, offset)
				gen.Start(sub.run)
			}
		}
	}()

	err := <-chErr
	if err != nil {
		return nil, err
	}

	go func() {
		for {
			select {
			case <-k.opts.Context.Done():
				return
			case err := <-chErr:
				if err != nil {
					logger.Error(err)
				}
			}
		}
	}()

	return sub, nil
}

func (s *subscriber) run(ctx context.Context) {
	// create reader for this partition.
	cfg := s.k.readerConfig
	cfg.Topic = s.t
	cfg.Partition = s.partition
	cfg.GroupID = ""
	p := &publication{t: s.t, ctx: s.k.opts.Context, gen: s.gen}
	p.mp = map[string]map[int]int64{p.t: {s.partition: s.offset}}

	reader := kafka.NewReader(cfg)
	defer reader.Close()
	logger.Info("set offset")
	// seek to the last committed offset for this partition.
	reader.SetOffset(s.offset)
	for {
		select {
		case <-ctx.Done():
			return
		default:
			msg, err := reader.ReadMessage(ctx)
			switch err {
			case kafka.ErrGenerationEnded:
				logger.Infof("reader err: %v", err)
				// generation has ended
				if logger.V(logger.DebugLevel, logger.DefaultLogger) {
					logger.Debug("[kafka] subscription closed")
				}
				return
			case nil:
				var m broker.Message
				eh := s.k.opts.ErrorHandler
				p.m = &m
				if err := s.k.opts.Codec.Unmarshal(msg.Value, &m); err != nil {
					p.err = err
					p.m.Body = msg.Value
					if eh != nil {
						eh(p)
					} else {
						if logger.V(logger.ErrorLevel, logger.DefaultLogger) {
							logger.Errorf("[kafka]: failed to unmarshal: %v", err)
						}
					}
					continue
				}
				err = s.handler(p)
				if err == nil && s.opts.AutoAck {
					if err = p.Ack(); err != nil {
						logger.Errorf("[kafka]: unable to commit msg: %v", err)
					}
				} else if err != nil {
					p.err = err
					if eh != nil {
						eh(p)
					} else {
						if logger.V(logger.ErrorLevel, logger.DefaultLogger) {
							logger.Errorf("[kafka]: subscriber error: %v", err)
						}
					}
				}
			}
		}
	}
}
*/
func (k *kBroker) String() string {
	return "kafka"
}

func NewBroker(opts ...broker.Option) broker.Broker {
	options := broker.Options{
		// default to json codec
		Codec:   json.Marshaler{},
		Context: context.Background(),
	}

	for _, o := range opts {
		o(&options)
	}

	var cAddrs []string
	for _, addr := range options.Addrs {
		if len(addr) == 0 {
			continue
		}
		cAddrs = append(cAddrs, addr)
	}
	if len(cAddrs) == 0 {
		cAddrs = []string{"127.0.0.1:9092"}
	}

	readerConfig := kafka.ReaderConfig{}
	if cfg, ok := options.Context.Value(readerConfigKey{}).(kafka.ReaderConfig); ok {
		readerConfig = cfg
	}
	if len(readerConfig.Brokers) == 0 {
		readerConfig.Brokers = cAddrs
	}
	readerConfig.WatchPartitionChanges = true

	writerConfig := kafka.WriterConfig{CompressionCodec: nil}
	if cfg, ok := options.Context.Value(writerConfigKey{}).(kafka.WriterConfig); ok {
		writerConfig = cfg
	}
	if len(writerConfig.Brokers) == 0 {
		writerConfig.Brokers = cAddrs
	}
	writerConfig.BatchSize = 1

	return &kBroker{
		readerConfig: readerConfig,
		writerConfig: writerConfig,
		writers:      make(map[string]*kafka.Writer),
		addrs:        cAddrs,
		opts:         options,
	}
}
