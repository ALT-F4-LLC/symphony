package main

import (
	"net"

	"github.com/erkrnt/symphony/protobuff"
	"github.com/erkrnt/symphony/services"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
)

const (
	port = ":50051"
)

type blockServer struct{}

func main() {
	flags := getFlags()

	if flags.Verbose {
		logrus.SetLevel(logrus.DebugLevel)
	}

	conn, err := grpc.Dial(flags.Manager, grpc.WithInsecure())

	if err != nil {
		panic(err)
	}

	defer conn.Close()

	client := protobuff.NewManagerClient(conn)

	service, err := services.Handshake(client, flags.Hostname, "block")

	if err != nil {
		panic(err)
	}

	lis, err := net.Listen("tcp", port)

	if err != nil {
		logrus.Fatal("Failed to listen: %v", err)
	}

	s := grpc.NewServer()

	protobuff.RegisterBlockServer(s, &blockServer{})

	logrus.WithFields(logrus.Fields{"ID": service.ID, "Port": port}).Info("Started block service.")

	if err := s.Serve(lis); err != nil {
		logrus.Fatal("Failed to serve: %v", err)
	}
}
