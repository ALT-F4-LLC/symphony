package utils

import (
	"errors"
	"fmt"
	"time"

	"github.com/erkrnt/symphony/api"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// APIServerConnHandler : function handler for returning listener
type APIServerConnHandler func(api.APIServerClient) (func() error, error)

// APIServerConn : listen for gRPC messages
type APIServerConn struct {
	Address string
	ErrorC  chan error
	Handler APIServerConnHandler

	reconnectAttempts int
}

const (
	// ReconnectMaxAttempts : default reconnect max attempts
	ReconnectMaxAttempts = 3

	// ReconnectTimeoutSeconds : default reconnect timeout seconds
	ReconnectTimeoutSeconds = 5
)

// Listen : starts listener for resource grpc stream
func (rm *APIServerConn) Listen() {
	conn, err := NewClientConnTcp(rm.Address)

	if err != nil {
		rm.handleGrpcReconnect(err)

		return
	}

	defer conn.Close()

	client := api.NewAPIServerClient(conn)

	listen, err := rm.Handler(client)

	if err != nil {
		rm.handleGrpcReconnect(err)

		return
	}

	rm.reconnectAttempts = 0

	logrus.Debug("ResourceManager connection established")

	listenErr := listen()

	if listenErr != nil {
		rm.handleGrpcReconnect(listenErr)

		return
	}
}

func (rm *APIServerConn) handleGrpcReconnect(err error) {
	st, ok := status.FromError(err)

	if ok && codes.Code(st.Code()) == codes.Unavailable {
		if rm.reconnectAttempts > ReconnectMaxAttempts {
			err := errors.New("ResourceManager unable to connect to APIServer")

			rm.ErrorC <- err

			return
		}

		if rm.reconnectAttempts <= ReconnectMaxAttempts {
			timeout := rm.reconnectAttempts * ReconnectTimeoutSeconds

			duration := time.Duration(timeout)

			attemptOutput := fmt.Sprintf("ResourceManager reconnect attempt in %ds...", timeout)

			logrus.Info(attemptOutput)

			time.Sleep(duration * time.Second)

			rm.reconnectAttempts = rm.reconnectAttempts + 1

			go rm.Listen()

			return
		}
	}

	rm.ErrorC <- err

	return
}
