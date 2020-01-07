package main

// LogicalVolumeReport : struct for LVDisplay output
// type LogicalVolumeReport struct {
// 	Report []struct {
// 		Lv []schemas.LogicalVolumeMetadata `json:"lv"`
// 	} `json:"report"`
// }

// // lvCreate : creates a logical volume
// func lvCreate(group string, name string, size string) (*LogicalVolumeMetadata, error) {
// 	// Check if logical volume exists
// 	exists, _ := getLogicalVolume(group, name)
// 	if exists != nil {
// 		return nil, status.Error(codes.AlreadyExists, "lv already exists")
// 	}
// 	// Create logical volume
// 	_, lvcreateErr := exec.Command("lvcreate", "-n", name, "-L", size, group).Output()
// 	if lvcreateErr != nil {
// 		return nil, HandleInternalError(lvcreateErr)
// 	}
// 	// Lookup new volume
// 	lv, err := getLogicalVolume(group, name)
// 	if err != nil {
// 		return nil, err
// 	}
// 	// Return new volume
// 	return lv, nil
// }

// // lvDisplay : displays all logical volumes
// func lvDisplay() (*LogicalVolumeReport, error) {
// 	// Handle lvdisplay command
// 	lvdisplay, err := exec.Command("lvdisplay", "--columns", "--reportformat", "json").Output()
// 	if err != nil {
// 		return nil, HandleInternalError(err)
// 	}
// 	// Handle output JSON
// 	output := LogicalVolumeReport{}
// 	if err := json.Unmarshal(lvdisplay, &output); err != nil {
// 		return nil, HandleInternalError(err)
// 	}
// 	// Return JSON data
// 	return &output, nil
// }

// // lvRemove : removes logical volume
// func lvRemove(group string, name string) error {
// 	// Check if logical volume exists
// 	exists, _ := getLogicalVolume(group, name)
// 	if exists == nil {
// 		err := status.Error(codes.NotFound, "lv not found")
// 		return err
// 	}
// 	// Remove logical volume
// 	path := fmt.Sprintf("/dev/%s/%s", group, name)
// 	_, err := exec.Command("lvremove", "--force", path).Output()
// 	if err != nil {
// 		return HandleInternalError(err)
// 	}
// 	return nil
// }

// getLogicalVolume : verifies if logical volume exists
// func getLogicalVolume(groupID uuid.UUID, id uuid.UUID) (*schemas.LogicalVolumeMetadata, error) {
// 	path := fmt.Sprintf("/dev/%s/%s", groupID.String(), id.String())
// 	cmd := exec.Command("lvdisplay", "--columns", "--reportformat", "json", path)
// 	lvd, lvdErr := cmd.CombinedOutput()
// 	output := fmt.Sprintf("Volume group \"%s\" not found", id)
// 	notExists := strings.Contains(string(lvd), output)
// 	if notExists {
// 		return nil, nil
// 	}
// 	if lvdErr != nil {
// 		return nil, lvdErr
// 	}

// 	res := LogicalVolumeReport{}
// 	if err := json.Unmarshal(lvd, &res); err != nil {
// 		return nil, err
// 	}

// 	var metadata schemas.LogicalVolumeMetadata
// 	if len(res.Report) == 1 && len(res.Report[0].Lv) == 1 {
// 		lv := res.Report[0].Lv[0]
// 		if lv.VgName == groupID.String() && lv.LvName == id.String() {
// 			metadata = lv
// 			logrus.WithFields(logrus.Fields{"id": id, "volume_group_id": groupID.String()}).Debug("Logical volume successfully discovered.")
// 		}
// 	}

// 	return &metadata, nil
// }
