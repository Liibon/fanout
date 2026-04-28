package main

import (
	"context"
	"sort"
	"sync"
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

func fanOut(
	ctx context.Context,
	tracer trace.Tracer,
	leaves []*leafClient,
	req *pb.SearchRequest,
	cfg *Config,
) ([]*pb.SearchResult, error) {
	ctx, span := tracer.Start(ctx, "root.fanout",
		trace.WithAttributes(attribute.Int("fan_out", len(leaves))))
	defer span.End()

	type hedgeKey struct{ idx int }

	deadline := cfg.PerLeafTimeout
	results := make(chan leafResult, len(leaves)*2)

	var wg sync.WaitGroup

	callLeaf := func(idx int) {
		wg.Add(1)
		go func() {
			defer wg.Done()
			lc := leaves[idx]
			lctx, cancel := context.WithTimeout(ctx, deadline)
			defer cancel()

			start := time.Now()
			resp, err := lc.client.Search(lctx, req)
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
				r.results = resp.Results
			}
			results <- r
		}()
	}

	for i := range leaves {
		callLeaf(i)
	}

	// Hedging: after HedgingDelay, duplicate the slowest outstanding leaf if enabled.
	if cfg.HedgingEnabled {
		go func() {
			select {
			case <-time.After(cfg.HedgingDelay):
				// Re-issue to a random leaf not yet heard from. Simplification: re-call leaf 0.
				callLeaf(0)
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

	return topK(all, int(req.TopK)), nil
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
