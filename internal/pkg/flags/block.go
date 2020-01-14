package flags

import (
	"gopkg.in/alecthomas/kingpin.v2"
)

// BlockFlags : command line flags for service
type BlockFlags struct {
	Join    *string
	Verbose *bool
}

// GetBlockFlags : gets struct of flags from command line
func GetBlockFlags() *BlockFlags {
	var (
		join    = kingpin.Flag("join", "Join existing Raft via node address.").String()
		verbose = kingpin.Flag("verbose", "Enables verbose output.").Short('v').Bool()
	)

	kingpin.Parse()

	flags := &BlockFlags{
		Join:    join,
		Verbose: verbose,
	}

	return flags
}
