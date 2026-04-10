package main

import (
	"context"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	pb "github.com/liibon/fanout/gen/hdsearchv1"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.21.0"
	"google.golang.org/grpc"
)

func main() {
	cfg, err := configFromEnv()
	if err != nil {
		log.Fatalf("config: %v", err)
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	// OTel tracer.
	tp, err := initTracer(ctx, cfg.OtelEndpoint, "usuite-root")
	if err != nil {
		log.Fatalf("tracer: %v", err)
	}
	defer func() { _ = tp.Shutdown(context.Background()) }()
	tracer := otel.Tracer("usuite-root")

	// Dial leaves.
	leaves, err := dialLeaves(cfg.LeafAddrs)
	if err != nil {
		log.Fatalf("dial leaves: %v", err)
	}
	log.Printf("connected to %d leaves", len(leaves))

	// gRPC server.
	lis, err := net.Listen("tcp", cfg.ListenAddr)
	if err != nil {
		log.Fatalf("listen: %v", err)
	}
	srv := grpc.NewServer(
		grpc.StatsHandler(otelgrpc.NewServerHandler()),
	)
	pb.RegisterHDSearchServer(srv, &hdSearchServer{
		cfg:    cfg,
		leaves: leaves,
		tracer: tracer,
	})

	// Prometheus metrics server.
	go func() {
		mux := http.NewServeMux()
		mux.Handle("/metrics", promhttp.Handler())
		if err := http.ListenAndServe(cfg.MetricsAddr, mux); err != nil {
			log.Printf("metrics server: %v", err)
		}
	}()

	log.Printf("root listening on %s (fan-out=%d, top-k=%d)", cfg.ListenAddr, cfg.FanOut, cfg.TopK)

	go func() {
		if err := srv.Serve(lis); err != nil {
			log.Printf("gRPC serve: %v", err)
		}
	}()

	<-ctx.Done()
	log.Println("shutting down")
	srv.GracefulStop()
}

func initTracer(ctx context.Context, endpoint, service string) (*sdktrace.TracerProvider, error) {
	exp, err := otlptracehttp.New(ctx,
		otlptracehttp.WithEndpoint(endpoint),
		otlptracehttp.WithInsecure(),
	)
	if err != nil {
		return nil, err
	}
	res, _ := resource.New(ctx,
		resource.WithAttributes(semconv.ServiceName(service)),
	)
	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exp),
		sdktrace.WithResource(res),
		sdktrace.WithSampler(sdktrace.AlwaysSample()),
	)
	otel.SetTracerProvider(tp)
	return tp, nil
}
