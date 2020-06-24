package block

import (
	"context"
	"time"

	"github.com/erkrnt/symphony/api"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type remoteServer struct {
	block *Block
}

// GetLogicalVolume : gets logical volume metadata from block host
func (s *remoteServer) GetLogicalVolume(ctx context.Context, in *api.BlockLogicalVolumeRequest) (*api.LogicalVolumeMetadata, error) {
	volumeGroupID, err := uuid.Parse(in.VolumeGroupID)

	if err != nil {
		return nil, api.ProtoError(err)
	}

	id, err := uuid.Parse(in.ID)

	if err != nil {
		return nil, api.ProtoError(err)
	}

	lv, lvErr := getLv(volumeGroupID, id)

	if lvErr != nil {
		return nil, api.ProtoError(lvErr)
	}

	if lv == nil {
		err := status.Error(codes.NotFound, "invalid_physical_volume")
		return nil, api.ProtoError(err)
	}

	metadata := &api.LogicalVolumeMetadata{
		LvName:          lv.LvName,
		VgName:          lv.VgName,
		LvAttr:          lv.LvAttr,
		LvSize:          lv.LvSize,
		PoolLv:          lv.PoolLv,
		Origin:          lv.Origin,
		DataPercent:     lv.DataPercent,
		MetadataPercent: lv.MetadataPercent,
		MovePv:          lv.MovePv,
		MirrorLog:       lv.MirrorLog,
		CopyPercent:     lv.CopyPercent,
		ConvertLv:       lv.ConvertLv,
	}

	logFields := logrus.Fields{
		"ID":            id.String(),
		"VolumeGroupID": volumeGroupID.String(),
	}

	logrus.WithFields(logFields).Info("GetLv")

	return metadata, nil
}

// GetPhysicalVolume : gets physical volume
func (s *remoteServer) GetPhysicalVolume(ctx context.Context, in *api.BlockPhysicalVolumeRequest) (*api.PhysicalVolumeMetadata, error) {
	metadata, pvErr := getPv(in.DeviceName)

	if pvErr != nil {
		return nil, api.ProtoError(pvErr)
	}

	if metadata == nil {
		err := status.Error(codes.NotFound, "invalid_physical_volume")

		return nil, api.ProtoError(err)
	}

	logrus.WithFields(logrus.Fields{"DeviceName": in.DeviceName}).Info("GetPv")

	return metadata, nil
}

// GetVolumeGroup : gets volume group
func (s *remoteServer) GetVolumeGroup(ctx context.Context, in *api.BlockVolumeGroupRequest) (*api.VolumeGroupMetadata, error) {
	id, err := uuid.Parse(in.ID)

	if err != nil {
		return nil, api.ProtoError(err)
	}

	vg, err := getVg(id)

	if err != nil {
		return nil, api.ProtoError(err)
	}

	metadata := &api.VolumeGroupMetadata{
		VgName:    vg.VgName,
		PvCount:   vg.PvCount,
		LvCount:   vg.LvCount,
		SnapCount: vg.SnapCount,
		VgAttr:    vg.VgAttr,
		VgSize:    vg.VgSize,
		VgFree:    vg.VgFree,
	}

	logrus.WithFields(logrus.Fields{"ID": id.String()}).Info("GetVg")

	return metadata, nil
}

// NewLogicalVolume : creates logical volume
func (s *remoteServer) NewLogicalVolume(ctx context.Context, in *api.BlockNewLogicalVolumeRequest) (*api.LogicalVolumeMetadata, error) {
	volumeGroupID, err := uuid.Parse(in.VolumeGroupID)

	if err != nil {
		return nil, api.ProtoError(err)
	}

	id, err := uuid.Parse(in.ID)

	if err != nil {
		return nil, api.ProtoError(err)
	}

	lv, err := newLv(volumeGroupID, id, in.Size)

	if err != nil {
		return nil, err
	}

	metadata := &api.LogicalVolumeMetadata{
		LvName:          lv.LvName,
		VgName:          lv.VgName,
		LvAttr:          lv.LvAttr,
		LvSize:          lv.LvSize,
		PoolLv:          lv.PoolLv,
		Origin:          lv.Origin,
		DataPercent:     lv.DataPercent,
		MetadataPercent: lv.MetadataPercent,
		MovePv:          lv.MovePv,
		MirrorLog:       lv.MirrorLog,
		CopyPercent:     lv.CopyPercent,
		ConvertLv:       lv.ConvertLv,
	}

	logFields := logrus.Fields{
		"ID":            id.String(),
		"Size":          in.Size,
		"VolumeGroupID": volumeGroupID.String(),
	}

	logrus.WithFields(logFields).Info("NewLv")

	return metadata, nil
}

// NewPhysicalVolume : creates physical volume
func (s *remoteServer) NewPhysicalVolume(ctx context.Context, in *api.BlockPhysicalVolumeRequest) (*api.PhysicalVolumeMetadata, error) {
	pv, err := newPv(in.DeviceName)

	if err != nil {
		return nil, api.ProtoError(err)
	}

	metadata := &api.PhysicalVolumeMetadata{
		PvName: pv.PvName,
		VgName: pv.VgName,
		PvFmt:  pv.PvFmt,
		PvAttr: pv.PvAttr,
		PvSize: pv.PvSize,
		PvFree: pv.PvFree,
	}

	logrus.WithFields(logrus.Fields{"DeviceName": in.DeviceName}).Info("NewPv")

	return metadata, nil
}

// NewVolumeGroup : creates volume group
func (s *remoteServer) NewVolumeGroup(ctx context.Context, in *api.BlockNewVolumeGroupRequest) (*api.VolumeGroupMetadata, error) {
	id, err := uuid.Parse(in.ID)

	if err != nil {
		return nil, api.ProtoError(err)
	}

	vg, err := newVg(in.DeviceName, id)

	if err != nil {
		return nil, api.ProtoError(err)
	}

	metadata := &api.VolumeGroupMetadata{
		VgName:    vg.VgName,
		PvCount:   vg.PvCount,
		LvCount:   vg.LvCount,
		SnapCount: vg.SnapCount,
		VgAttr:    vg.VgAttr,
		VgSize:    vg.VgSize,
		VgFree:    vg.VgFree,
	}

	logrus.WithFields(logrus.Fields{"ID": id.String()}).Info("NewVg")

	return metadata, nil
}

// RemoveLogicalVolume : removes logical volume
func (s *remoteServer) RemoveLogicalVolume(ctx context.Context, in *api.BlockLogicalVolumeRequest) (*api.SuccessStatusResponse, error) {
	volumeGroupID, err := uuid.Parse(in.VolumeGroupID)

	if err != nil {
		return nil, api.ProtoError(err)
	}

	id, err := uuid.Parse(in.ID)

	if err != nil {
		return nil, api.ProtoError(err)
	}

	rmErr := removeLv(volumeGroupID, id)

	if rmErr != nil {
		return nil, api.ProtoError(rmErr)
	}

	status := &api.SuccessStatusResponse{Success: true}

	logrus.WithFields(logrus.Fields{"Success": status.Success}).Info("RemoveLv")

	return status, nil
}

// RemovePhysicalVolume : removes physical volume
func (s *remoteServer) RemovePhysicalVolume(ctx context.Context, in *api.BlockPhysicalVolumeRequest) (*api.SuccessStatusResponse, error) {
	err := removePv(in.DeviceName)

	if err != nil {
		return nil, api.ProtoError(err)
	}

	status := &api.SuccessStatusResponse{Success: true}

	logrus.WithFields(logrus.Fields{"Success": status.Success}).Info("RemovePv")

	return status, nil
}

// RemoveVolumeGroup : removes volume group
func (s *remoteServer) RemoveVolumeGroup(ctx context.Context, in *api.BlockVolumeGroupRequest) (*api.SuccessStatusResponse, error) {
	id, err := uuid.Parse(in.ID)

	if err != nil {
		return nil, api.ProtoError(err)
	}

	rmErr := removeVg(id)

	if rmErr != nil {
		return nil, api.ProtoError(rmErr)
	}

	status := &api.SuccessStatusResponse{Success: true}

	logrus.WithFields(logrus.Fields{"Success": status.Success}).Info("RemoveVg")

	return status, nil
}

func (s *remoteServer) ServiceLeave(ctx context.Context, in *api.BlockServiceLeaveRequest) (*api.SuccessStatusResponse, error) {
	serviceID, err := uuid.Parse(in.ServiceID)

	if err != nil {
		st := status.New(codes.InvalidArgument, err.Error())

		return nil, st.Err()
	}

	key, err := s.block.key.Get(s.block.flags.configDir)

	if err != nil {
		st := status.New(codes.Internal, err.Error())

		return nil, st.Err()
	}

	if serviceID != *key.ServiceID {
		st := status.New(codes.PermissionDenied, err.Error())

		return nil, st.Err()
	}

	leaveErr := s.block.memberlist.Leave(5 * time.Second)

	if leaveErr != nil {
		st := status.New(codes.Internal, leaveErr.Error())

		return nil, st.Err()
	}

	shutdownErr := s.block.memberlist.Shutdown()

	if shutdownErr != nil {
		st := status.New(codes.Internal, shutdownErr.Error())

		return nil, st.Err()
	}

	logrus.Debug("Block service has left the cluster.")

	res := &api.SuccessStatusResponse{}

	return res, nil
}
