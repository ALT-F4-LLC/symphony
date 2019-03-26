package main

import (
	"context"
	"encoding/json"
	"log"
	"os/exec"

	pb "gitlab.drn.io/erikreinert/symphony/proto"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// PvStruct : struct for describing a pv in LVM
type PvStruct struct {
	PvName string `json:"pv_name"`
	VgName string `json:"vg_name"`
	PvFmt  string `json:"pv_fmt"`
	PvAttr string `json:"pv_attr"`
	PvSize string `json:"pv_size"`
	PvFree string `json:"pv_free"`
}

// PvDisplayStruct : struct for describing a series of pvs in LVM
type PvDisplayStruct struct {
	Report []struct {
		Pv []struct {
			*PvStruct
		} `json:"pv"`
	} `json:"report"`
}

func pvCreate(device string) (*PvStruct, error) {
	// Check if PV already exists
	exists, _ := pvExists(device)
	// If exists - return error
	if exists != nil {
		return nil, status.Error(codes.AlreadyExists, "pv already exists")
	}
	// Create PV with command
	_, pvCreateError := exec.Command("pvcreate", device).Output()
	if pvCreateError != nil {
		return nil, HandleInternalError(pvCreateError)
	}
	// Lookup new device
	pv, err := pvExists(device)
	if err != nil {
		return nil, err
	}
	// Return new device
	return pv, nil
}

func pvDisplay() (*PvDisplayStruct, error) {
	// Handle pvdisplay command
	pvdisplay, err := exec.Command("pvdisplay", "--columns", "--reportformat", "json").Output()
	if err != nil {
		return nil, HandleInternalError(err)
	}
	// Handle output JSON
	output := PvDisplayStruct{}
	if err := json.Unmarshal(pvdisplay, &output); err != nil {
		return nil, HandleInternalError(err)
	}
	// Return JSON data
	return &output, nil
}

func pvExists(device string) (*PvStruct, error) {
	// Handle pvdisplay command
	pvd, pvdErr := exec.Command("pvdisplay", "--columns", "--reportformat", "json", device).Output()
	if pvdErr != nil {
		return nil, HandleInternalError(pvdErr)
	}
	// Handle output JSON
	res := PvDisplayStruct{}
	if err := json.Unmarshal(pvd, &res); err != nil {
		return nil, HandleInternalError(err)
	}
	// Check if any volumes exist
	if len(res.Report) > 0 {
		// Display data for each volume
		for _, pv := range res.Report[0].Pv {
			if pv.PvName == device {
				output := PvStruct{
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

func pvRemove(device string) error {
	exists, _ := pvExists(device)
	// Handle if no PV exists
	if exists == nil {
		err := status.Error(codes.NotFound, "pv not found")
		return err
	}
	// Remove PV from LVM
	_, err := exec.Command("pvremove", "--force", device).Output()
	if err != nil {
		return HandleInternalError(err)
	}
	return nil
}

// CreatePv : implements proto.BlockServer CreatePv request
func (s *server) CreatePv(ctx context.Context, in *pb.PvRequest) (*pb.PvObject, error) {
	pv, err := pvCreate(in.Device)
	if err != nil {
		return nil, err
	}
	log.Printf("CreatePv: %s successfully created.", in.Device)
	return &pb.PvObject{
		PvName: pv.PvName,
		VgName: pv.VgName,
		PvFmt:  pv.PvFmt,
		PvAttr: pv.PvAttr,
		PvSize: pv.PvSize,
		PvFree: pv.PvFree,
	}, nil
}

// GetPv : implements proto.BlockServer GetPv request
func (s *server) GetPv(ctx context.Context, in *pb.PvRequest) (*pb.PvObject, error) {
	pv, err := pvExists(in.Device)
	if err != nil {
		return nil, err
	}
	log.Printf("GetPv: %s successfully found.", in.Device)
	return &pb.PvObject{
		PvName: pv.PvName,
		VgName: pv.VgName,
		PvFmt:  pv.PvFmt,
		PvAttr: pv.PvAttr,
		PvSize: pv.PvSize,
		PvFree: pv.PvFree,
	}, nil
}

// RemovePv : implements proto.BlockServer RemovePv request
func (s *server) RemovePv(ctx context.Context, in *pb.PvRequest) (*pb.PvMessage, error) {
	err := pvRemove(in.Device)
	if err != nil {
		return nil, err
	}
	log.Printf("RemovePv: %s successfully removed.", in.Device)
	return &pb.PvMessage{Message: "success"}, nil
}
