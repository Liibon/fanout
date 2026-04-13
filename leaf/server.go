package main

import (
	"context"
	"time"

	pb "github.com/liibon/fanout/gen/hdsearchv1"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

type hdSearchServer struct {
	pb.UnimplementedHDSearchServer
	cfg    *Config
	idx    Index
	tracer trace.Tracer
}

func (s *hdSearchServer) Search(ctx context.Context, req *pb.SearchRequest) (*pb.SearchResponse, error) {
	start := time.Now()

	_, span := s.tracer.Start(ctx, "leaf.Search",
		trace.WithAttributes(
			attribute.Int("leaf_id", s.cfg.LeafID),
			attribute.Int("top_k", int(req.TopK)),
			attribute.Bool("synthetic", s.cfg.Synthetic),
		))
	defer span.End()

	ids, dists, err := s.idx.Search(req.QueryVector, int(req.TopK))
	elapsed := time.Since(start)
	leafSearchDuration.Observe(elapsed.Seconds())
	leafSearchTotal.Inc()

	if err != nil {
		leafSearchErrors.Inc()
		return nil, err
	}

	results := make([]*pb.SearchResult, len(ids))
	for i := range ids {
		results[i] = &pb.SearchResult{
			VectorId: ids[i],
			Distance: dists[i],
		}
	}

	return &pb.SearchResponse{
		Results:        results,
		RespondingLeaf: getenv("HOSTNAME", "leaf"),
		LatencyUs:      elapsed.Microseconds(),
	}, nil
}
