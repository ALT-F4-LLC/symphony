package block

import (
	"encoding/json"
	"errors"
	"os/exec"
	"strings"

	"github.com/erkrnt/symphony/api"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

// VolumeGroupReport : output for "vgdisplay" command from LVM
type VolumeGroupReport struct {
	Report []struct {
		Vg []VolumeGroupReportResult `json:"vg"`
	} `json:"report"`
}

// VolumeGroupReportResult : output for individual entry from "vgdisplay" command from LVM
type VolumeGroupReportResult struct {
	VgName    string `json:"vg_name"`
	PvCount   string `json:"pv_count"`
	LvCount   string `json:"lv_count"`
	SnapCount string `json:"snap_count"`
	VgAttr    string `json:"vg_attr"`
	VgSize    string `json:"vg_size"`
	VgFree    string `json:"vg_free"`
}

func getVg(id uuid.UUID) (*api.Vg, error) {
	cmd := exec.Command("vgdisplay", "--columns", "--reportformat", "json", id.String())

	vgd, vgdErr := cmd.CombinedOutput()

	notExists := strings.Contains(string(vgd), "not found")

	if notExists {
		return nil, nil
	}

	if vgdErr != nil {
		return nil, vgdErr
	}

	response := &VolumeGroupReport{}

	if err := json.Unmarshal(vgd, &response); err != nil {
		return nil, err
	}

	var result VolumeGroupReportResult

	if len(response.Report) == 1 && len(response.Report[0].Vg) == 1 {
		vg := response.Report[0].Vg[0]

		if vg.VgName == id.String() {
			result = vg

			logrus.WithFields(logrus.Fields{"ID": id.String()}).Debug("Volume group successfully discovered.")
		}
	}

	metadata := &api.Vg{
		VgName:    result.VgName,
		PvCount:   result.PvCount,
		LvCount:   result.LvCount,
		SnapCount: result.SnapCount,
		VgAttr:    result.VgAttr,
		VgSize:    result.VgSize,
		VgFree:    result.VgFree,
	}

	return metadata, nil
}

func newVg(device string, id uuid.UUID) (*api.Vg, error) {
	exists, _ := getVg(id)

	if exists != nil {
		return nil, errors.New("vg_already_exists")
	}

	cmd := exec.Command("vgcreate", id.String(), device)

	vgc, vgcErr := cmd.CombinedOutput()

	if vgcErr != nil {
		return nil, errors.New(strings.TrimSpace(string(vgc)))
	}

	vg, err := getVg(id)

	if err != nil {
		return nil, err
	}

	logrus.WithFields(logrus.Fields{"ID": id.String()}).Debug("Volume group successfully created.")

	return vg, nil
}

func removeVg(id uuid.UUID) error {
	exists, _ := getVg(id)

	if exists == nil {
		return errors.New("vg_not_found")
	}

	cmd := exec.Command("vgremove", "--force", id.String())

	_, err := cmd.CombinedOutput()

	if err != nil {
		return err
	}

	logrus.WithFields(logrus.Fields{"ID": id.String()}).Debug("Volume group successfully removed.")

	return nil
}
