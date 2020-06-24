package manager

import (
	"context"
	"fmt"
	"net"
	"time"

	"github.com/erkrnt/symphony/api"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"go.etcd.io/etcd/clientv3"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type remoteServer struct {
	manager *Manager
}

func (s *remoteServer) GetLogicalVolume(ctx context.Context, in *api.ManagerLogicalVolumeRequest) (*api.LogicalVolume, error) {
	lvID, err := uuid.Parse(in.ID)

	if err != nil {
		st := status.New(codes.InvalidArgument, err.Error())

		return nil, st.Err()
	}

	lv, err := s.manager.getLogicalVolumeByID(lvID)

	if err != nil {
		st := status.New(codes.Internal, err.Error())

		return nil, st.Err()
	}

	if lv == nil {
		st := status.New(codes.NotFound, "invalid_logical_volume_id")

		return nil, st.Err()
	}

	vgID, err := uuid.Parse(lv.VolumeGroupID)

	if err != nil {
		st := status.New(codes.InvalidArgument, err.Error())

		return nil, st.Err()
	}

	vg, err := s.manager.getVolumeGroupByID(vgID)

	if err != nil {
		st := status.New(codes.Internal, err.Error())

		return nil, st.Err()
	}

	if vg == nil {
		st := status.New(codes.NotFound, "invalid_volume_group_id")

		return nil, st.Err()
	}

	pvID, err := uuid.Parse(vg.PhysicalVolumeID)

	if err != nil {
		st := status.New(codes.InvalidArgument, err.Error())

		return nil, st.Err()
	}

	pv, err := s.manager.getPhysicalVolumeByID(pvID)

	if err != nil {
		st := status.New(codes.Internal, err.Error())

		return nil, st.Err()
	}

	if pv == nil {
		st := status.New(codes.NotFound, "invalid_physical_volume_id")

		return nil, st.Err()
	}

	serviceID, err := uuid.Parse(pv.ServiceID)

	if err != nil {
		st := status.New(codes.InvalidArgument, err.Error())

		return nil, st.Err()
	}

	service, err := s.manager.getServiceByID(serviceID)

	if err != nil {
		st := status.New(codes.Internal, err.Error())

		return nil, st.Err()
	}

	if service == nil {
		st := status.New(codes.NotFound, "invalid_service_id")

		return nil, st.Err()
	}

	blockAddr, err := net.ResolveTCPAddr("tcp", service.Addr)

	if err != nil {
		return nil, err
	}

	conn, err := grpc.Dial(blockAddr.String(), grpc.WithInsecure())

	if err != nil {
		return nil, err
	}

	defer conn.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)

	defer cancel()

	remote := api.NewBlockRemoteClient(conn)

	opts := &api.BlockLogicalVolumeRequest{
		ID:            lv.ID,
		VolumeGroupID: lv.VolumeGroupID,
	}

	metadata, err := remote.GetLogicalVolume(ctx, opts)

	if err != nil {
		return nil, err
	}

	if metadata == nil {
		st := status.New(codes.NotFound, "invalid_metadata")

		return nil, st.Err()
	}

	lv.Metadata = metadata

	return lv, nil
}

func (s *remoteServer) GetPhysicalVolume(ctx context.Context, in *api.ManagerPhysicalVolumeRequest) (*api.PhysicalVolume, error) {
	pvID, err := uuid.Parse(in.ID)

	if err != nil {
		st := status.New(codes.InvalidArgument, err.Error())

		return nil, st.Err()
	}

	pv, err := s.manager.getPhysicalVolumeByID(pvID)

	if err != nil {
		st := status.New(codes.Internal, err.Error())

		return nil, st.Err()
	}

	if pv == nil {
		st := status.New(codes.NotFound, "invalid_physical_volume_id")

		return nil, st.Err()
	}

	serviceID, err := uuid.Parse(pv.ServiceID)

	if err != nil {
		st := status.New(codes.InvalidArgument, err.Error())

		return nil, st.Err()
	}

	service, err := s.manager.getServiceByID(serviceID)

	if err != nil {
		st := status.New(codes.Internal, err.Error())

		return nil, st.Err()
	}

	blockAddr, err := net.ResolveTCPAddr("tcp", service.Addr)

	if err != nil {
		return nil, err
	}

	conn, err := grpc.Dial(blockAddr.String(), grpc.WithInsecure())

	if err != nil {
		return nil, err
	}

	defer conn.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)

	defer cancel()

	remote := api.NewBlockRemoteClient(conn)

	opts := &api.BlockPhysicalVolumeRequest{
		DeviceName: pv.DeviceName,
	}

	metadata, err := remote.GetPhysicalVolume(ctx, opts)

	if err != nil {
		return nil, err
	}

	if metadata == nil {
		st := status.New(codes.NotFound, "invalid_metadata")

		return nil, st.Err()
	}

	pv.Metadata = metadata

	return pv, nil
}

