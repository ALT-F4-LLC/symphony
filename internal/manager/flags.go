package manager

import (
	"errors"
	"net"

	"github.com/erkrnt/symphony/internal/utils"
	"github.com/hashicorp/go-sockaddr"
	"github.com/sirupsen/logrus"
	"gopkg.in/alecthomas/kingpin.v2"
)

type flags struct {
	ConfigDir   string
	ConsulAddr  string
	ServiceAddr *net.TCPAddr
	Verbose     bool
}

var (
	bindInterface = kingpin.Flag("bind-interface", "Sets the bind interface for listening services.").Required().String()
	configDir     = kingpin.Flag("config-dir", "Sets configuration directory for manager.").Default(".").String()
	consulAddr    = kingpin.Flag("consul-addr", "Sets the etcd endpoints list.").Required().String()
	verbose       = kingpin.Flag("verbose", "Sets the lowest level of service output.").Bool()
)

func getFlags() (*flags, error) {
	kingpin.Parse()

	if bindInterface == nil {
		return nil, errors.New("invalid_bind_interface")
	}

	configDirPath, err := utils.GetDirPath(configDir)

	if err != nil {
		return nil, err
	}

	ipAddr, err := sockaddr.GetInterfaceIP(*bindInterface)

	if err != nil {
		return nil, err
	}

	serviceAddr, err := utils.GetListenAddr(ipAddr, 15760)

	if err != nil {
		return nil, err
	}

	f := &flags{
		ConfigDir:   *configDirPath,
		ConsulAddr:  *consulAddr,
		ServiceAddr: serviceAddr,
		Verbose:     *verbose,
	}

	fields := logrus.Fields{
		"ConfigDir":   f.ConfigDir,
		"ConsulAddr":  f.ConsulAddr,
		"ServiceAddr": f.ServiceAddr.String(),
		"Verbose":     f.Verbose,
	}

	logrus.WithFields(fields).Info("Service command-line flags loaded.")

	return f, nil
}
