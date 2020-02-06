package cli

import (
	"context"
	"log"
	"strings"
	"time"

	"github.com/erkrnt/symphony/api"
)

// InitHandler : handle the "init" command
func InitHandler(joinAddr *string, peers *string, socket *string) {
	if *joinAddr != "" && *peers != "" {
		log.Fatal("Cannot use --join-addr and --peers flags together.")
	}

	conn := NewConnection(socket)

	defer conn.Close()

	c := api.NewManagerControlClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)

	defer cancel()

	opts := &api.ManagerControlInitializeRequest{}

	if *joinAddr != "" {
		opts.JoinAddr = *joinAddr
	}

	if *peers != "" {
		opts.Peers = strings.Split(*peers, ",")
	}

	_, initErr := c.ManagerControlInitialize(ctx, opts)

	if initErr != nil {
		log.Fatal(initErr)
	}
}
