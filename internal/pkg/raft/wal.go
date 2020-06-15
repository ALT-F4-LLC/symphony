package raft

import "os"

// walExists : checks if a file exists
func walExists(path string) bool {
	_, err := os.Stat(path)

	return err == nil
}
