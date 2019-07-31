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

// VGStruct : struct for virtual group output
type VGStruct struct {
	VgName    string `json:"vg_name"`
	PvCount   string `json:"pv_count"`
	LvCount   string `json:"lv_count"`
	SnapCount string `json:"snap_count"`
	VgAttr    string `json:"vg_attr"`
	VgSize    string `json:"vg_size"`
	VgFree    string `json:"vg_free"`
}

// VGDisplayStruct : struct for VGDisplay output
type VGDisplayStruct struct {
	Report []struct {
		Vg []struct {
			*VGStruct
		} `json:"vg"`
	} `json:"report"`
}

// VGCreate : creates LVM volume group
func vgCreate(device string, group string) (*VGStruct, error) {
	// Check if a VG already exists
	exists, _ := vgExists(group)
	// If exists - return error
	if exists != nil {
		return nil, status.Error(codes.AlreadyExists, "vg already exists")
	}
	// Create VG with command
	_, vgcreateErr := exec.Command("vgcreate", group, device).Output()
	if vgcreateErr != nil {
		return nil, HandleInternalError(vgcreateErr)
	}
	// Lookup new device
	vg, err := vgExists(group)
	if err != nil {
		return nil, err
	}
	// Return new volume group
	return vg, nil
}

// VGDisplay : displays all LVM devices
func vgDisplay() (*VGDisplayStruct, error) {
	// Handle pvdisplay command
	vgdisplay, err := exec.Command("vgdisplay", "--columns", "--reportformat", "json").Output()
	if err != nil {
		return nil, HandleInternalError(err)
	}
	// Handle output JSON
	output := VGDisplayStruct{}
	if err := json.Unmarshal(vgdisplay, &output); err != nil {
		return nil, HandleInternalError(err)
	}
	// Return JSON data
	return &output, nil
}

// VGExists : verifies if volume group exists
func vgExists(group string) (*VGStruct, error) {
	// Handle vgdisplay command
	vgd, vgdErr := exec.Command("vgdisplay", "--columns", "--reportformat", "json", group).Output()
	if vgdErr != nil {
		return nil, HandleInternalError(vgdErr)
	}
	// Handle output JSON
	res := VGDisplayStruct{}
	if err := json.Unmarshal(vgd, &res); err != nil {
		return nil, HandleInternalError(err)
	}
	// Check if any volumes exist
	if len(res.Report) > 0 {
		// Display data for each volume
		for _, vg := range res.Report[0].Vg {
			if vg.VgName == group {
				output := VGStruct{
					VgName:    vg.VgName,
					PvCount:   vg.PvCount,
					LvCount:   vg.LvCount,
					SnapCount: vg.SnapCount,
					VgAttr:    vg.VgAttr,
					VgSize:    vg.VgSize,
					VgFree:    vg.VgFree,
				}
				return &output, nil
			}
		}
	}
	return nil, nil
}

// VGRemove : removes LVM volume group
func vgRemove(group string) error {
	exists, _ := vgExists(group)
	if exists == nil {
		err := status.Error(codes.NotFound, "vg not found")
		return err
	}
	_, err := exec.Command("vgremove", "--force", group).Output()
	if err != nil {
		return HandleInternalError(err)
	}
	return nil
}

// CreateVg : implements proto.BlockServer CreateVg request
func (s *blockServer) CreateVg(ctx context.Context, in *pb.CreateVgRequest) (*pb.VgObject, error) {
	vg, err := vgCreate(in.Device, in.Group)
	if err != nil {
		return nil, err
	}
	log.Printf("CreateVg: %s successfully created for %s.", in.Group, in.Device)
	return &pb.VgObject{
		VgName:    vg.VgName,
		PvCount:   vg.PvCount,
		LvCount:   vg.LvCount,
		SnapCount: vg.SnapCount,
		VgAttr:    vg.VgAttr,
		VgSize:    vg.VgSize,
		VgFree:    vg.VgFree,
	}, nil
}

// GetVg : implements proto.BlockServer GetVg request
func (s *blockServer) GetVg(ctx context.Context, in *pb.GetVgRequest) (*pb.VgObject, error) {
	vg, err := vgExists(in.Group)
	if err != nil {
		return nil, err
	}
	log.Printf("GetVg: %s successfully found.", in.Group)
	return &pb.VgObject{
		VgName:    vg.VgName,
		PvCount:   vg.PvCount,
		LvCount:   vg.LvCount,
		SnapCount: vg.SnapCount,
		VgAttr:    vg.VgAttr,
		VgSize:    vg.VgSize,
		VgFree:    vg.VgFree,
	}, nil
}

// RemoveVg : implements proto.BlockServer RemoveVg request
func (s *blockServer) RemoveVg(ctx context.Context, in *pb.RemoveVgRequest) (*pb.BlockMessage, error) {
	err := vgRemove(in.Group)
	if err != nil {
		return nil, err
	}
	log.Printf("RemovePv: %s successfully removed.", in.Group)
	return &pb.BlockMessage{Message: "success"}, nil
}
