package main

import (
	"log"
	"net"

	pb "github.com/liibon/fanout/gen/hdsearchv1"
	"google.golang.org/grpc"
)

func main() {
	cfg, err := configFromEnv()
	if err != nil {
		log.Fatalf("config: %v", err)
	}
	leaves, err := dialLeaves(cfg.LeafAddrs)
	if err != nil {
		log.Fatalf("dial leaves: %v", err)
	}
	log.Printf("connected to %d leaves", len(leaves))

	lis, err := net.Listen("tcp", cfg.ListenAddr)
	if err != nil {
		log.Fatalf("listen: %v", err)
	}
	srv := grpc.NewServer()
	pb.RegisterHDSearchServer(srv, &hdSearchServer{cfg: cfg, leaves: leaves})
	log.Printf("root listening on %s (fan-out=%d, top-k=%d)", cfg.ListenAddr, cfg.FanOut, cfg.TopK)
	if err := srv.Serve(lis); err != nil {
		log.Fatalf("serve: %v", err)
	}
}
