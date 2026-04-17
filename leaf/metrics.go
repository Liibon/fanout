package main

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	leafSearchTotal = promauto.NewCounter(prometheus.CounterOpts{
		Namespace: "usuite",
		Subsystem: "leaf",
		Name:      "searches_total",
		Help:      "Total search requests served by this leaf.",
	})
	leafSearchErrors = promauto.NewCounter(prometheus.CounterOpts{
		Namespace: "usuite",
		Subsystem: "leaf",
		Name:      "search_errors_total",
		Help:      "Total errors in leaf search handler.",
	})
	leafSearchDuration = promauto.NewHistogram(prometheus.HistogramOpts{
		Namespace: "usuite",
		Subsystem: "leaf",
		Name:      "search_duration_seconds",
		Help:      "Per-leaf search latency (includes FAISS or synthetic sleep).",
		Buckets:   prometheus.ExponentialBuckets(0.0001, 2, 20),
	})
)
