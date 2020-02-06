package cli

import (
	"fmt"
	"log"
	"path/filepath"

	"google.golang.org/grpc"
)

// NewConnection : creates a new grpc connection
func NewConnection(socket *string) *grpc.ClientConn {
	abs, err := filepath.Abs(".")

	if err != nil {
		log.Fatal(err)
	}

	s := fmt.Sprintf("%s/control.sock", abs)

	if *socket != "" {
		abs, err := filepath.Abs(*socket)

		if err != nil {
			log.Fatal(err)
		}

		s = abs
	}

	conn, err := grpc.Dial(fmt.Sprintf("unix://%s", s), grpc.WithInsecure())

	if err != nil {
		log.Fatal(err)
	}

	return conn
}
