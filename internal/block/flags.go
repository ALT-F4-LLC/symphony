package block

import (
	"net"

	"github.com/erkrnt/symphony/internal/pkg/config"
	"github.com/sirupsen/logrus"
	"gopkg.in/alecthomas/kingpin.v2"
)

type flags struct {
	configDir        string
	listenRemoteAddr *net.TCPAddr
}

var (
	configDir        = kingpin.Flag("config-dir", "Sets configuration directory for block service.").Default(".").String()
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

	ip := config.GetOutboundIP()

	listenRemoteAddr, err := config.GetListenAddr(27242, ip, listenRemotePort)

	if err != nil {
		return nil, err
	}

	flags.listenRemoteAddr = listenRemoteAddr

	fields := logrus.Fields{
		"ConfigDir":        flags.configDir,
		"ListenRemoteAddr": flags.listenRemoteAddr.String(),
	}

	logrus.WithFields(fields).Info("Loaded command-line flags")

	return flags, nil
}
