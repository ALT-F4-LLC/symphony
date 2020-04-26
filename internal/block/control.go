package block

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"

	"github.com/erkrnt/symphony/api"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
)

type controlServer struct {
	block *Block
}

func (s *controlServer) Join(ctx context.Context, in *api.BlockControlJoinRequest) (*api.BlockControlJoinResponse, error) {
	// TODO : Register the block service with the manager

	res := &api.BlockControlJoinResponse{}

	return res, nil
}

// ControlServer : starts block control server
func ControlServer(b *Block) {
	socketPath := fmt.Sprintf("%s/control.sock", b.Flags.configDir)

	if err := os.RemoveAll(socketPath); err != nil {
		log.Fatal(err)
	}

	lis, err := net.Listen("unix", socketPath)

	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	s := grpc.NewServer()

	cs := &controlServer{
		block: b,
	}

	logrus.Info("Started block control gRPC socket server.")

	api.RegisterBlockControlServer(s, cs)

	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
