package cli

import (
	"context"
	"log"
	"time"

	"github.com/erkrnt/symphony/api"
)

// RemoveHandler : handle the "remove" command
func RemoveHandler(nodeID *uint64, socket *string) {
	if *nodeID == 0 {
		log.Fatal("invalid_node_id")
	}

	conn := NewConnection(socket)

	defer conn.Close()

	c := api.NewManagerControlClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)

	defer cancel()

	opts := &api.ManagerControlRemoveRequest{
		NodeId: *nodeID,
	}

	_, err := c.ManagerControlRemove(ctx, opts)

	if err != nil {
		log.Fatal(err)
	}
}
