package main

import (
	"gopkg.in/alecthomas/kingpin.v2"
)

// Flags : command line flags for service
type Flags struct {
	Conductor string
	Config    string
	Hostname  string
	Verbose   bool
}

var (
	conductor = kingpin.Flag("conductor", "Conductor service address.").Required().String()
	config    = kingpin.Flag("config", "Path to configuration file.").Short('c').Default("config.yaml").String()
	hostname  = kingpin.Flag("hostname", "Service hostname.").Short('h').Required().String()
	verbose   = kingpin.Flag("verbose", "Enables debug output.").Short('v').Bool()
)

// GetFlags : gets command line flags at runtime
func GetFlags() Flags {
	kingpin.Parse()

	return Flags{Conductor: *conductor, Config: *config, Hostname: *hostname, Verbose: *verbose}
}
