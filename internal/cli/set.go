package cli

import (
	"context"
	"log"
	"time"

	"github.com/erkrnt/symphony/api"
)

// SetHandler : handles "set" command
func SetHandler(key *string, socket *string, value *string) {
	if *key == "" || *value == "" {
		log.Fatal("Invalid parameters")
	}

	conn := NewConnection(socket)

	defer conn.Close()

	c := api.NewManagerControlClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)

	defer cancel()

	opts := &api.ManagerControlSetReq{
		Key:   *key,
		Value: *value,
	}

	_, err := c.Set(ctx, opts)

	if err != nil {
		log.Fatal(err)
	}
}
