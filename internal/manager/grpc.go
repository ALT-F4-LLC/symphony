package manager

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"

	"github.com/erkrnt/symphony/api"
	"github.com/erkrnt/symphony/internal/utils"
	"github.com/google/uuid"
	consul "github.com/hashicorp/consul/api"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type grpcServerManager struct {
	image   *DiskImageStore
	manager *Manager
}

const maxImageSize = 1000 << 20 // 1GB max

// GetLogicalVolume : retrieves a logical volume from state
func (s *grpcServerManager) GetLogicalVolume(ctx context.Context, in *api.RequestLogicalVolume) (*api.LogicalVolume, error) {
	lvID, err := uuid.Parse(in.ID)

	if err != nil {
		st := status.New(codes.InvalidArgument, err.Error())

		return nil, st.Err()
	}

	lv, err := logicalVolumeByID(s.manager.flags.ConsulAddr, lvID)

	if err != nil {
		st := status.New(codes.Internal, err.Error())

		return nil, st.Err()
	}

	if lv == nil {
		st := status.New(codes.NotFound, "invalid_logical_volume_id")

		return nil, st.Err()
	}

	return lv, nil
}

// GetLogicalVolumes : retrieves all logical volumes from state
func (s *grpcServerManager) GetLogicalVolumes(ctx context.Context, in *api.RequestLogicalVolumes) (*api.ResponseLogicalVolumes, error) {
	lvs, err := logicalVolumes(s.manager.flags.ConsulAddr)

	if err != nil {
		st := status.New(codes.Internal, err.Error())

		return nil, st.Err()
	}

	results := lvs

	if in.ServiceID != "" {
		serviceID, err := uuid.Parse(in.ServiceID)

		if err != nil {
			st := status.New(codes.InvalidArgument, err.Error())

			return nil, st.Err()
		}

		filter := make([]*api.LogicalVolume, 0)

		for _, lv := range results {
			if serviceID.String() == lv.ServiceID {
				filter = append(filter, lv)
			}
		}

		results = filter
	}

	res := &api.ResponseLogicalVolumes{
		Results: results,
	}

	return res, nil
}

// GetPhysicalVolume : retrieves a physical volume from state
func (s *grpcServerManager) GetPhysicalVolume(ctx context.Context, in *api.RequestPhysicalVolume) (*api.PhysicalVolume, error) {
	pvID, err := uuid.Parse(in.ID)

	if err != nil {
		st := status.New(codes.InvalidArgument, err.Error())

		return nil, st.Err()
	}

	pv, err := physicalVolumeByID(s.manager.flags.ConsulAddr, pvID)

	if err != nil {
		st := status.New(codes.Internal, err.Error())

		return nil, st.Err()
	}

	if pv == nil {
		st := status.New(codes.NotFound, "invalid_physical_volume_id")

		return nil, st.Err()
	}

	return pv, nil
}

// GetPhysicalVolumes : retrieves all physical volumes from state
func (s *grpcServerManager) GetPhysicalVolumes(ctx context.Context, in *api.RequestPhysicalVolumes) (*api.ResponsePhysicalVolumes, error) {
	pvs, err := physicalVolumes(s.manager.flags.ConsulAddr)

	if err != nil {
		st := status.New(codes.Internal, err.Error())

		return nil, st.Err()
	}

	results := pvs

	if in.ServiceID != "" {
		serviceID, err := uuid.Parse(in.ServiceID)

		if err != nil {
			st := status.New(codes.InvalidArgument, err.Error())

			return nil, st.Err()
		}

		filter := make([]*api.PhysicalVolume, 0)

		for _, pv := range results {
			if serviceID.String() == pv.ServiceID {
				filter = append(filter, pv)
			}
		}

		results = filter
	}

	if in.Status != api.ResourceStatus_UNKNOWN_RESOURCE_STATUS {
		filter := make([]*api.PhysicalVolume, 0)

		for _, pv := range results {
			if in.Status == pv.Status {
				filter = append(filter, pv)
			}
		}

		results = filter
	}

	res := &api.ResponsePhysicalVolumes{
		Results: results,
	}

	return res, nil
}

