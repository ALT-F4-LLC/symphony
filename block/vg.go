package main

import (
	"encoding/json"
	"fmt"
	"os/exec"
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
func VGCreate(device string, group string) bool {
	vg := VGExists(group)
	if vg != nil {
		fmt.Println("Already existing LVM virtual group.")
		return false
	}
	_, err := exec.Command("vgcreate", group, device).Output()
	if err != nil {
		fmt.Println(err)
		return false
	}
	fmt.Printf("Successfully created volume group '%s' for '%s'.", group, device)
	return true
}

// VGDisplay : displays all LVM devices
func VGDisplay() *VGDisplayStruct {
	// Handle pvdisplay command
	vgd, err := exec.Command("vgdisplay", "--columns", "--reportformat", "json").Output()
	if err != nil {
		return nil
	}
	// Handle output JSON
	res := VGDisplayStruct{}
	if err := json.Unmarshal(vgd, &res); err != nil {
		return nil
	}
	// Return JSON data
	return &res
}

// VGExists : verifies if volume group exists
func VGExists(group string) *VGStruct {
	// Handle pvdisplay command
	vgd, err := exec.Command("vgdisplay", "--columns", "--reportformat", "json", group).Output()
	if err != nil {
		return nil
	}
	// Handle output JSON
	res := VGDisplayStruct{}
	if err := json.Unmarshal(vgd, &res); err != nil {
		return nil
	}
	// Check if any volumes exist
	if len(res.Report) == 0 {
		fmt.Println("No existing LVM volume groups.")
	} else {
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
				return &output
			}
		}
	}
	return nil
}

// VGRemove : removes LVM volume group
func VGRemove(group string) bool {
	vg := VGExists(group)
	if vg == nil {
		fmt.Println("No existing LVM volume group.")
		return false
	}
	_, err := exec.Command("vgremove", "--force", group).Output()
	if err != nil {
		fmt.Println(err)
		return false
	}
	fmt.Printf("Successfully removed '%s' volume group.", group)
	return true
}
