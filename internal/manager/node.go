package manager

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"os"
	"time"

	"github.com/erkrnt/symphony/api"
	"github.com/sirupsen/logrus"
	"go.etcd.io/etcd/clientv3"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// Manager : manager node
type Manager struct {
	Flags *Flags
}

// New : creates a new manager struct
func New() (*Manager, error) {
	flags, err := GetFlags()

	if err != nil {
		return nil, err
	}

	manager := &Manager{
		Flags: flags,
	}

	return manager, nil
}

// ControlServer : starts manager control server
func (m *Manager) ControlServer() {
	socketPath := fmt.Sprintf("%s/control.sock", m.Flags.ConfigPath)

	if err := os.RemoveAll(socketPath); err != nil {
		log.Fatal(err)
	}

	listen, err := net.Listen("unix", socketPath)

	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	server := grpc.NewServer()

	control := &controlServer{
		Manager: m,
	}

	api.RegisterManagerControlServer(server, control)

	logrus.Info("Started manager control gRPC socket server.")

	if err := server.Serve(listen); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}

func (m *Manager) GetServices() ([]*api.Service, error) {
	etcd, err := clientv3.New(clientv3.Config{
		Endpoints:   m.Flags.EtcdEndpoints,
		DialTimeout: 5 * time.Second,
	})

	if err != nil {
		st := status.New(codes.Internal, err.Error())

		return nil, st.Err()
	}

	defer etcd.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)

	defer cancel()

	results, err := etcd.Get(ctx, "/service", clientv3.WithPrefix())

	if err != nil {
		st := status.New(codes.Internal, err.Error())

		return nil, st.Err()
	}

	services := make([]*api.Service, 0)

	for _, ev := range results.Kvs {
		var s *api.Service

		err := json.Unmarshal(ev.Value, &s)

		if err != nil {
			st := status.New(codes.Internal, err.Error())

			return nil, st.Err()
		}

		services = append(services, s)
	}

	return services, nil
}

// RemoteServer : starts remote gRPC server
func (m *Manager) RemoteServer() {
	listen, err := net.Listen("tcp", m.Flags.ListenAddr.String())

	if err != nil {
		logrus.Fatal(err)
	}

	remote := &remoteServer{
		Manager: m,
	}

	server := grpc.NewServer()

	api.RegisterManagerRemoteServer(server, remote)

	logrus.Info("Started manager remote gRPC tcp server.")

	if err := server.Serve(listen); err != nil {
		logrus.Fatal(err)
	}
}
