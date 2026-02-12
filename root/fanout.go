package main

import (
	"context"
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
}

func fanOut(ctx context.Context, leaves []*leafClient, req *pb.SearchRequest, deadline time.Duration) ([]*pb.SearchResult, error) {
	ch := make(chan leafResult, len(leaves))
	var wg sync.WaitGroup
	for _, lc := range leaves {
		wg.Add(1)
		go func(lc *leafClient) {
			defer wg.Done()
			lctx, cancel := context.WithTimeout(ctx, deadline)
			defer cancel()
			resp, err := lc.client.Search(lctx, req)
			if err != nil {
				ch <- leafResult{err: err}
				return
			}
			ch <- leafResult{results: resp.Results}
		}(lc)
	}
	go func() { wg.Wait(); close(ch) }()

	var all []*pb.SearchResult
	for r := range ch {
		if r.err != nil {
			continue
		}
		all = append(all, r.results...)
	}
	return all, nil
}
