package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os/exec"
	"strings"

	"github.com/erkrnt/symphony/protobuff"
	"github.com/erkrnt/symphony/schemas"
	"github.com/erkrnt/symphony/services"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// LogicalVolumeReport : struct for LVDisplay output
type LogicalVolumeReport struct {
	Report []struct {
		Lv []schemas.LogicalVolumeMetadata `json:"lv"`
	} `json:"report"`
}

func getLv(volumeGroupID uuid.UUID, id uuid.UUID) (*schemas.LogicalVolumeMetadata, error) {
	path := fmt.Sprintf("/dev/%s/%s", volumeGroupID.String(), id.String())

	cmd := exec.Command("lvdisplay", "--columns", "--reportformat", "json", path)

	lvd, lvdErr := cmd.CombinedOutput()

	output := fmt.Sprintf("Volume group \"%s\" not found", id)

	notExists := strings.Contains(string(lvd), output)

	if notExists {
		return nil, nil
	}

	if lvdErr != nil {
		return nil, lvdErr
	}

	res := LogicalVolumeReport{}

	if err := json.Unmarshal(lvd, &res); err != nil {
		return nil, err
	}

	var metadata schemas.LogicalVolumeMetadata

	if len(res.Report) == 1 && len(res.Report[0].Lv) == 1 {
		lv := res.Report[0].Lv[0]

		if lv.VgName == volumeGroupID.String() && lv.LvName == id.String() {
			metadata = lv

			logFields := logrus.Fields{
				"ID":            id.String(),
				"VolumeGroupID": volumeGroupID.String(),
			}

			logrus.WithFields(logFields).Debug("Logical volume successfully discovered.")
		}
	}

	return &metadata, nil
}

func newLv(volumeGroupID uuid.UUID, id uuid.UUID, size string) (*schemas.LogicalVolumeMetadata, error) {
	exists, _ := getLv(volumeGroupID, id)

	if exists != nil {
		return nil, errors.New("invalid_logical_volume")
	}

	_, lvErr := exec.Command("lvcreate", "-n", id.String(), "-L", size, volumeGroupID.String()).Output()

	if lvErr != nil {
		return nil, lvErr
	}

	lv, err := getLv(volumeGroupID, id)

	if err != nil {
		return nil, err
	}

	return lv, nil
}

func removeLv(volumeGroupID uuid.UUID, id uuid.UUID) error {
	exists, _ := getLv(volumeGroupID, id)

	if exists == nil {
		err := status.Error(codes.NotFound, "lv not found")
		return err
	}

	path := fmt.Sprintf("/dev/%s/%s", volumeGroupID, id)

	_, err := exec.Command("lvremove", "--force", path).Output()

	if err != nil {
		return err
	}

	return nil
}

// // lvDisplay : displays all logical volumes
// func lvDisplay() (*LogicalVolumeReport, error) {
// 	lvdisplay, err := exec.Command("lvdisplay", "--columns", "--reportformat", "json").Output()
// 	if err != nil {
// 		return nil, HandleInternalError(err)
// 	}
// 	// Handle output JSON
// 	output := LogicalVolumeReport{}
// 	if err := json.Unmarshal(lvdisplay, &output); err != nil {
// 		return nil, HandleInternalError(err)
// 	}
// 	// Return JSON data
// 	return &output, nil
// }

func (s *blockServer) GetLv(ctx context.Context, in *protobuff.LvFields) (*protobuff.LvMetadata, error) {
	volumeGroupID, err := uuid.Parse(in.VolumeGroupID)

	if err != nil {
		return nil, services.HandleProtoError(err)
	}

	id, err := uuid.Parse(in.ID)

	if err != nil {
		return nil, services.HandleProtoError(err)
	}

	lv, lvErr := getLv(volumeGroupID, id)

	if lvErr != nil {
		return nil, services.HandleProtoError(lvErr)
	}

	if lv == nil {
		err := status.Error(codes.NotFound, "invalid_physical_volume")
		return nil, services.HandleProtoError(err)
	}

	metadata := &protobuff.LvMetadata{
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

func (s *blockServer) NewLv(ctx context.Context, in *protobuff.NewLvFields) (*protobuff.LvMetadata, error) {
	volumeGroupID, err := uuid.Parse(in.VolumeGroupID)

	if err != nil {
		return nil, services.HandleProtoError(err)
	}

	id, err := uuid.Parse(in.ID)

	if err != nil {
		return nil, services.HandleProtoError(err)
	}

	lv, err := newLv(volumeGroupID, id, in.Size)

	if err != nil {
		return nil, err
	}

	metadata := &protobuff.LvMetadata{
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

	logrus.WithFields(logFields).Debug("NewLv")

	return metadata, nil
}

func (s *blockServer) RemoveLv(ctx context.Context, in *protobuff.LvFields) (*protobuff.RemoveStatus, error) {
	volumeGroupID, err := uuid.Parse(in.VolumeGroupID)

	if err != nil {
		return nil, services.HandleProtoError(err)
	}

	id, err := uuid.Parse(in.ID)

	if err != nil {
		return nil, services.HandleProtoError(err)
	}

	rmErr := removeLv(volumeGroupID, id)

	if rmErr != nil {
		return nil, services.HandleProtoError(rmErr)
	}

	status := &protobuff.RemoveStatus{Success: true}

	logrus.WithFields(logrus.Fields{"Success": status.Success}).Info("RemoveLv")

	return status, nil
}