// GetService : retrieves a service from state
func (s *grpcServerManager) GetService(ctx context.Context, in *api.RequestService) (*api.Service, error) {
	serviceID, err := uuid.Parse(in.ID)

	if err != nil {
		st := status.New(codes.InvalidArgument, "invalid_service_id")

		return nil, st.Err()
	}

	agentService, err := agentServiceByID(s.manager.flags.ConsulAddr, serviceID)

	if err != nil {
		st := status.New(codes.Internal, err.Error())

		return nil, st.Err()
	}

	if agentService == nil {
		st := status.New(codes.NotFound, "invalid_service")

		return nil, st.Err()
	}

	res, resError := s.manager.service(agentService)

	if resError != nil {
		st := status.New(codes.Internal, resError.Error())

		return nil, st.Err()
	}

	return res, nil
}

// GetServices : retrieves all services from state
func (s *grpcServerManager) GetServices(ctx context.Context, in *api.RequestServices) (*api.ResponseServices, error) {
	agentServices, err := agentServices(s.manager.flags.ConsulAddr)

	if err != nil {
		st := status.New(codes.Internal, err.Error())

		return nil, st.Err()
	}

	services, err := s.manager.services(agentServices)

	if err != nil {
		st := status.New(codes.Internal, err.Error())

		return nil, st.Err()
	}

	res := &api.ResponseServices{
		Results: services,
	}

	return res, nil
}

// GetVolumeGroup : retrieves a volume group from state
func (s *grpcServerManager) GetVolumeGroup(ctx context.Context, in *api.RequestVolumeGroup) (*api.VolumeGroup, error) {
	vgID, err := uuid.Parse(in.ID)

	if err != nil {
		st := status.New(codes.InvalidArgument, err.Error())
		return nil, st.Err()
	}

	vg, err := volumeGroupByID(s.manager.flags.ConsulAddr, vgID)

	if err != nil {
		st := status.New(codes.Internal, err.Error())
		return nil, st.Err()
	}

	if vg == nil {
		st := status.New(codes.NotFound, "invalid_volume_group_id")
		return nil, st.Err()
	}

	return vg, nil
}

// GetVolumeGroups : retrieves all volume groups from state
func (s *grpcServerManager) GetVolumeGroups(ctx context.Context, in *api.RequestVolumeGroups) (*api.ResponseVolumeGroups, error) {
	vgs, err := volumeGroups(s.manager.flags.ConsulAddr)

	if err != nil {
		st := status.New(codes.Internal, err.Error())
		return nil, st.Err()
	}

	results := vgs

	if in.ServiceID != "" {
		serviceID, err := uuid.Parse(in.ServiceID)

		if err != nil {
			st := status.New(codes.InvalidArgument, err.Error())

			return nil, st.Err()
		}

		filter := make([]*api.VolumeGroup, 0)

		for _, vg := range results {
			if serviceID.String() == vg.ServiceID {
				filter = append(filter, vg)
			}
		}

		results = filter
	}

	res := &api.ResponseVolumeGroups{
		Results: results,
	}

	return res, nil
}

