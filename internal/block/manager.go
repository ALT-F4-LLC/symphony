package block

import (
	"context"
	"fmt"
	"net"

	"github.com/erkrnt/symphony/api"
	"github.com/erkrnt/symphony/internal/utils"
	"github.com/google/uuid"
)

func logicalVolume(managerAddr *net.TCPAddr, id uuid.UUID) (*api.LogicalVolume, error) {
	conn, err := utils.NewClientConnTcp(managerAddr.String())

	if err != nil {
		return nil, err
	}

	defer conn.Close()

	ctx, cancel := context.WithTimeout(context.Background(), utils.ContextTimeout)

	defer cancel()

	c := api.NewManagerClient(conn)

	requestOptions := &api.RequestLogicalVolume{
		ID: id.String(),
	}

	response, err := c.GetLogicalVolume(ctx, requestOptions)

	if err != nil {
		return nil, err
	}

	return response, nil
}

func logicalVolumes(managerAddr *net.TCPAddr, requestOptions *api.RequestLogicalVolumes) ([]*api.LogicalVolume, error) {
	conn, err := utils.NewClientConnTcp(managerAddr.String())

	if err != nil {
		return nil, err
	}

	defer conn.Close()

	ctx, cancel := context.WithTimeout(context.Background(), utils.ContextTimeout)

	defer cancel()

	c := api.NewManagerClient(conn)

	response, err := c.GetLogicalVolumes(ctx, requestOptions)

	if err != nil {
		return nil, err
	}

	return response.Results, nil
}

func physicalVolume(managerAddr *net.TCPAddr, id uuid.UUID) (*api.PhysicalVolume, error) {
	conn, err := utils.NewClientConnTcp(managerAddr.String())

	if err != nil {
		return nil, err
	}

	defer conn.Close()

	ctx, cancel := context.WithTimeout(context.Background(), utils.ContextTimeout)

	defer cancel()

	c := api.NewManagerClient(conn)

	requestOptions := &api.RequestPhysicalVolume{
		ID: id.String(),
	}

	response, err := c.GetPhysicalVolume(ctx, requestOptions)

	if err != nil {
		return nil, err
	}

	return response, nil
}

func physicalVolumes(managerAddr *net.TCPAddr, requestOptions *api.RequestPhysicalVolumes) ([]*api.PhysicalVolume, error) {
	conn, err := utils.NewClientConnTcp(managerAddr.String())

	if err != nil {
		return nil, err
	}

	defer conn.Close()

	ctx, cancel := context.WithTimeout(context.Background(), utils.ContextTimeout)

	defer cancel()

	c := api.NewManagerClient(conn)

	response, err := c.GetPhysicalVolumes(ctx, requestOptions)

	if err != nil {
		return nil, err
	}

	return response.Results, nil
}

func volumeGroup(managerAddr *net.TCPAddr, id uuid.UUID) (*api.VolumeGroup, error) {
	conn, err := utils.NewClientConnTcp(managerAddr.String())

	if err != nil {
		return nil, err
	}

	defer conn.Close()

	ctx, cancel := context.WithTimeout(context.Background(), utils.ContextTimeout)

	defer cancel()

	c := api.NewManagerClient(conn)

	requestOptions := &api.RequestVolumeGroup{
		ID: id.String(),
	}

	response, err := c.GetVolumeGroup(ctx, requestOptions)

	if err != nil {
		return nil, err
	}

	return response, nil
}

func volumeGroups(managerAddr *net.TCPAddr, requestOptions *api.RequestVolumeGroups) ([]*api.VolumeGroup, error) {
	conn, err := utils.NewClientConnTcp(managerAddr.String())

	if err != nil {
		return nil, err
	}

	defer conn.Close()

	ctx, cancel := context.WithTimeout(context.Background(), utils.ContextTimeout)

	defer cancel()

	c := api.NewManagerClient(conn)

	response, err := c.GetVolumeGroups(ctx, requestOptions)

	if err != nil {
		return nil, err
	}

	return response.Results, nil
}

func updateLogicalVolume(managerAddr *net.TCPAddr, id uuid.UUID, status api.ResourceStatus) (*api.LogicalVolume, error) {
	addr := fmt.Sprintf("%s:%d", managerAddr.IP.String(), managerAddr.Port)

	conn, err := utils.NewClientConnTcp(addr)

	if err != nil {
		return nil, err
	}

	defer conn.Close()

	ctx, cancel := context.WithTimeout(context.Background(), utils.ContextTimeout)

	defer cancel()

	c := api.NewManagerClient(conn)

	requestOptions := &api.RequestUpdateLogicalVolume{
		ID:     id.String(),
		Status: status,
	}

	response, err := c.UpdateLogicalVolume(ctx, requestOptions)

	if err != nil {
		return nil, err
	}

	return response, nil
}

func updatePhysicalVolume(managerAddr *net.TCPAddr, id uuid.UUID, status api.ResourceStatus) (*api.PhysicalVolume, error) {
	addr := fmt.Sprintf("%s:%d", managerAddr.IP.String(), managerAddr.Port)

	conn, err := utils.NewClientConnTcp(addr)

	if err != nil {
		return nil, err
	}

	defer conn.Close()

	ctx, cancel := context.WithTimeout(context.Background(), utils.ContextTimeout)

	defer cancel()

	c := api.NewManagerClient(conn)

	requestOptions := &api.RequestUpdatePhysicalVolume{
		ID:     id.String(),
		Status: status,
	}

	response, err := c.UpdatePhysicalVolume(ctx, requestOptions)

	if err != nil {
		return nil, err
	}

	return response, nil
}

func updateVolumeGroup(managerAddr *net.TCPAddr, id uuid.UUID, status api.ResourceStatus) (*api.VolumeGroup, error) {
	addr := fmt.Sprintf("%s:%d", managerAddr.IP.String(), managerAddr.Port)

	conn, err := utils.NewClientConnTcp(addr)

	if err != nil {
		return nil, err
	}

	defer conn.Close()

	ctx, cancel := context.WithTimeout(context.Background(), utils.ContextTimeout)

	defer cancel()

	c := api.NewManagerClient(conn)

	requestOptions := &api.RequestUpdateVolumeGroup{
		ID:     id.String(),
		Status: status,
	}

	response, err := c.UpdateVolumeGroup(ctx, requestOptions)

	if err != nil {
		return nil, err
	}

	return response, nil
}
