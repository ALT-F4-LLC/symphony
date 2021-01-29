package block

import (
	"encoding/json"
	"errors"
	"fmt"
	"os/exec"
	"strings"

	"github.com/erkrnt/symphony/api"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// LogicalVolumeReport : output for "lvdisplay" command from LVM
type LogicalVolumeReport struct {
	Report []struct {
		Lv []LogicalVolumeReportResult `json:"lv"`
	} `json:"report"`
}

// LogicalVolumeReportResult : output for individual entry from "lvdisplay" command from LVM
type LogicalVolumeReportResult struct {
	LvName          string `json:"lv_name"`
	VgName          string `json:"vg_name"`
	LvAttr          string `json:"lv_attr"`
	LvSize          string `json:"lv_size"`
	PoolLv          string `json:"pool_lv"`
	Origin          string `json:"origin"`
	DataPercent     string `json:"data_percent"`
	MetadataPercent string `json:"metadata_percent"`
	MovePv          string `json:"move_pv"`
	MirrorLog       string `json:"mirror_log"`
	CopyPercent     string `json:"copy_percent"`
	ConvertLv       string `json:"convert_lv"`
}

func getLv(volumeGroupID uuid.UUID, id uuid.UUID) (*api.Lv, error) {
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

	var result LogicalVolumeReportResult

	if len(res.Report) == 1 && len(res.Report[0].Lv) == 1 {
		lv := res.Report[0].Lv[0]

		if lv.VgName == volumeGroupID.String() && lv.LvName == id.String() {
			result = lv

			logFields := logrus.Fields{
				"ID":            id.String(),
				"VolumeGroupID": volumeGroupID.String(),
			}

			logrus.WithFields(logFields).Debug("Logical volume successfully discovered.")
		}
	}

	metadata := &api.Lv{
		LvName:          result.LvName,
		VgName:          result.VgName,
		LvAttr:          result.LvAttr,
		LvSize:          result.LvSize,
		PoolLv:          result.PoolLv,
		Origin:          result.Origin,
		DataPercent:     result.DataPercent,
		MetadataPercent: result.MetadataPercent,
		MovePv:          result.MovePv,
		MirrorLog:       result.MirrorLog,
		CopyPercent:     result.CopyPercent,
		ConvertLv:       result.ConvertLv,
	}

	return metadata, nil
}

func newLv(volumeGroupID uuid.UUID, id uuid.UUID, size int64) (*api.Lv, error) {
	exists, _ := getLv(volumeGroupID, id)

	if exists != nil {
		return nil, errors.New("invalid_logical_volume")
	}

	sizeG := fmt.Sprintf("%dG", size)

	_, lvErr := exec.Command("lvcreate", "-n", id.String(), "-L", sizeG, volumeGroupID.String()).Output()

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

// lvDisplay : displays all logical volumes
//func lvDisplay() (*LogicalVolumeReport, error) {
//lvdisplay, err := exec.Command("lvdisplay", "--columns", "--reportformat", "json").Output()

//if err != nil {
//return nil, HandleInternalError(err)
//}

//// Handle output JSON
//output := LogicalVolumeReport{}
//if err := json.Unmarshal(lvdisplay, &output); err != nil {
//return nil, HandleInternalError(err)
//}

//// Return JSON data
//return &output, nil
//}
