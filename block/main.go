package main

import (
	"log"
	"net"

	pb "gitlab.drn.io/erikreinert/symphony/proto"

	"google.golang.org/grpc"
)

const (
	device = "/dev/sdb"
	group  = "virtual"
	name   = "test"
	port   = ":50051"
	size   = "1G"
)

// server is used to implement proto.BlockServer.
type server struct{}

func main() {
	// Set static variables
	lis, err := net.Listen("tcp", port)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	s := grpc.NewServer()
	pb.RegisterBlockServer(s, &server{})
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
