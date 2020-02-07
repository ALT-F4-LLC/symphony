package cli

import (
	"context"
	"log"
	"time"

	"github.com/erkrnt/symphony/api"
)

// BlockJoinHandler : handle the "join" command
func BlockJoinHandler(joinAddr *string, socket *string) {
	conn := NewConnection(socket)

	defer conn.Close()

	c := api.NewBlockControlClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)

	defer cancel()

	opts := &api.BlockControlJoinReq{JoinAddr: *joinAddr}

	_, joinErr := c.Join(ctx, opts)

	if joinErr != nil {
		log.Fatal(joinErr)
	}
}

// ManagerJoinHandler : handle the "join" command
func ManagerJoinHandler(joinAddr *string, socket *string) {
	conn := NewConnection(socket)

	defer conn.Close()

	c := api.NewManagerControlClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)

	defer cancel()

	opts := &api.ManagerControlJoinReq{JoinAddr: *joinAddr}

	_, joinErr := c.Join(ctx, opts)

	if joinErr != nil {
		log.Fatal(joinErr)
	}
}
