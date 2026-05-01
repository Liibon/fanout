// loadgen is an open-loop load generator for fanout.
//
// Methodology:
//   - Arrivals follow a Poisson process at the target QPS (inter-arrival = Exp(1/QPS)).
//   - Warmup phase issues WarmupReqs requests; latencies are discarded.
//   - Measurement window issues MeasureReqs requests; latencies are recorded.
//   - Coordinated omission is avoided: clock for each request starts at the
//     scheduled send time, not the actual send time.
//   - If the achieved QPS is < 95% of the target, the run is marked invalid
//     and the process exits 1.
//   - All inputs that produced the result are printed alongside the percentiles.
package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"math"
	"math/rand"
	"os"
	"runtime"
	"sort"
	"time"

	pb "github.com/liibon/fanout/gen/hdsearchv1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {
	rootAddr := flag.String("root", "root:50051", "root gRPC address")
	qps := flag.Float64("qps", 100, "target queries per second (open-loop Poisson)")
	warmup := flag.Int("warmup", 500, "warmup requests (latencies discarded)")
	measure := flag.Int("measure", 5000, "measurement window requests")
	dim := flag.Int("dim", 128, "query vector dimension")
	topK := flag.Int("top-k", 10, "top-K to request")
	seed := flag.Int64("seed", 7, "RNG seed for query vectors")
	flag.Parse()

	conn, err := grpc.NewClient(*rootAddr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		log.Fatalf("dial root: %v", err)
	}
	defer conn.Close()
	client := pb.NewHDSearchClient(conn)

	rng := rand.New(rand.NewSource(*seed))

	// Print run inputs.
	hostname, _ := os.Hostname()
	fmt.Printf("=== fanout loadgen ===\n")
	fmt.Printf("root:       %s\n", *rootAddr)
	fmt.Printf("qps_target: %.1f\n", *qps)
	fmt.Printf("warmup:     %d requests\n", *warmup)
	fmt.Printf("measure:    %d requests\n", *measure)
	fmt.Printf("dim:        %d\n", *dim)
	fmt.Printf("top_k:      %d\n", *topK)
	fmt.Printf("seed:       %d\n", *seed)
	fmt.Printf("go:         %s\n", runtime.Version())
	fmt.Printf("host:       %s\n", hostname)
	fmt.Println()

	interArrival := time.Duration(float64(time.Second) / *qps)

	// Warmup.
	log.Printf("warmup: %d requests at %.0f QPS", *warmup, *qps)
	runPhase(client, rng, *warmup, *dim, *topK, interArrival, false)

	// Measurement.
	log.Printf("measuring: %d requests at %.0f QPS", *measure, *qps)
	start := time.Now()
	latencies := runPhase(client, rng, *measure, *dim, *topK, interArrival, true)
	elapsed := time.Since(start)

	achievedQPS := float64(len(latencies)) / elapsed.Seconds()
	valid := achievedQPS >= *qps*0.95

	printResults(latencies, achievedQPS, *qps, valid)

	if !valid {
		fmt.Fprintf(os.Stderr,
			"\nINVALID RUN: achieved QPS %.1f < 95%% of target %.1f\n",
			achievedQPS, *qps)
		os.Exit(1)
	}
}

func runPhase(
	client pb.HDSearchClient,
	rng *rand.Rand,
	n, dim, topK int,
	interArrival time.Duration,
	record bool,
) []time.Duration {
	latencies := make([]time.Duration, 0, n)
	next := time.Now()

	for i := 0; i < n; i++ {
		scheduleTime := next
		now := time.Now()
		if scheduleTime.After(now) {
			time.Sleep(scheduleTime.Sub(now))
		}

		query := randomVector(rng, dim)
		reqID := fmt.Sprintf("lg-%d", i)

		// Latency starts at scheduled time to avoid coordinated omission.
		t0 := scheduleTime
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		_, err := client.Search(ctx, &pb.SearchRequest{
			QueryVector: query,
			TopK:        int32(topK),
			RequestId:   reqID,
		})
		cancel()
		latency := time.Since(t0)

		if err != nil {
			// Count timeout as full 2-second latency to expose tail.
			latency = 2 * time.Second
		}

		if record {
			latencies = append(latencies, latency)
		}

		next = scheduleTime.Add(poissonInterval(rng, interArrival))
	}
	return latencies
}

func poissonInterval(rng *rand.Rand, mean time.Duration) time.Duration {
	// Exponential inter-arrival for Poisson process.
	return time.Duration(-math.Log(1-rng.Float64()) * float64(mean))
}

func randomVector(rng *rand.Rand, dim int) []float32 {
	v := make([]float32, dim)
	for i := range v {
		v[i] = float32(rng.NormFloat64())
	}
	return v
}

func printResults(latencies []time.Duration, achievedQPS, targetQPS float64, valid bool) {
	sort.Slice(latencies, func(i, j int) bool { return latencies[i] < latencies[j] })
	n := len(latencies)

	pct := func(p float64) time.Duration {
		idx := int(math.Ceil(p/100*float64(n))) - 1
		if idx < 0 {
			idx = 0
		}
		if idx >= n {
			idx = n - 1
		}
		return latencies[idx]
	}

	status := "VALID"
	if !valid {
		status = "INVALID"
	}

	fmt.Printf("=== Results (%s) ===\n", status)
	fmt.Printf("samples:      %d\n", n)
	fmt.Printf("qps_target:   %.1f\n", targetQPS)
	fmt.Printf("qps_achieved: %.1f\n", achievedQPS)
	fmt.Printf("p50:          %v\n", pct(50).Round(time.Microsecond))
	fmt.Printf("p90:          %v\n", pct(90).Round(time.Microsecond))
	fmt.Printf("p95:          %v\n", pct(95).Round(time.Microsecond))
	fmt.Printf("p99:          %v\n", pct(99).Round(time.Microsecond))
	fmt.Printf("p99.9:        %v\n", pct(99.9).Round(time.Microsecond))
	fmt.Printf("max:          %v\n", latencies[n-1].Round(time.Microsecond))
}