func (s *remoteServer) GetVolumeGroup(ctx context.Context, in *api.ManagerVolumeGroupRequest) (*api.VolumeGroup, error) {
	vgID, err := uuid.Parse(in.ID)

	if err != nil {
		st := status.New(codes.InvalidArgument, err.Error())

		return nil, st.Err()
	}

	vg, err := s.manager.getVolumeGroupByID(vgID)

	if err != nil {
		st := status.New(codes.Internal, err.Error())

		return nil, st.Err()
	}

	if vg == nil {
		st := status.New(codes.NotFound, "invalid_volume_group_id")

		return nil, st.Err()
	}

	pvID, err := uuid.Parse(vg.PhysicalVolumeID)

	if err != nil {
		st := status.New(codes.InvalidArgument, err.Error())

		return nil, st.Err()
	}

	pv, err := s.manager.getPhysicalVolumeByID(pvID)

	if err != nil {
		st := status.New(codes.Internal, err.Error())

		return nil, st.Err()
	}

	if pv == nil {
		st := status.New(codes.NotFound, "invalid_physical_volume_id")

		return nil, st.Err()
	}

	serviceID, err := uuid.Parse(pv.ServiceID)

	if err != nil {
		st := status.New(codes.InvalidArgument, err.Error())

		return nil, st.Err()
	}

	service, err := s.manager.getServiceByID(serviceID)

	if err != nil {
		st := status.New(codes.Internal, err.Error())

		return nil, st.Err()
	}

	if service == nil {
		st := status.New(codes.NotFound, "invalid_service_id")

		return nil, st.Err()
	}

	blockAddr, err := net.ResolveTCPAddr("tcp", service.Addr)

	if err != nil {
		return nil, err
	}

	conn, err := grpc.Dial(blockAddr.String(), grpc.WithInsecure())

	if err != nil {
		return nil, err
	}

	defer conn.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)

	defer cancel()

	remote := api.NewBlockRemoteClient(conn)

	opts := &api.BlockVolumeGroupRequest{
		ID: vg.ID,
	}

	metadata, err := remote.GetVolumeGroup(ctx, opts)

	if err != nil {
		return nil, err
	}

	if metadata == nil {
		st := status.New(codes.NotFound, "invalid_metadata")

		return nil, st.Err()
	}

	vg.Metadata = metadata

	return vg, nil
}

