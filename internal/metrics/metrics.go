package metrics

import "github.com/prometheus/client_golang/prometheus"

var ()

// Metricer is the interface for business metrics.
type Metricer interface {
	RecordMetricRequests(path, method string) func()
}

type metricer struct {
	requestCount    *prometheus.CounterVec
	requestDuration *prometheus.HistogramVec
}

func NewMetricer(name, subname string) Metricer {
	if name == "" {
		name = "default"
	}

	m := metricer{
		requestCount: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: name,
				Subsystem: subname,
				Name:      "http_requests_total",
				Help:      "Total number of HTTP requests",
			},
			[]string{"path", "method"},
		),
		requestDuration: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Namespace: name,
				Subsystem: subname,
				Name:      "http_request_duration_seconds",
				Help:      "HTTP request duration in seconds",
				Buckets:   prometheus.DefBuckets,
			},
			[]string{"path", "method"},
		),
	}

	// Register metrics with Prometheus default registry
	prometheus.MustRegister(m.requestCount)
	prometheus.MustRegister(m.requestDuration)

	return &m
}

func (m *metricer) RecordMetricRequests(path, method string) func() {
	m.requestCount.WithLabelValues(path, method).Inc()
	timer := prometheus.NewTimer(m.requestDuration.WithLabelValues(path, method))
	return func() {
		timer.ObserveDuration()
	}
}
