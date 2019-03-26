package main

import (
	"context"
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

// CreatePv implements proto.BlockServer
func (s *server) CreatePv(ctx context.Context, in *pb.PvRequest) (*pb.PvReply, error) {
	log.Printf("CreatePv: received %s create request.", in.Device)
	pv, err := PVCreate(in.Device)
	if err != nil {
		return nil, err
	}
	log.Printf("CreatePv: %s successfully created.", in.Device)
	return &pb.PvReply{Message: pv.PvName}, nil
}

// CreatePv implements proto.BlockServer
func (s *server) RemovePv(ctx context.Context, in *pb.PvRequest) (*pb.PvReply, error) {
	log.Printf("RemovePv: received %s remove request.", in.Device)
	err := PVRemove(in.Device)
	if err != nil {
		return nil, err
	}
	log.Printf("RemovePv: %s successfully removed.", in.Device)
	return &pb.PvReply{Message: "success"}, nil
}

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
