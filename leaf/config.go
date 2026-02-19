package main

import (
	"fmt"
	"os"
	"strconv"
)

type Config struct {
	ListenAddr  string
	LeafID      int
	NumLeaves   int
	DatasetPath string
	NumVectors  int
	Dim         int
}

func configFromEnv() (*Config, error) {
	c := &Config{
		ListenAddr:  ":50051",
		DatasetPath: "/data/dataset.bin",
		NumVectors:  1_000_000,
		Dim:         128,
		NumLeaves:   16,
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
	return c, nil
}

func getenv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}
