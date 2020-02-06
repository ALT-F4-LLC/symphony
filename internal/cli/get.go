package cli

import (
	"context"
	"log"
	"time"

	"github.com/erkrnt/symphony/api"
)

// GetHandler : handle the "get" cli command
func GetHandler(key *string, socket *string) {
	if *key == "" {
		log.Fatal("Invalid parameters")
	}

	conn := NewConnection(socket)

	defer conn.Close()

	c := api.NewManagerControlClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)

	defer cancel()

	opts := &api.ManagerControlGetValueRequest{
		Key: *key,
	}

	res, err := c.ManagerControlGetValue(ctx, opts)

	if err != nil {
		log.Fatal(err)
	}

	log.Print(res.Value)
}
