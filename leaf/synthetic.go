package main

import (
	"math"
	"math/rand"
	"sync"
	"time"
)

// syntheticIndex satisfies Index with configurable latency distributions.
// Latency draws from lognormal(mu, sigma) in milliseconds.
// A heavy-tail mixture models GC pauses and stragglers:
//
//	with probability HeavyPct, draw from lognormal(HeavyMu, HeavySigma) instead.
//
// gRPC handlers run concurrently, so rng must be guarded; *rand.Rand is not
// safe for concurrent use.
type syntheticIndex struct {
	mu         float64
	sigma      float64
	heavyPct   float64
	heavyMu    float64
	heavySigma float64
	rngMu      sync.Mutex
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
	s.rngMu.Lock()
	for i := range ids {
		ids[i] = s.rng.Int63()
		dists[i] = s.rng.Float32()
	}
	s.rngMu.Unlock()
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
	s.rngMu.Lock()
	for i := range ids {
		ids[i] = candidateIDs[i]
		dists[i] = s.rng.Float32()
	}
	s.rngMu.Unlock()
	return ids, dists, nil
}

func (s *syntheticIndex) Close() {}

func (s *syntheticIndex) sampleLatencyMs() float64 {
	s.rngMu.Lock()
	defer s.rngMu.Unlock()
	if s.rng.Float64() < s.heavyPct {
		return math.Exp(s.heavyMu + s.heavySigma*s.rng.NormFloat64())
	}
	return math.Exp(s.mu + s.sigma*s.rng.NormFloat64())
}
