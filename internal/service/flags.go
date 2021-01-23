package service

import (
	"errors"
	"net"

	"github.com/erkrnt/symphony/internal/utils"
	"github.com/hashicorp/go-sockaddr"
	"github.com/sirupsen/logrus"
	"gopkg.in/alecthomas/kingpin.v2"
)

// Flags : service flags
type Flags struct {
	APIServerAddr     *net.TCPAddr
	ConfigDir         string
	HealthServiceAddr *net.TCPAddr
	Verbose           bool
}

// GetFlags : retrieves flags from CLI
func GetFlags() (*Flags, error) {
	var (
		apiserverAddrFlag = kingpin.Flag("apiserver-addr", "Sets the API server instance to communicate with.").Required().String()
		bindInterfaceFlag = kingpin.Flag("bind-interface", "Sets the bind interface for listening services.").Required().String()
		configDirFlag     = kingpin.Flag("config-dir", "Sets configuration directory for block service.").Default(".").String()
		verboseFlag       = kingpin.Flag("verbose", "Sets the lowest level of service output.").Bool()
	)

	kingpin.Parse()

	if bindInterfaceFlag == nil {
		return nil, errors.New("invalid_bind_interface")
	}

	configDirPath, err := utils.GetDirPath(configDirFlag)

	if err != nil {
		return nil, err
	}

	listenIPAddr, err := sockaddr.GetInterfaceIP(*bindInterfaceFlag)

	if err != nil {
		return nil, err
	}

	healthServiceAddr, err := utils.GetListenAddr(listenIPAddr, 15761)

	if err != nil {
		return nil, err
	}

	apiserverAddr, err := net.ResolveTCPAddr("tcp", *apiserverAddrFlag)

	if err != nil {
		return nil, err
	}

	flags := &Flags{
		APIServerAddr:     apiserverAddr,
		ConfigDir:         *configDirPath,
		HealthServiceAddr: healthServiceAddr,
		Verbose:           *verboseFlag,
	}

	fields := logrus.Fields{
		"APIServerAddr":     flags.APIServerAddr.String(),
		"ConfigDir":         flags.ConfigDir,
		"HealthServiceAddr": flags.HealthServiceAddr.String(),
		"Verbose":           flags.Verbose,
	}

	logrus.WithFields(fields).Info("Service command-line flags loaded.")

	return flags, nil
}