// NewImage : uploads and creates image
func (s *grpcServerManager) NewImage(stream api.Manager_NewImageServer) error {
	req, err := stream.Recv()

	if err != nil {
		st := status.New(codes.Internal, err.Error())

		return st.Err()
	}

	description := req.GetDetails().Description

	file := req.GetDetails().File

	name := req.GetDetails().Name

	imageData := bytes.Buffer{}

	imageSize := 0

	fields := logrus.Fields{
		"Description": description,
		"File":        file,
		"Name":        name,
	}

	logrus.WithFields(fields).Debug("NEWIMAGE_STARTED")

	for {
		req, err := stream.Recv()

		if err == io.EOF {
			break
		}

		if err != nil {
			st := status.New(codes.Internal, err.Error())

			return st.Err()
		}

		chunk := req.GetChunkData()

		size := len(chunk)

		imageSize += size

		if imageSize > maxImageSize {
			st := status.New(codes.InvalidArgument, "invalid_file_size")

			return st.Err()
		}

		_, writeErr := imageData.Write(chunk)

		if writeErr != nil {
			st := status.New(codes.Internal, writeErr.Error())

			return st.Err()
		}
	}

	contentType := http.DetectContentType(imageData.Bytes())

	if contentType != "application/octet-stream" {
		st := status.New(codes.InvalidArgument, "invalid_content_type")

		return st.Err()
	}

	saveOptions := ImageStoreSaveOptions{
		Data:        imageData,
		Description: description,
		File:        file,
		Name:        name,
		Size:        int64(imageSize),
	}

	storeErr := s.image.Save(saveOptions)

	if storeErr != nil {
		st := status.New(codes.Internal, storeErr.Error())

		return st.Err()
	}

	imageID := uuid.New()

	image := &api.Image{
		Description: description,
		File:        file,
		ID:          imageID.String(),
		Name:        name,
		Size:        int64(imageSize),
	}

	key := utils.KvKey(imageID, api.ResourceType_IMAGE)

	value, err := json.Marshal(image)

	if err != nil {
		st := status.New(codes.InvalidArgument, err.Error())

		return st.Err()
	}

	putErr := putKVPair(s.manager.flags.ConsulAddr, key, value)

	if putErr != nil {
		st := status.New(codes.Internal, putErr.Error())

		return st.Err()
	}

	closeErr := stream.SendAndClose(image)

	if closeErr != nil {
		st := status.New(codes.Internal, closeErr.Error())

		return st.Err()
	}

	logrus.WithFields(fields).Debug("NEWIMAGE_COMPLETED")

	return nil
}

// NewLogicalVolume : creates a new logical volume in state
func (s *grpcServerManager) NewLogicalVolume(ctx context.Context, in *api.RequestNewLogicalVolume) (*api.LogicalVolume, error) {
	vgID, err := uuid.Parse(in.VolumeGroupID)

	if err != nil {
		st := status.New(codes.InvalidArgument, err.Error())

		return nil, st.Err()
	}

	vg, err := volumeGroupByID(s.manager.flags.ConsulAddr, vgID)

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

	pv, err := physicalVolumeByID(s.manager.flags.ConsulAddr, pvID)

	if err != nil {
		st := status.New(codes.Internal, err.Error())

		return nil, st.Err()
	}

	if pv == nil {
		st := status.New(codes.NotFound, "invalid_physical_volume_id")

		return nil, st.Err()
	}

	lvID := uuid.New()

	lv := &api.LogicalVolume{
		ID:            lvID.String(),
		ServiceID:     vg.ServiceID,
		Size:          in.Size,
		Status:        api.ResourceStatus_REVIEW_IN_PROGRESS,
		VolumeGroupID: vg.ID,
	}

	key := utils.KvKey(lvID, api.ResourceType_LOGICAL_VOLUME)

	value, err := json.Marshal(lv)

	if err != nil {
		st := status.New(codes.InvalidArgument, err.Error())

		return nil, st.Err()
	}

	putErr := putKVPair(s.manager.flags.ConsulAddr, key, value)

	if putErr != nil {
		st := status.New(codes.Internal, putErr.Error())

		return nil, st.Err()
	}

	return lv, nil
}

