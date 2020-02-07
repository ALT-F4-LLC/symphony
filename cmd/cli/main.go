package main

import (
	"github.com/erkrnt/symphony/internal/cli"
	"gopkg.in/alecthomas/kingpin.v2"
)

var (
	blockCommand                 = kingpin.Command("block", "Block service commands.")
	blockCommandJoin             = blockCommand.Command("join", "Joins an existing initialized cluster.")
	blockCommandJoinAddr         = blockCommandJoin.Arg("addr", "Manager address of existing initialized cluster.").Required().String()
	getCommand                   = kingpin.Command("get", "Get a value from the raft state.")
	getCommandKey                = getCommand.Arg("key", "Specific key to retrieve for value.").Required().String()
	managerCommand               = kingpin.Command("manager", "Manager service commands.")
	managerCommandInit           = managerCommand.Command("init", "Initializes a new cluster with optional peers list (minimum of 3)")
	managerCommandInitJoinAddr   = managerCommandInit.Flag("join-addr", "Sets address of node in existing cluster.").String()
	managerCommandInitPeers      = managerCommandInit.Flag("peers", "Sets peers of nodes in existing cluster. (eg. http://<host>:<port>,http://<host>:<port>)").String()
	managerCommandJoin           = managerCommand.Command("join", "Joins an existing initialized cluster.")
	managerCommandJoinAddr       = managerCommandJoin.Arg("addr", "Manager address of existing initialized cluster.").Required().String()
	managerCommandRemove         = managerCommand.Command("remove", "Removes a node from the cluster.")
	managerCommandRemoveMemberID = managerCommandRemove.Arg("member-id", "Raft member ID to be removed from the cluster.").Required().Uint64()
	setCommand                   = kingpin.Command("set", "Sets a value into the raft state.")
	setCommandKey                = setCommand.Arg("key", "Specific key to set for value.").Required().String()
	setCommandValue              = setCommand.Arg("value", "Specific value to set for key.").Required().String()
	socket                       = kingpin.Flag("socket", "Sets the socket connection for the client.").String()
)

func main() {
	switch kingpin.Parse() {
	case blockCommandJoin.FullCommand():
		cli.BlockJoinHandler(blockCommandJoinAddr, socket)
	case getCommand.FullCommand():
		cli.GetHandler(getCommandKey, socket)
	case managerCommandInit.FullCommand():
		cli.ManagerInitHandler(managerCommandInitJoinAddr, managerCommandInitPeers, socket)
	case managerCommandJoin.FullCommand():
		cli.ManagerJoinHandler(managerCommandJoinAddr, socket)
	case managerCommandRemove.FullCommand():
		cli.ManagerRemoveHandler(managerCommandRemoveMemberID, socket)
	case setCommand.FullCommand():
		cli.SetHandler(setCommandKey, socket, setCommandValue)
	}
}
