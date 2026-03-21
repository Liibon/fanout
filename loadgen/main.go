package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"math"
	"math/rand"
	"time"

	pb "github.com/liibon/fanout/gen/hdsearchv1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {
	rootAddr := flag.String("root", "root:50051", "root gRPC address")
	qps := flag.Float64("qps", 100, "target QPS")
	dim := flag.Int("dim", 128, "query vector dimension")
	topK := flag.Int("top-k", 10, "top-K")
	flag.Parse()

	conn, err := grpc.NewClient(*rootAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("dial: %v", err)
	}
	defer conn.Close()
	client := pb.NewHDSearchClient(conn)
	rng := rand.New(rand.NewSource(7))
	interArrival := time.Duration(float64(time.Second) / *qps)
	next := time.Now()
	fmt.Printf("open-loop load at %.0f QPS\n", *qps)
	for i := 0; ; i++ {
		scheduleTime := next
		if now := time.Now(); scheduleTime.After(now) {
			time.Sleep(scheduleTime.Sub(now))
		}
		query := make([]float32, *dim)
		for j := range query { query[j] = float32(rng.NormFloat64()) }
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		_, err := client.Search(ctx, &pb.SearchRequest{QueryVector: query, TopK: int32(*topK)})
		cancel()
		if err != nil { log.Printf("req %d: %v", i, err) }
		next = scheduleTime.Add(time.Duration(-math.Log(1-rng.Float64()) * float64(interArrival)))
	}
}
