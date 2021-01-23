package scheduler

import (
	"errors"
	"net"

	"github.com/sirupsen/logrus"
	"gopkg.in/alecthomas/kingpin.v2"
)

// Flags : service flags
type Flags struct {
	APIServerAddr *net.TCPAddr
	ConfigDir     string
	Verbose       bool
}

func getFlags() (*Flags, error) {
	var (
		apiserverAddrFlag = kingpin.Flag("apiserver-addr", "Sets the API server instance to communicate with.").Required().String()
		bindInterfaceFlag = kingpin.Flag("bind-interface", "Sets the bind interface for listening services.").Required().String()
		verboseFlag       = kingpin.Flag("verbose", "Sets the lowest level of service output.").Bool()
	)

	kingpin.Parse()

	if bindInterfaceFlag == nil {
		return nil, errors.New("invalid_bind_interface")
	}

	apiserverAddr, err := net.ResolveTCPAddr("tcp", *apiserverAddrFlag)

	if err != nil {
		return nil, err
	}

	flags := &Flags{
		APIServerAddr: apiserverAddr,
		Verbose:       *verboseFlag,
	}

	fields := logrus.Fields{
		"APIServerAddr": flags.APIServerAddr.String(),
		"Verbose":       flags.Verbose,
	}

	logrus.WithFields(fields).Info("Service command-line flags loaded.")

	return flags, nil
}
