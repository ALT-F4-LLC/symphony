package manager

import (
	"encoding/json"
	"fmt"

	"github.com/erkrnt/symphony/api"
	"github.com/erkrnt/symphony/internal/utils"
	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func logicalVolumes(consulAddr string) ([]*api.LogicalVolume, error) {
	resourceType := api.ResourceType_LOGICAL_VOLUME

	key := fmt.Sprintf("%s/", resourceType.String())

	results, err := kvPairs(consulAddr, key)

	if err != nil {
		st := status.New(codes.Internal, err.Error())

		return nil, st.Err()
	}

	lvs := make([]*api.LogicalVolume, 0)

	for _, ev := range results {
		var lv *api.LogicalVolume

		err := json.Unmarshal(ev.Value, &lv)

		if err != nil {
			st := status.New(codes.Internal, err.Error())

			return nil, st.Err()
		}

		lvs = append(lvs, lv)
	}

	return lvs, nil
}

func logicalVolumeByID(consulAddr string, id uuid.UUID) (*api.LogicalVolume, error) {
	key := utils.KvKey(id, api.ResourceType_LOGICAL_VOLUME)

	result, err := kvPair(consulAddr, key)

	if err != nil {
		st := status.New(codes.Internal, err.Error())

		return nil, st.Err()
	}

	var lv *api.LogicalVolume

	jsonErr := json.Unmarshal(result.Value, &lv)

	if jsonErr != nil {
		st := status.New(codes.Internal, err.Error())

		return nil, st.Err()
	}

	return lv, nil
}

func physicalVolumes(consulAddr string) ([]*api.PhysicalVolume, error) {
	resourceType := api.ResourceType_PHYSICAL_VOLUME

	key := fmt.Sprintf("%s/", resourceType.String())

	results, err := kvPairs(consulAddr, key)

	if err != nil {
		st := status.New(codes.Internal, err.Error())

		return nil, st.Err()
	}

	pvs := make([]*api.PhysicalVolume, 0)

	for _, ev := range results {
		var s *api.PhysicalVolume

		err := json.Unmarshal(ev.Value, &s)

		if err != nil {
			st := status.New(codes.Internal, err.Error())

			return nil, st.Err()
		}

		pvs = append(pvs, s)
	}

	return pvs, nil
}

func physicalVolumeByID(consulAddr string, id uuid.UUID) (*api.PhysicalVolume, error) {
	key := utils.KvKey(id, api.ResourceType_PHYSICAL_VOLUME)

	result, err := kvPair(consulAddr, key)

	if err != nil {
		st := status.New(codes.Internal, err.Error())

		return nil, st.Err()
	}

	var pv *api.PhysicalVolume

	jsonErr := json.Unmarshal(result.Value, &pv)

	if jsonErr != nil {
		st := status.New(codes.Internal, err.Error())

		return nil, st.Err()
	}

	return pv, nil
}

func volumeGroups(consulAddr string) ([]*api.VolumeGroup, error) {
	resourceType := api.ResourceType_VOLUME_GROUP

	key := fmt.Sprintf("%s/", resourceType.String())

	results, err := kvPairs(consulAddr, key)

	if err != nil {
		st := status.New(codes.Internal, err.Error())

		return nil, st.Err()
	}

	vgs := make([]*api.VolumeGroup, 0)

	for _, ev := range results {
		var s *api.VolumeGroup

		err := json.Unmarshal(ev.Value, &s)

		if err != nil {
			st := status.New(codes.Internal, err.Error())

			return nil, st.Err()
		}

		vgs = append(vgs, s)
	}

	return vgs, nil
}

func volumeGroupByID(consulAddr string, id uuid.UUID) (*api.VolumeGroup, error) {
	key := utils.KvKey(id, api.ResourceType_VOLUME_GROUP)

	result, err := kvPair(consulAddr, key)

	if err != nil {
		st := status.New(codes.Internal, err.Error())

		return nil, st.Err()
	}

	var vg *api.VolumeGroup

	jsonErr := json.Unmarshal(result.Value, &vg)

	if jsonErr != nil {
		st := status.New(codes.Internal, jsonErr.Error())

		return nil, st.Err()
	}

	return vg, nil
}
