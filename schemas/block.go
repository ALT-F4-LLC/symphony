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
