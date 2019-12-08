package main

import (
	"encoding/json"
	"errors"
	"net/http"
	"os/exec"
	"strings"

	"github.com/erkrnt/symphony/schemas"
	"github.com/sirupsen/logrus"
)

// PhysicalVolumeReport : describes a series of physical volumes in LVM
type PhysicalVolumeReport struct {
	Report []struct {
		Pv []schemas.PhysicalVolumeMetadata `json:"pv"`
	} `json:"report"`
}

// getPhysicalVolume : if exists - gets physical volume in LVM
func getPhysicalVolume(device string) (*schemas.PhysicalVolumeMetadata, error) {
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
			logrus.WithFields(logrus.Fields{"device": device}).Debug("Physical volume successfully discovered.")
		}
	}

	return &metadata, nil
}

// newPhysicalVolume : creates a physical device in LVM
func newPhysicalVolume(device string) (*schemas.PhysicalVolumeMetadata, error) {
	exists, _ := getPhysicalVolume(device)
	if exists != nil {
		return nil, errors.New("pv_already_exists")
	}

	cmd := exec.Command("pvcreate", device)
	pvc, pvcErr := cmd.CombinedOutput()
	if pvcErr != nil {
		return nil, errors.New(strings.TrimSpace(string(pvc)))
	}

	pv, err := getPhysicalVolume(device)
	if err != nil {
		return nil, err
	}

	logrus.WithFields(logrus.Fields{"device": device}).Debug("Physical volume successfully created.")

	return pv, nil
}

// removePhysicalVolme : if exists - removes physical volume in LVM
func removePhysicalVolme(device string) error {
	exists, _ := getPhysicalVolume(device)
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

// GetPhysicalVolumeByDeviceHandler : handles getting specific physical volumes
func GetPhysicalVolumeByDeviceHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		device := r.FormValue("device")
		json := make([]schemas.PhysicalVolumeMetadata, 0)

		pv, pvErr := getPhysicalVolume(device)
		if pvErr != nil {
			HandleErrorResponse(w, pvErr)
			return
		}
		if pv != nil {
			json = append(json, *pv)
		}

		HandleResponse(w, json)
	}
}

// PostPhysicalVolumeHandler : creates a physical volume on host
func PostPhysicalVolumeHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		body := RequestBody{}
		_ = json.NewDecoder(r.Body).Decode(&body)

		pv, err := newPhysicalVolume(body.Device)
		if err != nil {
			HandleErrorResponse(w, err)
			return
		}

		json := append([]schemas.PhysicalVolumeMetadata{}, *pv)
		HandleResponse(w, json)
	}
}

// DeletePhysicalVolumeHandler : delete a physical volume on host
func DeletePhysicalVolumeHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		body := RequestBody{}
		_ = json.NewDecoder(r.Body).Decode(&body)

		err := removePhysicalVolme(body.Device)
		if err != nil {
			HandleErrorResponse(w, err)
			return
		}

		json := make([]schemas.PhysicalVolumeMetadata, 0)
		HandleResponse(w, json)
	}
}
