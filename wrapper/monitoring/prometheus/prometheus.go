package prometheus

import (
	"context"
	"fmt"

	"github.com/micro/go-micro/v2/client"
	"github.com/micro/go-micro/v2/logger"
	"github.com/micro/go-micro/v2/registry"
	"github.com/micro/go-micro/v2/server"
	"github.com/prometheus/client_golang/prometheus"
)

var (
	// default metric prefix
	DefaultMetricPrefix = "micro"
	// default label prefix
	DefaultLabelPrefix = "micro"
)

type wrapper struct {
	opsCounter           *prometheus.CounterVec
	timeCounterSummary   *prometheus.SummaryVec
	timeCounterHistogram *prometheus.HistogramVec
	callFunc             client.CallFunc
	client.Client
}

func getLabels(opts ...server.Option) prometheus.Labels {
	sopts := server.Options{}

	for _, opt := range opts {
		opt(&sopts)
	}

	labels := make(prometheus.Labels, len(sopts.Metadata))
	for k, v := range sopts.Metadata {
		labels[fmt.Sprintf("%s_%s", DefaultLabelPrefix, k)] = v
	}
	if len(sopts.Name) > 0 {
		labels[fmt.Sprintf("%s_%s", DefaultLabelPrefix, "name")] = sopts.Name
	}
	if len(sopts.Id) > 0 {
		labels[fmt.Sprintf("%s_%s", DefaultLabelPrefix, "id")] = sopts.Id
	}
	if len(sopts.Version) > 0 {
		labels[fmt.Sprintf("%s_%s", DefaultLabelPrefix, "version")] = sopts.Version
	}

	return labels
}

func NewClientWrapper() client.Wrapper {
	return func(c client.Client) client.Client {
		opsCounter := prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: fmt.Sprintf("%s_client_request_total", DefaultMetricPrefix),
				Help: "How many requests called, partitioned by method and status",
			},
			[]string{"method", "status"},
		)

		timeCounterSummary := prometheus.NewSummaryVec(
			prometheus.SummaryOpts{
				Name: fmt.Sprintf("%s_client_latency_microseconds", DefaultMetricPrefix),
				Help: "Service client request latencies in microseconds",
			},
			[]string{"method"},
		)

		timeCounterHistogram := prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Name: fmt.Sprintf("%s_client_request_duration_seconds", DefaultMetricPrefix),
				Help: "Service client request time in seconds",
			},
			[]string{"method"},
		)

		for _, collector := range []prometheus.Collector{opsCounter, timeCounterSummary, timeCounterHistogram} {
			if err := prometheus.DefaultRegisterer.Register(collector); err != nil {
				// if already registered, skip fatal
				if are, ok := err.(prometheus.AlreadyRegisteredError); ok {
					collector = are.ExistingCollector
				} else {
					logger.Fatal(err)
				}
			}
		}

		handler := &wrapper{
			timeCounterHistogram: timeCounterHistogram,
			timeCounterSummary:   timeCounterSummary,
			opsCounter:           opsCounter,
			Client:               c,
		}

		return handler
	}
}

func NewCallWrapper() client.CallWrapper {
	return func(fn client.CallFunc) client.CallFunc {
		opsCounter := prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: fmt.Sprintf("%s_call_request_total", DefaultMetricPrefix),
				Help: "How many requests called, partitioned by method and status",
			},
			[]string{"method", "status"},
		)

		timeCounterSummary := prometheus.NewSummaryVec(
			prometheus.SummaryOpts{
				Name: fmt.Sprintf("%s_call_latency_microseconds", DefaultMetricPrefix),
				Help: "Service client request latencies in microseconds",
			},
			[]string{"method"},
		)

		timeCounterHistogram := prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Name: fmt.Sprintf("%s_call_request_duration_seconds", DefaultMetricPrefix),
				Help: "Service client request time in seconds",
			},
			[]string{"method"},
		)

		for _, collector := range []prometheus.Collector{opsCounter, timeCounterSummary, timeCounterHistogram} {
			if err := prometheus.DefaultRegisterer.Register(collector); err != nil {
				// if already registered, skip fatal
				if are, ok := err.(prometheus.AlreadyRegisteredError); ok {
					collector = are.ExistingCollector
				} else {
					logger.Fatal(err)
				}
			}
		}

		handler := &wrapper{
			timeCounterHistogram: timeCounterHistogram,
			timeCounterSummary:   timeCounterSummary,
			opsCounter:           opsCounter,
			callFunc:             fn,
		}

		return handler.CallFunc
	}
}