func (s *remoteServer) NewLogicalVolume(ctx context.Context, in *api.ManagerNewLogicalVolumeRequest) (*api.LogicalVolume, error) {
	vgID, err := uuid.Parse(in.VolumeGroupID)

	if err != nil {
		st := status.New(codes.InvalidArgument, err.Error())

		return nil, st.Err()
	}

	vg, err := s.manager.getVolumeGroupByID(vgID)

	if err != nil {
		st := status.New(codes.Internal, err.Error())

		return nil, st.Err()
	}

	if vg == nil {
		st := status.New(codes.NotFound, "invalid_volume_group_id")

		return nil, st.Err()
	}

	pvID, err := uuid.Parse(vg.PhysicalVolumeID)

	if err != nil {
		st := status.New(codes.InvalidArgument, "invalid_service_id")

		return nil, st.Err()
	}

	pv, err := s.manager.getPhysicalVolumeByID(pvID)

	if err != nil {
		st := status.New(codes.Internal, err.Error())

		return nil, st.Err()
	}

	if pv == nil {
		st := status.New(codes.NotFound, "invalid_physical_volume_id")

		return nil, st.Err()
	}

	serviceID, err := uuid.Parse(pv.ServiceID)

	if err != nil {
		st := status.New(codes.InvalidArgument, "invalid_service_id")

		return nil, st.Err()
	}

	service, err := s.manager.getServiceByID(serviceID)

	if service == nil {
		st := status.New(codes.NotFound, "invalid_service_id")

		return nil, st.Err()
	}

	newLvAddr, err := net.ResolveTCPAddr("tcp", service.Addr)

	if err != nil {
		return nil, err
	}

	conn, err := grpc.Dial(newLvAddr.String(), grpc.WithInsecure())

	if err != nil {
		return nil, err
	}

	defer conn.Close()

	remote := api.NewBlockRemoteClient(conn)

	lvID := uuid.New()

	opts := &api.BlockNewLogicalVolumeRequest{
		ID:            lvID.String(),
		Size:          in.Size,
		VolumeGroupID: vg.ID,
	}

	metadata, err := remote.NewLogicalVolume(ctx, opts)

	if err != nil {
		return nil, err
	}

	if metadata == nil {
		st := status.New(codes.NotFound, "invalid_metadata")

		return nil, st.Err()
	}

	lv := &api.LogicalVolume{
		ID:            opts.ID,
		Size:          opts.Size,
		VolumeGroupID: opts.VolumeGroupID,
	}

	saveErr := s.manager.saveLogicalVolume(lv)

	if saveErr != nil {
		return nil, saveErr
	}

	lv.Metadata = metadata

	return lv, nil
}

func (s *remoteServer) NewPhysicalVolume(ctx context.Context, in *api.ManagerNewPhysicalVolumeRequest) (*api.PhysicalVolume, error) {
	serviceID, err := uuid.Parse(in.ServiceID)

	if err != nil {
		st := status.New(codes.InvalidArgument, err.Error())

		return nil, st.Err()
	}

	service, err := s.manager.getServiceByID(serviceID)

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

	opts := &api.BlockPhysicalVolumeRequest{
		DeviceName: in.DeviceName,
	}

	metadata, err := remote.NewPhysicalVolume(ctx, opts)

	if err != nil {
		return nil, err
	}

	if metadata == nil {
		st := status.New(codes.NotFound, "invalid_metadata")

		return nil, st.Err()
	}

	pvID := uuid.New()

	pv := &api.PhysicalVolume{
		DeviceName: in.DeviceName,
		ID:         pvID.String(),
		ServiceID:  service.ID,
	}

	saveErr := s.manager.savePhysicalVolume(pv)

	if saveErr != nil {
		return nil, saveErr
	}

	pv.Metadata = metadata

	return pv, nil
}

func (s *remoteServer) NewVolumeGroup(ctx context.Context, in *api.ManagerNewVolumeGroupRequest) (*api.VolumeGroup, error) {
	physicalVolumeID, err := uuid.Parse(in.PhysicalVolumeID)

	if err != nil {
		st := status.New(codes.InvalidArgument, err.Error())

		return nil, st.Err()
	}

	physicalVolume, err := s.manager.getPhysicalVolumeByID(physicalVolumeID)

	if err != nil {
		st := status.New(codes.Internal, err.Error())

		return nil, st.Err()
	}

	if physicalVolume == nil {
		st := status.New(codes.NotFound, "invalid_physical_volume_id")

		return nil, st.Err()
	}

	physicalVolumeServiceID, err := uuid.Parse(physicalVolume.ServiceID)

	if err != nil {
		st := status.New(codes.InvalidArgument, "invalid_service_id")

		return nil, st.Err()
	}

	service, err := s.manager.getServiceByID(physicalVolumeServiceID)

	if service == nil {
		st := status.New(codes.NotFound, "invalid_service_id")

		return nil, st.Err()
	}

	newVgAddr, err := net.ResolveTCPAddr("tcp", service.Addr)

	if err != nil {
		return nil, err
	}

	conn, err := grpc.Dial(newVgAddr.String(), grpc.WithInsecure())

	if err != nil {
		return nil, err
	}

	defer conn.Close()

	remote := api.NewBlockRemoteClient(conn)

	volumeGroupID := uuid.New()

	opts := &api.BlockNewVolumeGroupRequest{
		DeviceName: physicalVolume.DeviceName,
		ID:         volumeGroupID.String(),
	}

	metadata, err := remote.NewVolumeGroup(ctx, opts)

	if err != nil {
		return nil, err
	}

	if metadata == nil {
		st := status.New(codes.NotFound, "invalid_metadata")

		return nil, st.Err()
	}

	vg := &api.VolumeGroup{
		ID:               volumeGroupID.String(),
		PhysicalVolumeID: physicalVolume.ID,
	}

	saveErr := s.manager.saveVolumeGroup(vg)

	if saveErr != nil {
		return nil, saveErr
	}

	vg.Metadata = metadata

	return vg, nil
}

