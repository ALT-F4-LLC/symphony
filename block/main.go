package main

import (
	"log"
	"net"

	pb "gitlab.drn.io/erikreinert/symphony/proto"

	"google.golang.org/grpc"
)

const (
	port = ":50051"
)

// server is used to implement proto.BlockServer.
type blockServer struct {
}

func main() {
	// Set static variables
	lis, err := net.Listen("tcp", port)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	s := grpc.NewServer()
	pb.RegisterBlockServer(s, &blockServer{})
	log.Printf("Started block service.")
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
