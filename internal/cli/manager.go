package cli

import (
	"context"
	"log"
	"time"

	"github.com/erkrnt/symphony/api"
	"google.golang.org/grpc"
)

// ManagerInit : handle the "init" command
func ManagerInit(serviceAddr *string, socket *string) {
	conn := NewConnControl(socket)

	defer conn.Close()

	c := api.NewManagerControlClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)

	defer cancel()

	opts := &api.ManagerControlInitRequest{
		ServiceAddr: *serviceAddr,
	}

	_, initErr := c.Init(ctx, opts)

	if initErr != nil {
		log.Fatal(initErr)
	}
}

// ManagerRemove : handle the "remove" command
func ManagerRemove(id *string, socket *string) {
	if *id == "" {
		log.Fatal("invalid_service_id")
	}

	conn := NewConnControl(socket)

	defer conn.Close()

	c := api.NewManagerControlClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)

	defer cancel()

	opts := &api.ManagerControlRemoveRequest{
		ServiceID: *id,
	}

	_, removeErr := c.Remove(ctx, opts)

	if removeErr != nil {
		log.Fatal(removeErr)
	}
}

// ManagerLvCreate : handles creation of new logical volume
func ManagerLvCreate(endpoint *string, size *int64, volumeGroupID *string) {
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

	opts := &api.ManagerRemoteNewLvRequest{
		Size:          *size,
		VolumeGroupID: *volumeGroupID,
	}

	lv, err := c.NewLv(ctx, opts)

	if err != nil {
		log.Fatal(err)
	}

	log.Print(*lv)
}

// ManagerPvCreate : handles creation of new physical volume
func ManagerPvCreate(deviceName *string, endpoint *string, serviceID *string) {
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

	opts := &api.ManagerRemoteNewPvRequest{
		DeviceName: *deviceName,
		ServiceID:  *serviceID,
	}

	pv, err := c.NewPv(ctx, opts)

	if err != nil {
		log.Fatal(err)
	}

	log.Print(*pv)
}

// ManagerVgCreate : handles creation of new volume group
func ManagerVgCreate(endpoint *string, physicalVolumeID *string) {
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

	opts := &api.ManagerRemoteNewVgRequest{
		PhysicalVolumeID: *physicalVolumeID,
	}

	vg, err := c.NewVg(ctx, opts)

	if err != nil {
		log.Fatal(err)
	}

	log.Print(*vg)
}

// ManagerLvGet : gets logical volume
// func ManagerLvGet(id *string, remoteAddr *string) {
// 	if *id == "" {
// 		log.Fatal("invalid_parameters")
// 	}

// 	conn, err := grpc.Dial(*remoteAddr, grpc.WithInsecure())

// 	if err != nil {
// 		log.Fatal(err)
// 	}

// 	defer conn.Close()

// 	c := api.NewManagerRemoteClient(conn)

// 	ctx, cancel := context.WithTimeout(context.Background(), time.Second)

// 	defer cancel()

// 	opts := &api.ManagerRemoteLvRequest{ID: *id}

// 	lv, err := c.GetLv(ctx, opts)

// 	if err != nil {
// 		log.Fatal(err)
// 	}

// 	log.Print(*lv)
// }

// ManagerPvGet : gets physical volume
func ManagerPvGet(endpoint *string, id *string) {
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

	opts := &api.ManagerRemotePvRequest{ID: *id}

	pv, err := c.GetPv(ctx, opts)

	if err != nil {
		log.Fatal(err)
	}

	log.Print(pv)
}

// ManagerVgGet : gets volume group
func ManagerVgGet(endpoint *string, id *string) {
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

	opts := &api.ManagerRemoteVgRequest{ID: *id}

	vg, err := c.GetVg(ctx, opts)

	if err != nil {
		log.Fatal(err)
	}

	log.Print(*vg)
}

// ManagerLvRemove : removes logical volume
// func ManagerLvRemove(id *string, remoteAddr *string) {
// 	if *id == "" {
// 		log.Fatal("invalid_parameters")
// 	}

// 	conn, err := grpc.Dial(*remoteAddr, grpc.WithInsecure())

// 	if err != nil {
// 		log.Fatal(err)
// 	}

// 	defer conn.Close()

// 	c := api.NewManagerRemoteClient(conn)

// 	ctx, cancel := context.WithTimeout(context.Background(), time.Second)

// 	defer cancel()

// 	opts := &api.ManagerRemoteLvRequest{ID: *id}

// 	lv, err := c.RemoveLv(ctx, opts)

// 	if err != nil {
// 		log.Fatal(err)
// 	}

// 	log.Print(*lv)
// }

// ManagerPvRemove : removes physical volume
func ManagerPvRemove(endpoint *string, id *string) {
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

	opts := &api.ManagerRemotePvRequest{ID: *id}

	pv, err := c.RemovePv(ctx, opts)

	if err != nil {
		log.Fatal(err)
	}

	log.Print(*pv)
}

// ManagerVgRemove : removes volume group
func ManagerVgRemove(endpoint *string, id *string) {
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

	opts := &api.ManagerRemoteVgRequest{ID: *id}

	vg, err := c.RemoveVg(ctx, opts)

	if err != nil {
		log.Fatal(err)
	}

	log.Print(*vg)
}
