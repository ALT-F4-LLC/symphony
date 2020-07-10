package cli

import (
	"context"
	"log"
	"time"

	"github.com/erkrnt/symphony/api"
)

// BlockServiceInit : handles joining block node to an existing cluster
func BlockServiceInit(serviceAddr *string, socket *string) {
	conn := NewConnControl(socket)

	defer conn.Close()

	c := api.NewBlockClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)

	defer cancel()

	opts := &api.BlockServiceInitRequest{ServiceAddr: *serviceAddr}

	_, joinErr := c.ServiceInit(ctx, opts)

	if joinErr != nil {
		log.Fatal(joinErr)
	}
}