func (s *remoteServer) RemoveLogicalVolume(ctx context.Context, in *api.ManagerLogicalVolumeRequest) (*api.SuccessStatusResponse, error) {
	lvID, err := uuid.Parse(in.ID)

	if err != nil {
		st := status.New(codes.InvalidArgument, err.Error())

		return nil, st.Err()
	}

	lv, err := s.manager.getLogicalVolumeByID(lvID)

	if err != nil {
		st := status.New(codes.Internal, err.Error())

		return nil, st.Err()
	}

	if lv == nil {
		st := status.New(codes.NotFound, "invalid_logical_volume_id")

		return nil, st.Err()
	}

	vgID, err := uuid.Parse(lv.VolumeGroupID)

	if err != nil {
		st := status.New(codes.InvalidArgument, err.Error())

		return nil, st.Err()
	}

	vg, err := s.manager.getVolumeGroupByID(vgID)

	if err != nil {
		st := status.New(codes.Internal, err.Error())

		return nil, st.Err()
	}

	if vg == nil {
		st := status.New(codes.NotFound, "invalid_volume_group_id")

		return nil, st.Err()
	}

	pvID, err := uuid.Parse(vg.PhysicalVolumeID)

	if err != nil {
		st := status.New(codes.InvalidArgument, err.Error())

		return nil, st.Err()
	}

	pv, err := s.manager.getPhysicalVolumeByID(pvID)

	if err != nil {
		st := status.New(codes.Internal, err.Error())

		return nil, st.Err()
	}

	if pv == nil {
		st := status.New(codes.NotFound, "invalid_physical_volume_id")

		return nil, st.Err()
	}

	serviceID, err := uuid.Parse(pv.ServiceID)

	if err != nil {
		st := status.New(codes.InvalidArgument, err.Error())

		return nil, st.Err()
	}

	service, err := s.manager.getServiceByID(serviceID)

	if err != nil {
		st := status.New(codes.Internal, err.Error())

		return nil, st.Err()
	}

	if service == nil {
		st := status.New(codes.NotFound, "invalid_service_id")

		return nil, st.Err()
	}

	blockAddr, err := net.ResolveTCPAddr("tcp", service.Addr)

	if err != nil {
		return nil, err
	}

	conn, err := grpc.Dial(blockAddr.String(), grpc.WithInsecure())

	if err != nil {
		return nil, err
	}

	defer conn.Close()

	removeCtx, removeCancel := context.WithTimeout(context.Background(), 5*time.Second)

	defer removeCancel()

	remote := api.NewBlockRemoteClient(conn)

	opts := &api.BlockLogicalVolumeRequest{
		ID:            lv.ID,
		VolumeGroupID: lv.VolumeGroupID,
	}

	_, removeErr := remote.RemoveLogicalVolume(removeCtx, opts)

	if removeErr != nil {
		return nil, removeErr
	}

	etcd, err := clientv3.New(clientv3.Config{
		Endpoints:   s.manager.flags.etcdEndpoints,
		DialTimeout: 5 * time.Second,
	})

	if err != nil {
		st := status.New(codes.Internal, err.Error())

		return nil, st.Err()
	}

	defer etcd.Close()

	etcdCtx, etcdCancel := context.WithTimeout(context.Background(), 5*time.Second)

	defer etcdCancel()

	etcdKey := fmt.Sprintf("/logicalvolume/%s", lv.ID)

	_, delRes := etcd.Delete(etcdCtx, etcdKey)

	if delRes != nil {
		st := status.New(codes.Internal, delRes.Error())

		return nil, st.Err()
	}

	res := &api.SuccessStatusResponse{Success: true}

	return res, nil
}

