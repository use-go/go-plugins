package rabbitmq_test

import (
	"context"
	"os"
	"testing"

	micro "github.com/micro/go-micro/v2"
	broker "github.com/micro/go-micro/v2/broker"
	server "github.com/micro/go-micro/v2/server"
	rabbitmq "github.com/micro/go-plugins/broker/rabbitmq/v2"
)

type Example struct{}

func (e *Example) Handler(ctx context.Context, r interface{}) error {
	return nil
}

func TestDurable(t *testing.T) {
	if tr := os.Getenv("TRAVIS"); len(tr) > 0 {
		t.Skip()
	}
	rabbitmq.DefaultRabbitURL = "amqp://rabbitmq:rabbitmq@172.18.0.2:5672"
	brkrSub := broker.NewSubscribeOptions(
		broker.Queue("queue.default"),
		broker.DisableAutoAck(),
		rabbitmq.DurableQueue(),
	)

	b := rabbitmq.NewBroker()
	b.Init()
	b.Connect()
	s := server.NewServer(server.Broker(b))

	service := micro.NewService(
		micro.Server(s),
		micro.Broker(b),
	)
	h := &Example{}
	// Register a subscriber
	micro.RegisterSubscriber(
		"topic",
		service.Server(),
		h.Handler,
		server.SubscriberContext(brkrSub.Context),
		server.SubscriberQueue("queue.default"),
	)

	//service.Init()

	if err := service.Run(); err != nil {
		t.Fatal(err)
	}

}
