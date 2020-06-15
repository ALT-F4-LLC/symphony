package block

import (
	"net"

	"github.com/erkrnt/symphony/internal/pkg/config"
	"github.com/sirupsen/logrus"
	"gopkg.in/alecthomas/kingpin.v2"
)

type flags struct {
	configDir        string
	listenGossipAddr *net.TCPAddr
	listenRemoteAddr *net.TCPAddr
}

var (
	configDir        = kingpin.Flag("config-dir", "Sets configuration directory for block service.").Default(".").String()
	listenGossipPort = kingpin.Flag("listen-gossip-port", "Sets the gossip listener port.").Int()
	listenRemotePort = kingpin.Flag("listen-remote-port", "Sets the remote listener port.").Int()
)

func getFlags() (*flags, error) {
	kingpin.Parse()

	configPath, err := config.GetDirPath(configDir)

	if err != nil {
		return nil, err
	}

	flags := &flags{
		configDir: *configPath,
	}

	ip, err := config.GetOutboundIP()

	if err != nil {
		return nil, err
	}

	listenGossipAddr, err := config.GetListenAddr(37065, ip, listenGossipPort)

	if err != nil {
		return nil, err
	}

	flags.listenGossipAddr = listenGossipAddr

	listenRemoteAddr, err := config.GetListenAddr(27242, ip, listenRemotePort)

	if err != nil {
		return nil, err
	}

	flags.listenRemoteAddr = listenRemoteAddr

	fields := logrus.Fields{
		"ConfigDir":        flags.configDir,
		"ListenGossipAddr": flags.listenGossipAddr.String(),
		"ListenRemoteAddr": flags.listenRemoteAddr.String(),
	}

	logrus.WithFields(fields).Info("Loaded command-line flags")

	return flags, nil
}
