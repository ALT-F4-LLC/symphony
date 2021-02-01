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
	ConfigDir   string
	ManagerAddr *net.TCPAddr
	ServiceAddr *net.TCPAddr
	Verbose     bool
}

// GetFlags : retrieves flags from CLI
func GetFlags() (*Flags, error) {
	var (
		managerAddrFlag   = kingpin.Flag("manager-addr", "Sets the manager to communicate with.").Required().String()
		bindInterfaceFlag = kingpin.Flag("bind-interface", "Sets the bind interface for listening services.").Required().String()
		configDirFlag     = kingpin.Flag("config-dir", "Sets configuration directory for block service.").Default(".").String()
		verboseFlag       = kingpin.Flag("verbose", "Sets the lowest level of service output.").Bool()
	)

	kingpin.Parse()

	if bindInterfaceFlag == nil {
		return nil, errors.New("invalid_bind_interface")
	}

	managerAddr, err := net.ResolveTCPAddr("tcp", *managerAddrFlag)

	if err != nil {
		return nil, err
	}

	configDirPath, err := utils.GetDirPath(*configDirFlag)

	if err != nil {
		return nil, err
	}

	listenIPAddr, err := sockaddr.GetInterfaceIP(*bindInterfaceFlag)

	if err != nil {
		return nil, err
	}

	serviceAddr, err := utils.GetListenAddr(listenIPAddr, 15760)

	if err != nil {
		return nil, err
	}

	flags := &Flags{
		ConfigDir:   *configDirPath,
		ManagerAddr: managerAddr,
		ServiceAddr: serviceAddr,
		Verbose:     *verboseFlag,
	}

	fields := logrus.Fields{
		"ConfigDir":   flags.ConfigDir,
		"ManagerAddr": flags.ManagerAddr.String(),
		"ServiceAddr": flags.ServiceAddr.String(),
		"Verbose":     flags.Verbose,
	}

	logrus.WithFields(fields).Info("Service command-line flags loaded.")

	return flags, nil
}