func (s *remoteServer) RemovePhysicalVolume(ctx context.Context, in *api.ManagerPhysicalVolumeRequest) (*api.SuccessStatusResponse, error) {
	volumeID, err := uuid.Parse(in.ID)

	if err != nil {
		st := status.New(codes.InvalidArgument, err.Error())

		return nil, st.Err()
	}

	volume, err := s.manager.getPhysicalVolumeByID(volumeID)

	if err != nil {
		st := status.New(codes.Internal, err.Error())

		return nil, st.Err()
	}

	if volume == nil {
		st := status.New(codes.NotFound, "invalid_physical_volume_id")

		return nil, st.Err()
	}

	serviceID, err := uuid.Parse(volume.ServiceID)

	if err != nil {
		st := status.New(codes.InvalidArgument, err.Error())

		return nil, st.Err()
	}

	service, err := s.manager.getServiceByID(serviceID)

	if err != nil {
		st := status.New(codes.Internal, err.Error())

		return nil, st.Err()
	}

	blockAddr, err := net.ResolveTCPAddr("tcp", service.Addr)

	if err != nil {
		return nil, err
	}

	conn, err := grpc.Dial(blockAddr.String(), grpc.WithInsecure())

	if err != nil {
		return nil, err
	}

	defer conn.Close()

	removeCtx, removeCancel := context.WithTimeout(context.Background(), 5*time.Second)

	defer removeCancel()

	remote := api.NewBlockRemoteClient(conn)

	opts := &api.BlockPhysicalVolumeRequest{
		DeviceName: volume.DeviceName,
	}

	_, removeErr := remote.RemovePhysicalVolume(removeCtx, opts)

	if removeErr != nil {
		return nil, removeErr
	}

	etcd, err := clientv3.New(clientv3.Config{
		Endpoints:   s.manager.flags.etcdEndpoints,
		DialTimeout: 5 * time.Second,
	})

	if err != nil {
		st := status.New(codes.Internal, err.Error())

		return nil, st.Err()
	}

	defer etcd.Close()

	etcdCtx, etcdCancel := context.WithTimeout(context.Background(), 5*time.Second)

	defer etcdCancel()

	etcdKey := fmt.Sprintf("/physicalvolume/%s", volume.ID)

	_, delRes := etcd.Delete(etcdCtx, etcdKey)

	if delRes != nil {
		st := status.New(codes.Internal, delRes.Error())

		return nil, st.Err()
	}

	res := &api.SuccessStatusResponse{Success: true}

	return res, nil
}

func (s *remoteServer) RemoveVolumeGroup(ctx context.Context, in *api.ManagerVolumeGroupRequest) (*api.SuccessStatusResponse, error) {
	volumeGroupID, err := uuid.Parse(in.ID)

	if err != nil {
		st := status.New(codes.InvalidArgument, err.Error())

		return nil, st.Err()
	}

	volumeGroup, err := s.manager.getVolumeGroupByID(volumeGroupID)

	if err != nil {
		st := status.New(codes.Internal, err.Error())

		return nil, st.Err()
	}

	if volumeGroup == nil {
		st := status.New(codes.NotFound, "invalid_volume_group_id")

		return nil, st.Err()
	}

	physicalVolumeID, err := uuid.Parse(volumeGroup.PhysicalVolumeID)

	if err != nil {
		st := status.New(codes.InvalidArgument, err.Error())

		return nil, st.Err()
	}

	physicalVolume, err := s.manager.getPhysicalVolumeByID(physicalVolumeID)

	if err != nil {
		st := status.New(codes.Internal, err.Error())

		return nil, st.Err()
	}

	if physicalVolume == nil {
		st := status.New(codes.NotFound, "invalid_physical_volume_id")

		return nil, st.Err()
	}

	serviceID, err := uuid.Parse(physicalVolume.ServiceID)

	if err != nil {
		st := status.New(codes.InvalidArgument, err.Error())

		return nil, st.Err()
	}

	service, err := s.manager.getServiceByID(serviceID)

	if err != nil {
		st := status.New(codes.Internal, err.Error())

		return nil, st.Err()
	}

	blockAddr, err := net.ResolveTCPAddr("tcp", service.Addr)

	if err != nil {
		return nil, err
	}

	conn, err := grpc.Dial(blockAddr.String(), grpc.WithInsecure())

	if err != nil {
		return nil, err
	}

	defer conn.Close()

	removeCtx, removeCancel := context.WithTimeout(context.Background(), 5*time.Second)

	defer removeCancel()

	remote := api.NewBlockRemoteClient(conn)

	opts := &api.BlockVolumeGroupRequest{
		ID: volumeGroup.ID,
	}

	_, removeErr := remote.RemoveVolumeGroup(removeCtx, opts)

	if removeErr != nil {
		return nil, removeErr
	}

	etcd, err := clientv3.New(clientv3.Config{
		Endpoints:   s.manager.flags.etcdEndpoints,
		DialTimeout: 5 * time.Second,
	})

	if err != nil {
		st := status.New(codes.Internal, err.Error())

		return nil, st.Err()
	}

	defer etcd.Close()

	etcdCtx, etcdCancel := context.WithTimeout(context.Background(), 5*time.Second)

	defer etcdCancel()

	etcdKey := fmt.Sprintf("/volumegroup/%s", volumeGroup.ID)

	_, delRes := etcd.Delete(etcdCtx, etcdKey)

	if delRes != nil {
		st := status.New(codes.Internal, delRes.Error())

		return nil, st.Err()
	}

	res := &api.SuccessStatusResponse{Success: true}

	return res, nil
}

