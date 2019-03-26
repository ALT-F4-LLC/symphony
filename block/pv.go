package main

import (
	"encoding/json"
	"os/exec"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// PVStruct : struct for PVDisplay output
type PVStruct struct {
	PvName string `json:"pv_name"`
	VgName string `json:"vg_name"`
	PvFmt  string `json:"pv_fmt"`
	PvAttr string `json:"pv_attr"`
	PvSize string `json:"pv_size"`
	PvFree string `json:"pv_free"`
}

// PVDisplayStruct : struct for PVDisplay output
type PVDisplayStruct struct {
	Report []struct {
		Pv []struct {
			*PVStruct
		} `json:"pv"`
	} `json:"report"`
}

// PVCreate : creates LVM physical device
func PVCreate(device string) (*PVStruct, error) {
	// Check if PV already exists
	pvExists, _ := PVExists(device)
	// If exists - return error
	if pvExists != nil {
		return nil, status.Error(codes.AlreadyExists, "pv already exists")
	}
	// Create PV with command
	_, pvCreateError := exec.Command("pvcreate", device).Output()
	if pvCreateError != nil {
		return nil, HandleInternalError(pvCreateError)
	}
	// Lookup new device
	pv, err := PVExists(device)
	if err != nil {
		return nil, err
	}
	// Return new device
	return pv, nil
}

// PVDisplay : displays all LVM devices
func PVDisplay() (*PVDisplayStruct, error) {
	// Handle pvdisplay command
	pvDisplay, pvDisplayError := exec.Command("pvdisplay", "--columns", "--reportformat", "json").Output()
	if pvDisplayError != nil {
		return nil, HandleInternalError(pvDisplayError)
	}
	// Handle output JSON
	output := PVDisplayStruct{}
	if err := json.Unmarshal(pvDisplay, &output); err != nil {
		return nil, HandleInternalError(pvDisplayError)
	}
	// Return JSON data
	return &output, nil
}

// PVExists : verifies if PV device exists
func PVExists(device string) (*PVStruct, error) {
	// Handle pvdisplay command
	pvd, pvdErr := exec.Command("pvdisplay", "--columns", "--reportformat", "json", device).Output()
	if pvdErr != nil {
		return nil, HandleInternalError(pvdErr)
	}
	// Handle output JSON
	res := PVDisplayStruct{}
	if err := json.Unmarshal(pvd, &res); err != nil {
		return nil, HandleInternalError(err)
	}
	// Check if any volumes exist
	if len(res.Report) > 0 {
		// Display data for each volume
		for _, pv := range res.Report[0].Pv {
			if pv.PvName == device {
				output := PVStruct{
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

// PVRemove : removes LVM physical device
func PVRemove(device string) error {
	pvExists, _ := PVExists(device)
	// Handle if no LVM exists
	if pvExists == nil {
		err := status.Error(codes.NotFound, "pv not found")
		return err
	}
	_, pvRemoveError := exec.Command("pvremove", "--force", device).Output()
	if pvRemoveError != nil {
		return HandleInternalError(pvRemoveError)
	}
	return nil
}
