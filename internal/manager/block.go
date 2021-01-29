package manager

import (
	"context"
	"errors"
	"fmt"

	"github.com/erkrnt/symphony/api"
	"github.com/erkrnt/symphony/internal/utils"
	"github.com/google/uuid"
	consul "github.com/hashicorp/consul/api"
	"github.com/sirupsen/logrus"
)

func handleCreateErr(err error, rID string, rType api.ResourceType) error {
	fields := logrus.Fields{
		"ResourceError": err.Error(),
		"ResourceID":    rID,
		"ResourceType":  rType.String(),
	}

	logrus.WithFields(fields).Error("CREATE_FAILED")

	return err
}

func (m *Manager) lvCreate(id uuid.UUID) error {
	lv, err := m.logicalVolumeByID(id)

	if err != nil {
		return handleCreateErr(err, lv.ID, api.ResourceType_LOGICAL_VOLUME)
	}

	volumeGroupID, err := uuid.Parse(lv.VolumeGroupID)

	if err != nil {
		return handleCreateErr(err, lv.ID, api.ResourceType_LOGICAL_VOLUME)
	}

	vg, err := m.volumeGroupByID(volumeGroupID)

	if err != nil {
		return handleCreateErr(err, lv.ID, api.ResourceType_LOGICAL_VOLUME)
	}

	physicalVolumeID, err := uuid.Parse(vg.PhysicalVolumeID)

	if err != nil {
		return handleCreateErr(err, lv.ID, api.ResourceType_LOGICAL_VOLUME)
	}

	pv, err := m.physicalVolumeByID(physicalVolumeID)

	if err != nil {
		return handleCreateErr(err, lv.ID, api.ResourceType_LOGICAL_VOLUME)
	}

	serviceID, err := uuid.Parse(pv.ServiceID)

	if err != nil {
		return handleCreateErr(err, lv.ID, api.ResourceType_LOGICAL_VOLUME)
	}

	agentService, err := m.agentServiceByID(serviceID)

	if err != nil {
		return handleCreateErr(err, lv.ID, api.ResourceType_LOGICAL_VOLUME)
	}

	agentServiceHealth, err := m.agentServiceHealth(agentService.ID)

	if err != nil {
		return handleCreateErr(err, lv.ID, api.ResourceType_LOGICAL_VOLUME)
	}

	if agentServiceHealth == "critical" {
		lv.Status = api.ResourceStatus_CREATE_FAILED

		saveErr := m.saveLogicalVolume(lv)

		if saveErr != nil {
			return handleCreateErr(saveErr, lv.ID, api.ResourceType_LOGICAL_VOLUME)
		}

		return handleCreateErr(errors.New("service_unavailable"), lv.ID, api.ResourceType_LOGICAL_VOLUME)
	}

	blockAddr := fmt.Sprintf("%s:%d", agentService.Address, agentService.Port)

	conn, err := utils.NewClientConnTcp(blockAddr)

	if err != nil {
		return handleCreateErr(err, lv.ID, api.ResourceType_LOGICAL_VOLUME)
	}

	defer conn.Close()

	ctx, cancel := context.WithTimeout(context.Background(), utils.ContextTimeout)

	defer cancel()

	c := api.NewBlockClient(conn)

	lvCreateOptions := &api.RequestLvCreate{
		ID:            lv.ID,
		Size:          lv.Size,
		VolumeGroupID: lv.VolumeGroupID,
	}

	_, lvCreateErr := c.LvCreate(ctx, lvCreateOptions)

	if lvCreateErr != nil {
		lv.Status = api.ResourceStatus_CREATE_FAILED

		saveErr := m.saveLogicalVolume(lv)

		if saveErr != nil {
			return handleCreateErr(saveErr, lv.ID, api.ResourceType_LOGICAL_VOLUME)
		}

		return handleCreateErr(lvCreateErr, lv.ID, api.ResourceType_LOGICAL_VOLUME)
	}

	lv.Status = api.ResourceStatus_CREATE_COMPLETED

	completedErr := m.saveLogicalVolume(lv)

	if completedErr != nil {
		return handleCreateErr(completedErr, lv.ID, api.ResourceType_LOGICAL_VOLUME)
	}

	return nil
}

