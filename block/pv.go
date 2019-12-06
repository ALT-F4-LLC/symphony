package main

import (
	"encoding/json"
	"errors"
	"os/exec"
	"strings"
)

// PvTable : struct for describing a pv table entry
// type PvTable struct {
// 	ID     string
// 	Device string
// }

// PvEntry : database entry for Pv data
// type PvEntry struct {
// 	Device string
// 	ID     uuid.UUID
// }

// PvMetadata : struct for describing a pv in LVM
type PvMetadata struct {
	PvName string `json:"pv_name"`
	VgName string `json:"vg_name"`
	PvFmt  string `json:"pv_fmt"`
	PvAttr string `json:"pv_attr"`
	PvSize string `json:"pv_size"`
	PvFree string `json:"pv_free"`
}

// PvMetadataJSON : struct for describing a series of pvs in LVM
type PvMetadataJSON struct {
	Report []struct {
		Pv []struct {
			*PvMetadata
		} `json:"pv"`
	} `json:"report"`
}

// // pvCreate : creates a pv from a physical device
// func pvCreate(device string) (*PvMetadata, error) {
// 	// Check if PV already exists
// 	exists, _ := pvExists(device)
// 	// If exists - return error
// 	if exists != nil {
// 		return nil, status.Error(codes.AlreadyExists, "pv already exists")
// 	}
// 	// Create PV with command
// 	_, pvCreateError := exec.Command("pvcreate", device).Output()
// 	if pvCreateError != nil {
// 		return nil, HandleInternalError(pvCreateError)
// 	}
// 	// Lookup new device
// 	pv, err := pvExists(device)
// 	if err != nil {
// 		return nil, err
// 	}
// 	// Return new device
// 	return pv, nil
// }

// // pvDisplay : displays all pv devices
// func pvDisplay() (*PvMetadataJSON, error) {
// 	// Handle pvdisplay command
// 	pvdisplay, err := exec.Command("pvdisplay", "--columns", "--reportformat", "json").Output()
// 	if err != nil {
// 		return nil, HandleInternalError(err)
// 	}
// 	// Handle output JSON
// 	output := PvMetadataJSON{}
// 	if err := json.Unmarshal(pvdisplay, &output); err != nil {
// 		return nil, HandleInternalError(err)
// 	}
// 	// Return JSON data
// 	return &output, nil
// }

// pvExists : verifies if pv exists
func pvExists(device string) (*PvMetadata, error) {
	// Handle pvdisplay command
	cmd := exec.Command("pvdisplay", "--columns", "--reportformat", "json", "--quiet", device)
	pvd, pvdErr := cmd.CombinedOutput()
	notExists := strings.Contains(string(pvd), "Failed to find physical volume")
	if notExists {
		return nil, errors.New("invalid_pv_device")
	}
	if pvdErr != nil {
		return nil, pvdErr
	}
	// Handle output JSONs
	res := &PvMetadataJSON{}
	if err := json.Unmarshal(pvd, &res); err != nil {
		return nil, err
	}
	// Check if any volumes exist
	if len(res.Report) > 0 {
		// Display data for each volume
		for _, pv := range res.Report[0].Pv {
			if pv.PvName == device {
				output := PvMetadata{
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

// // pvRemove : removes pv if exists
// func pvRemove(device string) error {
// 	exists, _ := pvExists(device)
// 	// Handle if no PV exists
// 	if exists == nil {
// 		err := status.Error(codes.NotFound, "pv not found")
// 		return err
// 	}
// 	// Remove PV from LVM
// 	_, err := exec.Command("pvremove", "--force", device).Output()
// 	if err != nil {
// 		return HandleInternalError(err)
// 	}
// 	return nil
// }

// // CreatePvHandler : creates a physical volume entry in database and on host
// func CreatePvHandler(db *sql.DB) http.HandlerFunc {
// 	return func(w http.ResponseWriter, r *http.Request) {
// 		params := mux.Vars(r)
// 		device := params["device"]

// 		// TODO: Lookup pv in database based off of device

// 		pv, err := pvCreate(device)
// 		if err != nil {
// 			return nil, err
// 		}

// 		// TODO: Create PV entry in database for reference

// 		id := uuid.New()
// 		sql := `INSERT INTO pv(id, device) VALUES ($1, $2) RETURNING id`
// 		err = s.db.QueryRow(sql, id, in.Device).Scan(&id)
// 		if err != nil {
// 			panic(err)
// 		}
// 		log.Printf("CreatePv: %s successfully created.", in.Device)
// 		return id, nil
// 	}
// }

// // RemovePv : implements proto.BlockServer RemovePv request
// func (s *blockServer) RemovePv(ctx context.Context, in *pb.RemovePvRequest) (*pb.GenericResponse, error) {
// 	err := pvRemove(in.Device)
// 	if err != nil {
// 		return nil, err
// 	}
// 	log.Printf("RemovePv: %s successfully removed.", in.Device)
// 	return &pb.BlockMessage{Message: "success"}, nil
// }
