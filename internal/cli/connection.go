package cli

import (
	"fmt"
	"path/filepath"

	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
)

// NewConnControl : creates a new control connection
func NewConnControl(socketPath *string) *grpc.ClientConn {
	socket, err := filepath.Abs(*socketPath)

	if err != nil {
		logrus.Fatal(err)
	}

	conn, err := grpc.Dial(fmt.Sprintf("unix://%s", socket), grpc.WithInsecure())

	if err != nil {
		logrus.Fatal(err)
	}

	return conn
}
