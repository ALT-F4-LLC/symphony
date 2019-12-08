package main

import (
	"encoding/json"
	"errors"
	"log"
	"os/exec"
	"strings"

	"github.com/erkrnt/symphony/schemas"
)

// PvReport : struct for describing a series of pvs in LVM
type PvReport struct {
	Report []struct {
		Pv []struct {
			*schemas.PhysicalVolumeMetadata
		} `json:"pv"`
	} `json:"report"`
}

// // pvDisplay : displays all pv devices
// func pvDisplay() (*PvReport, error) {
// 	// Handle pvdisplay command
// 	pvdisplay, err := exec.Command("pvdisplay", "--columns", "--reportformat", "json").Output()
// 	if err != nil {
// 		return nil, HandleInternalError(err)
// 	}
// 	// Handle output JSON
// 	output := PvReport{}
// 	if err := json.Unmarshal(pvdisplay, &output); err != nil {
// 		return nil, HandleInternalError(err)
// 	}
// 	// Return JSON data
// 	return &output, nil
// }

// pvRemove : removes pv if exists
func pvRemove(device string) error {
	exists, _ := pvExists(device)
	if exists == nil {
		err := errors.New("pv_not_found")
		return err
	}

	cmd := exec.Command("pvremove", "--force", device)
	_, pvrErr := cmd.CombinedOutput()
	if pvrErr != nil {
		return pvrErr
	}

	log.Printf("pvRemove: %s successfully remove.", device)

	return nil
}

// pvExists : verifies if pv exists
func pvExists(device string) (*schemas.PhysicalVolumeMetadata, error) {
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

	if len(res.Report) > 0 {
		for _, pv := range res.Report[0].Pv {
			if pv.PvName == device {
				output := schemas.PhysicalVolumeMetadata{
					PvName: pv.PvName,
					VgName: pv.VgName,
					PvFmt:  pv.PvFmt,
					PvAttr: pv.PvAttr,
					PvSize: pv.PvSize,
					PvFree: pv.PvFree,
				}
				return &output, nil
			}
		}
	}
	return nil, nil
}

// pvCreate : creates a pv from a physical device
func pvCreate(device string) (*schemas.PhysicalVolumeMetadata, error) {
	exists, _ := pvExists(device)
	if exists != nil {
		err := "pv_already_exists"
		log.Printf("[ERROR] pvExists: %s", err)
		return nil, errors.New(err)
	}

	cmd := exec.Command("pvcreate", device)
	pvc, pvcErr := cmd.CombinedOutput()
	if pvcErr != nil {
		return nil, errors.New(strings.TrimSpace(string(pvc)))
	}

	pv, err := pvExists(device)
	if err != nil {
		return nil, err
	}

	log.Printf("pvCreate: %s successfully created.", device)

	return pv, nil
}
