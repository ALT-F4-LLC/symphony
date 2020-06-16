package block

import (
	"net"

	"github.com/erkrnt/symphony/internal/pkg/config"
	"github.com/sirupsen/logrus"
	"gopkg.in/alecthomas/kingpin.v2"
)

type flags struct {
	configDir  string
	listenAddr *net.TCPAddr
}

var (
	configDir  = kingpin.Flag("config-dir", "Sets configuration directory for block service.").Default(".").String()
	listenPort = kingpin.Flag("listen-port", "Sets the remote listener port.").Int()
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

	listenAddr, err := config.GetListenAddr(27242, ip, listenPort)

	if err != nil {
		return nil, err
	}

	flags := &flags{
		configDir:  *configPath,
		listenAddr: listenAddr,
	}

	fields := logrus.Fields{
		"ConfigDir":        flags.configDir,
		"ListenRemoteAddr": flags.listenAddr.String(),
	}

	logrus.WithFields(fields).Info("Loaded command-line flags")

	return flags, nil
}
