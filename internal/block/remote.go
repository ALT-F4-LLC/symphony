package block

import (
	"context"
	"log"
	"net"

	"github.com/erkrnt/symphony/api"
	"github.com/erkrnt/symphony/internal/pkg/cluster"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// RemoteServer : block remote requests
type RemoteServer struct {
	Node *cluster.Node
}

// GetLv : gets logical volume metadata from block host
func (s *RemoteServer) GetLv(ctx context.Context, in *api.BlockLvFields) (*api.BlockLvMetadata, error) {
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

	metadata := &api.BlockLvMetadata{
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
func (s *RemoteServer) GetPv(ctx context.Context, in *api.BlockPvFields) (*api.BlockPvMetadata, error) {
	pv, pvErr := getPv(in.Device)

	if pvErr != nil {
		return nil, api.ProtoError(pvErr)
	}

	if pv == nil {
		err := status.Error(codes.NotFound, "invalid_physical_volume")

		return nil, api.ProtoError(err)
	}

	metadata := &api.BlockPvMetadata{
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
func (s *RemoteServer) GetVg(ctx context.Context, in *api.BlockVgFields) (*api.BlockVgMetadata, error) {
	id, err := uuid.Parse(in.ID)

	if err != nil {
		return nil, api.ProtoError(err)
	}

	vg, err := getVg(id)

	if err != nil {
		return nil, api.ProtoError(err)
	}

	metadata := &api.BlockVgMetadata{
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

// NewLv : creates logical volume
func (s *RemoteServer) NewLv(ctx context.Context, in *api.BlockNewLvFields) (*api.BlockLvMetadata, error) {
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

	metadata := &api.BlockLvMetadata{
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
func (s *RemoteServer) NewPv(ctx context.Context, in *api.BlockPvFields) (*api.BlockPvMetadata, error) {
	pv, err := newPv(in.Device)

	if err != nil {
		return nil, api.ProtoError(err)
	}

	metadata := &api.BlockPvMetadata{
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
func (s *RemoteServer) NewVg(ctx context.Context, in *api.BlockNewVgFields) (*api.BlockVgMetadata, error) {
	id, err := uuid.Parse(in.ID)

	if err != nil {
		return nil, api.ProtoError(err)
	}

	vg, err := newVg(in.Device, id)

	if err != nil {
		return nil, api.ProtoError(err)
	}

	metadata := &api.BlockVgMetadata{
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
func (s *RemoteServer) RemoveLv(ctx context.Context, in *api.BlockLvFields) (*api.RemoveStatus, error) {
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

	status := &api.RemoveStatus{Success: true}

	logrus.WithFields(logrus.Fields{"Success": status.Success}).Info("RemoveLv")

	return status, nil
}

// RemovePv : removes physical volume
func (s *RemoteServer) RemovePv(ctx context.Context, in *api.BlockPvFields) (*api.RemoveStatus, error) {
	err := removePv(in.Device)

	if err != nil {
		return nil, api.ProtoError(err)
	}

	status := &api.RemoveStatus{Success: true}

	logrus.WithFields(logrus.Fields{"Success": status.Success}).Info("RemovePv")

	return status, nil
}

// RemoveVg : removes volume group
func (s *RemoteServer) RemoveVg(ctx context.Context, in *api.BlockVgFields) (*api.RemoveStatus, error) {
	id, err := uuid.Parse(in.ID)

	if err != nil {
		return nil, api.ProtoError(err)
	}

	rmErr := removeVg(id)

	if rmErr != nil {
		return nil, api.ProtoError(rmErr)
	}

	status := &api.RemoveStatus{Success: true}

	logrus.WithFields(logrus.Fields{"Success": status.Success}).Info("RemoveVg")

	return status, nil
}

// StartRemoteServer : starts Raft memebership server
func StartRemoteServer(node *cluster.Node) {
	lis, err := net.Listen("tcp", node.Flags.ListenRemoteAddr.String())

	if err != nil {
		log.Fatal("Failed to listen")
	}

	s := grpc.NewServer()

	server := &RemoteServer{
		Node: node,
	}

	logrus.Info("Started block remote gRPC tcp server.")

	api.RegisterBlockRemoteServer(s, server)

	if err := s.Serve(lis); err != nil {
		log.Fatal("Failed to serve")
	}
}
