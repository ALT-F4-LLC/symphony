package config

import (
	"net"

	"github.com/erkrnt/symphony/api"
	"github.com/hashicorp/serf/serf"
)

// Node : manager node
type Node struct {
	ErrorC chan string
	Key    *Key
	Serf   *serf.Serf
}

// GetMember : retrieves a gossip member from a service
func (n *Node) GetMember(service *api.Service) (*serf.Member, error) {
	var member *serf.Member

	members := n.Serf.Members()

	for _, m := range members {
		if m.Tags["ServiceID"] == service.ID {
			member = &m
		}
	}

	return member, nil
}

// StopSerf : stops Serf instance
func (n *Node) StopSerf() error {
	leaveErr := n.Serf.Leave()

	if leaveErr != nil {
		return leaveErr
	}

	shutdownErr := n.Serf.Shutdown()

	if shutdownErr != nil {
		return shutdownErr
	}

	n.Serf = nil

	return nil
}

// RestartSerf : restarts Serf instance
func (n *Node) RestartSerf(serviceAddr *net.TCPAddr, serviceType api.ServiceType) error {
	stopSerfErr := n.StopSerf()

	if stopSerfErr != nil {
		return stopSerfErr
	}

	startSerfErr := n.startSerf(serviceAddr, serviceType)

	if startSerfErr != nil {
		return startSerfErr
	}

	return nil
}

func (n *Node) startSerf(serviceAddr *net.TCPAddr, serviceType api.ServiceType) error {
	conf := serf.DefaultConfig()

	tags := make(map[string]string)

	tags["ServiceAddr"] = serviceAddr.String()

	if n.Key.ServiceID != nil {
		tags["ServiceID"] = n.Key.ServiceID.String()
	}

	tags["ServiceType"] = serviceType.String()

	conf.Tags = tags

	conf.Init()

	cluster, err := serf.Create(conf)

	if err != nil {
		return err
	}

	n.Serf = cluster

	return nil
}
