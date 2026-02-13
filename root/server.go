package main

import (
	"context"
	"time"

	pb "github.com/liibon/fanout/gen/hdsearchv1"
)

type hdSearchServer struct {
	pb.UnimplementedHDSearchServer
	cfg    *Config
	leaves []*leafClient
}

func (s *hdSearchServer) Search(ctx context.Context, req *pb.SearchRequest) (*pb.SearchResponse, error) {
	start := time.Now()
	leaves := s.leaves[:s.cfg.FanOut]
	results, err := fanOut(ctx, leaves, req, s.cfg.PerLeafTimeout)
	if err != nil {
		return nil, err
	}
	return &pb.SearchResponse{
		Results:        results,
		RespondingLeaf: "root",
		LatencyUs:      time.Since(start).Microseconds(),
	}, nil
}
