package block

import (
	"fmt"
	"log"
	"net"
	"os"

	"github.com/erkrnt/symphony/api"
	"github.com/erkrnt/symphony/internal/pkg/config"
	"github.com/hashicorp/memberlist"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
)

// Block : block node
type Block struct {
	flags      *flags
	key        *config.Key
	memberlist *memberlist.Memberlist
}

// ControlServer : starts block control server
func (b *Block) ControlServer() {
	socketPath := fmt.Sprintf("%s/control.sock", b.flags.configDir)

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

// New : creates new block node
func New() (*Block, error) {
	flags, err := getFlags()

	if err != nil {
		return nil, err
	}

	if flags.verbose {
		logrus.SetLevel(logrus.DebugLevel)
	}

	block := &Block{
		flags: flags,
	}

	return block, nil
}

// RemoteServer : starts remote grpc endpoints
func (b *Block) RemoteServer() {
	lis, err := net.Listen("tcp", b.flags.listenServiceAddr.String())

	if err != nil {
		log.Fatal("Failed to listen")
	}

	s := grpc.NewServer()

	server := &remoteServer{
		block: b,
	}

	logrus.Info("Started block remote gRPC tcp server.")

	api.RegisterBlockRemoteServer(s, server)

	if err := s.Serve(lis); err != nil {
		log.Fatal("Failed to serve")
	}
}
