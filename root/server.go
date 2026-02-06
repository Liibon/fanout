package main

import (
	"context"

	pb "github.com/liibon/fanout/gen/hdsearchv1"
)

type hdSearchServer struct {
	pb.UnimplementedHDSearchServer
	leaves []*leafClient
}

func (s *hdSearchServer) Search(ctx context.Context, req *pb.SearchRequest) (*pb.SearchResponse, error) {
	return &pb.SearchResponse{}, nil
}
