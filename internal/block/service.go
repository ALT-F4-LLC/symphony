package block

import (
	"fmt"
	"net"
	"os"

	"github.com/erkrnt/symphony/api"
	"github.com/erkrnt/symphony/internal/service"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
)

// Block : block service
type Block struct {
	ErrorC chan error
	Flags  *Flags
	Key    *service.Key
}

// New : creates new block service
func New() (*Block, error) {
	flags, err := getFlags()

	if err != nil {
		return nil, err
	}

	if flags.Verbose {
		logrus.SetLevel(logrus.DebugLevel)
	}

	key, err := service.GetKey(flags.ConfigDir)

	if err != nil {
		return nil, err
	}

	block := &Block{
		ErrorC: make(chan error),
		Flags:  flags,
		Key:    key,
	}

	return block, nil
}

// Start : handles start of block service
func (b *Block) Start() {
	go b.startControl()

	go b.startHealth()
}

func (b *Block) startControl() {
	socketPath := fmt.Sprintf("%s/control.sock", b.Flags.ConfigDir)

	if err := os.RemoveAll(socketPath); err != nil {
		b.ErrorC <- err
	}

	lis, err := net.Listen("unix", socketPath)

	if err != nil {
		b.ErrorC <- err
	}

	grpcServer := grpc.NewServer()

	controlServer := &GRPCServerControl{
		Block: b,
	}

	api.RegisterControlServer(grpcServer, controlServer)

	logrus.Info("Started block control gRPC socket server.")

	serveErr := grpcServer.Serve(lis)

	if serveErr != nil {
		b.ErrorC <- serveErr
	}
}

func (b *Block) startHealth() {
	listenAddr := b.Flags.HealthServiceAddr.String()

	listen, err := net.Listen("tcp", listenAddr)

	if err != nil {
		b.ErrorC <- err
	}

	grpcServer := grpc.NewServer()

	healthServer := &service.GRPCServerHealth{}

	api.RegisterHealthServer(grpcServer, healthServer)

	logrus.Info("Started health grpc server.")

	serveErr := grpcServer.Serve(listen)

	if serveErr != nil {
		b.ErrorC <- serveErr
	}
}
