package manager

import (
	"context"
	"errors"
	"fmt"

	"github.com/erkrnt/symphony/api"
	"github.com/erkrnt/symphony/internal/utils"
	"github.com/google/uuid"
	consul "github.com/hashicorp/consul/api"
)

func (m *Manager) lvCreate(as *consul.AgentService, lv *api.LogicalVolume) error {
	blockAddr := fmt.Sprintf("%s:%d", as.Address, as.Port)

	conn, err := utils.NewClientConnTcp(blockAddr)

	if err != nil {
		return err
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
		return lvCreateErr
	}

	return nil
}

func (m *Manager) pvCreate(as *consul.AgentService, pv *api.PhysicalVolume) error {
	healthStatus, err := m.agentServiceHealth(as.ID)

	if err != nil {
		return err
	}

	if healthStatus == "critical" {
		return errors.New("service_unavailable")
	}

	pv.Status = api.ResourceStatus_CREATE_IN_PROGRESS

	inProgressErr := m.savePhysicalVolume(pv)

	if inProgressErr != nil {
		return inProgressErr
	}

	blockAddr := fmt.Sprintf("%s:%d", as.Address, as.Port)

	conn, err := utils.NewClientConnTcp(blockAddr)

	if err != nil {
		return err
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
		return pvCreateErr
	}

	return nil
}

func (m *Manager) vgCreate(as *consul.AgentService, vg *api.VolumeGroup) error {
	physicalVolumeID, err := uuid.Parse(vg.PhysicalVolumeID)

	if err != nil {
		return err
	}

	pv, err := m.physicalVolumeByID(physicalVolumeID)

	if err != nil {
		return err
	}

	pvServiceID, err := uuid.Parse(pv.ServiceID)

	if err != nil {
		return err
	}

	agentService, err := m.agentServiceByID(pvServiceID)

	if err != nil {
		return err
	}

	agentServiceHealth, err := m.agentServiceHealth(agentService.ID)

	if err != nil {
		return err
	}

	if agentServiceHealth == "critical" {
		return errors.New("service_unavailable")
	}

	vg.Status = api.ResourceStatus_CREATE_IN_PROGRESS

	inProgressErr := m.saveVolumeGroup(vg)

	if inProgressErr != nil {
		return inProgressErr
	}

	blockAddr := fmt.Sprintf("%s:%d", agentService.Address, agentService.Port)

	conn, err := utils.NewClientConnTcp(blockAddr)

	if err != nil {
		return err
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
		return vgCreateErr
	}

	return nil
}
