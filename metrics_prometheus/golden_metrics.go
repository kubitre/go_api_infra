package metricsprometheus

import (
	"context"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

type MetricRecorder struct {
	httpRequestDurHistogram   *prometheus.HistogramVec
	httpResponseSizeHistogram *prometheus.HistogramVec
	httpRequestsInflight      *prometheus.GaugeVec
}

type MetricConfig struct {
	Prefix          string
	DurationBuckets []float64
	SizeBuckets     []float64
	Registry        prometheus.Registerer
	HandlerIDLabel  string
	StatusCodeLabel string
	MethodLabel     string
	ServiceLabel    string
}

type HTTPReqProperties struct {
	// Service is the service that has served the request.
	Service string
	// ID is the id of the request handler.
	ID string
	// Method is the method of the request.
	Method string
	// Code is the response of the request.
	Code string
}

func (c *MetricConfig) defaults() {
	if len(c.DurationBuckets) == 0 {
		c.DurationBuckets = prometheus.DefBuckets
	}

	if len(c.SizeBuckets) == 0 {
		c.SizeBuckets = prometheus.ExponentialBuckets(100, 10, 8)
	}

	if c.Registry == nil {
		c.Registry = prometheus.DefaultRegisterer
	}

	if c.HandlerIDLabel == "" {
		c.HandlerIDLabel = "handler"
	}

	if c.StatusCodeLabel == "" {
		c.StatusCodeLabel = "code"
	}

	if c.MethodLabel == "" {
		c.MethodLabel = "method"
	}

	if c.ServiceLabel == "" {
		c.ServiceLabel = "service"
	}
}

func NewRecorder(cfg MetricConfig) *MetricRecorder {
	cfg.defaults()

	r := &MetricRecorder{
		httpRequestDurHistogram: prometheus.NewHistogramVec(prometheus.HistogramOpts{
			Namespace: cfg.Prefix,
			Subsystem: "http",
			Name:      "request_duration_seconds",
			Help:      "The latency of the HTTP requests.",
			Buckets:   cfg.DurationBuckets,
		}, []string{cfg.ServiceLabel, cfg.HandlerIDLabel, cfg.MethodLabel, cfg.StatusCodeLabel}),

		httpResponseSizeHistogram: prometheus.NewHistogramVec(prometheus.HistogramOpts{
			Namespace: cfg.Prefix,
			Subsystem: "http",
			Name:      "response_size_bytes",
			Help:      "The size of the HTTP responses.",
			Buckets:   cfg.SizeBuckets,
		}, []string{cfg.ServiceLabel, cfg.HandlerIDLabel, cfg.MethodLabel, cfg.StatusCodeLabel}),

		httpRequestsInflight: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Namespace: cfg.Prefix,
			Subsystem: "http",
			Name:      "requests_inflight",
			Help:      "The number of inflight requests being handled at the same time.",
		}, []string{cfg.ServiceLabel, cfg.HandlerIDLabel}),
	}

	cfg.Registry.MustRegister(
		r.httpRequestDurHistogram,
		r.httpResponseSizeHistogram,
		r.httpRequestsInflight,
	)

	return r
}

func (recorder MetricRecorder) ObserveHTTPRequestDuration(_ context.Context, p HTTPReqProperties, duration time.Duration) {
	recorder.httpRequestDurHistogram.WithLabelValues(p.Service, p.ID, p.Method, p.Code).Observe(duration.Seconds())
}

func (recorder MetricRecorder) ObserveHTTPResponseSize(_ context.Context, p HTTPReqProperties, sizeBytes int64) {
	recorder.httpResponseSizeHistogram.WithLabelValues(p.Service, p.ID, p.Method, p.Code).Observe(float64(sizeBytes))
}

func (recorder MetricRecorder) AddInflightRequests(_ context.Context, p HTTPReqProperties, quantity int) {
	recorder.httpRequestsInflight.WithLabelValues(p.Service, p.ID).Add(float64(quantity))
}
