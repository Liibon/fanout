package main

import (
	"context"
	"time"

	pb "github.com/liibon/fanout/gen/hdsearchv1"
)

type hdSearchServer struct {
	pb.UnimplementedHDSearchServer
	cfg *Config
	idx Index
}

func (s *hdSearchServer) Search(ctx context.Context, req *pb.SearchRequest) (*pb.SearchResponse, error) {
	start := time.Now()
	ids, dists, err := s.idx.Search(req.QueryVector, int(req.TopK))
	if err != nil {
		return nil, err
	}
	results := make([]*pb.SearchResult, len(ids))
	for i := range ids {
		results[i] = &pb.SearchResult{VectorId: ids[i], Distance: dists[i]}
	}
	return &pb.SearchResponse{
		Results:        results,
		RespondingLeaf: getenv("HOSTNAME", "leaf"),
		LatencyUs:      time.Since(start).Microseconds(),
	}, nil
}
