package block

import (
	"net"

	"github.com/erkrnt/symphony/internal/pkg/config"
	"github.com/sirupsen/logrus"
	"gopkg.in/alecthomas/kingpin.v2"
)

type flags struct {
	configDir         string
	listenGossipAddr  *net.TCPAddr
	listenServiceAddr *net.TCPAddr
	verbose           bool
}

var (
	configDir         = kingpin.Flag("config-dir", "Sets configuration directory for block service.").Default(".").String()
	listenGossipPort  = kingpin.Flag("listen-gossip-port", "Sets the remote gossip listener port.").Int()
	listenServicePort = kingpin.Flag("listen-service-port", "Sets the remote service listener port.").Int()
	verbose           = kingpin.Flag("verbose", "Sets the lowest level of service output.").Bool()
)

func getFlags() (*flags, error) {
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

	flags := &flags{
		configDir:         *configPath,
		listenGossipAddr:  listenGossipAddr,
		listenServiceAddr: listenServiceAddr,
		verbose:           *verbose,
	}

	fields := logrus.Fields{
		"ConfigDir":         flags.configDir,
		"ListenGossipAddr":  flags.listenGossipAddr.String(),
		"ListenServiceAddr": flags.listenServiceAddr.String(),
		"Verbose":           flags.verbose,
	}

	logrus.WithFields(fields).Info("Loaded command-line flags")

	return flags, nil
}
