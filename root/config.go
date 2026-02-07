package main

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	ListenAddr     string
	LeafAddrs      []string
	FanOut         int
	TopK           int
	PerLeafTimeout time.Duration
}

func configFromEnv() (*Config, error) {
	c := &Config{
		ListenAddr:     ":50051",
		TopK:           10,
		FanOut:         4,
		PerLeafTimeout: 100 * time.Millisecond,
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
	return c, nil
}

func getenv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}
