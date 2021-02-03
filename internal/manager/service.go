package manager

import (
	"errors"
	"math/rand"
	"net"
	"time"

	"github.com/erkrnt/symphony/api"
	"github.com/erkrnt/symphony/internal/utils"
	"github.com/google/uuid"
	consul "github.com/hashicorp/consul/api"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
)

// Manager : defines manager service
type Manager struct {
	ErrorC chan error
	State  *utils.StateMachine

	flags *flags
}

const maxTimeoutMs = 10000

// New : creates a new apiserver instance
func New() (*Manager, error) {
	flags, err := getFlags()

	if err != nil {
		return nil, err
	}

	if flags.Verbose {
		logrus.SetLevel(logrus.DebugLevel)
	}

	rand.Seed(time.Now().UnixNano())

	manager := &Manager{
		ErrorC: make(chan error),
		State:  newState(flags.ConsulAddr),

		flags: flags,
	}

	return manager, nil
}

// Start : handles start of manager service
func (m *Manager) Start() {
	go m.listenGRPC()

	go m.watchLogicalVolumes()
	go m.watchPhysicalVolumes()
	go m.watchVolumeGroups()
}

func (m *Manager) listenGRPC() {
	listen, err := net.Listen("tcp", m.flags.ServiceAddr.String())

	if err != nil {
		m.ErrorC <- err

		return
	}

	grpcServer := grpc.NewServer()

	i, err := NewDiskImageStore("/data")

	if err != nil {
		m.ErrorC <- err

		return
	}

	managerServer := &grpcServerManager{
		image:   i,
		manager: m,
	}

	api.RegisterManagerServer(grpcServer, managerServer)

	logrus.Info("Started TCP server")

	serveErr := grpcServer.Serve(listen)

	if serveErr != nil {
		m.ErrorC <- serveErr

		return
	}
}

func (m *Manager) service(agentService *consul.AgentService) (*api.Service, error) {
	var serviceType api.ServiceType

	switch agentService.Meta["ServiceType"] {
	case "BLOCK":
		serviceType = api.ServiceType_BLOCK
	}

	if serviceType == api.ServiceType_UNKNOWN_SERVICE_TYPE {
		return nil, errors.New("invalid_service_type")
	}

	service := &api.Service{
		ID:   agentService.ID,
		Type: serviceType,
	}

	return service, nil
}

func (m *Manager) services(agentServices map[string]*consul.AgentService) ([]*api.Service, error) {
	services := make([]*api.Service, 0)

	for _, ev := range agentServices {
		s, err := m.service(ev)
		if err != nil {
			return nil, err
		}

		services = append(services, s)
	}

	return services, nil
}

func (m *Manager) watchLogicalVolumes() {
	status := api.ResourceStatus_REVIEW_IN_PROGRESS

	lvs, err := logicalVolumes(m.flags.ConsulAddr)

	if err != nil {
		m.ErrorC <- err

		return
	}

	for _, lv := range lvs {
		if lv.Status == status {
			resourceID, err := uuid.Parse(lv.ID)

			if err != nil {
				m.ErrorC <- err

				return
			}

			eventContext := &ReviewEventContext{
				ResourceID:   resourceID,
				ResourceType: api.ResourceType_LOGICAL_VOLUME,
			}

			go m.State.Send(ReviewInProgressLogicalVolume, eventContext)
		}
	}

	time.Sleep(utils.RandomTimeout(maxTimeoutMs))

	m.watchLogicalVolumes()
}

func (m *Manager) watchPhysicalVolumes() {
	status := api.ResourceStatus_REVIEW_IN_PROGRESS

	pvs, err := physicalVolumes(m.flags.ConsulAddr)

	if err != nil {
		m.ErrorC <- err

		return
	}

	for _, pv := range pvs {
		if pv.Status == status {
			resourceID, err := uuid.Parse(pv.ID)

			if err != nil {
				m.ErrorC <- err

				return
			}

			eventContext := &ReviewEventContext{
				ResourceID:   resourceID,
				ResourceType: api.ResourceType_PHYSICAL_VOLUME,
			}

			go m.State.Send(ReviewInProgressPhysicalVolume, eventContext)
		}
	}

	time.Sleep(utils.RandomTimeout(maxTimeoutMs))

	m.watchPhysicalVolumes()
}

func (m *Manager) watchVolumeGroups() {
	status := api.ResourceStatus_REVIEW_IN_PROGRESS

	vgs, err := volumeGroups(m.flags.ConsulAddr)

	if err != nil {
		m.ErrorC <- err

		return
	}

	for _, vg := range vgs {
		if vg.Status == status {
			resourceID, err := uuid.Parse(vg.ID)

			if err != nil {
				m.ErrorC <- err

				return
			}

			eventContext := &ReviewEventContext{
				ResourceID:   resourceID,
				ResourceType: api.ResourceType_VOLUME_GROUP,
			}

			go m.State.Send(ReviewInProgressVolumeGroup, eventContext)
		}
	}

	time.Sleep(utils.RandomTimeout(maxTimeoutMs))

	m.watchVolumeGroups()
}
