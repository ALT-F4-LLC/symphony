package main

import (
	"encoding/json"
	"fmt"
	"os/exec"
)

// LVStruct : struct for LVM volume
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

// LVCreate : creates LVM volume group
func LVCreate(group string, name string, size string) bool {
	lv := LVExists(group, name)
	if lv != nil {
		fmt.Println("LVM volume already exists.")
		return false
	}
	_, err := exec.Command("lvcreate", "-n", name, "-L", size, group).Output()
	if err != nil {
		fmt.Println(err)
		return false
	}
	fmt.Printf("Successfully created volume '%s' in '%s' group.", name, group)
	return true
}

// LVDisplay : displays all LVM volumes
func LVDisplay() *LVDisplayStruct {
	// Handle pvdisplay command
	lvd, err := exec.Command("lvdisplay", "--columns", "--reportformat", "json").Output()
	if err != nil {
		return nil
	}
	// Handle output JSON
	res := LVDisplayStruct{}
	if err := json.Unmarshal(lvd, &res); err != nil {
		return nil
	}
	// Return JSON data
	return &res
}

// LVExists : verifies if volume group exists
func LVExists(group string, name string) *LVStruct {
	// Handle pvdisplay command
	path := fmt.Sprintf("/dev/%s/%s", group, name)
	lvd, err := exec.Command("lvdisplay", "--columns", "--reportformat", "json", path).Output()
	if err != nil {
		return nil
	}
	// Handle output JSON
	res := LVDisplayStruct{}
	if err := json.Unmarshal(lvd, &res); err != nil {
		return nil
	}
	// Check if any volumes exist
	if len(res.Report) == 0 {
		fmt.Println("No existing LVM volumes.")
	} else {
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
				return &output
			}
		}
	}
	return nil
}

// LVRemove : removes LVM volume group
func LVRemove(group string, name string) bool {
	lv := LVExists(group, name)
	if lv == nil {
		fmt.Println("No existing LVM volume group.")
		return false
	}
	path := fmt.Sprintf("/dev/%s/%s", group, name)
	_, err := exec.Command("lvremove", "--force", path).Output()
	if err != nil {
		fmt.Println(err)
		return false
	}
	fmt.Printf("Successfully removed '%s' volume from '%s' group.", name, group)
	return true
}
