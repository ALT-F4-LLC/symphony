package block

import (
	"context"
	"errors"
	"fmt"
	"net"
	"os"
	"time"

	"github.com/erkrnt/symphony/api"
	"github.com/erkrnt/symphony/internal/service"
	"github.com/erkrnt/symphony/internal/utils"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
)

// Block : block service
type Block struct {
	Service *service.Service
	State   *utils.StateMachine
}

const maxTimeoutMs = 10000

// New : creates new block service
func New() (*Block, error) {
	service, err := service.New()

	if err != nil {
		return nil, err
	}

	state := newState(service.Flags.ManagerAddr)

	block := &Block{
		Service: service,
		State:   state,
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

	blockServer := &GRPCServerBlock{
		Block: b,
	}

	healthServer := &service.GRPCServerHealth{}

	api.RegisterBlockServer(grpcServer, blockServer)

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

func (b *Block) watchLogicalVolumes() {
	managerAddr := b.Service.Flags.ManagerAddr

	serviceID := b.Service.Key.ServiceID

	status := api.ResourceStatus_REVIEW_COMPLETED

	reqOpts := &api.RequestLogicalVolumes{
		ServiceID: serviceID.String(),
		Status:    status,
	}

	pvs, err := logicalVolumes(managerAddr, reqOpts)

	if err != nil {
		b.Service.ErrorC <- err

		return
	}

	for _, lv := range pvs {
		resourceID, err := uuid.Parse(lv.ID)

		if err != nil {
			b.Service.ErrorC <- err

			return
		}

		eventContext := &CreateEventContext{
			ResourceID:   resourceID,
			ResourceType: api.ResourceType_LOGICAL_VOLUME,
		}

		go b.State.Send(CreateInProgressLogicalVolume, eventContext)
	}

	time.Sleep(utils.RandomTimeout(maxTimeoutMs))

	b.watchLogicalVolumes()
}

func (b *Block) watchPhysicalVolumes() {
	managerAddr := b.Service.Flags.ManagerAddr

	serviceID := b.Service.Key.ServiceID

	status := api.ResourceStatus_REVIEW_COMPLETED

	reqOpts := &api.RequestPhysicalVolumes{
		ServiceID: serviceID.String(),
		Status:    status,
	}

	pvs, err := physicalVolumes(managerAddr, reqOpts)

	if err != nil {
		b.Service.ErrorC <- err

		return
	}

	for _, pv := range pvs {
		resourceID, err := uuid.Parse(pv.ID)

		if err != nil {
			b.Service.ErrorC <- err

			return
		}

		eventContext := &CreateEventContext{
			ResourceID:   resourceID,
			ResourceType: api.ResourceType_PHYSICAL_VOLUME,
		}

		go b.State.Send(CreateInProgressPhysicalVolume, eventContext)
	}

	time.Sleep(utils.RandomTimeout(maxTimeoutMs))

	b.watchPhysicalVolumes()
}

func (b *Block) watchVolumeGroups() {
	managerAddr := b.Service.Flags.ManagerAddr

	serviceID := b.Service.Key.ServiceID

	status := api.ResourceStatus_REVIEW_COMPLETED

	reqOpts := &api.RequestVolumeGroups{
		ServiceID: serviceID.String(),
		Status:    status,
	}

	vgs, err := volumeGroups(managerAddr, reqOpts)

	if err != nil {
		b.Service.ErrorC <- err

		return
	}

	for _, vg := range vgs {
		resourceID, err := uuid.Parse(vg.ID)

		if err != nil {
			b.Service.ErrorC <- err

			return
		}

		eventContext := &CreateEventContext{
			ResourceID:   resourceID,
			ResourceType: api.ResourceType_VOLUME_GROUP,
		}

		go b.State.Send(CreateInProgressVolumeGroup, eventContext)
	}

	time.Sleep(utils.RandomTimeout(maxTimeoutMs))

	b.watchVolumeGroups()
}

func (b *Block) restart() error {
	apiserverAddr := b.Service.Flags.ManagerAddr

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

	c := api.NewManagerClient(conn)

	serviceID := b.Service.Key.ServiceID.String()

	serviceOptions := &api.RequestService{
		ID: serviceID,
	}

	service, err := c.GetService(ctx, serviceOptions)

	if err != nil {
		return err
	}

	if service == nil {
		return errors.New("invalid_service_id")
	}

	go b.watchLogicalVolumes()
	go b.watchPhysicalVolumes()
	go b.watchVolumeGroups()

	return nil
}
