package manager

import (
	"net"
	"strings"

	"github.com/erkrnt/symphony/internal/pkg/config"
	"github.com/sirupsen/logrus"
	"gopkg.in/alecthomas/kingpin.v2"
)

// Flags : command line flags for service
type Flags struct {
	ConfigPath        string
	EtcdEndpoints     []string
	ListenGossipAddr  *net.TCPAddr
	ListenServiceAddr *net.TCPAddr
	Verbose           bool
}

var (
	configDir         = kingpin.Flag("config-dir", "Sets configuration directory for manager.").Default(".").String()
	etcdEndpoints     = kingpin.Flag("etcd-endpoints", "Sets the etcd endpoints list.").Required().String()
	listenGossipPort  = kingpin.Flag("listen-gossip-port", "Sets the remote gossip listener port.").Int()
	listenServicePort = kingpin.Flag("listen-service-port", "Sets the remote service port.").Int()
	verbose           = kingpin.Flag("verbose", "Sets the lowest level of service output.").Bool()
)

// GetFlags : gets struct of flags from command line
func GetFlags() (*Flags, error) {
	kingpin.Parse()

	configPath, err := config.GetDirPath(configDir)

	if err != nil {
		return nil, err
	}

	ip, err := config.GetOutboundIP()

	if err != nil {
		return nil, err
	}

	listenGossipAddr, err := config.GetListenAddr(37065, ip, listenGossipPort)

	if err != nil {
		return nil, err
	}

	listenServiceAddr, err := config.GetListenAddr(15760, ip, listenServicePort)

	if err != nil {
		return nil, err
	}

	flags := &Flags{
		ConfigPath:        *configPath,
		EtcdEndpoints:     strings.Split(*etcdEndpoints, ","),
		ListenGossipAddr:  listenGossipAddr,
		ListenServiceAddr: listenServiceAddr,
		Verbose:           *verbose,
	}

	fields := logrus.Fields{
		"ConfigPath":        flags.ConfigPath,
		"EtcdEndpoints":     flags.EtcdEndpoints,
		"ListenGossipAddr":  flags.ListenGossipAddr.String(),
		"ListenServiceAddr": flags.ListenServiceAddr.String(),
		"Verbose":           flags.Verbose,
	}

	logrus.WithFields(fields).Info("Loaded command-line flags")

	return flags, nil
}