// NewPhysicalVolume : creates a new physical volume in state
func (s *grpcServerManager) NewPhysicalVolume(ctx context.Context, in *api.RequestNewPhysicalVolume) (*api.PhysicalVolume, error) {
	serviceID, err := uuid.Parse(in.ServiceID)

	if err != nil {
		st := status.New(codes.InvalidArgument, err.Error())

		return nil, st.Err()
	}

	as, err := agentServiceByID(s.manager.flags.ConsulAddr, serviceID)

	if err != nil {
		st := status.New(codes.NotFound, err.Error())

		return nil, st.Err()
	}

	volumes, err := physicalVolumes(s.manager.flags.ConsulAddr)

	if err != nil {
		st := status.New(codes.Internal, err.Error())

		return nil, st.Err()
	}

	var volume *api.PhysicalVolume

	for _, v := range volumes {
		if in.DeviceName == v.DeviceName && as.ID == v.ServiceID {
			volume = v
		}
	}

	if volume != nil {
		st := status.New(codes.AlreadyExists, "physical_volume_exists")

		return nil, st.Err()
	}

	pvID := uuid.New()

	pv := &api.PhysicalVolume{
		DeviceName: in.DeviceName,
		ID:         pvID.String(),
		ServiceID:  as.ID,
		Status:     api.ResourceStatus_REVIEW_IN_PROGRESS,
	}

	key := utils.KvKey(pvID, api.ResourceType_PHYSICAL_VOLUME)

	value, err := json.Marshal(pv)

	if err != nil {
		st := status.New(codes.InvalidArgument, err.Error())

		return nil, st.Err()
	}

	putErr := putKVPair(s.manager.flags.ConsulAddr, key, value)

	if putErr != nil {
		st := status.New(codes.Internal, putErr.Error())

		return nil, st.Err()
	}

	return pv, nil
}

// NewService : creates a new service in state
func (s *grpcServerManager) NewService(ctx context.Context, in *api.RequestNewService) (*api.Service, error) {
	client, err := utils.NewConsulClient(s.manager.flags.ConsulAddr)

	if err != nil {
		st := status.New(codes.Internal, err.Error())

		return nil, st.Err()
	}

	agent := client.Agent()

	regName := uuid.New()

	regAddr, err := net.ResolveTCPAddr("tcp", in.ServiceAddr)

	if err != nil {
		st := status.New(codes.InvalidArgument, err.Error())

		return nil, st.Err()
	}

	regCheck := &consul.AgentServiceCheck{
		CheckID:    regName.String(),
		GRPC:       regAddr.String(),
		GRPCUseTLS: false,
		Interval:   "10s",
		Name:       regName.String(),
	}

	regMeta := make(map[string]string)

	regMeta["ServiceType"] = in.ServiceType.String()

	reg := &consul.AgentServiceRegistration{
		Address: regAddr.IP.String(),
		Check:   regCheck,
		Meta:    regMeta,
		Name:    regName.String(),
		Port:    regAddr.Port,
	}

	regErr := agent.ServiceRegister(reg)

	if regErr != nil {
		st := status.New(codes.Internal, regErr.Error())

		return nil, st.Err()
	}

	serviceQueryOpts := &consul.QueryOptions{}

	regService, _, err := agent.Service(reg.Name, serviceQueryOpts)

	if err != nil {
		st := status.New(codes.Internal, err.Error())

		return nil, st.Err()
	}

	service := &api.Service{
		ID:   regService.ID,
		Type: in.ServiceType,
	}

	return service, nil
}

// NewVolumeGroup : creates a new volume group in state
func (s *grpcServerManager) NewVolumeGroup(ctx context.Context, in *api.RequestNewVolumeGroup) (*api.VolumeGroup, error) {
	physicalVolumeID, err := uuid.Parse(in.PhysicalVolumeID)

	if err != nil {
		st := status.New(codes.InvalidArgument, err.Error())

		return nil, st.Err()
	}

	pv, err := physicalVolumeByID(s.manager.flags.ConsulAddr, physicalVolumeID)

	if err != nil {
		st := status.New(codes.Internal, err.Error())

		return nil, st.Err()
	}

	if pv == nil {
		st := status.New(codes.NotFound, "invalid_physical_volume_id")

		return nil, st.Err()
	}

	vgID := uuid.New()

	vg := &api.VolumeGroup{
		ID:               vgID.String(),
		PhysicalVolumeID: pv.ID,
		ServiceID:        pv.ServiceID,
		Status:           api.ResourceStatus_REVIEW_IN_PROGRESS,
	}

	kvKey := utils.KvKey(vgID, api.ResourceType_VOLUME_GROUP)

	kvValue, err := json.Marshal(vg)

	if err != nil {
		st := status.New(codes.InvalidArgument, err.Error())

		return nil, st.Err()
	}

	saveErr := putKVPair(s.manager.flags.ConsulAddr, kvKey, kvValue)

	if saveErr != nil {
		st := status.New(codes.Internal, saveErr.Error())

		return nil, st.Err()
	}

	return vg, nil
}

