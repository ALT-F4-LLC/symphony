package main

import (
	"gopkg.in/alecthomas/kingpin.v2"
)

// Flags : command line flags for service
type Flags struct {
	Config   string
	Hostname string
	Manager  string
	Verbose  bool
}

func getFlags() Flags {
	var (
		config   = kingpin.Flag("config", "Path to configuration file.").Short('c').Default("config.yaml").String()
		hostname = kingpin.Flag("hostname", "Set service hostname.").Short('h').Required().String()
		manager  = kingpin.Flag("manager", "Set manager service hostname.").Required().String()
		verbose  = kingpin.Flag("verbose", "Enables debug output.").Short('v').Bool()
	)

	kingpin.Parse()

	return Flags{
		Config:   *config,
		Hostname: *hostname,
		Manager:  *manager,
		Verbose:  *verbose,
	}
}
