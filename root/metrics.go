package main

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	requestsTotal = promauto.NewCounter(prometheus.CounterOpts{
		Namespace: "usuite",
		Subsystem: "root",
		Name:      "requests_total",
		Help:      "Total search requests received by root.",
	})
	requestErrors = promauto.NewCounter(prometheus.CounterOpts{
		Namespace: "usuite",
		Subsystem: "root",
		Name:      "request_errors_total",
		Help:      "Total failed search requests at root.",
	})
	requestDuration = promauto.NewHistogram(prometheus.HistogramOpts{
		Namespace: "usuite",
		Subsystem: "root",
		Name:      "request_duration_seconds",
		Help:      "End-to-end search latency at root.",
		Buckets:   prometheus.ExponentialBuckets(0.0001, 2, 20),
	})
	leafTimeouts = promauto.NewCounterVec(prometheus.CounterOpts{
		Namespace: "usuite",
		Subsystem: "root",
		Name:      "leaf_timeouts_total",
		Help:      "Per-leaf RPC timeouts observed by root.",
	}, []string{"leaf"})
	leafErrorsTotal = promauto.NewCounter(prometheus.CounterOpts{
		Namespace: "usuite",
		Subsystem: "root",
		Name:      "leaf_errors_total",
		Help:      "Total leaf RPC errors across all fan-out calls.",
	})
)
