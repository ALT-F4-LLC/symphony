package cli

import (
	"context"
	"log"
	"time"

	"github.com/erkrnt/symphony/api"
	"google.golang.org/grpc"
)

// ManagerServiceInit : handle the "init" command
func ManagerServiceInit(serviceAddr *string, socket *string) {
	conn := NewConnControl(socket)

	defer conn.Close()

	c := api.NewManagerControlClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)

	defer cancel()

	opts := &api.ManagerControlServiceInitRequest{
		ServiceAddr: *serviceAddr,
	}

	_, initErr := c.ServiceInit(ctx, opts)

	if initErr != nil {
		log.Fatal(initErr)
	}
}

// ManagerServiceRemove : handle the "remove" command
func ManagerServiceRemove(id *string, socket *string) {
	if *id == "" {
		log.Fatal("invalid_service_id")
	}

	conn := NewConnControl(socket)

	defer conn.Close()

	c := api.NewManagerControlClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)

	defer cancel()

	opts := &api.ManagerControlServiceRemoveRequest{
		ServiceID: *id,
	}

	_, removeErr := c.ServiceRemove(ctx, opts)

	if removeErr != nil {
		log.Fatal(removeErr)
	}
}

// ManagerNewLogicalVolume : handles creation of new logical volume
func ManagerNewLogicalVolume(endpoint *string, size *int64, volumeGroupID *string) {
	if *endpoint == "" || *size == 0 || *volumeGroupID == "" {
		log.Fatal("invalid_parameters")
	}

	conn, err := grpc.Dial(*endpoint, grpc.WithInsecure())

	if err != nil {
		log.Fatal(err)
	}

	defer conn.Close()

	c := api.NewManagerRemoteClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)

	defer cancel()

	opts := &api.ManagerNewLogicalVolumeRequest{
		Size:          *size,
		VolumeGroupID: *volumeGroupID,
	}

	lv, err := c.NewLogicalVolume(ctx, opts)

	if err != nil {
		log.Fatal(err)
	}

	log.Print(*lv)
}

// ManagerNewPhysicalVolume : handles creation of new physical volume
func ManagerNewPhysicalVolume(deviceName *string, endpoint *string, serviceID *string) {
	if *deviceName == "" || *endpoint == "" || *serviceID == "" {
		log.Fatal("invalid_device_parameter")
	}

	conn, err := grpc.Dial(*endpoint, grpc.WithInsecure())

	if err != nil {
		log.Fatal(err)
	}

	defer conn.Close()

	c := api.NewManagerRemoteClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)

	defer cancel()

	opts := &api.ManagerNewPhysicalVolumeRequest{
		DeviceName: *deviceName,
		ServiceID:  *serviceID,
	}

	pv, err := c.NewPhysicalVolume(ctx, opts)

	if err != nil {
		log.Fatal(err)
	}

	log.Print(*pv)
}

// ManagerNewVolumeGroup : handles creation of new volume group
func ManagerNewVolumeGroup(endpoint *string, physicalVolumeID *string) {
	if *endpoint == "" || *physicalVolumeID == "" {
		log.Fatal("invalid_parameters")
	}

	conn, err := grpc.Dial(*endpoint, grpc.WithInsecure())

	if err != nil {
		log.Fatal(err)
	}

	defer conn.Close()

	c := api.NewManagerRemoteClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)

	defer cancel()

	opts := &api.ManagerNewVolumeGroupRequest{
		PhysicalVolumeID: *physicalVolumeID,
	}

	vg, err := c.NewVolumeGroup(ctx, opts)

	if err != nil {
		log.Fatal(err)
	}

	log.Print(*vg)
}

// ManagerGetLogicalVolume : gets logical volume
func ManagerGetLogicalVolume(endpoint *string, id *string) {
	if *endpoint == "" || *id == "" {
		log.Fatal("invalid_parameters")
	}

	conn, err := grpc.Dial(*endpoint, grpc.WithInsecure())

	if err != nil {
		log.Fatal(err)
	}

	defer conn.Close()

	c := api.NewManagerRemoteClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)

	defer cancel()

	opts := &api.ManagerLogicalVolumeRequest{ID: *id}

	lv, err := c.GetLogicalVolume(ctx, opts)

	if err != nil {
		log.Fatal(err)
	}

	log.Print(*lv)
}

// ManagerGetPhysicalVolume : gets physical volume
func ManagerGetPhysicalVolume(endpoint *string, id *string) {
	if *id == "" {
		log.Fatal("invalid_device_parameter")
	}

	conn, err := grpc.Dial(*endpoint, grpc.WithInsecure())

	if err != nil {
		log.Fatal(err)
	}

	defer conn.Close()

	c := api.NewManagerRemoteClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)

	defer cancel()

	opts := &api.ManagerPhysicalVolumeRequest{ID: *id}

	pv, err := c.GetPhysicalVolume(ctx, opts)

	if err != nil {
		log.Fatal(err)
	}

	log.Print(pv)
}

// ManagerGetVolumeGroup : gets volume group
func ManagerGetVolumeGroup(endpoint *string, id *string) {
	if *endpoint == "" || *id == "" {
		log.Fatal("invalid_parameters")
	}

	conn, err := grpc.Dial(*endpoint, grpc.WithInsecure())

	if err != nil {
		log.Fatal(err)
	}

	defer conn.Close()

	c := api.NewManagerRemoteClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)

	defer cancel()

	opts := &api.ManagerVolumeGroupRequest{ID: *id}

	vg, err := c.GetVolumeGroup(ctx, opts)

	if err != nil {
		log.Fatal(err)
	}

	log.Print(*vg)
}

// ManagerRemoveLogicalVolume : removes logical volume
func ManagerRemoveLogicalVolume(endpoint *string, id *string) {
	if *endpoint == "" || *id == "" {
		log.Fatal("invalid_parameters")
	}

	conn, err := grpc.Dial(*endpoint, grpc.WithInsecure())

	if err != nil {
		log.Fatal(err)
	}

	defer conn.Close()

	c := api.NewManagerRemoteClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)

	defer cancel()

	opts := &api.ManagerLogicalVolumeRequest{ID: *id}

	lv, err := c.RemoveLogicalVolume(ctx, opts)

	if err != nil {
		log.Fatal(err)
	}

	log.Print(*lv)
}

// ManagerRemovePhysicalVolume : removes physical volume
func ManagerRemovePhysicalVolume(endpoint *string, id *string) {
	if *endpoint == "" || *id == "" {
		log.Fatal("invalid_parameter")
	}

	conn, err := grpc.Dial(*endpoint, grpc.WithInsecure())

	if err != nil {
		log.Fatal(err)
	}

	defer conn.Close()

	c := api.NewManagerRemoteClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)

	defer cancel()

	opts := &api.ManagerPhysicalVolumeRequest{ID: *id}

	pv, err := c.RemovePhysicalVolume(ctx, opts)

	if err != nil {
		log.Fatal(err)
	}

	log.Print(*pv)
}

// ManagerRemoveVolumeGroup : removes volume group
func ManagerRemoveVolumeGroup(endpoint *string, id *string) {
	if *endpoint == "" || *id == "" {
		log.Fatal("invalid_parameter")
	}

	conn, err := grpc.Dial(*endpoint, grpc.WithInsecure())

	if err != nil {
		log.Fatal(err)
	}

	defer conn.Close()

	c := api.NewManagerRemoteClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)

	defer cancel()

	opts := &api.ManagerVolumeGroupRequest{ID: *id}

	vg, err := c.RemoveVolumeGroup(ctx, opts)

	if err != nil {
		log.Fatal(err)
	}

	log.Print(*vg)
}
