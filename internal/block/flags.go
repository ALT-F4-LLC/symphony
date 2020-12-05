package block

import (
	"net"

	"github.com/sirupsen/logrus"
	"gopkg.in/alecthomas/kingpin.v2"
)

type flags struct {
	configDir  string
	listenAddr *net.TCPAddr
	verbose    bool
}

var (
	configDir  = kingpin.Flag("config-dir", "Sets configuration directory for block service.").Default(".").String()
	listenPort = kingpin.Flag("listen-port", "Sets the remote service listener port.").Int()
	verbose    = kingpin.Flag("verbose", "Sets the lowest level of service output.").Bool()
)

func getFlags() (*flags, error) {
	kingpin.Parse()

	configPath, err := config.GetDirPath(configDir)

	if err != nil {
		return nil, err
	}

	listenAddr, err := config.GetListenAddr(15760, listenPort)

	if err != nil {
		return nil, err
	}

	flags := &flags{
		configDir:  *configPath,
		listenAddr: listenAddr,
		verbose:    *verbose,
	}

	fields := logrus.Fields{
		"ConfigDir":  flags.configDir,
		"ListenAddr": flags.listenAddr.String(),
		"Verbose":    flags.verbose,
	}

	logrus.WithFields(fields).Info("Service command-line flags loaded.")

	return flags, nil
}