func (s *remoteServer) ServiceInit(ctx context.Context, in *api.ManagerServiceInitRequest) (*api.ManagerServiceInitResponse, error) {
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

	res := &api.ManagerServiceInitResponse{
		ClusterID:  cluster.ID,
		Endpoints:  endpoints,
		GossipAddr: gossipAddr.String(),
		ServiceID:  service.ID,
	}

	return res, nil
}

func (s *remoteServer) ServiceJoin(ctx context.Context, in *api.ManagerServiceJoinRequest) (*api.ManagerServiceInitResponse, error) {
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

	service, err := s.manager.getServiceByID(serviceID)

	if err != nil {
		st := status.New(codes.Internal, err.Error())

		return nil, st.Err()
	}

	if service == nil {
		st := status.New(codes.InvalidArgument, "invalid_service_id")

		return nil, st.Err()
	}

	endpoints := make([]string, 0)

	services, err := s.manager.getServices()

	if err != nil {
		st := status.New(codes.Internal, err.Error())

		return nil, st.Err()
	}

	for _, service := range services {
		if service.Type == api.ServiceType_MANAGER.String() {
			endpoints = append(endpoints, service.Addr)
		}
	}

	gossipAddr := s.manager.flags.listenGossipAddr

	res := &api.ManagerServiceInitResponse{
		ClusterID:  cluster.ID,
		Endpoints:  endpoints,
		GossipAddr: gossipAddr.String(),
		ServiceID:  service.ID,
	}

	return res, nil
}

func (s *remoteServer) ServiceLeave(ctx context.Context, in *api.ManagerServiceLeaveRequest) (*api.SuccessStatusResponse, error) {
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

	leaveErr := s.manager.memberlist.Leave(5 * time.Second)

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

func (s *remoteServer) ServiceRemove(ctx context.Context, in *api.ManagerServiceRemoveRequest) (*api.SuccessStatusResponse, error) {
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

	service, err := s.manager.getServiceByID(serviceID)

	if err != nil {
		st := status.New(codes.Internal, err.Error())

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

		opts := &api.BlockServiceLeaveRequest{
			ServiceID: service.ID,
		}

		_, err := remote.ServiceLeave(ctx, opts)

		if err != nil {
			return nil, err
		}
	}

	if service.Type == api.ServiceType_MANAGER.String() {
		remote := api.NewManagerRemoteClient(conn)

		opts := &api.ManagerServiceLeaveRequest{
			ServiceID: service.ID,
		}

		_, err := remote.ServiceLeave(ctx, opts)

		if err != nil {
			return nil, err
		}
	}

	serviceKey := fmt.Sprintf("/service/%s", service.ID)

	etcd, err := clientv3.New(clientv3.Config{
		Endpoints:   s.manager.flags.etcdEndpoints,
		DialTimeout: 5 * time.Second,
	})

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
