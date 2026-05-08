package main

import (
	"context"
	"log"
	"time"
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

	tp, err := initTracer(ctx, cfg.OtelEndpoint, "usuite-leaf")
	if err != nil {
		log.Fatalf("tracer: %v", err)
	}
	defer func() { _ = tp.Shutdown(context.Background()) }()
	tracer := otel.Tracer("usuite-leaf")

	// Choose backend.
	t0 := time.Now()
	var idx Index
	if cfg.Synthetic {
		log.Printf("leaf %d: synthetic mode (mu=%.2f sigma=%.2f heavyPct=%.3f)",
			cfg.LeafID, cfg.SyntheticMu, cfg.SyntheticSigma, cfg.SyntheticHeavyPct)
		idx, err = NewSyntheticIndex(cfg)
	} else {
		log.Printf("leaf %d: FAISS mode", cfg.LeafID)
		idx, err = NewFaissIndex(cfg)
	}
	if err != nil {
		log.Fatalf("index init: %v", err)
	}
	defer idx.Close()
	log.Printf("leaf %d: index loaded in %v", cfg.LeafID, time.Since(t0))

	lis, err := net.Listen("tcp", cfg.ListenAddr)
	if err != nil {
		log.Fatalf("listen: %v", err)
	}
	srv := grpc.NewServer(
		grpc.StatsHandler(otelgrpc.NewServerHandler()),
	)
	pb.RegisterHDSearchServer(srv, &hdSearchServer{
		cfg:    cfg,
		idx:    idx,
		tracer: tracer,
	})

	go func() {
		mux := http.NewServeMux()
		mux.Handle("/metrics", promhttp.Handler())
		if err := http.ListenAndServe(cfg.MetricsAddr, mux); err != nil {
			log.Printf("metrics: %v", err)
		}
	}()

	log.Printf("leaf %d listening on %s", cfg.LeafID, cfg.ListenAddr)

	go func() {
		if err := srv.Serve(lis); err != nil {
			log.Printf("gRPC serve: %v", err)
		}
	}()

	<-ctx.Done()
	log.Printf("leaf %d shutting down", cfg.LeafID)
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
