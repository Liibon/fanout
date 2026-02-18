package main

import (
	"log"
	"net"

	pb "github.com/liibon/fanout/gen/hdsearchv1"
	"google.golang.org/grpc"
)

func main() {
	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalf("listen: %v", err)
	}
	srv := grpc.NewServer()
	pb.RegisterHDSearchServer(srv, &hdSearchServer{})
	log.Printf("leaf listening on :50051")
	if err := srv.Serve(lis); err != nil {
		log.Fatalf("serve: %v", err)
	}
}
