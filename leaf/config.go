package main

import (
	"fmt"
	"os"
	"strconv"
)

type Config struct {
	ListenAddr  string
	MetricsAddr string
	OtelEndpoint string
	LeafID      int
	NumLeaves   int

	// Dataset
	DatasetPath string
	NumVectors  int
	Dim         int

	// Synthetic mode
	Synthetic            bool
	SyntheticMu          float64
	SyntheticSigma       float64
	SyntheticHeavyPct    float64
	SyntheticHeavyMu     float64
	SyntheticHeavySigma  float64
}

func configFromEnv() (*Config, error) {
	c := &Config{
		ListenAddr:          getenv("LEAF_LISTEN_ADDR", ":50051"),
		MetricsAddr:         getenv("LEAF_METRICS_ADDR", ":9103"),
		OtelEndpoint:        getenv("OTEL_EXPORTER_OTLP_ENDPOINT", "http://otel-collector:4318"),
		DatasetPath:         getenv("DATASET_PATH", "/data/dataset.bin"),
		NumVectors:          1_000_000,
		Dim:                 128,
		NumLeaves:           16,
		SyntheticMu:         2.5,
		SyntheticSigma:      0.6,
		SyntheticHeavyPct:   0.01,
		SyntheticHeavyMu:    5.0,
		SyntheticHeavySigma: 0.3,
	}

	if s := os.Getenv("LEAF_ID"); s != "" {
		v, err := strconv.Atoi(s)
		if err != nil {
			return nil, fmt.Errorf("LEAF_ID: %w", err)
		}
		c.LeafID = v
	}
	if s := os.Getenv("NUM_LEAVES"); s != "" {
		v, err := strconv.Atoi(s)
		if err != nil {
			return nil, fmt.Errorf("NUM_LEAVES: %w", err)
		}
		c.NumLeaves = v
	}
	if s := os.Getenv("NUM_VECTORS"); s != "" {
		v, err := strconv.Atoi(s)
		if err != nil {
			return nil, fmt.Errorf("NUM_VECTORS: %w", err)
		}
		c.NumVectors = v
	}
	if s := os.Getenv("DIM"); s != "" {
		v, err := strconv.Atoi(s)
		if err != nil {
			return nil, fmt.Errorf("DIM: %w", err)
		}
		c.Dim = v
	}
	if s := os.Getenv("SYNTHETIC"); s == "true" || s == "1" {
		c.Synthetic = true
	}
	if s := os.Getenv("SYNTHETIC_MU"); s != "" {
		v, err := strconv.ParseFloat(s, 64)
		if err != nil {
			return nil, fmt.Errorf("SYNTHETIC_MU: %w", err)
		}
		c.SyntheticMu = v
	}
	if s := os.Getenv("SYNTHETIC_SIGMA"); s != "" {
		v, err := strconv.ParseFloat(s, 64)
		if err != nil {
			return nil, fmt.Errorf("SYNTHETIC_SIGMA: %w", err)
		}
		c.SyntheticSigma = v
	}
	if s := os.Getenv("SYNTHETIC_HEAVY_PCT"); s != "" {
		v, err := strconv.ParseFloat(s, 64)
		if err != nil {
			return nil, fmt.Errorf("SYNTHETIC_HEAVY_PCT: %w", err)
		}
		c.SyntheticHeavyPct = v
	}

	return c, nil
}

func getenv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}
