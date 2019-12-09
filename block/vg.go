package main

import (
	"encoding/json"
	"errors"
	"net/http"
	"os/exec"
	"strings"

	"github.com/erkrnt/symphony/schemas"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
)

// VolumeGroupReport : struct for VGDisplay output
type VolumeGroupReport struct {
	Report []struct {
		Vg []schemas.VolumeGroupMetadata `json:"vg"`
	} `json:"report"`
}

// getVolumeGroup : verifies if volume group exists
func getVolumeGroup(id uuid.UUID) (*schemas.VolumeGroupMetadata, error) {
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

	var metadata schemas.VolumeGroupMetadata
	if len(res.Report) == 1 && len(res.Report[0].Vg) == 1 {
		vg := res.Report[0].Vg[0]
		if vg.VgName == id.String() {
			metadata = vg
			logrus.WithFields(logrus.Fields{"id": id.String()}).Debug("VolumeGroup successfully discovered.")
		}
	}

	return &metadata, nil
}

// newVolumeGroup : creates volume group
func newVolumeGroup(device string, id uuid.UUID) (*schemas.VolumeGroupMetadata, error) {
	exists, _ := getVolumeGroup(id)
	if exists != nil {
		return nil, errors.New("vg_already_exists")
	}

	cmd := exec.Command("vgcreate", id.String(), device)
	vgc, vgcErr := cmd.CombinedOutput()
	if vgcErr != nil {
		return nil, errors.New(strings.TrimSpace(string(vgc)))
	}

	vg, err := getVolumeGroup(id)
	if err != nil {
		return nil, err
	}

	logrus.WithFields(logrus.Fields{"id": id.String()}).Debug("VolumeGroup successfully created.")

	return vg, nil
}

// removeVolumeGroup : removes volume group
func removeVolumeGroup(id uuid.UUID) error {
	exists, _ := getVolumeGroup(id)
	if exists == nil {
		return errors.New("vg_not_found")
	}

	cmd := exec.Command("vgremove", "--force", id.String())
	_, err := cmd.CombinedOutput()
	if err != nil {
		return err
	}

	logrus.WithFields(logrus.Fields{"id": id.String()}).Debug("VolumeGroup successfully removed.")

	return nil
}

// GetVolumeGroupByIDHandler : handles getting specific physical volumes
func GetVolumeGroupByIDHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		params := mux.Vars(r)
		groupID := params["id"]

		id, err := uuid.Parse(groupID)
		if err != nil {
			HandleErrorResponse(w, err)
			return
		}

		virtualGroup, err := getVolumeGroup(id)
		if err != nil {
			HandleErrorResponse(w, err)
			return
		}

		json := make([]schemas.VolumeGroupMetadata, 0)
		if virtualGroup != nil {
			json = append(json, *virtualGroup)
		}

		HandleResponse(w, json)
	}
}

// PostVolumeGroupHandler : creates a physical volume on host
func PostVolumeGroupHandler() http.HandlerFunc {
	type reqBody struct {
		Device string    `json:"device"`
		ID     uuid.UUID `json:"id"`
	}
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		body := reqBody{}
		_ = json.NewDecoder(r.Body).Decode(&body)

		vg, err := newVolumeGroup(body.Device, body.ID)
		if err != nil {
			HandleErrorResponse(w, err)
			return
		}

		json := append([]schemas.VolumeGroupMetadata{}, *vg)
		HandleResponse(w, json)
	}
}

// DeleteVolumeGroupHandler : delete a volume group on host
func DeleteVolumeGroupHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		params := mux.Vars(r)
		volumeGroupID := params["id"]

		id, err := uuid.Parse(volumeGroupID)
		if err != nil {
			HandleErrorResponse(w, err)
			return
		}

		rmErr := removeVolumeGroup(id)
		if rmErr != nil {
			HandleErrorResponse(w, rmErr)
			return
		}

		json := make([]schemas.PhysicalVolumeMetadata, 0)
		HandleResponse(w, json)
	}
}
