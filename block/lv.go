package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os/exec"

	pb "gitlab.drn.io/erikreinert/symphony/proto"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// LVStruct : struct for logical volume
type LVStruct struct {
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

// LVDisplayStruct : struct for LVDisplay output
type LVDisplayStruct struct {
	Report []struct {
		Lv []struct {
			*LVStruct
		} `json:"lv"`
	} `json:"report"`
}

// lvCreate : creates a logical volume
func lvCreate(group string, name string, size string) (*LVStruct, error) {
	// Check if logical volume exists
	exists, _ := lvExists(group, name)
	if exists != nil {
		return nil, status.Error(codes.AlreadyExists, "lv already exists")
	}
	// Create logical volume
	_, lvcreateErr := exec.Command("lvcreate", "-n", name, "-L", size, group).Output()
	if lvcreateErr != nil {
		return nil, HandleInternalError(lvcreateErr)
	}
	// Lookup new volume
	lv, err := lvExists(group, name)
	if err != nil {
		return nil, err
	}
	// Return new volume
	return lv, nil
}

// lvDisplay : displays all logical volumes
func lvDisplay() (*LVDisplayStruct, error) {
	// Handle lvdisplay command
	lvdisplay, err := exec.Command("lvdisplay", "--columns", "--reportformat", "json").Output()
	if err != nil {
		return nil, HandleInternalError(err)
	}
	// Handle output JSON
	output := LVDisplayStruct{}
	if err := json.Unmarshal(lvdisplay, &output); err != nil {
		return nil, HandleInternalError(err)
	}
	// Return JSON data
	return &output, nil
}

// lvExists : verifies if logical volume exists
func lvExists(group string, name string) (*LVStruct, error) {
	// Handle lvdisplay command
	path := fmt.Sprintf("/dev/%s/%s", group, name)
	lvd, lvdError := exec.Command("lvdisplay", "--columns", "--reportformat", "json", path).Output()
	if lvdError != nil {
		return nil, HandleInternalError(lvdError)
	}
	// Handle output JSON
	res := LVDisplayStruct{}
	if err := json.Unmarshal(lvd, &res); err != nil {
		return nil, HandleInternalError(err)
	}
	// Check if any logical volumes exist
	if len(res.Report) > 0 {
		// Display data for each volume
		for _, lv := range res.Report[0].Lv {
			if lv.VgName == group && lv.LvName == name {
				output := LVStruct{
					LvName:          lv.LvName,
					VgName:          lv.VgName,
					LvAttr:          lv.LvAttr,
					LvSize:          lv.LvSize,
					PoolLv:          lv.PoolLv,
					Origin:          lv.Origin,
					DataPercent:     lv.DataPercent,
					MetadataPercent: lv.MetadataPercent,
					MovePv:          lv.MovePv,
					MirrorLog:       lv.MirrorLog,
					CopyPercent:     lv.CopyPercent,
					ConvertLv:       lv.ConvertLv,
				}
				return &output, nil
			}
		}
	}
	return nil, nil
}

// lvRemove : removes logical volume
func lvRemove(group string, name string) error {
	// Check if logical volume exists
	exists, _ := lvExists(group, name)
	if exists == nil {
		err := status.Error(codes.NotFound, "lv not found")
		return err
	}
	// Remove logical volume
	path := fmt.Sprintf("/dev/%s/%s", group, name)
	_, err := exec.Command("lvremove", "--force", path).Output()
	if err != nil {
		return HandleInternalError(err)
	}
	return nil
}

// CreateLv : implements proto.BlockServer CreateLv request
func (s *blockServer) CreateLv(ctx context.Context, in *pb.CreateLvRequest) (*pb.LvObject, error) {
	lv, err := lvCreate(in.Group, in.Name, in.Size)
	if err != nil {
		return nil, err
	}
	log.Printf("CreateLv: %s successfully created in %s with %s space.", in.Name, in.Group, in.Size)
	return &pb.LvObject{
		LvName:          lv.LvName,
		VgName:          lv.VgName,
		LvAttr:          lv.LvAttr,
		LvSize:          lv.LvSize,
		PoolLv:          lv.PoolLv,
		Origin:          lv.Origin,
		DataPercent:     lv.DataPercent,
		MetadataPercent: lv.MetadataPercent,
		MovePv:          lv.MovePv,
		MirrorLog:       lv.MirrorLog,
		CopyPercent:     lv.CopyPercent,
		ConvertLv:       lv.ConvertLv,
	}, nil
}

// GetLv : implements proto.BlockServer GetLv request
func (s *blockServer) GetLv(ctx context.Context, in *pb.GetLvRequest) (*pb.LvObject, error) {
	lv, err := lvExists(in.Group, in.Name)
	if err != nil {
		return nil, err
	}
	log.Printf("GetLv: %s successfully found in %s.", in.Name, in.Group)
	return &pb.LvObject{
		LvName:          lv.LvName,
		VgName:          lv.VgName,
		LvAttr:          lv.LvAttr,
		LvSize:          lv.LvSize,
		PoolLv:          lv.PoolLv,
		Origin:          lv.Origin,
		DataPercent:     lv.DataPercent,
		MetadataPercent: lv.MetadataPercent,
		MovePv:          lv.MovePv,
		MirrorLog:       lv.MirrorLog,
		CopyPercent:     lv.CopyPercent,
		ConvertLv:       lv.ConvertLv,
	}, nil
}

// RemoveLv : implements proto.BlockServer RemoveLv request
func (s *blockServer) RemoveLv(ctx context.Context, in *pb.RemoveLvRequest) (*pb.BlockMessage, error) {
	err := lvRemove(in.Group, in.Name)
	if err != nil {
		return nil, err
	}
	log.Printf("RemoveLv: %s successfully removed from %s.", in.Name, in.Group)
	return &pb.BlockMessage{Message: "success"}, nil
}
