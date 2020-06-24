package block

import (
	"encoding/json"
	"errors"
	"os/exec"
	"strings"

	"github.com/erkrnt/symphony/api"
	"github.com/sirupsen/logrus"
)

// PhysicalVolumeReport : Output for "pvdisplay" command from LVM
type PhysicalVolumeReport struct {
	Report []struct {
		Pv []*PhysicalVolumeReportResult `json:"pv"`
	} `json:"report"`
}

// PhysicalVolumeReportResult : Output for individual pv from "pvdisplay" command from LVM
type PhysicalVolumeReportResult struct {
	PvName string `json:"pv_name"`
	VgName string `json:"vg_name"`
	PvFmt  string `json:"pv_fmt"`
	PvAttr string `json:"pv_attr"`
	PvSize string `json:"pv_size"`
	PvFree string `json:"pv_free"`
}

func getPv(deviceName string) (*api.PhysicalVolumeMetadata, error) {
	cmd := exec.Command("pvdisplay", "--columns", "--reportformat", "json", "--quiet", deviceName)

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

	var result *PhysicalVolumeReportResult

	if len(res.Report) == 1 && len(res.Report[0].Pv) == 1 {
		pv := res.Report[0].Pv[0]

		if pv.PvName == deviceName {
			result = pv

			logrus.WithFields(logrus.Fields{"DeviceName": deviceName}).Debug("Physical volume successfully discovered.")
		}
	}

	metadata := &api.PhysicalVolumeMetadata{
		PvName: result.PvName,
		VgName: result.VgName,
		PvFmt:  result.PvFmt,
		PvAttr: result.PvAttr,
		PvSize: result.PvSize,
		PvFree: result.PvFree,
	}

	return metadata, nil
}

func newPv(device string) (*api.PhysicalVolumeMetadata, error) {
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

func removePv(deviceName string) error {
	exists, _ := getPv(deviceName)

	if exists == nil {
		err := errors.New("pv_not_found")
		return err
	}

	cmd := exec.Command("pvremove", "--force", deviceName)

	_, pvrErr := cmd.CombinedOutput()

	if pvrErr != nil {
		return pvrErr
	}

	logrus.WithFields(logrus.Fields{"DeviceName": deviceName}).Debug("Physical volume successfully removed.")

	return nil
}