// RemoveLogicalVolume : removes a logical volume from state
func (s *grpcServerManager) RemoveLogicalVolume(ctx context.Context, in *api.RequestLogicalVolume) (*api.ResponseStatus, error) {
	logicalVolumeID, err := uuid.Parse(in.ID)

	if err != nil {
		st := status.New(codes.InvalidArgument, err.Error())
		return nil, st.Err()
	}

	lv, err := logicalVolumeByID(s.manager.flags.ConsulAddr, logicalVolumeID)

	if err != nil {
		st := status.New(codes.Internal, err.Error())
		return nil, st.Err()
	}

	if lv == nil {
		st := status.New(codes.NotFound, "invalid_logical_volume_id")
		return nil, st.Err()
	}

	resourceKey := fmt.Sprintf("/logicalvolume/%s", lv.ID)

	delErr := deleteKvPair(s.manager.flags.ConsulAddr, resourceKey)

	if delErr != nil {
		st := status.New(codes.Internal, delErr.Error())

		return nil, st.Err()
	}

	res := &api.ResponseStatus{SUCCESS: true}

	return res, nil
}

// RemovePhysicalVolume : removes a physical volume from state
func (s *grpcServerManager) RemovePhysicalVolume(ctx context.Context, in *api.RequestPhysicalVolume) (*api.ResponseStatus, error) {
	physicalVolumeID, err := uuid.Parse(in.ID)

	if err != nil {
		st := status.New(codes.InvalidArgument, err.Error())

		return nil, st.Err()
	}

	pv, err := physicalVolumeByID(s.manager.flags.ConsulAddr, physicalVolumeID)

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

	_, agentServiceErr := agentServiceByID(s.manager.flags.ConsulAddr, serviceID)

	if agentServiceErr != nil {
		st := status.New(codes.NotFound, agentServiceErr.Error())

		return nil, st.Err()
	}

	resourceKey := fmt.Sprintf("physicalvolume/%s", pv.ID)

	delRes := deleteKvPair(s.manager.flags.ConsulAddr, resourceKey)

	if delRes != nil {
		st := status.New(codes.Internal, delRes.Error())

		return nil, st.Err()
	}

	res := &api.ResponseStatus{SUCCESS: true}

	return res, nil
}

// RemoveService : removes a service from state
func (s *grpcServerManager) RemoveService(ctx context.Context, in *api.RequestService) (*api.ResponseStatus, error) {
	serviceID, err := uuid.Parse(in.ID)

	if err != nil {
		st := status.New(codes.InvalidArgument, err.Error())

		return nil, st.Err()
	}

	_, agentServiceErr := agentServiceByID(s.manager.flags.ConsulAddr, serviceID)

	if err != nil {
		st := status.New(codes.Internal, agentServiceErr.Error())

		return nil, st.Err()
	}

	// TODO : deregister service

	res := &api.ResponseStatus{SUCCESS: true}

	return res, nil
}

// RemoveVolumeGroup : removes a volume group from state
func (s *grpcServerManager) RemoveVolumeGroup(ctx context.Context, in *api.RequestVolumeGroup) (*api.ResponseStatus, error) {
	volumeGroupID, err := uuid.Parse(in.ID)

	if err != nil {
		st := status.New(codes.InvalidArgument, err.Error())

		return nil, st.Err()
	}

	vg, err := volumeGroupByID(s.manager.flags.ConsulAddr, volumeGroupID)

	if err != nil {
		st := status.New(codes.Internal, err.Error())

		return nil, st.Err()
	}

	if vg == nil {
		st := status.New(codes.NotFound, "invalid_volume_group_id")

		return nil, st.Err()
	}

	resourceKey := fmt.Sprintf("volumegroup/%s", vg.ID)

	delErr := deleteKvPair(s.manager.flags.ConsulAddr, resourceKey)

	if delErr != nil {
		st := status.New(codes.Internal, delErr.Error())

		return nil, st.Err()
	}

	res := &api.ResponseStatus{SUCCESS: true}

	return res, nil
}

