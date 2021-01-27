package block

import (
	"context"
	"errors"
	"fmt"
	"net"
	"os"

	"github.com/erkrnt/symphony/api"
	"github.com/erkrnt/symphony/internal/service"
	"github.com/erkrnt/symphony/internal/utils"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
)

// Block : block service
type Block struct {
	Service *service.Service
}

// New : creates new block service
func New() (*Block, error) {
	service, err := service.New()

	if err != nil {
		return nil, err
	}

	block := &Block{
		Service: service,
	}

	return block, nil
}

// Start : handles start of block service
func (b *Block) Start() {
	if b.Service.Key.ServiceID != nil {
		err := b.restart()

		if err != nil {
			b.Service.ErrorC <- err
		}
	}

	go b.listenGRPC()

	go b.listenSocket()
}

func (b *Block) listenGRPC() {
	listenAddr := b.Service.Flags.ServiceAddr.String()

	listen, err := net.Listen("tcp", listenAddr)

	if err != nil {
		b.Service.ErrorC <- err
	}

	grpcServer := grpc.NewServer()

	healthServer := &service.GRPCServerHealth{}

	api.RegisterHealthServer(grpcServer, healthServer)

	logrus.Info("Started TCP server")

	serveErr := grpcServer.Serve(listen)

	if serveErr != nil {
		b.Service.ErrorC <- serveErr
	}
}

func (b *Block) listenSocket() {
	configDir := b.Service.Flags.ConfigDir

	socketPath := fmt.Sprintf("%s/control.sock", configDir)

	if err := os.RemoveAll(socketPath); err != nil {
		b.Service.ErrorC <- err
	}

	listen, err := net.Listen("unix", socketPath)

	if err != nil {
		b.Service.ErrorC <- err
	}

	grpcServer := grpc.NewServer()

	controlServer := &GRPCServerControl{
		Block: b,
	}

	api.RegisterControlServer(grpcServer, controlServer)

	logrus.Info("Started socket server")

	serveErr := grpcServer.Serve(listen)

	if serveErr != nil {
		b.Service.ErrorC <- serveErr
	}
}

func (b *Block) restart() error {
	apiserverAddr := b.Service.Flags.APIServerAddr

	if apiserverAddr == nil {
		return errors.New("invalid_apiserver_addr")
	}

	conn, err := utils.NewClientConnTcp(apiserverAddr.String())

	if err != nil {
		return err
	}

	defer conn.Close()

	ctx, cancel := context.WithTimeout(context.Background(), utils.ContextTimeout)

	defer cancel()

	c := api.NewAPIServerClient(conn)

	serviceID := b.Service.Key.ServiceID.String()

	serviceOptions := &api.RequestService{
		ServiceID: serviceID,
	}

	service, err := c.GetService(ctx, serviceOptions)

	if err != nil {
		return err
	}

	if service == nil {
		return errors.New("invalid_service_id")
	}

	return nil
}
