package main

import (
	"gopkg.in/alecthomas/kingpin.v2"
)

// Flags : command line flags for service
type Flags struct {
	Config  string
	Preseed bool
	Verbose bool
}

var (
	config  = kingpin.Flag("config", "Path to configuration file.").Short('c').Default("config.yaml").String()
	preseed = kingpin.Flag("preseed", "Enables preseeding of database.").Short('p').Bool()
	verbose = kingpin.Flag("verbose", "Enables debug output.").Short('v').Bool()
)

// GetFlags : gets command line flags at runtime
func GetFlags() Flags {
	kingpin.Parse()
	return Flags{Config: *config, Preseed: *preseed, Verbose: *verbose}
}
