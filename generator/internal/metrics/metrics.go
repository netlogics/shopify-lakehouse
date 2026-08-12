// Package metrics exposes generator counters and gauges in Prometheus
// exposition format on a dedicated HTTP server.
package metrics

import (
	"log/slog"
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	// EventsProduced counts successfully produced events, by topic.
	EventsProduced = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "generator_events_produced_total",
		Help: "Total number of events successfully produced, by topic.",
	}, []string{"topic"})

	// DeliveryErrors counts Kafka publish/delivery errors, by topic.
	DeliveryErrors = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "generator_delivery_errors_total",
		Help: "Total number of Kafka publish or delivery errors, by topic.",
	}, []string{"topic"})

	// BackpressurePauses counts how many times the generator entered a
	// backpressure pause due to a burst of delivery errors.
	BackpressurePauses = promauto.NewCounter(prometheus.CounterOpts{
		Name: "generator_backpressure_pauses_total",
		Help: "Total number of times the generator paused publishing due to backpressure.",
	})

	// BackpressureActive is 1 while the generator is paused for backpressure, 0 otherwise.
	BackpressureActive = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "generator_backpressure_active",
		Help: "1 if the generator is currently paused due to backpressure, 0 otherwise.",
	})
)

// Serve starts the Prometheus /metrics HTTP endpoint on addr (e.g. ":2112").
// It blocks, so call it in a goroutine.
func Serve(addr string) {
	mux := http.NewServeMux()
	mux.Handle("/metrics", promhttp.Handler())
	slog.Info("metrics server listening", "addr", addr)
	if err := http.ListenAndServe(addr, mux); err != nil {
		slog.Error("metrics server stopped", "error", err)
	}
}
