package service

import (
	"net"

	"github.com/erkrnt/symphony/api"
	"github.com/hashicorp/serf/serf"
)

// Service : service member
type Service struct {
	ErrorC chan string
	Key    *Key
	Serf   *serf.Serf
}

// GetSerfMember : retrieves a gossip member from a service
func (s *Service) GetSerfMember(service *api.Service) (*serf.Member, error) {
	var serfMember *serf.Member

	serfMembers := s.Serf.Members()

	for _, sm := range serfMembers {
		if sm.Tags["ServiceID"] == service.ID {
			serfMember = &sm
		}
	}

	return serfMember, nil
}

// StopSerf : stops Serf instance
func (s *Service) StopSerf() error {
	leaveErr := s.Serf.Leave()

	if leaveErr != nil {
		return leaveErr
	}

	shutdownErr := s.Serf.Shutdown()

	if shutdownErr != nil {
		return shutdownErr
	}

	s.Serf = nil

	return nil
}

// RestartSerf : restarts Serf instance
func (s *Service) RestartSerf(serviceAddr *net.TCPAddr, serviceType api.ServiceType) error {
	stopSerfErr := s.StopSerf()

	if stopSerfErr != nil {
		return stopSerfErr
	}

	startSerfErr := s.startSerf(serviceAddr, serviceType)

	if startSerfErr != nil {
		return startSerfErr
	}

	return nil
}

func (s *Service) startSerf(serviceAddr *net.TCPAddr, serviceType api.ServiceType) error {
	conf := serf.DefaultConfig()

	tags := make(map[string]string)

	tags["ServiceAddr"] = serviceAddr.String()

	if s.Key.ServiceID != nil {
		tags["ServiceID"] = s.Key.ServiceID.String()
	}

	tags["ServiceType"] = serviceType.String()

	conf.Tags = tags

	conf.Init()

	cluster, err := serf.Create(conf)

	if err != nil {
		return err
	}

	s.Serf = cluster

	return nil
}
