package config

import (
	"log"
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

func (n *Node) startSerf(serviceAddr *net.TCPAddr) error {
	conf := serf.DefaultConfig()

	confTags := make(map[string]string)

	confTags["ServiceAddr"] = serviceAddr.String()

	if n.Key.ServiceID != nil {
		confTags["ServiceID"] = n.Key.ServiceID.String()
	}

	confTags["ServiceType"] = api.ServiceType_MANAGER.String()

	log.Print(confTags)

	conf.Tags = confTags

	conf.Init()

	cluster, err := serf.Create(conf)

	if err != nil {
		return err
	}

	n.Serf = cluster

	return nil

}

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

func (n *Node) RestartSerf(serviceAddr *net.TCPAddr) error {
	stopSerfErr := n.StopSerf()

	if stopSerfErr != nil {
		return stopSerfErr
	}

	startSerfErr := n.startSerf(serviceAddr)

	if startSerfErr != nil {
		return startSerfErr
	}

	return nil
}