// UpdateLogicalVolume : updates a logical volume from state
func (s *grpcServerManager) UpdateLogicalVolume(ctx context.Context, in *api.RequestUpdateLogicalVolume) (*api.LogicalVolume, error) {
	lvID, err := uuid.Parse(in.ID)

	if err != nil {
		st := status.New(codes.InvalidArgument, err.Error())

		return nil, st.Err()
	}

	lv, err := logicalVolumeByID(s.manager.flags.ConsulAddr, lvID)

	if err != nil {
		st := status.New(codes.Internal, err.Error())

		return nil, st.Err()
	}

	if lv == nil {
		st := status.New(codes.NotFound, "invalid_logical_volume_id")

		return nil, st.Err()
	}

	if in.Status != api.ResourceStatus_UNKNOWN_RESOURCE_STATUS {
		lv.Status = in.Status
	}

	key := utils.KvKey(lvID, api.ResourceType_LOGICAL_VOLUME)

	value, err := json.Marshal(lv)

	if err != nil {
		st := status.New(codes.InvalidArgument, err.Error())

		return nil, st.Err()
	}

	putErr := putKVPair(s.manager.flags.ConsulAddr, key, value)

	if putErr != nil {
		st := status.New(codes.Internal, putErr.Error())

		return nil, st.Err()
	}

	return lv, nil
}

// UpdatePhysicalVolume : updates a physical volume from state
func (s *grpcServerManager) UpdatePhysicalVolume(ctx context.Context, in *api.RequestUpdatePhysicalVolume) (*api.PhysicalVolume, error) {
	pvID, err := uuid.Parse(in.ID)

	if err != nil {
		st := status.New(codes.InvalidArgument, err.Error())

		return nil, st.Err()
	}

	pv, err := physicalVolumeByID(s.manager.flags.ConsulAddr, pvID)

	if err != nil {
		st := status.New(codes.Internal, err.Error())

		return nil, st.Err()
	}

	if pv == nil {
		st := status.New(codes.NotFound, "invalid_physical_volume_id")

		return nil, st.Err()
	}

	if in.Status != api.ResourceStatus_UNKNOWN_RESOURCE_STATUS {
		pv.Status = in.Status
	}

	key := utils.KvKey(pvID, api.ResourceType_PHYSICAL_VOLUME)

	value, err := json.Marshal(pv)

	if err != nil {
		st := status.New(codes.InvalidArgument, err.Error())

		return nil, st.Err()
	}

	putErr := putKVPair(s.manager.flags.ConsulAddr, key, value)

	if putErr != nil {
		st := status.New(codes.Internal, putErr.Error())

		return nil, st.Err()
	}

	return pv, nil
}

// UpdateVolumeGroup : updates a volume group from state
func (s *grpcServerManager) UpdateVolumeGroup(ctx context.Context, in *api.RequestUpdateVolumeGroup) (*api.VolumeGroup, error) {
	vgID, err := uuid.Parse(in.ID)

	if err != nil {
		st := status.New(codes.InvalidArgument, err.Error())

		return nil, st.Err()
	}

	vg, err := volumeGroupByID(s.manager.flags.ConsulAddr, vgID)

	if err != nil {
		st := status.New(codes.Internal, err.Error())

		return nil, st.Err()
	}

	if vg == nil {
		st := status.New(codes.NotFound, "invalid_volume_group_id")

		return nil, st.Err()
	}

	if in.Status != api.ResourceStatus_UNKNOWN_RESOURCE_STATUS {
		vg.Status = in.Status
	}

	key := utils.KvKey(vgID, api.ResourceType_VOLUME_GROUP)

	value, err := json.Marshal(vg)

	if err != nil {
		st := status.New(codes.InvalidArgument, err.Error())

		return nil, st.Err()
	}

	putErr := putKVPair(s.manager.flags.ConsulAddr, key, value)

	if putErr != nil {
		st := status.New(codes.Internal, putErr.Error())

		return nil, st.Err()
	}

	return vg, nil
}
