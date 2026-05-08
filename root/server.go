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
	leaves []*leafClient
	tracer trace.Tracer
}

func (s *hdSearchServer) Search(ctx context.Context, req *pb.SearchRequest) (*pb.SearchResponse, error) {
	start := time.Now()

	ctx, span := s.tracer.Start(ctx, "root.Search",
		trace.WithAttributes(
			attribute.String("request_id", req.RequestId),
			attribute.Int("top_k", int(req.TopK)),
			attribute.Int("fan_out", s.cfg.FanOut),
		))
	defer span.End()

	leaves := s.leaves[:s.cfg.FanOut]
	fo, err := fanOut(ctx, s.tracer, leaves, req, s.cfg)

	elapsed := time.Since(start)
	requestDuration.Observe(elapsed.Seconds())
	requestsTotal.Inc()

	if err != nil {
		requestErrors.Inc()
		return nil, err
	}

	return &pb.SearchResponse{
		Results:         fo.results,
		RespondingLeaf:  "root",
		LatencyUs:       elapsed.Microseconds(),
		ShardsQueried:   int32(fo.shardsQueried),
		ShardsResponded: int32(fo.shardsResponded),
		IndexUs:         fo.indexUs,
		MergeUs:         fo.mergeUs,
	}, nil
}
