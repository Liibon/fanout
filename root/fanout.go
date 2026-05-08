package main

import (
	"context"
	"math"
	"math/rand"
	"sort"
	"sync"
	"sync/atomic"
	"time"

	pb "github.com/liibon/fanout/gen/hdsearchv1"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type leafClient struct {
	addr   string
	client pb.HDSearchClient
}

func dialLeaves(addrs []string) ([]*leafClient, error) {
	clients := make([]*leafClient, 0, len(addrs))
	for _, addr := range addrs {
		conn, err := grpc.NewClient(addr,
			grpc.WithTransportCredentials(insecure.NewCredentials()),
		)
		if err != nil {
			return nil, err
		}
		clients = append(clients, &leafClient{addr: addr, client: pb.NewHDSearchClient(conn)})
	}
	return clients, nil
}

type leafResult struct {
	results []*pb.SearchResult
	err     error
	leaf    string
	latency time.Duration
}

type fanOutResult struct {
	results         []*pb.SearchResult
	shardsQueried   int
	shardsResponded int
	indexUs         int64
	mergeUs         int64
}

// candidatesForLeaf generates a query-seeded random candidate window within
// the given leaf's shard. This mirrors MicroSuite's index stage, which uses
// FLANN/LSH to produce a candidate list per bucket. Here we use a
// deterministic random walk seeded from the query so the root controls
// candidate volume without holding the full dataset. Replace with an IVF or
// LSH index at the root for production-grade recall guarantees.
func candidatesForLeaf(query []float32, leafIdx, numLeaves, numVectors, numCandidates int) []int64 {
	shardSize := numVectors / numLeaves
	offset := int64(leafIdx * shardSize)
	thisShardSize := int64(shardSize)
	if leafIdx == numLeaves-1 {
		thisShardSize = int64(numVectors) - offset
	}

	nc := int64(numCandidates)
	if nc > thisShardSize {
		nc = thisShardSize
	}

	// Seed a per-leaf RNG from the query vector using a simple hash, so the
	// candidate window is deterministic for a given (query, leaf) pair.
	var h uint64 = 14695981039346656037
	for _, f := range query {
		bits := math.Float32bits(f)
		h ^= uint64(bits)
		h *= 1099511628211
	}
	// Mix in the leaf index so each leaf gets a different window.
	h ^= uint64(leafIdx) * 6364136223846793005

	rng := rand.New(rand.NewSource(int64(h)))
	start := rng.Int63n(thisShardSize-nc+1)

	ids := make([]int64, nc)
	for i := range ids {
		ids[i] = offset + start + int64(i)
	}
	return ids
}

func fanOut(
	ctx context.Context,
	tracer trace.Tracer,
	leaves []*leafClient,
	req *pb.SearchRequest,
	cfg *Config,
) (fanOutResult, error) {
	ctx, span := tracer.Start(ctx, "root.fanout",
		trace.WithAttributes(attribute.Int("fan_out", len(leaves))))
	defer span.End()

	// Candidate selection: if NumCandidates > 0, generate per-leaf candidate ID
	// windows so each leaf scores only that subset — the MicroSuite FIXEDCOMP /
	// PercentDataSent mechanism. indexUs records the time spent here.
	var perLeafCandidates [][]int64
	t0Index := time.Now()
	if cfg.NumCandidates > 0 {
		perLeafCandidates = make([][]int64, len(leaves))
		for i := range leaves {
			perLeafCandidates[i] = candidatesForLeaf(
				req.QueryVector, i, cfg.FanOut, cfg.NumVectors, cfg.NumCandidates,
			)
		}
	}
	indexUs := time.Since(t0Index).Microseconds()

	deadline := cfg.PerLeafTimeout
	results := make(chan leafResult, len(leaves)*2)

	// responded[i] is set to 1 when leaf i returns without error.
	responded := make([]atomic.Int32, len(leaves))

	var wg sync.WaitGroup

	callLeaf := func(idx int) {
		wg.Add(1)
		go func() {
			defer wg.Done()
			lc := leaves[idx]
			lctx, cancel := context.WithTimeout(ctx, deadline)
			defer cancel()

			// Build a per-leaf request carrying candidate IDs when available.
			leafReq := req
			if perLeafCandidates != nil {
				leafReq = &pb.SearchRequest{
					QueryVector:  req.QueryVector,
					TopK:         req.TopK,
					RequestId:    req.RequestId,
					CandidateIds: perLeafCandidates[idx],
				}
			}

			start := time.Now()
			resp, err := lc.client.Search(lctx, leafReq)
			elapsed := time.Since(start)

			if err != nil && cfg.RetryEnabled {
				for i := 0; i < cfg.MaxRetries; i++ {
					lctx2, cancel2 := context.WithTimeout(ctx, deadline)
					resp, err = lc.client.Search(lctx2, req)
					cancel2()
					if err == nil {
						break
					}
				}
			}

			var r leafResult
			r.leaf = lc.addr
			r.latency = elapsed
			if err != nil {
				r.err = err
				leafTimeouts.WithLabelValues(lc.addr).Inc()
			} else {
				responded[idx].Store(1)
				r.results = resp.Results
			}
			results <- r
		}()
	}

	for i := range leaves {
		callLeaf(i)
	}

	// Hedging: after HedgingDelay, re-issue to the first leaf that has not yet
	// responded. wg.Add(1) is called here (before the closer starts) so that
	// wg.Wait() cannot return while the hedge goroutine is still pending.
	if cfg.HedgingEnabled {
		wg.Add(1)
		go func() {
			defer wg.Done()
			select {
			case <-time.After(cfg.HedgingDelay):
				for i := range leaves {
					if responded[i].Load() == 0 {
						callLeaf(i)
						return
					}
				}
			case <-ctx.Done():
			}
		}()
	}

	// Close results channel once all goroutines finish.
	go func() {
		wg.Wait()
		close(results)
	}()

	var all []*pb.SearchResult
	errCount := 0
	for r := range results {
		if r.err != nil {
			errCount++
			span.AddEvent("leaf_timeout", trace.WithAttributes(
				attribute.String("leaf", r.leaf),
				attribute.String("error", r.err.Error()),
			))
			continue
		}
		all = append(all, r.results...)
	}

	leafErrorsTotal.Add(float64(errCount))

	t0Merge := time.Now()
	merged := topK(all, int(req.TopK))
	mergeUs := time.Since(t0Merge).Microseconds()

	return fanOutResult{
		results:         merged,
		shardsQueried:   len(leaves),
		shardsResponded: len(leaves) - errCount,
		indexUs:         indexUs,
		mergeUs:         mergeUs,
	}, nil
}

func topK(results []*pb.SearchResult, k int) []*pb.SearchResult {
	sort.Slice(results, func(i, j int) bool {
		return results[i].Distance < results[j].Distance
	})
	if k > len(results) {
		k = len(results)
	}
	return results[:k]
}
