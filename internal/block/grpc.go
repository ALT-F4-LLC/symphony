package block

import (
	"context"

	"github.com/erkrnt/symphony/api"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type GRPCServerBlock struct {
	Block *Block
}

// LvDisplay : gets logical volume metadata from block host
func (b *GRPCServerBlock) LvDisplay(ctx context.Context, in *api.RequestLv) (*api.Lv, error) {
	volumeGroupID, err := uuid.Parse(in.VolumeGroupID)

	if err != nil {
		st := status.New(codes.InvalidArgument, err.Error())

		return nil, st.Err()
	}

	id, err := uuid.Parse(in.ID)

	if err != nil {
		st := status.New(codes.InvalidArgument, err.Error())

		return nil, st.Err()
	}

	lv, lvErr := getLv(volumeGroupID, id)

	if lvErr != nil {
		st := status.New(codes.InvalidArgument, err.Error())

		return nil, st.Err()
	}

	if lv == nil {
		st := status.New(codes.NotFound, "invalid_logical_volume")

		return nil, st.Err()
	}

	metadata := &api.Lv{
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

// PvDisplay : gets physical volume
func (b *GRPCServerBlock) PvDisplay(ctx context.Context, in *api.RequestPv) (*api.Pv, error) {
	metadata, pvErr := getPv(in.DeviceName)

	if pvErr != nil {
		st := status.New(codes.InvalidArgument, pvErr.Error())

		return nil, st.Err()
	}

	if metadata == nil {
		st := status.New(codes.NotFound, "invalid_physical_volume")

		return nil, st.Err()
	}

	logrus.WithFields(logrus.Fields{"DeviceName": in.DeviceName}).Info("GetPv")

	return metadata, nil
}

// VgDisplay : gets volume group
func (b *GRPCServerBlock) VgDisplay(ctx context.Context, in *api.RequestVg) (*api.Vg, error) {
	id, err := uuid.Parse(in.ID)

	if err != nil {
		st := status.New(codes.InvalidArgument, err.Error())

		return nil, st.Err()
	}

	vg, err := getVg(id)

	if err != nil {
		st := status.New(codes.Internal, err.Error())

		return nil, st.Err()
	}

	metadata := &api.Vg{
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

// LvCreate : creates logical volume
func (b *GRPCServerBlock) LvCreate(ctx context.Context, in *api.RequestLvCreate) (*api.Lv, error) {
	volumeGroupID, err := uuid.Parse(in.VolumeGroupID)

	if err != nil {
		st := status.New(codes.InvalidArgument, err.Error())

		return nil, st.Err()
	}

	id, err := uuid.Parse(in.ID)

	if err != nil {
		st := status.New(codes.InvalidArgument, err.Error())

		return nil, st.Err()
	}

	lv, err := newLv(volumeGroupID, id, in.Size)

	if err != nil {
		st := status.New(codes.Internal, err.Error())

		return nil, st.Err()
	}

	metadata := &api.Lv{
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

// PvCreate : creates physical volume
func (b *GRPCServerBlock) PvCreate(ctx context.Context, in *api.RequestPvCreate) (*api.Pv, error) {
	pv, err := newPv(in.DeviceName)

	if err != nil {
		st := status.New(codes.Internal, err.Error())

		return nil, st.Err()
	}

	metadata := &api.Pv{
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

// VgCreate : creates volume group
func (b *GRPCServerBlock) VgCreate(ctx context.Context, in *api.RequestVgCreate) (*api.Vg, error) {
	id, err := uuid.Parse(in.ID)

	if err != nil {
		st := status.New(codes.Internal, err.Error())

		return nil, st.Err()
	}

	vg, err := newVg(in.DeviceName, id)

	if err != nil {
		st := status.New(codes.Internal, err.Error())

		return nil, st.Err()
	}

	metadata := &api.Vg{
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

// LvRemove : removes logical volume
func (b *GRPCServerBlock) LvRemove(ctx context.Context, in *api.RequestLv) (*api.ResponseStatus, error) {
	volumeGroupID, err := uuid.Parse(in.VolumeGroupID)

	if err != nil {
		st := status.New(codes.InvalidArgument, err.Error())

		return nil, st.Err()
	}

	id, err := uuid.Parse(in.ID)

	if err != nil {
		st := status.New(codes.InvalidArgument, err.Error())

		return nil, st.Err()
	}

	rmErr := removeLv(volumeGroupID, id)

	if rmErr != nil {
		st := status.New(codes.Internal, err.Error())

		return nil, st.Err()
	}

	status := &api.ResponseStatus{SUCCESS: true}

	logrus.WithFields(logrus.Fields{"SUCCESS": status.SUCCESS}).Info("RemoveLv")

	return status, nil
}

// PvRemove : removes physical volume
func (b *GRPCServerBlock) PvRemove(ctx context.Context, in *api.RequestPv) (*api.ResponseStatus, error) {
	err := removePv(in.DeviceName)

	if err != nil {
		st := status.New(codes.Internal, err.Error())

		return nil, st.Err()
	}

	status := &api.ResponseStatus{SUCCESS: true}

	logrus.WithFields(logrus.Fields{"SUCCESS": status.SUCCESS}).Info("RemovePv")

	return status, nil
}

// VgRemove : removes volume group
func (b *GRPCServerBlock) VgRemove(ctx context.Context, in *api.RequestVg) (*api.ResponseStatus, error) {
	id, err := uuid.Parse(in.ID)

	if err != nil {
		st := status.New(codes.InvalidArgument, err.Error())

		return nil, st.Err()
	}

	rmErr := removeVg(id)

	if rmErr != nil {
		st := status.New(codes.Internal, err.Error())

		return nil, st.Err()
	}

	status := &api.ResponseStatus{SUCCESS: true}

	logrus.WithFields(logrus.Fields{"SUCCESS": status.SUCCESS}).Info("RemoveVg")

	return status, nil
}
