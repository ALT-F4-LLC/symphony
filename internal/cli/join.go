package cli

import (
	"context"
	"log"
	"time"

	"github.com/erkrnt/symphony/api"
)

// JoinHandler : handle the "join" command
func JoinHandler(joinAddr *string, socket *string) {
	conn := NewConnection(socket)

	defer conn.Close()

	c := api.NewManagerControlClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)

	defer cancel()

	opts := &api.ManagerControlJoinRequest{JoinAddr: *joinAddr}

	_, joinErr := c.ManagerControlJoin(ctx, opts)

	if joinErr != nil {
		log.Fatal(joinErr)
	}
}
