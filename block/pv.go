package main

import (
	"encoding/json"
	"errors"
	"os/exec"
	"strings"

	"github.com/erkrnt/symphony/schemas"
	"github.com/sirupsen/logrus"
)

// PvReport : describes a series of physical volumes in LVM
type PvReport struct {
	Report []struct {
		Pv []schemas.PhysicalVolumeMetadata `json:"pv"`
	} `json:"report"`
}

// getPv : if exists - gets physical volume in LVM
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

	res := &PvReport{}
	if err := json.Unmarshal(pvd, &res); err != nil {
		return nil, err
	}

	var metadata schemas.PhysicalVolumeMetadata
	if len(res.Report) == 1 && len(res.Report[0].Pv) == 1 {
		pv := res.Report[0].Pv[0]
		if pv.PvName == device {
			metadata = pv
			logrus.WithFields(logrus.Fields{"device": device}).Debug("Physical volume successfully discovered.")
		}
	}

	return &metadata, nil
}

// newPv : creates a physical device in LVM
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

	logrus.WithFields(logrus.Fields{"device": device}).Debug("Physical volume successfully created.")

	return pv, nil
}

// removePv : if exists - removes physical volume in LVM
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

	logrus.WithFields(logrus.Fields{"device": device}).Debug("Physical volume successfully removed.")

	return nil
}
