package main

import (
	"gopkg.in/alecthomas/kingpin.v2"
)

// Flags : command line flags for service
type Flags struct {
	Join    *string
	Verbose *bool
}

func getFlags() Flags {
	var (
		join    = kingpin.Flag("join", "Join existing Raft via node address.").String()
		verbose = kingpin.Flag("verbose", "Enables verbose output.").Short('v').Bool()
	)

	kingpin.Parse()

	flags := Flags{
		Join:    join,
		Verbose: verbose,
	}

	return flags
}
