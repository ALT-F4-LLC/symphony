package schema

import "github.com/google/uuid"

// LogicalVolume : database entry for LogicalVolume data
type LogicalVolume struct {
	Size          string
	VolumeGroupID uuid.UUID
	ServiceID     uuid.UUID
}

// LogicalVolumeMetadata : struct for logical volume
type LogicalVolumeMetadata struct {
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
