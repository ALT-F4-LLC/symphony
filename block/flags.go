package main

import "flag"

// Flags : command line flags for service
type Flags struct {
	Debug   bool
	Preseed bool
}

// GetFlags : gets command line flags at runtime
func GetFlags() Flags {
	debug := flag.Bool("debug", false, "Enables debug output.")
	preseed := flag.Bool("preseed", false, "Enables preseeding of database.")
	flag.Parse()
	return Flags{Debug: *debug, Preseed: *preseed}
}
