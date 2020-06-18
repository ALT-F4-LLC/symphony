package cli

import (
	"context"
	"log"
	"time"

	"github.com/erkrnt/symphony/api"
)

// BlockInit : handles joining block node to an existing cluster
func BlockInit(serviceAddr *string, socket *string) {
	conn := NewConnControl(socket)

	defer conn.Close()

	c := api.NewBlockControlClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)

	defer cancel()

	opts := &api.BlockControlInitRequest{ServiceAddr: *serviceAddr}

	_, joinErr := c.Init(ctx, opts)

	if joinErr != nil {
		log.Fatal(joinErr)
	}
}
