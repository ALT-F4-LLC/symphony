package block

import (
	"context"
	"errors"
	"fmt"
	"io"
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

	go b.listenControl()

	go b.listenHealth()
}

func (b *Block) listenControl() {
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

	logrus.Info("Started block control gRPC socket server.")

	serveErr := grpcServer.Serve(listen)

	if serveErr != nil {
		b.Service.ErrorC <- serveErr
	}
}

func (b *Block) listenHealth() {
	listenAddr := b.Service.Flags.HealthServiceAddr.String()

	listen, err := net.Listen("tcp", listenAddr)

	if err != nil {
		b.Service.ErrorC <- err
	}

	grpcServer := grpc.NewServer()

	healthServer := &service.GRPCServerHealth{}

	api.RegisterHealthServer(grpcServer, healthServer)

	logrus.Info("Started health gRPC server.")

	serveErr := grpcServer.Serve(listen)

	if serveErr != nil {
		b.Service.ErrorC <- serveErr
	}
}

func (b *Block) listenLogicalVolumes(client api.APIServerClient) (func() error, error) {
	serviceID := b.Service.Key.ServiceID.String()

	in := &api.RequestState{
		ServiceID: serviceID,
	}

	lvs, err := client.StateLogicalVolumes(context.Background(), in)

	if err != nil {
		return nil, err
	}

	listener := func() error {
		for {
			lv, err := lvs.Recv()

			if err == io.EOF {
				break
			}

			if err != nil {
				return err
			}

			if lv != nil {
				logrus.Print(lv)
			}
		}

		return nil
	}

	return listener, nil
}

func (b *Block) listenPhysicalVolumes(client api.APIServerClient) (func() error, error) {
	serviceID := b.Service.Key.ServiceID.String()

	in := &api.RequestState{
		ServiceID: serviceID,
	}

	pvs, err := client.StatePhysicalVolumes(context.Background(), in)

	if err != nil {
		return nil, err
	}

	listener := func() error {
		for {
			pv, err := pvs.Recv()

			if err == io.EOF {
				break
			}

			if err != nil {
				return err
			}

			if pv != nil {
				logrus.Print(pv)
			}
		}

		return nil
	}

	return listener, nil
}

func (b *Block) listenResources() {
	apiserverAddr := b.Service.Flags.APIServerAddr.String()

	logicalVolumes := service.APIServerConn{
		Address: apiserverAddr,
		ErrorC:  b.Service.ErrorC,
		Handler: b.listenLogicalVolumes,
	}

	physicalVolumes := service.APIServerConn{
		Address: apiserverAddr,
		ErrorC:  b.Service.ErrorC,
		Handler: b.listenPhysicalVolumes,
	}

	volumeGroups := service.APIServerConn{
		Address: apiserverAddr,
		ErrorC:  b.Service.ErrorC,
		Handler: b.listenVolumeGroups,
	}

	go logicalVolumes.Listen()

	go physicalVolumes.Listen()

	go volumeGroups.Listen()
}

func (b *Block) listenVolumeGroups(client api.APIServerClient) (func() error, error) {
	serviceID := b.Service.Key.ServiceID.String()

	in := &api.RequestState{
		ServiceID: serviceID,
	}

	vgs, err := client.StateVolumeGroups(context.Background(), in)

	if err != nil {
		return nil, err
	}

	listener := func() error {
		for {
			vg, err := vgs.Recv()

			if err == io.EOF {
				break
			}

			if err != nil {
				return err
			}

			if vg != nil {
				logrus.Print(vg)
			}
		}

		return nil
	}

	return listener, nil
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

	b.listenResources()

	return nil
}
