package cli

import (
	"context"
	"log"
	"time"

	"github.com/erkrnt/symphony/api"
)

// ManagerRemoveHandler : handle the "remove" command
func ManagerRemoveHandler(memberID *uint64, socket *string) {
	if *memberID == 0 {
		log.Fatal("invalid_member_id")
	}

	conn := NewConnection(socket)

	defer conn.Close()

	c := api.NewManagerControlClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)

	defer cancel()

	opts := &api.ManagerControlRemoveReq{
		MemberId: *memberID,
	}

	_, err := c.Remove(ctx, opts)

	if err != nil {
		log.Fatal(err)
	}
}
