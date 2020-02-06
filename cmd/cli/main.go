package main

import (
	"github.com/erkrnt/symphony/internal/cli"
	"gopkg.in/alecthomas/kingpin.v2"
)

var (
	getCommand          = kingpin.Command("get", "Get a value from the raft state.")
	getCommandKey       = getCommand.Arg("key", "Specific key to retrieve for value.").Required().String()
	initCommand         = kingpin.Command("init", "Initializes a new cluster with optional peers list (minimum of 3)")
	initCommandJoinAddr = initCommand.Flag("join-addr", "Sets value to use for state.").String()
	initCommandPeers    = initCommand.Flag("peers", "Sets value to use for state.").String()
	joinCommand         = kingpin.Command("join", "Joins an existing initialized cluster.")
	joinCommandAddr     = joinCommand.Arg("addr", "Manager address of existing initialized cluster.").Required().String()
	setCommand          = kingpin.Command("set", "Sets a value into the raft state.")
	setCommandKey       = setCommand.Arg("key", "Specific key to set for value.").Required().String()
	setCommandValue     = setCommand.Arg("value", "Specific value to set for key.").Required().String()
	removeCommand       = kingpin.Command("remove", "Removes a node from the cluster.")
	removeCommandNodeID = removeCommand.Arg("node-id", "Raft node id to be removed from the cluster.").Required().Uint64()
	socket              = kingpin.Flag("socket", "Sets the socket connection for the client.").String()
)

func main() {
	switch kingpin.Parse() {
	case getCommand.FullCommand():
		cli.GetHandler(getCommandKey, socket)
	case initCommand.FullCommand():
		cli.InitHandler(initCommandJoinAddr, initCommandPeers, socket)
	case joinCommand.FullCommand():
		cli.JoinHandler(joinCommandAddr, socket)
	case removeCommand.FullCommand():
		cli.RemoveHandler(removeCommandNodeID, socket)
	case setCommand.FullCommand():
		cli.SetHandler(setCommandKey, socket, setCommandValue)
	}
}
