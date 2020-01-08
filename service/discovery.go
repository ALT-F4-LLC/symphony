package service

import (
	"context"
	"time"

	"github.com/erkrnt/symphony/protobuff"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc/status"
)

func getServiceByHostname(c protobuff.ManagerClient, hostname string) (*protobuff.Service, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)

	defer cancel()

	fields := &protobuff.GetServiceByHostnameFields{Hostname: hostname}

	service, err := c.GetServiceByHostname(ctx, fields)

	if err != nil {
		return nil, err
	}

	logrus.WithFields(logrus.Fields{"ID": service.ID}).Debug("getServiceByHostname")

	return service, nil
}

func getServiceTypeByName(c protobuff.ManagerClient, typeName string) (*protobuff.ServiceType, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)

	defer cancel()

	fields := &protobuff.GetServiceTypeByNameFields{Name: "block"}

	serviceType, err := c.GetServiceTypeByName(ctx, fields)

	if err != nil {
		return nil, err
	}

	logrus.WithFields(logrus.Fields{"ID": serviceType.ID}).Debug("getServiceTypeByName")

	return serviceType, nil
}

func newService(c protobuff.ManagerClient, hostname string, typeID string) (*protobuff.Service, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)

	defer cancel()

	srvc := &protobuff.NewServiceFields{
		Hostname:      hostname,
		ServiceTypeID: typeID,
	}

	service, err := c.NewService(ctx, srvc)

	if err != nil {
		return nil, err
	}

	logrus.WithFields(logrus.Fields{"ID": service.ID}).Debug("newService")

	return service, nil
}

// Handshake : handles initialization and clustering of instance
func Handshake(c protobuff.ManagerClient, hostname string, typeName string) (*protobuff.Service, error) {
	service, err := getServiceByHostname(c, hostname)

	if err != nil {
		st, _ := status.FromError(err)

		if st.Message() == "invalid_service" {
			serviceType, err := getServiceTypeByName(c, typeName)

			if err != nil {
				return nil, err
			}

			service, err := newService(c, hostname, serviceType.ID)

			if err != nil {
				return nil, err
			}

			return service, nil
		}

		return nil, err
	}

	return service, nil
}
