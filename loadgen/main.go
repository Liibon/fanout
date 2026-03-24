package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"math"
	"math/rand"
	"sort"
	"time"

	pb "github.com/liibon/fanout/gen/hdsearchv1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {
	rootAddr := flag.String("root", "root:50051", "root gRPC address")
	qps := flag.Float64("qps", 100, "target QPS")
	warmup := flag.Int("warmup", 500, "warmup requests")
	measure := flag.Int("measure", 5000, "measurement requests")
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

	log.Printf("warmup: %d requests", *warmup)
	runPhase(client, rng, *warmup, *dim, *topK, interArrival, false)

	log.Printf("measuring: %d requests", *measure)
	start := time.Now()
	latencies := runPhase(client, rng, *measure, *dim, *topK, interArrival, true)
	elapsed := time.Since(start)

	sort.Slice(latencies, func(i, j int) bool { return latencies[i] < latencies[j] })
	n := len(latencies)
	pct := func(p float64) time.Duration {
		idx := int(math.Ceil(p/100*float64(n))) - 1
		if idx < 0 { idx = 0 }
		if idx >= n { idx = n - 1 }
		return latencies[idx]
	}
	fmt.Printf("qps_achieved: %.1f\n", float64(n)/elapsed.Seconds())
	fmt.Printf("p50: %v\np99: %v\nmax: %v\n",
		pct(50).Round(time.Microsecond),
		pct(99).Round(time.Microsecond),
		latencies[n-1].Round(time.Microsecond))
}

func runPhase(client pb.HDSearchClient, rng *rand.Rand, n, dim, topK int, interArrival time.Duration, record bool) []time.Duration {
	latencies := make([]time.Duration, 0, n)
	next := time.Now()
	for i := 0; i < n; i++ {
		scheduleTime := next
		if now := time.Now(); scheduleTime.After(now) {
			time.Sleep(scheduleTime.Sub(now))
		}
		query := make([]float32, dim)
		for j := range query { query[j] = float32(rng.NormFloat64()) }
		t0 := scheduleTime
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		_, err := client.Search(ctx, &pb.SearchRequest{QueryVector: query, TopK: int32(topK)})
		cancel()
		latency := time.Since(t0)
		if err != nil { latency = 2 * time.Second }
		if record { latencies = append(latencies, latency) }
		next = scheduleTime.Add(time.Duration(-math.Log(1-rng.Float64()) * float64(interArrival)))
	}
	return latencies
}
