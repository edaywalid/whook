// Package metrics exposes Prometheus counters and histograms for ingest and
// delivery, plus an HTTP handler that serves them.
package metrics

import (
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// Metrics holds the gateway's collectors registered on a private registry.
type Metrics struct {
	registry         *prometheus.Registry
	eventsCaptured   *prometheus.CounterVec
	ingestDuration   prometheus.Histogram
	deliveryAttempts *prometheus.CounterVec
	deadLetters      prometheus.Counter
}

// New builds and registers the collectors.
func New() *Metrics {
	reg := prometheus.NewRegistry()
	m := &Metrics{
		registry: reg,
		eventsCaptured: prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: "whook_events_captured_total",
			Help: "Inbound webhooks captured, by source and status.",
		}, []string{"source", "status"}),
		ingestDuration: prometheus.NewHistogram(prometheus.HistogramOpts{
			Name:    "whook_ingest_duration_seconds",
			Help:    "Time to capture and persist an inbound request.",
			Buckets: prometheus.DefBuckets,
		}),
		deliveryAttempts: prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: "whook_delivery_attempts_total",
			Help: "Delivery attempts, by outcome.",
		}, []string{"outcome"}),
		deadLetters: prometheus.NewCounter(prometheus.CounterOpts{
			Name: "whook_dead_letters_total",
			Help: "Deliveries moved to dead-letter.",
		}),
	}
	reg.MustRegister(m.eventsCaptured, m.ingestDuration, m.deliveryAttempts, m.deadLetters)
	reg.MustRegister(collectors.NewGoCollector())
	return m
}

// Handler serves the metrics in Prometheus exposition format.
func (m *Metrics) Handler() http.Handler {
	return promhttp.HandlerFor(m.registry, promhttp.HandlerOpts{})
}

// RecordCapture counts a captured event and observes how long capture took.
func (m *Metrics) RecordCapture(source, status string, seconds float64) {
	m.eventsCaptured.WithLabelValues(source, status).Inc()
	m.ingestDuration.Observe(seconds)
}

// RecordDelivery counts a delivery attempt by outcome.
func (m *Metrics) RecordDelivery(outcome string) {
	m.deliveryAttempts.WithLabelValues(outcome).Inc()
}

// RecordDeadLetter counts a delivery moved to dead-letter.
func (m *Metrics) RecordDeadLetter() {
	m.deadLetters.Inc()
}
