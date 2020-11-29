package service

import (
	"net"

	"github.com/erkrnt/symphony/api"
	"github.com/google/uuid"
	"github.com/hashicorp/serf/serf"
)

// NewSerf : creates a new serf instance
func NewSerf(serviceAddr *net.TCPAddr, serviceID *uuid.UUID, serviceType api.ServiceType) (*serf.Serf, error) {
	conf := serf.DefaultConfig()

	confTags := make(map[string]string)

	confTags["ServiceAddr"] = serviceAddr.String()

	if serviceID != nil {
		confTags["ServiceID"] = serviceID.String()
	}

	confTags["ServiceType"] = serviceType.String()

	conf.Tags = confTags

	conf.Init()

	serfInstance, err := serf.Create(conf)

	if err != nil {
		return nil, err
	}

	return serfInstance, nil
}
