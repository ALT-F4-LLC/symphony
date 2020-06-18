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

// VolumeGroupReport : struct for VGDisplay output
type VolumeGroupReport struct {
	Report []struct {
		Vg []api.VolumeGroupMetadata `json:"vg"`
	} `json:"report"`
}

func getVg(id uuid.UUID) (*api.VolumeGroupMetadata, error) {
	cmd := exec.Command("vgdisplay", "--columns", "--reportformat", "json", id.String())

	vgd, vgdErr := cmd.CombinedOutput()

	notExists := strings.Contains(string(vgd), "not found")

	if notExists {
		return nil, nil
	}

	if vgdErr != nil {
		return nil, vgdErr
	}

	res := &VolumeGroupReport{}

	if err := json.Unmarshal(vgd, &res); err != nil {
		return nil, err
	}

	var metadata api.VolumeGroupMetadata

	if len(res.Report) == 1 && len(res.Report[0].Vg) == 1 {
		vg := res.Report[0].Vg[0]

		if vg.VgName == id.String() {
			metadata = vg

			logrus.WithFields(logrus.Fields{"ID": id.String()}).Debug("Volume group successfully discovered.")
		}
	}

	return &metadata, nil
}

func newVg(device string, id uuid.UUID) (*api.VolumeGroupMetadata, error) {
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
