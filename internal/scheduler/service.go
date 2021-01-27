package scheduler

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

// Scheduler : scheduler service
type Scheduler struct {
	Service *service.Service
}

// New : creates new scheduler service
func New() (*Scheduler, error) {
	service, err := service.New()

	if err != nil {
		return nil, err
	}

	scheduler := &Scheduler{
		Service: service,
	}

	return scheduler, nil
}

// Start : handles start of block service
func (s *Scheduler) Start() {
	if s.Service.Key.ServiceID != nil {
		err := s.restart()

		if err != nil {
			s.Service.ErrorC <- err
		}
	}

	go s.listenControl()

	go s.listenGRPC()
}

func (s *Scheduler) listenControl() {
	configDir := s.Service.Flags.ConfigDir

	socketPath := fmt.Sprintf("%s/control.sock", configDir)

	if err := os.RemoveAll(socketPath); err != nil {
		s.Service.ErrorC <- err
	}

	listen, err := net.Listen("unix", socketPath)

	if err != nil {
		s.Service.ErrorC <- err
	}

	grpcServer := grpc.NewServer()

	controlServer := &GRPCServerControl{
		Scheduler: s,
	}

	api.RegisterControlServer(grpcServer, controlServer)

	logrus.Info("Started socket server")

	serveErr := grpcServer.Serve(listen)

	if serveErr != nil {
		s.Service.ErrorC <- serveErr
	}
}

func (s *Scheduler) listenGRPC() {
	listenAddr := s.Service.Flags.ServiceAddr.String()

	listen, err := net.Listen("tcp", listenAddr)

	if err != nil {
		s.Service.ErrorC <- err
	}

	grpcServer := grpc.NewServer()

	healthServer := &service.GRPCServerHealth{}

	schedulerServer := &GRPCServerScheduler{
		Scheduler: s,
	}

	api.RegisterHealthServer(grpcServer, healthServer)

	api.RegisterSchedulerServer(grpcServer, schedulerServer)

	logrus.Info("Started TCP server")

	serveErr := grpcServer.Serve(listen)

	if serveErr != nil {
		s.Service.ErrorC <- serveErr
	}
}

func (s *Scheduler) restart() error {
	apiserverAddr := s.Service.Flags.APIServerAddr

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

	serviceID := s.Service.Key.ServiceID.String()

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
