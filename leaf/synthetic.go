package main

import (
	"context"
	"math"
	"math/rand"
	"time"
)

type syntheticIndex struct {
	mu    float64
	sigma float64
	rng   *rand.Rand
}

func NewSyntheticIndex(mu, sigma float64) *syntheticIndex {
	return &syntheticIndex{mu: mu, sigma: sigma, rng: rand.New(rand.NewSource(42))}
}

func (s *syntheticIndex) Search(ctx context.Context, query []float32, k int) ([]int64, []float32, error) {
	delay := time.Duration(math.Exp(s.mu+s.sigma*s.rng.NormFloat64()) * float64(time.Millisecond))
	select {
	case <-time.After(delay):
	case <-ctx.Done():
		return nil, nil, ctx.Err()
	}
	ids := make([]int64, k)
	dists := make([]float32, k)
	for i := range ids {
		ids[i] = int64(i)
		dists[i] = float32(i)
	}
	return ids, dists, nil
}

func (s *syntheticIndex) Close() {}
