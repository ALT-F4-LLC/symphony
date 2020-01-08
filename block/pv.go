package main

import (
	"context"
	"encoding/json"
	"errors"
	"os/exec"
	"strings"

	"github.com/erkrnt/symphony/protobuff"
	"github.com/erkrnt/symphony/schemas"
	"github.com/erkrnt/symphony/services"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// PhysicalVolumeReport : describes a series of physical volumes in LVM
type PhysicalVolumeReport struct {
	Report []struct {
		Pv []schemas.PhysicalVolumeMetadata `json:"pv"`
	} `json:"report"`
}

func getPv(device string) (*schemas.PhysicalVolumeMetadata, error) {
	cmd := exec.Command("pvdisplay", "--columns", "--reportformat", "json", "--quiet", device)

	pvd, pvdErr := cmd.CombinedOutput()

	notExists := strings.Contains(string(pvd), "Failed to find physical volume")

	if notExists {
		return nil, nil
	}

	if pvdErr != nil {
		return nil, pvdErr
	}

	res := &PhysicalVolumeReport{}

	if err := json.Unmarshal(pvd, &res); err != nil {
		return nil, err
	}

	var metadata schemas.PhysicalVolumeMetadata

	if len(res.Report) == 1 && len(res.Report[0].Pv) == 1 {
		pv := res.Report[0].Pv[0]

		if pv.PvName == device {
			metadata = pv

			logrus.WithFields(logrus.Fields{"Device": device}).Debug("Physical volume successfully discovered.")
		}
	}

	return &metadata, nil
}

func newPv(device string) (*schemas.PhysicalVolumeMetadata, error) {
	exists, _ := getPv(device)

	if exists != nil {
		return nil, errors.New("pv_already_exists")
	}

	cmd := exec.Command("pvcreate", device)

	pvc, pvcErr := cmd.CombinedOutput()

	if pvcErr != nil {
		return nil, errors.New(strings.TrimSpace(string(pvc)))
	}

	pv, err := getPv(device)

	if err != nil {
		return nil, err
	}

	logrus.WithFields(logrus.Fields{"Device": device}).Debug("Physical volume successfully created.")

	return pv, nil
}

func removePv(device string) error {
	exists, _ := getPv(device)

	if exists == nil {
		err := errors.New("pv_not_found")
		return err
	}

	cmd := exec.Command("pvremove", "--force", device)

	_, pvrErr := cmd.CombinedOutput()

	if pvrErr != nil {
		return pvrErr
	}

	logrus.WithFields(logrus.Fields{"Device": device}).Debug("Physical volume successfully removed.")

	return nil
}

func (s *blockServer) GetPv(ctx context.Context, in *protobuff.PvFields) (*protobuff.PvMetadata, error) {
	pv, pvErr := getPv(in.Device)

	if pvErr != nil {
		return nil, services.HandleProtoError(pvErr)
	}

	if pv == nil {
		err := status.Error(codes.NotFound, "invalid_physical_volume")
		return nil, services.HandleProtoError(err)
	}

	metadata := &protobuff.PvMetadata{
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

func (s *blockServer) NewPv(ctx context.Context, in *protobuff.PvFields) (*protobuff.PvMetadata, error) {
	pv, err := newPv(in.Device)

	if err != nil {
		return nil, services.HandleProtoError(err)
	}

	metadata := &protobuff.PvMetadata{
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

func (s *blockServer) RemovePv(ctx context.Context, in *protobuff.PvFields) (*protobuff.RemoveStatus, error) {
	err := removePv(in.Device)

	if err != nil {
		return nil, services.HandleProtoError(err)
	}

	status := &protobuff.RemoveStatus{Success: true}

	logrus.WithFields(logrus.Fields{"Success": status.Success}).Info("RemovePv")

	return status, nil
}
