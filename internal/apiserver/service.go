package apiserver

import (
	"context"
	"fmt"
	"net"
	"os"

	"github.com/erkrnt/symphony/api"
	"github.com/erkrnt/symphony/internal/service"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
)

// APIServer : defines apiserver service
type APIServer struct {
	Flags     *Flags
	Resources *Resources
	Service   *service.Service
}

// New : creates a new apiserver instance
func New() (*APIServer, error) {
	flags, err := getFlags()

	if err != nil {
		return nil, err
	}

	if flags.Verbose {
		logrus.SetLevel(logrus.DebugLevel)
	}

	resources := &Resources{
		EtcdEndpoints: flags.EtcdEndpoints,
	}

	key, err := service.GetKey(flags.ConfigDir)

	if err != nil {
		return nil, err
	}

	serviceType := api.ServiceType_APISERVER

	serf, err := service.NewSerf(flags.ListenAddr, key.ServiceID, serviceType)

	if err != nil {
		return nil, err
	}

	service := &service.Service{
		ErrorC: make(chan string),
		Key:    key,
		Serf:   serf,
	}

	apiserver := &APIServer{
		Flags:     flags,
		Resources: resources,
		Service:   service,
	}

	return apiserver, nil
}

// Start : handles start of apiserver service
func (apiserver *APIServer) Start() {
	key := apiserver.Service.Key

	if key.ClusterID != nil && key.ServiceID != nil {
		peerAddr := apiserver.Flags.PeerAddr

		restartErr := apiserver.restart(peerAddr)

		if restartErr != nil {
			errorC := apiserver.Service.ErrorC

			errorC <- restartErr.Error()
		}
	}

	go apiserver.listenControl()

	go apiserver.listenAPIServer()
}

func (apiserver *APIServer) listenControl() {
	socketPath := fmt.Sprintf("%s/control.sock", apiserver.Flags.ConfigDir)

	if err := os.RemoveAll(socketPath); err != nil {
		apiserver.Service.ErrorC <- err.Error()
	}

	listen, err := net.Listen("unix", socketPath)

	if err != nil {
		apiserver.Service.ErrorC <- err.Error()
	}

	server := grpc.NewServer()

	api.RegisterControlServer(server, &GRPCServerControl{
		APIServer: apiserver,
	})

	logrus.Info("Started control grpc server.")

	serveErr := server.Serve(listen)

	if serveErr != nil {
		apiserver.Service.ErrorC <- serveErr.Error()
	}
}

func (apiserver *APIServer) listenAPIServer() {
	listen, err := net.Listen("tcp", apiserver.Flags.ListenAddr.String())

	if err != nil {
		apiserver.Service.ErrorC <- err.Error()
	}

	server := grpc.NewServer()

	api.RegisterAPIServerServer(server, &GRPCServerAPIServer{
		APIServer: apiserver,
	})

	logrus.Info("Started apiserver grpc server.")

	serveErr := server.Serve(listen)

	if serveErr != nil {
		apiserver.Service.ErrorC <- serveErr.Error()
	}
}

func (apiserver *APIServer) restart(peerAddr *net.TCPAddr) error {
	cluster, err := apiserver.Resources.getClusterByID(*apiserver.Service.Key.ClusterID)

	if err != nil {
		return err
	}

	srvc, err := apiserver.Resources.getServiceByID(*apiserver.Service.Key.ServiceID)

	if err != nil {
		return err
	}

	localAddr := apiserver.Flags.ListenAddr.String()

	logrus.Debug(localAddr)

	logrus.Debug(peerAddr)

	if cluster != nil && srvc != nil && localAddr != peerAddr.String() {
		serfError := apiserver.Service.RestartSerf(apiserver.Flags.ListenAddr, srvc.Type)

		if serfError != nil {
			return serfError
		}

		conn, err := grpc.Dial(peerAddr.String(), grpc.WithInsecure())

		if err != nil {
			return err
		}

		defer conn.Close()

		ctx, cancel := context.WithTimeout(context.Background(), service.ContextTimeout)

		defer cancel()

		r := api.NewAPIServerClient(conn)

		opts := &api.RequestServiceJoin{
			ClusterID: cluster.ID,
			ServiceID: srvc.ID,
		}

		res, err := r.ServiceJoin(ctx, opts)

		if err != nil {
			return err
		}

		_, serfErr := apiserver.Service.Serf.Join([]string{res.PeerAddr}, true)

		if serfErr != nil {
			return serfErr
		}
	}

	return nil
}
