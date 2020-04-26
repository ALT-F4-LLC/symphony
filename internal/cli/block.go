package cli

import (
	"context"
	"log"
	"time"

	"github.com/erkrnt/symphony/api"
	"google.golang.org/grpc"
)

// BlockJoin : handles joining block node to an existing cluster
func BlockJoin(joinAddr *string, socket *string) {
	conn := NewConnSocket(socket)

	defer conn.Close()

	c := api.NewBlockControlClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)

	defer cancel()

	opts := &api.BlockControlJoinRequest{JoinAddr: *joinAddr}

	_, joinErr := c.Join(ctx, opts)

	if joinErr != nil {
		log.Fatal(joinErr)
	}
}

// BlockLvCreate : handles creation of new logical volume
func BlockLvCreate(id *string, volumeGroupID *string, size *string, remoteAddr *string) {
	if *id == "" || *volumeGroupID == "" || *size == "" {
		log.Fatal("invalid_parameters")
	}

	conn, err := grpc.Dial(*remoteAddr, grpc.WithInsecure())

	if err != nil {
		log.Fatal(err)
	}

	defer conn.Close()

	c := api.NewBlockRemoteClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)

	defer cancel()

	opts := &api.BlockNewLvFields{ID: *id, VolumeGroupID: *volumeGroupID, Size: *size}

	lv, err := c.NewLv(ctx, opts)

	if err != nil {
		log.Fatal(err)
	}

	log.Print(*lv)
}

// BlockPvCreate : handles creation of new physical volume
func BlockPvCreate(device *string, remoteAddr *string) {
	if *device == "" {
		log.Fatal("invalid_device_parameter")
	}

	conn, err := grpc.Dial(*remoteAddr, grpc.WithInsecure())

	if err != nil {
		log.Fatal(err)
	}

	defer conn.Close()

	c := api.NewBlockRemoteClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)

	defer cancel()

	opts := &api.BlockPvFields{Device: *device}

	pv, err := c.NewPv(ctx, opts)

	if err != nil {
		log.Fatal(err)
	}

	log.Print(*pv)
}

// BlockVgCreate : handles creation of new volume group
func BlockVgCreate(device *string, id *string, remoteAddr *string) {
	if *device == "" || *id == "" {
		log.Fatal("invalid_parameters")
	}

	conn, err := grpc.Dial(*remoteAddr, grpc.WithInsecure())

	if err != nil {
		log.Fatal(err)
	}

	defer conn.Close()

	c := api.NewBlockRemoteClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)

	defer cancel()

	opts := &api.BlockNewVgFields{Device: *device, ID: *id}

	vg, err := c.NewVg(ctx, opts)

	if err != nil {
		log.Fatal(err)
	}

	log.Print(*vg)
}

// BlockLvGet : gets logical volume
func BlockLvGet(id *string, volumeGroupID *string, remoteAddr *string) {
	if *id == "" || *volumeGroupID == "" {
		log.Fatal("invalid_parameters")
	}

	conn, err := grpc.Dial(*remoteAddr, grpc.WithInsecure())

	if err != nil {
		log.Fatal(err)
	}

	defer conn.Close()

	c := api.NewBlockRemoteClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)

	defer cancel()

	opts := &api.BlockLvFields{ID: *id, VolumeGroupID: *volumeGroupID}

	lv, err := c.GetLv(ctx, opts)

	if err != nil {
		log.Fatal(err)
	}

	log.Print(*lv)
}

// BlockPvGet : gets physical volume
func BlockPvGet(device *string, remoteAddr *string) {
	if *device == "" {
		log.Fatal("invalid_device_parameter")
	}

	conn, err := grpc.Dial(*remoteAddr, grpc.WithInsecure())

	if err != nil {
		log.Fatal(err)
	}

	defer conn.Close()

	c := api.NewBlockRemoteClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)

	defer cancel()

	opts := &api.BlockPvFields{Device: *device}

	pv, err := c.GetPv(ctx, opts)

	if err != nil {
		log.Fatal(err)
	}

	log.Print(*pv)
}

// BlockVgGet : gets volume group
func BlockVgGet(id *string, remoteAddr *string) {
	if *id == "" {
		log.Fatal("invalid_parameters")
	}

	conn, err := grpc.Dial(*remoteAddr, grpc.WithInsecure())

	if err != nil {
		log.Fatal(err)
	}

	defer conn.Close()

	c := api.NewBlockRemoteClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)

	defer cancel()

	opts := &api.BlockVgFields{ID: *id}

	vg, err := c.GetVg(ctx, opts)

	if err != nil {
		log.Fatal(err)
	}

	log.Print(*vg)
}

// BlockLvRemove : removes logical volume
func BlockLvRemove(id *string, volumeGroupID *string, remoteAddr *string) {
	if *id == "" || *volumeGroupID == "" {
		log.Fatal("invalid_parameters")
	}

	conn, err := grpc.Dial(*remoteAddr, grpc.WithInsecure())

	if err != nil {
		log.Fatal(err)
	}

	defer conn.Close()

	c := api.NewBlockRemoteClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)

	defer cancel()

	opts := &api.BlockLvFields{ID: *id, VolumeGroupID: *volumeGroupID}

	lv, err := c.RemoveLv(ctx, opts)

	if err != nil {
		log.Fatal(err)
	}

	log.Print(*lv)
}

// BlockPvRemove : removes physical volume
func BlockPvRemove(device *string, remoteAddr *string) {
	if *device == "" {
		log.Fatal("invalid_parameter")
	}

	conn, err := grpc.Dial(*remoteAddr, grpc.WithInsecure())

	if err != nil {
		log.Fatal(err)
	}

	defer conn.Close()

	c := api.NewBlockRemoteClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)

	defer cancel()

	opts := &api.BlockPvFields{Device: *device}

	pv, err := c.RemovePv(ctx, opts)

	if err != nil {
		log.Fatal(err)
	}

	log.Print(*pv)
}

// BlockVgRemove : removes volume group
func BlockVgRemove(id *string, remoteAddr *string) {
	if *id == "" {
		log.Fatal("invalid_parameter")
	}

	conn, err := grpc.Dial(*remoteAddr, grpc.WithInsecure())

	if err != nil {
		log.Fatal(err)
	}

	defer conn.Close()

	c := api.NewBlockRemoteClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)

	defer cancel()

	opts := &api.BlockVgFields{ID: *id}

	vg, err := c.RemoveVg(ctx, opts)

	if err != nil {
		log.Fatal(err)
	}

	log.Print(*vg)
}
