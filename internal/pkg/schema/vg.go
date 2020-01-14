package schema

import "github.com/google/uuid"

// VolumeGroup : database entry for VolumeGroup data
type VolumeGroup struct {
	Metadata         *VolumeGroupMetadata
	PhysicalVolumeID uuid.UUID `gorm:"unique_index:volume_groups_physical_volume_id_service_id;not null;type:uuid"`
	ServiceID        uuid.UUID `gorm:"unique_index:volume_groups_physical_volume_id_service_id;not null;type:uuid"`
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
