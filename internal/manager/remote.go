package manager

import (
	"context"
	"fmt"
	"net"
	"time"

	"github.com/erkrnt/symphony/api"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type remoteServer struct {
	manager *Manager
}

func (s *remoteServer) Init(ctx context.Context, in *api.ManagerRemoteInitRequest) (*api.ManagerRemoteInitResponse, error) {
	cluster, err := s.manager.getCluster()

	if err != nil {
		st := status.New(codes.Internal, err.Error())

		return nil, st.Err()
	}

	isLocalInit := in.ServiceAddr == s.manager.flags.listenServiceAddr.String()

	if cluster != nil && isLocalInit {
		st := status.New(codes.AlreadyExists, "cluster_already_initialized")

		return nil, st.Err()
	}

	if cluster == nil && !isLocalInit {
		st := status.New(codes.InvalidArgument, "cluster_not_initialized")

		return nil, st.Err()
	}

	if cluster == nil && isLocalInit {
		clusterID := uuid.New()

		cluster = &api.Cluster{
			ID: clusterID.String(),
		}

		err := s.manager.saveCluster(cluster)

		if err != nil {
			st := status.New(codes.Internal, err.Error())

			return nil, st.Err()
		}
	}

	services, err := s.manager.getServices()

	if err != nil {
		st := status.New(codes.Internal, err.Error())

		return nil, st.Err()
	}

	for _, service := range services {
		if service.Addr == in.ServiceAddr {
			st := status.New(codes.AlreadyExists, codes.AlreadyExists.String())

			return nil, st.Err()
		}
	}

	serviceID := uuid.New()

	service := &api.Service{
		Addr: in.ServiceAddr,
		ID:   serviceID.String(),
		Type: in.ServiceType.String(),
	}

	saveErr := s.manager.saveService(service)

	if saveErr != nil {
		st := status.New(codes.Internal, saveErr.Error())

		return nil, st.Err()
	}

	endpoints := make([]string, 0)

	for _, service := range services {
		if service.Type == api.ServiceType_MANAGER.String() {
			endpoints = append(endpoints, service.Addr)
		}
	}

	gossipAddr := s.manager.flags.listenGossipAddr

	res := &api.ManagerRemoteInitResponse{
		ClusterID:  cluster.ID,
		Endpoints:  endpoints,
		GossipAddr: gossipAddr.String(),
		ServiceID:  service.ID,
	}

	return res, nil
}

func (s *remoteServer) Join(ctx context.Context, in *api.ManagerRemoteJoinRequest) (*api.ManagerRemoteInitResponse, error) {
	cluster, err := s.manager.getCluster()

	if err != nil {
		st := status.New(codes.Internal, err.Error())

		return nil, st.Err()
	}

	clusterID, err := uuid.Parse(in.ClusterID)

	if err != nil {
		st := status.New(codes.InvalidArgument, err.Error())

		return nil, st.Err()
	}

	if cluster.ID != clusterID.String() {
		st := status.New(codes.InvalidArgument, "invalid_cluster_id")

		return nil, st.Err()
	}

	serviceID, err := uuid.Parse(in.ServiceID)

	if err != nil {
		st := status.New(codes.InvalidArgument, err.Error())

		return nil, st.Err()
	}

	services, err := s.manager.getServices()

	if err != nil {
		st := status.New(codes.Internal, err.Error())

		return nil, st.Err()
	}

	service := GetServiceByID(services, serviceID)

	if service == nil {
		st := status.New(codes.InvalidArgument, "invalid_service_id")

		return nil, st.Err()
	}

	endpoints := make([]string, 0)

	for _, service := range services {
		if service.Type == api.ServiceType_MANAGER.String() {
			endpoints = append(endpoints, service.Addr)
		}
	}

	gossipAddr := s.manager.flags.listenGossipAddr

	res := &api.ManagerRemoteInitResponse{
		ClusterID:  cluster.ID,
		Endpoints:  endpoints,
		GossipAddr: gossipAddr.String(),
		ServiceID:  service.ID,
	}

	return res, nil
}

func (s *remoteServer) Leave(ctx context.Context, in *api.ManagerRemoteLeaveRequest) (*api.SuccessStatusResponse, error) {
	serviceID, err := uuid.Parse(in.ServiceID)

	if err != nil {
		st := status.New(codes.InvalidArgument, err.Error())

		return nil, st.Err()
	}

	key, err := s.manager.key.Get(s.manager.flags.configDir)

	if err != nil {
		st := status.New(codes.Internal, err.Error())

		return nil, st.Err()
	}

	if serviceID != *key.ServiceID {
		st := status.New(codes.PermissionDenied, err.Error())

		return nil, st.Err()
	}

	leaveErr := s.manager.memberlist.Leave(time.Second)

	if leaveErr != nil {
		st := status.New(codes.Internal, leaveErr.Error())

		return nil, st.Err()
	}

	shutdownErr := s.manager.memberlist.Shutdown()

	if shutdownErr != nil {
		st := status.New(codes.Internal, shutdownErr.Error())

		return nil, st.Err()
	}

	logrus.Debug("Manager service has left the cluster.")

	res := &api.SuccessStatusResponse{Success: true}

	return res, nil
}

func (s *remoteServer) Remove(ctx context.Context, in *api.ManagerRemoteRemoveRequest) (*api.SuccessStatusResponse, error) {
	serviceID, err := uuid.Parse(in.ServiceID)

	if err != nil {
		st := status.New(codes.InvalidArgument, err.Error())

		return nil, st.Err()
	}

	cluster, err := s.manager.getCluster()

	if err != nil {
		st := status.New(codes.InvalidArgument, err.Error())

		return nil, st.Err()
	}

	if cluster == nil {
		st := status.New(codes.NotFound, "cluster_not_initialized")

		return nil, st.Err()
	}

	services, err := s.manager.getServices()

	if err != nil {
		st := status.New(codes.Internal, err.Error())

		return nil, st.Err()
	}

	service := GetServiceByID(services, serviceID)

	if service == nil {
		st := status.New(codes.NotFound, "service_not_found")

		return nil, st.Err()
	}

	leaveAddr, err := net.ResolveTCPAddr("tcp", service.Addr)

	if err != nil {
		return nil, err
	}

	conn, err := grpc.Dial(leaveAddr.String(), grpc.WithInsecure())

	if err != nil {
		return nil, err
	}

	defer conn.Close()

	if service.Type == api.ServiceType_BLOCK.String() {
		remote := api.NewBlockRemoteClient(conn)

		opts := &api.BlockRemoteLeaveRequest{
			ServiceID: service.ID,
		}

		_, err := remote.Leave(ctx, opts)

		if err != nil {
			return nil, err
		}
	}

	if service.Type == api.ServiceType_MANAGER.String() {
		remote := api.NewManagerRemoteClient(conn)

		opts := &api.ManagerRemoteLeaveRequest{
			ServiceID: service.ID,
		}

		_, err := remote.Leave(ctx, opts)

		if err != nil {
			return nil, err
		}
	}

	serviceKey := fmt.Sprintf("/service/%s", service.ID)

	etcd, err := NewEtcdClient(s.manager.flags.etcdEndpoints)

	if err != nil {
		st := status.New(codes.Internal, err.Error())

		return nil, st.Err()
	}

	defer etcd.Close()

	_, delErr := etcd.Delete(ctx, serviceKey)

	if delErr != nil {
		st := status.New(codes.Internal, delErr.Error())

		return nil, st.Err()
	}

	res := &api.SuccessStatusResponse{Success: true}

	return res, nil
}

func (s *remoteServer) NewPv(ctx context.Context, in *api.ManagerRemoteNewPvRequest) (*api.ManagerRemotePvResponse, error) {
	services, err := s.manager.getServices()

	if err != nil {
		st := status.New(codes.Internal, err.Error())

		return nil, st.Err()
	}

	serviceID, err := uuid.Parse(in.ServiceID)

	if err != nil {
		st := status.New(codes.InvalidArgument, err.Error())

		return nil, st.Err()
	}

	service := GetServiceByID(services, serviceID)

	if service == nil {
		st := status.New(codes.NotFound, "invalid_service_id")

		return nil, st.Err()
	}

	volumes, err := s.manager.getPhysicalVolumes()

	if err != nil {
		st := status.New(codes.Internal, err.Error())

		return nil, st.Err()
	}

	var volume *api.PhysicalVolume

	for _, v := range volumes {
		if in.DeviceName == v.DeviceName && service.ID == v.ServiceID {
			volume = v
		}
	}

	if volume != nil {
		st := status.New(codes.AlreadyExists, "physical_volume_exists")

		return nil, st.Err()
	}

	newPvAddr, err := net.ResolveTCPAddr("tcp", service.Addr)

	if err != nil {
		return nil, err
	}

	conn, err := grpc.Dial(newPvAddr.String(), grpc.WithInsecure())

	if err != nil {
		return nil, err
	}

	defer conn.Close()

	remote := api.NewBlockRemoteClient(conn)

	opts := &api.BlockRemotePvRequest{
		DeviceName: in.DeviceName,
	}

	newPvRes, err := remote.NewPv(ctx, opts)

	if err != nil {
		return nil, err
	}

	id := uuid.New()

	v := &api.PhysicalVolume{
		DeviceName: in.DeviceName,
		ID:         id.String(),
		ServiceID:  service.ID,
	}

	saveErr := s.manager.savePhysicalVolume(v)

	if saveErr != nil {
		return nil, saveErr
	}

	res := &api.ManagerRemotePvResponse{
		DeviceName: v.DeviceName,
		ID:         v.ID,
		Metadata:   newPvRes,
		ServiceID:  v.ServiceID,
	}

	return res, nil
}
