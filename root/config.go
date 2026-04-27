package main

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	ListenAddr      string
	MetricsAddr     string
	OtelEndpoint    string
	LeafAddrs       []string
	FanOut          int
	TopK            int
	PerLeafTimeout  time.Duration
	HedgingEnabled  bool
	HedgingDelay    time.Duration
	RetryEnabled    bool
	MaxRetries      int
}

func configFromEnv() (*Config, error) {
	c := &Config{
		ListenAddr:     getenv("ROOT_LISTEN_ADDR", ":50051"),
		MetricsAddr:    getenv("ROOT_METRICS_ADDR", ":9102"),
		OtelEndpoint:   getenv("OTEL_EXPORTER_OTLP_ENDPOINT", "http://otel-collector:4318"),
		TopK:           10,
		FanOut:         16,
		PerLeafTimeout: 100 * time.Millisecond,
		HedgingEnabled: false,
		HedgingDelay:   20 * time.Millisecond,
		RetryEnabled:   false,
		MaxRetries:     1,
	}

	if s := os.Getenv("LEAF_ADDRS"); s != "" {
		c.LeafAddrs = strings.Split(s, ",")
	}
	if s := os.Getenv("FAN_OUT"); s != "" {
		v, err := strconv.Atoi(s)
		if err != nil {
			return nil, fmt.Errorf("FAN_OUT: %w", err)
		}
		c.FanOut = v
	}
	if s := os.Getenv("TOP_K"); s != "" {
		v, err := strconv.Atoi(s)
		if err != nil {
			return nil, fmt.Errorf("TOP_K: %w", err)
		}
		c.TopK = v
	}
	if s := os.Getenv("PER_LEAF_TIMEOUT_MS"); s != "" {
		v, err := strconv.Atoi(s)
		if err != nil {
			return nil, fmt.Errorf("PER_LEAF_TIMEOUT_MS: %w", err)
		}
		c.PerLeafTimeout = time.Duration(v) * time.Millisecond
	}
	if s := os.Getenv("HEDGING"); s == "true" || s == "1" {
		c.HedgingEnabled = true
	}
	if s := os.Getenv("HEDGING_DELAY_MS"); s != "" {
		v, err := strconv.Atoi(s)
		if err != nil {
			return nil, fmt.Errorf("HEDGING_DELAY_MS: %w", err)
		}
		c.HedgingDelay = time.Duration(v) * time.Millisecond
	}
	if s := os.Getenv("RETRY"); s == "true" || s == "1" {
		c.RetryEnabled = true
	}
	if s := os.Getenv("MAX_RETRIES"); s != "" {
		v, err := strconv.Atoi(s)
		if err != nil {
			return nil, fmt.Errorf("MAX_RETRIES: %w", err)
		}
		c.MaxRetries = v
	}

	if c.FanOut > len(c.LeafAddrs) {
		return nil, fmt.Errorf("FAN_OUT=%d but only %d leaf addresses provided", c.FanOut, len(c.LeafAddrs))
	}

	return c, nil
}

func getenv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}