func (w *wrapper) CallFunc(ctx context.Context, node *registry.Node, req client.Request, rsp interface{}, opts client.CallOptions) error {
	name := fmt.Sprintf("%s.%s", req.Service(), req.Endpoint())

	timer := prometheus.NewTimer(prometheus.ObserverFunc(func(v float64) {
		us := v * 1000000 // make microseconds
		w.timeCounterSummary.WithLabelValues(name).Observe(us)
		w.timeCounterHistogram.WithLabelValues(name).Observe(v)
	}))
	defer timer.ObserveDuration()

	err := w.callFunc(ctx, node, req, rsp, opts)
	if err == nil {
		w.opsCounter.WithLabelValues(name, "success").Inc()
	} else {
		w.opsCounter.WithLabelValues(name, "fail").Inc()
	}

	return err

}

func (w *wrapper) Call(ctx context.Context, req client.Request, rsp interface{}, opts ...client.CallOption) error {
	name := fmt.Sprintf("%s.%s", req.Service(), req.Endpoint())

	timer := prometheus.NewTimer(prometheus.ObserverFunc(func(v float64) {
		us := v * 1000000 // make microseconds
		w.timeCounterSummary.WithLabelValues(name).Observe(us)
		w.timeCounterHistogram.WithLabelValues(name).Observe(v)
	}))
	defer timer.ObserveDuration()

	err := w.Client.Call(ctx, req, rsp, opts...)
	if err == nil {
		w.opsCounter.WithLabelValues(name, "success").Inc()
	} else {
		w.opsCounter.WithLabelValues(name, "fail").Inc()
	}

	return err
}

func (w *wrapper) Stream(ctx context.Context, req client.Request, opts ...client.CallOption) (client.Stream, error) {
	name := fmt.Sprintf("%s.%s", req.Service(), req.Endpoint())

	timer := prometheus.NewTimer(prometheus.ObserverFunc(func(v float64) {
		us := v * 1000000 // make microseconds
		w.timeCounterSummary.WithLabelValues(name).Observe(us)
		w.timeCounterHistogram.WithLabelValues(name).Observe(v)
	}))
	defer timer.ObserveDuration()

	stream, err := w.Client.Stream(ctx, req, opts...)
	if err == nil {
		w.opsCounter.WithLabelValues(name, "success").Inc()
	} else {
		w.opsCounter.WithLabelValues(name, "fail").Inc()
	}

	return stream, err
}

func (w *wrapper) Publish(ctx context.Context, p client.Message, opts ...client.PublishOption) error {
	name := p.Topic()

	timer := prometheus.NewTimer(prometheus.ObserverFunc(func(v float64) {
		us := v * 1000000 // make microseconds
		w.timeCounterSummary.WithLabelValues(name).Observe(us)
		w.timeCounterHistogram.WithLabelValues(name).Observe(v)
	}))
	defer timer.ObserveDuration()

	err := w.Client.Publish(ctx, p, opts...)
	if err == nil {
		w.opsCounter.WithLabelValues(name, "success").Inc()
	} else {
		w.opsCounter.WithLabelValues(name, "fail").Inc()
	}

	return err
}

