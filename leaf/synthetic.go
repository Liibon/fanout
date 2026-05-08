package main

import (
	"math"
	"math/rand"
	"time"
)

// syntheticIndex satisfies Index with configurable latency distributions.
// Latency draws from lognormal(mu, sigma) in milliseconds.
// A heavy-tail mixture models GC pauses and stragglers:
//
//	with probability HeavyPct, draw from lognormal(HeavyMu, HeavySigma) instead.
type syntheticIndex struct {
	mu         float64
	sigma      float64
	heavyPct   float64
	heavyMu    float64
	heavySigma float64
	rng        *rand.Rand
}

func NewSyntheticIndex(cfg *Config) (Index, error) {
	return &syntheticIndex{
		mu:         cfg.SyntheticMu,
		sigma:      cfg.SyntheticSigma,
		heavyPct:   cfg.SyntheticHeavyPct,
		heavyMu:    cfg.SyntheticHeavyMu,
		heavySigma: cfg.SyntheticHeavySigma,
		rng:        rand.New(rand.NewSource(time.Now().UnixNano())),
	}, nil
}

func (s *syntheticIndex) Search(query []float32, k int) ([]int64, []float32, error) {
	delayMs := s.sampleLatencyMs()
	time.Sleep(time.Duration(delayMs*1e6) * time.Nanosecond)

	ids := make([]int64, k)
	dists := make([]float32, k)
	for i := range ids {
		ids[i] = s.rng.Int63()
		dists[i] = s.rng.Float32()
	}
	return ids, dists, nil
}

func (s *syntheticIndex) SearchByIDs(query []float32, candidateIDs []int64, k int) ([]int64, []float32, error) {
	delayMs := s.sampleLatencyMs()
	time.Sleep(time.Duration(delayMs*1e6) * time.Nanosecond)

	if k > len(candidateIDs) {
		k = len(candidateIDs)
	}
	ids := make([]int64, k)
	dists := make([]float32, k)
	for i := range ids {
		ids[i] = candidateIDs[i]
		dists[i] = s.rng.Float32()
	}
	return ids, dists, nil
}

func (s *syntheticIndex) Close() {}

func (s *syntheticIndex) sampleLatencyMs() float64 {
	if s.rng.Float64() < s.heavyPct {
		return lognormal(s.rng, s.heavyMu, s.heavySigma)
	}
	return lognormal(s.rng, s.mu, s.sigma)
}

func lognormal(rng *rand.Rand, mu, sigma float64) float64 {
	return math.Exp(mu + sigma*rng.NormFloat64())
}