func (m *Manager) pvCreate(as *consul.AgentService, pv *api.PhysicalVolume) error {
	healthStatus, err := m.agentServiceHealth(as.ID)

	if err != nil {
		return handleCreateErr(err, pv.ID, api.ResourceType_PHYSICAL_VOLUME)
	}

	if healthStatus == "critical" {
		pv.Status = api.ResourceStatus_CREATE_FAILED

		saveErr := m.savePhysicalVolume(pv)

		if saveErr != nil {
			return handleCreateErr(saveErr, pv.ID, api.ResourceType_PHYSICAL_VOLUME)
		}

		return handleCreateErr(errors.New("service_unavailable"), pv.ID, api.ResourceType_PHYSICAL_VOLUME)
	}

	pv.Status = api.ResourceStatus_CREATE_IN_PROGRESS

	inProgressErr := m.savePhysicalVolume(pv)

	if inProgressErr != nil {
		return handleCreateErr(inProgressErr, pv.ID, api.ResourceType_PHYSICAL_VOLUME)
	}

	blockAddr := fmt.Sprintf("%s:%d", as.Address, as.Port)

	conn, err := utils.NewClientConnTcp(blockAddr)

	if err != nil {
		return handleCreateErr(err, pv.ID, api.ResourceType_PHYSICAL_VOLUME)
	}

	defer conn.Close()

	ctx, cancel := context.WithTimeout(context.Background(), utils.ContextTimeout)

	defer cancel()

	c := api.NewBlockClient(conn)

	pvCreateOptions := &api.RequestPvCreate{
		DeviceName: pv.DeviceName,
	}

	_, pvCreateErr := c.PvCreate(ctx, pvCreateOptions)

	if pvCreateErr != nil {
		pv.Status = api.ResourceStatus_CREATE_FAILED

		saveErr := m.savePhysicalVolume(pv)

		if saveErr != nil {
			return handleCreateErr(saveErr, pv.ID, api.ResourceType_PHYSICAL_VOLUME)
		}

		return handleCreateErr(pvCreateErr, pv.ID, api.ResourceType_PHYSICAL_VOLUME)
	}

	pv.Status = api.ResourceStatus_CREATE_COMPLETED

	completedErr := m.savePhysicalVolume(pv)

	if completedErr != nil {
		return handleCreateErr(completedErr, pv.ID, api.ResourceType_PHYSICAL_VOLUME)
	}

	return nil
}

func (m *Manager) vgCreate(id uuid.UUID) error {
	vg, err := m.volumeGroupByID(id)

	if err != nil {
		return handleCreateErr(err, vg.ID, api.ResourceType_VOLUME_GROUP)
	}

	physicalVolumeID, err := uuid.Parse(vg.PhysicalVolumeID)

	if err != nil {
		return handleCreateErr(err, vg.ID, api.ResourceType_VOLUME_GROUP)
	}

	pv, err := m.physicalVolumeByID(physicalVolumeID)

	if err != nil {
		return handleCreateErr(err, vg.ID, api.ResourceType_VOLUME_GROUP)
	}

	pvServiceID, err := uuid.Parse(pv.ServiceID)

	if err != nil {
		return handleCreateErr(err, vg.ID, api.ResourceType_VOLUME_GROUP)
	}

	agentService, err := m.agentServiceByID(pvServiceID)

	if err != nil {
		return handleCreateErr(err, vg.ID, api.ResourceType_VOLUME_GROUP)
	}

	agentServiceHealth, err := m.agentServiceHealth(agentService.ID)

	if err != nil {
		return handleCreateErr(err, vg.ID, api.ResourceType_VOLUME_GROUP)
	}

	if agentServiceHealth == "critical" {
		vg.Status = api.ResourceStatus_CREATE_FAILED

		saveErr := m.saveVolumeGroup(vg)

		if saveErr != nil {
			return handleCreateErr(saveErr, vg.ID, api.ResourceType_VOLUME_GROUP)
		}

		return handleCreateErr(errors.New("service_unavailable"), vg.ID, api.ResourceType_VOLUME_GROUP)
	}

	vg.Status = api.ResourceStatus_CREATE_IN_PROGRESS

	inProgressErr := m.saveVolumeGroup(vg)

	if inProgressErr != nil {
		return handleCreateErr(inProgressErr, vg.ID, api.ResourceType_VOLUME_GROUP)
	}

	blockAddr := fmt.Sprintf("%s:%d", agentService.Address, agentService.Port)

	conn, err := utils.NewClientConnTcp(blockAddr)

	if err != nil {
		return handleCreateErr(err, vg.ID, api.ResourceType_VOLUME_GROUP)
	}

	defer conn.Close()

	ctx, cancel := context.WithTimeout(context.Background(), utils.ContextTimeout)

	defer cancel()

	c := api.NewBlockClient(conn)

	vgCreateOptions := &api.RequestVgCreate{
		DeviceName: pv.DeviceName,
		ID:         vg.ID,
	}

	_, vgCreateErr := c.VgCreate(ctx, vgCreateOptions)

	if vgCreateErr != nil {
		vg.Status = api.ResourceStatus_CREATE_FAILED

		saveErr := m.saveVolumeGroup(vg)

		if saveErr != nil {
			return handleCreateErr(saveErr, vg.ID, api.ResourceType_VOLUME_GROUP)
		}

		return handleCreateErr(vgCreateErr, vg.ID, api.ResourceType_VOLUME_GROUP)
	}

	vg.Status = api.ResourceStatus_CREATE_COMPLETED

	completedErr := m.saveVolumeGroup(vg)

	if completedErr != nil {
		return handleCreateErr(completedErr, vg.ID, api.ResourceType_VOLUME_GROUP)
	}

	return nil
}
