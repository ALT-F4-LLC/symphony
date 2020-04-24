package cli

import (
	"context"
	"log"
	"time"

	"github.com/erkrnt/symphony/api"
)

// ManagerInit : handle the "init" command
func ManagerInit(socket *string) {
	conn := NewConnSocket(socket)

	defer conn.Close()

	c := api.NewManagerControlClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)

	defer cancel()

	opts := &api.ManagerControlInitRequest{}

	_, initErr := c.Init(ctx, opts)

	if initErr != nil {
		log.Fatal(initErr)
	}
}

// ManagerJoin : handle the "join" command
func ManagerJoin(joinAddr *string, socket *string) {
	conn := NewConnSocket(socket)

	defer conn.Close()

	c := api.NewManagerControlClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)

	defer cancel()

	opts := &api.ManagerControlJoinRequest{
		JoinAddr: *joinAddr,
	}

	_, joinErr := c.Join(ctx, opts)

	if joinErr != nil {
		log.Fatal(joinErr)
	}
}

// ManagerRemove : handle the "remove" command
func ManagerRemove(raftID *uint64, socket *string) {
	if *raftID == 0 {
		log.Fatal("invalid_member_id")
	}

	conn := NewConnSocket(socket)

	defer conn.Close()

	c := api.NewManagerControlClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)

	defer cancel()

	opts := &api.ManagerControlRemoveRequest{
		RaftId: *raftID,
	}

	_, removeErr := c.Remove(ctx, opts)

	if removeErr != nil {
		log.Fatal(removeErr)
	}
}
