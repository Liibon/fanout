package main

import (
	"testing"

	pb "github.com/liibon/fanout/gen/hdsearchv1"
)

func TestTopK(t *testing.T) {
	results := []*pb.SearchResult{
		{VectorId: 3, Distance: 0.9},
		{VectorId: 1, Distance: 0.1},
		{VectorId: 2, Distance: 0.5},
	}
	got := topK(results, 2)
	if len(got) != 2 { t.Fatalf("want 2, got %d", len(got)) }
	if got[0].VectorId != 1 || got[1].VectorId != 2 { t.Errorf("wrong order") }
}

func TestTopKFewerThanK(t *testing.T) {
	got := topK([]*pb.SearchResult{{VectorId: 1, Distance: 0.5}}, 10)
	if len(got) != 1 { t.Fatalf("want 1, got %d", len(got)) }
}

func TestTopKEmpty(t *testing.T) {
	if len(topK(nil, 5)) != 0 { t.Fatal("want empty") }
}

func TestTopKEqualDistances(t *testing.T) {
	results := []*pb.SearchResult{
		{VectorId: 10, Distance: 1.0},
		{VectorId: 20, Distance: 1.0},
		{VectorId: 30, Distance: 1.0},
	}
	got := topK(results, 2)
	if len(got) != 2 {
		t.Fatalf("want 2, got %d", len(got))
	}
}

func TestTopKAllSameDistance(t *testing.T) {
	var results []*pb.SearchResult
	for i := 0; i < 20; i++ {
		results = append(results, &pb.SearchResult{VectorId: int64(i), Distance: 0.5})
	}
	got := topK(results, 5)
	if len(got) != 5 {
		t.Fatalf("want 5, got %d", len(got))
	}
}
