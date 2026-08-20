package handler

import (
	"net/http"
	"strconv"

	"github.com/wqyjh/zero-contrib/handler/response"
	"github.com/zeromicro/go-zero/core/metric"
	"github.com/zeromicro/go-zero/core/timex"
)

// PrometheusHandler returns a middleware that reports stats to prometheus.
func PrometheusHandler(opts ...PrometheusOption) func(http.Handler) http.Handler {
	options := prometheusOptions{
		namespace: "http_server",
		subsystem: "requests",
		buckets:   []float64{5, 10, 25, 50, 100, 250, 500, 750, 1000},
	}
	for _, o := range opts {
		o(&options)
	}

	metricServerReqDur := metric.NewHistogramVec(&metric.HistogramVecOpts{
		Namespace: options.namespace,
		Subsystem: options.subsystem,
		Name:      "duration_ms",
		Help:      "http server requests duration(ms).",
		Labels:    []string{"path", "method", "code"},
		Buckets:   options.buckets,
	})

	metricServerReqCodeTotal := metric.NewCounterVec(&metric.CounterVecOpts{
		Namespace: options.namespace,
		Subsystem: options.subsystem,
		Name:      "code_total",
		Help:      "http server requests error count.",
		Labels:    []string{"path", "method", "code"},
	})

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			startTime := timex.Now()
			cw := response.NewWithCodeResponseWriter(w)
			defer func() {
				code := strconv.Itoa(cw.Code)
				path := r.URL.Path
				method := r.Method
				metricServerReqDur.Observe(timex.Since(startTime).Milliseconds(), path, method, code)
				metricServerReqCodeTotal.Inc(path, method, code)
			}()

			next.ServeHTTP(cw, r)
		})
	}
}

type prometheusOptions struct {
	namespace string
	subsystem string
	buckets   []float64
}

// PrometheusOption allows for managing prometheus options.
type PrometheusOption func(*prometheusOptions)

// WithBuckets sets the buckets for the prometheus metrics.
func WithBuckets(buckets []float64) PrometheusOption {
	return func(o *prometheusOptions) {
		o.buckets = buckets
	}
}

// WithNamespace sets the namespace for the prometheus metrics.
func WithNamespace(namespace string) PrometheusOption {
	return func(o *prometheusOptions) {
		o.namespace = namespace
	}
}

// WithSubsystem sets the subsystem for the prometheus metrics.
func WithSubsystem(subsystem string) PrometheusOption {
	return func(o *prometheusOptions) {
		o.subsystem = subsystem
	}
}
