package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/erkrnt/symphony/schemas"
	"github.com/google/uuid"
	"github.com/jinzhu/gorm"
)

// CreatePhysicalVolumeOnHost : creates a PhysicalVolume on a block host
func CreatePhysicalVolumeOnHost(serviceHostname string, device string) (*schemas.PhysicalVolumeMetadata, error) {
	data := fmt.Sprintf(`{"device":"%s"}`, device)
	url := fmt.Sprintf("http://%s:50051/physicalvolume", serviceHostname)
	res, err := http.Post(url, "application/json", bytes.NewBuffer([]byte(data)))
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	if res.StatusCode != 200 {
		data := ErrorResponse{}
		if err := json.Unmarshal(body, &data); err != nil {
			return nil, err
		}
		return nil, errors.New(data.Message)
	}

	pvMetadata := []schemas.PhysicalVolumeMetadata{}
	if err := json.Unmarshal(body, &pvMetadata); err != nil {
		return nil, err
	}

	return &pvMetadata[0], nil
}

// GetPhysicalVolumeMetadataByDeviceOnHost : retrieves PhysicalVolumeMetadata from host
func GetPhysicalVolumeMetadataByDeviceOnHost(db *gorm.DB, device string, serviceHostname string) (*schemas.PhysicalVolumeMetadata, error) {
	base := fmt.Sprintf("http://%s:50051/physicalvolume", serviceHostname)
	url := fmt.Sprintf("%s?device=%s", base, device)
	res, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	pvMetadata := make([]schemas.PhysicalVolumeMetadata, 0)
	if err := json.Unmarshal(body, &pvMetadata); err != nil {
		return nil, err
	}
	if len(pvMetadata) < 1 {
		return nil, errors.New("invalid_physical_volume")
	}

	return &pvMetadata[0], nil
}

// DeletePhysicalVolumeOnHost : deletes a PhysicalVolume on block host
func DeletePhysicalVolumeOnHost(serviceHostname string, device string) error {
	data := fmt.Sprintf(`{"device":"%s"}`, device)
	url := fmt.Sprintf("http://%s:50051/physicalvolume", serviceHostname)

	req, err := http.NewRequest("DELETE", url, strings.NewReader(data))
	if err != nil {
		return err
	}

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return err
	}

	if res.StatusCode != 200 {
		data := ErrorResponse{}
		if err := json.Unmarshal(body, &data); err != nil {
			return err
		}
		return errors.New(data.Message)
	}

	return nil
}

// CreateVolumeGroupOnHost : creates a VolumeGroup on block host
func CreateVolumeGroupOnHost(serviceHostname string, physicalVolumeDevice string, volumeGroupID uuid.UUID) (*schemas.VolumeGroupMetadata, error) {
	data := fmt.Sprintf(`{"device":"%s","id":"%s"}`, physicalVolumeDevice, volumeGroupID.String())
	url := fmt.Sprintf("http://%s:50051/volumegroup", serviceHostname)
	res, err := http.Post(url, "application/json", bytes.NewBuffer([]byte(data)))
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	if res.StatusCode != 200 {
		data := ErrorResponse{}
		if err := json.Unmarshal(body, &data); err != nil {
			return nil, err
		}
		return nil, errors.New(data.Message)
	}

	volumeGroupMetadata := []schemas.VolumeGroupMetadata{}
	if err := json.Unmarshal(body, &volumeGroupMetadata); err != nil {
		return nil, err
	}

	return &volumeGroupMetadata[0], nil
}

// GetVolumeGroupMetadataByIDOnHost : retrieves PhysicalVolumeMetadata from host
func GetVolumeGroupMetadataByIDOnHost(db *gorm.DB, serviceHostname string, id uuid.UUID) (*schemas.VolumeGroupMetadata, error) {
	url := fmt.Sprintf("http://%s:50051/volumegroup/%s", serviceHostname, id.String())
	res, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	metadata := make([]schemas.VolumeGroupMetadata, 0)
	if err := json.Unmarshal(body, &metadata); err != nil {
		return nil, err
	}
	if len(metadata) < 1 {
		return nil, errors.New("invalid_volume_group")
	}

	return &metadata[0], nil
}

// DeleteVolumeGroupOnHost : deletes a PhysicalVolume on block host
func DeleteVolumeGroupOnHost(serviceHostname string, volumeGroupID uuid.UUID) error {
	url := fmt.Sprintf("http://%s:50051/volumegroup/%s", serviceHostname, volumeGroupID.String())
	req, err := http.NewRequest("DELETE", url, nil)
	if err != nil {
		return err
	}

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return err
	}

	if res.StatusCode != 200 {
		data := ErrorResponse{}
		if err := json.Unmarshal(body, &data); err != nil {
			return err
		}
		return errors.New(data.Message)
	}

	return nil
}
