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

	log.Printf("leaf %d: FAISS mode", cfg.LeafID)
	idx, err := NewFaissIndex(cfg)
	if err != nil {
		log.Fatalf("index: %v", err)
	}
	defer idx.Close()

	lis, err := net.Listen("tcp", cfg.ListenAddr)
	if err != nil {
		log.Fatalf("listen: %v", err)
	}
	srv := grpc.NewServer()
	pb.RegisterHDSearchServer(srv, &hdSearchServer{cfg: cfg, idx: idx})
	log.Printf("leaf %d listening on %s", cfg.LeafID, cfg.ListenAddr)
	if err := srv.Serve(lis); err != nil {
		log.Fatalf("serve: %v", err)
	}
}
