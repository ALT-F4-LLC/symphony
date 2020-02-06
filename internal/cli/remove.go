package cli

import (
	"context"
	"log"
	"time"

	"github.com/erkrnt/symphony/api"
)

// RemoveHandler : handle the "remove" command
func RemoveHandler(memberID *uint64, socket *string) {
	if *memberID == 0 {
		log.Fatal("invalid_member_id")
	}

	conn := NewConnection(socket)

	defer conn.Close()

	c := api.NewManagerControlClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)

	defer cancel()

	opts := &api.ManagerControlRemoveRequest{
		MemberId: *memberID,
	}

	_, err := c.ManagerControlRemove(ctx, opts)

	if err != nil {
		log.Fatal(err)
	}
}
