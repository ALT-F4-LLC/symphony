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

// GetLv : gets logical volume metadata from block host
func (s *remoteServer) GetLv(ctx context.Context, in *api.BlockRemoteLvRequest) (*api.BlockRemoteLvResponse, error) {
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

	metadata := &api.BlockRemoteLvResponse{
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

// GetPv : gets physical volume
func (s *remoteServer) GetPv(ctx context.Context, in *api.BlockRemotePvRequest) (*api.BlockRemotePvResponse, error) {
	pv, pvErr := getPv(in.Device)

	if pvErr != nil {
		return nil, api.ProtoError(pvErr)
	}

	if pv == nil {
		err := status.Error(codes.NotFound, "invalid_physical_volume")

		return nil, api.ProtoError(err)
	}

	metadata := &api.BlockRemotePvResponse{
		PvName: pv.PvName,
		VgName: pv.VgName,
		PvFmt:  pv.PvFmt,
		PvAttr: pv.PvAttr,
		PvSize: pv.PvSize,
		PvFree: pv.PvFree,
	}

	logrus.WithFields(logrus.Fields{"Device": in.Device}).Info("GetPv")

	return metadata, nil
}

// GetVg : gets volume group
func (s *remoteServer) GetVg(ctx context.Context, in *api.BlockRemoteVgRequest) (*api.BlockRemoteVgResponse, error) {
	id, err := uuid.Parse(in.ID)

	if err != nil {
		return nil, api.ProtoError(err)
	}

	vg, err := getVg(id)

	if err != nil {
		return nil, api.ProtoError(err)
	}

	metadata := &api.BlockRemoteVgResponse{
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

func (s *remoteServer) Leave(ctx context.Context, in *api.BlockRemoteLeaveRequest) (*api.SuccessStatusResponse, error) {
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

	leaveErr := s.block.memberlist.Leave(time.Second)

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

// NewLv : creates logical volume
func (s *remoteServer) NewLv(ctx context.Context, in *api.BlockRemoteNewLvRequest) (*api.BlockRemoteLvResponse, error) {
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

	metadata := &api.BlockRemoteLvResponse{
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

// NewPv : creates physical volume
func (s *remoteServer) NewPv(ctx context.Context, in *api.BlockRemotePvRequest) (*api.BlockRemotePvResponse, error) {
	pv, err := newPv(in.Device)

	if err != nil {
		return nil, api.ProtoError(err)
	}

	metadata := &api.BlockRemotePvResponse{
		PvName: pv.PvName,
		VgName: pv.VgName,
		PvFmt:  pv.PvFmt,
		PvAttr: pv.PvAttr,
		PvSize: pv.PvSize,
		PvFree: pv.PvFree,
	}

	logrus.WithFields(logrus.Fields{"Device": in.Device}).Info("NewPv")

	return metadata, nil
}

// NewVg : creates volume group
func (s *remoteServer) NewVg(ctx context.Context, in *api.BlockRemoteNewVgRequest) (*api.BlockRemoteVgResponse, error) {
	id, err := uuid.Parse(in.ID)

	if err != nil {
		return nil, api.ProtoError(err)
	}

	vg, err := newVg(in.Device, id)

	if err != nil {
		return nil, api.ProtoError(err)
	}

	metadata := &api.BlockRemoteVgResponse{
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

// RemoveLv : removes logical volume
func (s *remoteServer) RemoveLv(ctx context.Context, in *api.BlockRemoteLvRequest) (*api.SuccessStatusResponse, error) {
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

// RemovePv : removes physical volume
func (s *remoteServer) RemovePv(ctx context.Context, in *api.BlockRemotePvRequest) (*api.SuccessStatusResponse, error) {
	err := removePv(in.Device)

	if err != nil {
		return nil, api.ProtoError(err)
	}

	status := &api.SuccessStatusResponse{Success: true}

	logrus.WithFields(logrus.Fields{"Success": status.Success}).Info("RemovePv")

	return status, nil
}

// RemoveVg : removes volume group
func (s *remoteServer) RemoveVg(ctx context.Context, in *api.BlockRemoteVgRequest) (*api.SuccessStatusResponse, error) {
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
