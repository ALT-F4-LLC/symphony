package utils

import (
	"fmt"
	"net"
	"path/filepath"
	"time"

	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
)

const (
	// ContextTimeout : default context timeout
	ContextTimeout = 5 * time.Second

	// DialTimeout : default dial timeout
	DialTimeout = 5 * time.Second
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

// NewClientConnTcp : creates new TCP ClientConn
func NewClientConnTcp(addr string) (*grpc.ClientConn, error) {
	a, err := net.ResolveTCPAddr("tcp", addr)

	if err != nil {
		return nil, err
	}

	conn, err := grpc.Dial(a.String(), grpc.WithInsecure())

	if err != nil {
		return nil, err
	}

	return conn, nil
}
