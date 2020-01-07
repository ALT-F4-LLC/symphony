package schemas

import (
	"github.com/google/uuid"
)

// PhysicalVolume : database entry for PhysicalVolume data
type PhysicalVolume struct {
	PrimaryKey
	PrimaryTimestamps
	Device    string `gorm:"unique_index:physical_volumes_device_service_id;not null;type:string"`
	Metadata  *PhysicalVolumeMetadata
	ServiceID uuid.UUID `gorm:"unique_index:physical_volumes_device_service_id;not null;type:uuid"`
}

// PhysicalVolumeMetadata : struct for describing a pv in LVM
type PhysicalVolumeMetadata struct {
	PvName string `json:"pv_name"`
	VgName string `json:"vg_name"`
	PvFmt  string `json:"pv_fmt"`
	PvAttr string `json:"pv_attr"`
	PvSize string `json:"pv_size"`
	PvFree string `json:"pv_free"`
}

// VolumeGroup : database entry for VolumeGroup data
type VolumeGroup struct {
	PrimaryKey
	PrimaryTimestamps
	Metadata         *VolumeGroupMetadata
	PhysicalVolumeID uuid.UUID `gorm:"unique_index:volume_groups_physical_volume_id_service_id;not null;type:uuid"`
	ServiceID        uuid.UUID `gorm:"unique_index:volume_groups_physical_volume_id_service_id;not null;type:uuid"`
}

// LogicalVolume : database entry for LogicalVolume data
type LogicalVolume struct {
	PrimaryKey
	PrimaryTimestamps
	Size          int64
	VolumeGroupID uuid.UUID
	ServiceID     uuid.UUID
}

// VolumeGroupMetadata : struct for virtual group output
type VolumeGroupMetadata struct {
	VgName    string `json:"vg_name"`
	PvCount   string `json:"pv_count"`
	LvCount   string `json:"lv_count"`
	SnapCount string `json:"snap_count"`
	VgAttr    string `json:"vg_attr"`
	VgSize    string `json:"vg_size"`
	VgFree    string `json:"vg_free"`
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
