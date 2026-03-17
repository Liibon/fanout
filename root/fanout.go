package main

import (
	"context"
	"sort"
	"sync"
	"time"

	pb "github.com/liibon/fanout/gen/hdsearchv1"
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
		conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
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

func fanOut(ctx context.Context, leaves []*leafClient, req *pb.SearchRequest, cfg *Config) ([]*pb.SearchResult, error) {
	ch := make(chan leafResult, len(leaves))
	var wg sync.WaitGroup
	for _, lc := range leaves {
		wg.Add(1)
		go func(lc *leafClient) {
			defer wg.Done()
			lctx, cancel := context.WithTimeout(ctx, cfg.PerLeafTimeout)
			defer cancel()
			start := time.Now()
			resp, err := lc.client.Search(lctx, req)
			r := leafResult{leaf: lc.addr, latency: time.Since(start)}
			if err != nil {
				r.err = err
			} else {
				for _, res := range resp.Results {
					if res.VectorId >= 0 {
						r.results = append(r.results, res)
					}
				}
			}
			ch <- r
		}(lc)
	}
	go func() { wg.Wait(); close(ch) }()

	var all []*pb.SearchResult
	for r := range ch {
		if r.err == nil {
			all = append(all, r.results...)
		}
	}
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
