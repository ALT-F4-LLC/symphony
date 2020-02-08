package cli

import (
	"context"
	"log"
	"strings"
	"time"

	"github.com/erkrnt/symphony/api"
)

// ManagerGet : handle the "get" cli command
func ManagerGet(key *string, socket *string) {
	if *key == "" {
		log.Fatal("Invalid parameters")
	}

	conn := NewConnSocket(socket)

	defer conn.Close()

	c := api.NewManagerControlClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)

	defer cancel()

	opts := &api.ManagerControlGetReq{
		Key: *key,
	}

	res, err := c.Get(ctx, opts)

	if err != nil {
		log.Fatal(err)
	}

	log.Print(res.Value)
}

// ManagerInit : handle the "init" command
func ManagerInit(joinAddr *string, peers *string, socket *string) {
	if *joinAddr != "" && *peers != "" {
		log.Fatal("Cannot use --join-addr and --peers flags together.")
	}

	conn := NewConnSocket(socket)

	defer conn.Close()

	c := api.NewManagerControlClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)

	defer cancel()

	opts := &api.ManagerControlInitReq{}

	if *joinAddr != "" {
		opts.JoinAddr = *joinAddr
	}

	if *peers != "" {
		peersList := strings.Split(*peers, ",")

		members := make([]*api.Member, len(peersList))

		for i := range peersList {
			base := i + 1
			members[i] = &api.Member{ID: uint64(base + 1), Addr: peersList[i]}
		}

		opts.Members = members
	}

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

	opts := &api.ManagerControlJoinReq{JoinAddr: *joinAddr}

	_, joinErr := c.Join(ctx, opts)

	if joinErr != nil {
		log.Fatal(joinErr)
	}
}

// ManagerRemove : handle the "remove" command
func ManagerRemove(memberID *uint64, socket *string) {
	if *memberID == 0 {
		log.Fatal("invalid_member_id")
	}

	conn := NewConnSocket(socket)

	defer conn.Close()

	c := api.NewManagerControlClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)

	defer cancel()

	opts := &api.ManagerControlRemoveReq{
		MemberId: *memberID,
	}

	_, removeErr := c.Remove(ctx, opts)

	if removeErr != nil {
		log.Fatal(removeErr)
	}
}

// ManagerSet : handles "set" command
func ManagerSet(key *string, socket *string, value *string) {
	if *key == "" || *value == "" {
		log.Fatal("Invalid parameters")
	}

	conn := NewConnSocket(socket)

	defer conn.Close()

	c := api.NewManagerControlClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)

	defer cancel()

	opts := &api.ManagerControlSetReq{
		Key:   *key,
		Value: *value,
	}

	_, setErr := c.Set(ctx, opts)

	if setErr != nil {
		log.Fatal(setErr)
	}
}
