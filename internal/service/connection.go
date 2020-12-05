package service

import (
	"fmt"
	"net"
	"path/filepath"

	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
)

// NewClientConnSocket : creates new socket ClientConn
func NewClientConnSocket(socket string) *grpc.ClientConn {
	s, err := filepath.Abs(socket)

	if err != nil {
		logrus.Fatal(err)
	}

	conn, err := grpc.Dial(fmt.Sprintf("unix://%s", s), grpc.WithInsecure())

	if err != nil {
		logrus.Fatal(err)
	}

	return conn
}

// NewClientConnTCP : creates new TCP ClientConn
func NewClientConnTCP(addr string) *grpc.ClientConn {
	a, err := net.ResolveTCPAddr("tcp", addr)

	if err != nil {
		logrus.Fatal(err)
	}

	conn, err := grpc.Dial(a.String(), grpc.WithInsecure())

	if err != nil {
		logrus.Fatal(err)
	}

	return conn
}
