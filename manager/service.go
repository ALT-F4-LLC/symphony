package main

import (
	"context"

	"github.com/erkrnt/symphony/protobuff"
	"github.com/erkrnt/symphony/schemas"
	"github.com/erkrnt/symphony/services"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/golang/protobuf/ptypes"
)

func (s *managerServer) GetServiceByHostname(ctx context.Context, in *protobuff.GetServiceByHostnameFields) (*protobuff.Service, error) {
	service, err := getServiceByHostname(s.db, in.Hostname)

	if err != nil {
		return nil, services.HandleProtoError(err)
	}

	if service == nil {
		err := status.Error(codes.NotFound, "invalid_service")
		return nil, services.HandleProtoError(err)
	}

	createdAt, err := ptypes.TimestampProto(service.CreatedAt)

	if err != nil {
		return nil, services.HandleProtoError(err)
	}

	updatedAt, err := ptypes.TimestampProto(service.UpdatedAt)

	if err != nil {
		return nil, services.HandleProtoError(err)
	}

	srvc := &protobuff.Service{
		CreatedAt:     createdAt,
		UpdatedAt:     updatedAt,
		ID:            service.ID.String(),
		Hostname:      service.Hostname,
		ServiceTypeID: service.ServiceTypeID.String(),
	}

	logrus.WithFields(logrus.Fields{"Hostname": srvc.Hostname}).Info("GetServiceByHostname")

	return srvc, nil
}

func (s *managerServer) GetServiceByID(ctx context.Context, in *protobuff.GetServiceByIDFields) (*protobuff.Service, error) {
	id, err := uuid.Parse(in.ID)

	if err != nil {
		return nil, services.HandleProtoError(err)
	}

	service, err := getServiceByID(s.db, id)

	if err != nil {
		return nil, services.HandleProtoError(err)
	}

	if service == nil {
		err := status.Error(codes.NotFound, "invalid_service")
		return nil, services.HandleProtoError(err)
	}

	createdAt, err := ptypes.TimestampProto(service.CreatedAt)

	if err != nil {
		return nil, services.HandleProtoError(err)
	}

	updatedAt, err := ptypes.TimestampProto(service.UpdatedAt)

	if err != nil {
		return nil, services.HandleProtoError(err)
	}

	srvc := &protobuff.Service{
		CreatedAt:     createdAt,
		UpdatedAt:     updatedAt,
		ID:            service.ID.String(),
		Hostname:      service.Hostname,
		ServiceTypeID: service.ServiceTypeID.String(),
	}

	logrus.WithFields(logrus.Fields{"ID": srvc.ID}).Info("GetServiceByID")

	return srvc, nil
}

func (s *managerServer) GetServiceTypeByName(ctx context.Context, in *protobuff.GetServiceTypeByNameFields) (*protobuff.ServiceType, error) {
	servicetype, err := getServiceTypeByName(s.db, in.Name)

	if err != nil {
		return nil, services.HandleProtoError(err)
	}

	if servicetype == nil {
		err := status.Error(codes.NotFound, "invalid_service_type")
		return nil, services.HandleProtoError(err)
	}

	st := &protobuff.ServiceType{
		ID:   servicetype.ID.String(),
		Name: servicetype.Name,
	}

	logrus.WithFields(logrus.Fields{"Name": st.Name}).Info("GetServiceTypeByName")

	return st, nil
}

func (s *managerServer) NewService(ctx context.Context, in *protobuff.NewServiceFields) (*protobuff.Service, error) {
	serviceTypeID, err := uuid.Parse(in.ServiceTypeID)

	if err != nil {
		return nil, services.HandleProtoError(err)
	}

	serviceType, err := getServiceTypeByID(s.db, serviceTypeID)

	if err != nil {
		return nil, services.HandleProtoError(err)
	}

	if serviceType == nil {
		err := status.Error(codes.NotFound, "invalid_service_type")
		return nil, services.HandleProtoError(err)
	}

	service := schemas.Service{Hostname: in.Hostname, ServiceTypeID: serviceType.ID}

	if err := s.db.Create(&service).Error; err != nil {
		return nil, services.HandleProtoError(err)
	}

	createdAt, err := ptypes.TimestampProto(service.CreatedAt)

	if err != nil {
		return nil, services.HandleProtoError(err)
	}

	updatedAt, err := ptypes.TimestampProto(service.UpdatedAt)

	if err != nil {
		return nil, services.HandleProtoError(err)
	}

	svc := &protobuff.Service{
		CreatedAt:     createdAt,
		UpdatedAt:     updatedAt,
		ID:            service.ID.String(),
		Hostname:      service.Hostname,
		ServiceTypeID: service.ServiceTypeID.String(),
	}

	logFields := logrus.Fields{"Hostname": svc.Hostname, "ServiceTypeID": svc.ServiceTypeID}

	logrus.WithFields(logFields).Info("NewService")

	return svc, nil
}
