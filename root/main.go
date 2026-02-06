package main

import (
	"log"
	"net"
	"strings"

	pb "github.com/liibon/fanout/gen/hdsearchv1"
	"google.golang.org/grpc"
)

func main() {
	addrs := strings.Split("leaf-0:50051,leaf-1:50051,leaf-2:50051,leaf-3:50051", ",")
	leaves, err := dialLeaves(addrs)
	if err != nil {
		log.Fatalf("dial leaves: %v", err)
	}

	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalf("listen: %v", err)
	}
	srv := grpc.NewServer()
	pb.RegisterHDSearchServer(srv, &hdSearchServer{leaves: leaves})
	log.Printf("root listening on :50051")
	if err := srv.Serve(lis); err != nil {
		log.Fatalf("serve: %v", err)
	}
}
