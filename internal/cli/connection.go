package cli

import (
	"net"

	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
)

// NewClient : creates a new api client
func NewClient(serviceAddr string) *grpc.ClientConn {
	addr, err := net.ResolveTCPAddr("tcp", serviceAddr)

	if err != nil {
		logrus.Fatal(err)
	}

	conn, err := grpc.Dial(addr.String(), grpc.WithInsecure())

	if err != nil {
		logrus.Fatal(err)
	}

	return conn
}
