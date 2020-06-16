package cli

import (
	"context"
	"log"
	"time"

	"github.com/erkrnt/symphony/api"
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