func NewHandlerWrapper(opts ...server.Option) server.HandlerWrapper {
	labels := getLabels(opts...)

	opsCounter := prometheus.NewCounterVec(
		prometheus.CounterOpts{
			ConstLabels: labels,
			Name:        fmt.Sprintf("%s_handler_request_total", DefaultMetricPrefix),
			Help:        "How many requests processed, partitioned by method and status",
		},
		[]string{"method", "status"},
	)

	timeCounterSummary := prometheus.NewSummaryVec(
		prometheus.SummaryOpts{
			ConstLabels: labels,
			Name:        fmt.Sprintf("%s_handler_latency_microseconds", DefaultMetricPrefix),
			Help:        "Service handler request latencies in microseconds",
		},
		[]string{"method"},
	)

	timeCounterHistogram := prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			ConstLabels: labels,
			Name:        fmt.Sprintf("%s_handler_request_duration_seconds", DefaultMetricPrefix),
			Help:        "Service handler request time in seconds",
		},
		[]string{"method"},
	)

	for _, collector := range []prometheus.Collector{opsCounter, timeCounterSummary, timeCounterHistogram} {
		if err := prometheus.DefaultRegisterer.Register(collector); err != nil {
			// if already registered, skip fatal
			if are, ok := err.(prometheus.AlreadyRegisteredError); ok {
				collector = are.ExistingCollector
			} else {
				logger.Fatal(err)
			}
		}
	}

	handler := &wrapper{
		timeCounterHistogram: timeCounterHistogram,
		timeCounterSummary:   timeCounterSummary,
		opsCounter:           opsCounter,
	}

	return handler.HandlerFunc
}

func (w *wrapper) HandlerFunc(fn server.HandlerFunc) server.HandlerFunc {
	return func(ctx context.Context, req server.Request, rsp interface{}) error {
		name := req.Endpoint()

		timer := prometheus.NewTimer(prometheus.ObserverFunc(func(v float64) {
			us := v * 1000000 // make microseconds
			w.timeCounterSummary.WithLabelValues(name).Observe(us)
			w.timeCounterHistogram.WithLabelValues(name).Observe(v)
		}))
		defer timer.ObserveDuration()

		err := fn(ctx, req, rsp)
		if err == nil {
			w.opsCounter.WithLabelValues(name, "success").Inc()
		} else {
			w.opsCounter.WithLabelValues(name, "fail").Inc()
		}

		return err
	}
}

func NewSubscriberWrapper(opts ...server.Option) server.SubscriberWrapper {
	labels := getLabels(opts...)

	opsCounter := prometheus.NewCounterVec(
		prometheus.CounterOpts{
			ConstLabels: labels,
			Name:        fmt.Sprintf("%s_subscriber_request_total", DefaultMetricPrefix),
			Help:        "How many requests processed, partitioned by topic and status",
		},
		[]string{"topic", "status"},
	)

	timeCounterSummary := prometheus.NewSummaryVec(
		prometheus.SummaryOpts{
			ConstLabels: labels,
			Name:        fmt.Sprintf("%s_subscriber_latency_microseconds", DefaultMetricPrefix),
			Help:        "Service subscriber request latencies in microseconds",
		},
		[]string{"topic"},
	)

	timeCounterHistogram := prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			ConstLabels: labels,
			Name:        fmt.Sprintf("%s_subscriber_request_duration_seconds", DefaultMetricPrefix),
			Help:        "Service subscriber request time in seconds",
		},
		[]string{"topic"},
	)

	for _, collector := range []prometheus.Collector{opsCounter, timeCounterSummary, timeCounterHistogram} {
		if err := prometheus.DefaultRegisterer.Register(collector); err != nil {
			// if already registered, skip fatal
			if are, ok := err.(prometheus.AlreadyRegisteredError); ok {
				collector = are.ExistingCollector
			} else {
				logger.Fatal(err)
			}
		}
	}

	handler := &wrapper{
		timeCounterHistogram: timeCounterHistogram,
		timeCounterSummary:   timeCounterSummary,
		opsCounter:           opsCounter,
	}

	return handler.SubscriberFunc
}

func (w *wrapper) SubscriberFunc(fn server.SubscriberFunc) server.SubscriberFunc {
	return func(ctx context.Context, msg server.Message) error {
		topic := msg.Topic()

		timer := prometheus.NewTimer(prometheus.ObserverFunc(func(v float64) {
			us := v * 1000000 // make microseconds
			w.timeCounterSummary.WithLabelValues(topic).Observe(us)
			w.timeCounterHistogram.WithLabelValues(topic).Observe(v)
		}))
		defer timer.ObserveDuration()

		err := fn(ctx, msg)
		if err == nil {
			w.opsCounter.WithLabelValues(topic, "success").Inc()
		} else {
			w.opsCounter.WithLabelValues(topic, "fail").Inc()
		}

		return err
	}
}
