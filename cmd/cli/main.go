package main

import (
	"github.com/erkrnt/symphony/internal/cli"
	"gopkg.in/alecthomas/kingpin.v2"
)

var (
	// block                      = kingpin.Command("block", "Block service commands.")
	// blockJoin                  = block.Command("join", "Joins an existing initialized cluster.")
	// blockJoinAddr              = blockJoin.Arg("addr", "Manager address of existing initialized cluster.").Required().String()
	// blockLv                    = block.Command("lv", "Block service logical volume commands.")
	// blockLvCreate              = blockLv.Command("create", "Creates a logical volume.")
	// blockLvCreateID            = blockLvCreate.Arg("id", "Set logical volume id.").Required().String()
	// blockLvCreateVolumeGroupID = blockLvCreate.Arg("volume-group-id", "Set logical volume group id.").Required().String()
	// blockLvCreateSize          = blockLvCreate.Arg("size", "Set logical volume size.").Required().String()
	// blockLvGet                 = blockLv.Command("get", "Gets a logical volume.")
	// blockLvGetID               = blockLvGet.Arg("id", "Sets logical volume id.").Required().String()
	// blockLvGetVolumeGroupID    = blockLvGet.Arg("volume-group-id", "Set logical volume group id.").Required().String()
	// blockLvRemove              = blockLv.Command("remove", "Removes a logical volume.")
	// blockLvRemoveID            = blockLvRemove.Arg("id", "Sets logical volume id.").Required().String()
	// blockLvRemoveVolumeGroupID = blockLvRemove.Arg("volume-group-id", "Set logical volume group id.").Required().String()
	// blockPv                    = block.Command("pv", "Block service physical volume commands.")
	// blockPvCreate              = blockPv.Command("create", "Creates a physical volume.")
	// blockPvCreateDevice        = blockPvCreate.Arg("device", "Specific physical volume to create.").Required().String()
	// blockPvGet                 = blockPv.Command("get", "Gets a physical volume's metadata.")
	// blockPvGetDevice           = blockPvGet.Arg("device", "Specific physical volume to retrieve.").Required().String()
	// blockPvRemove              = blockPv.Command("remove", "Remove a physical volume.")
	// blockPvRemoveDevice        = blockPvRemove.Arg("device", "Specific physical volume to remove.").Required().String()
	// blockVg                    = block.Command("vg", "Block service volume group commands.")
	// blockVgCreate              = blockVg.Command("create", "Creates a volume group.")
	// blockVgCreateDevice        = blockVgCreate.Arg("device", "Set volume group device.").Required().String()
	// blockVgCreateID            = blockVgCreate.Arg("id", "Set volume group id.").Required().String()
	// blockVgGet                 = blockVg.Command("get", "Gets a volume group.")
	// blockVgGetID               = blockVgGet.Arg("id", "Sets volume group id.").Required().String()
	// blockVgRemove              = blockVg.Command("remove", "Removes a volume group.")
	// blockVgRemoveID            = blockVgRemove.Arg("id", "Sets volume group id.").Required().String()
	manager               = kingpin.Command("manager", "Manager service commands.")
	managerInit           = manager.Command("init", "Initializes a new cluster.")
	managerJoin           = manager.Command("join", "Joins an existing initialized cluster.")
	managerJoinAddr       = managerJoin.Arg("addr", "Manager address of existing initialized cluster.").Required().String()
	managerMembers        = manager.Command("members", "Retrieves active cluster members.")
	managerRemove         = manager.Command("remove", "Removes a node from the cluster.")
	managerRemoveMemberID = managerRemove.Arg("member-id", "Raft member ID to be removed from the cluster.").Required().Uint64()
	remoteAddr            = kingpin.Flag("remote-addr", "Block service commands.").Default("127.0.0.1:27242").String()
	socket                = kingpin.Flag("socket", "Sets the socket connection for the client.").Default("./control.sock").String()
)

func main() {
	switch kingpin.Parse() {
	// case blockJoin.FullCommand():
	// 	cli.BlockJoin(blockJoinAddr, socket)
	// case blockLvCreate.FullCommand():
	// 	cli.BlockLvCreate(blockLvCreateID, blockLvCreateVolumeGroupID, blockLvCreateSize, remoteAddr)
	// case blockLvGet.FullCommand():
	// 	cli.BlockLvGet(blockLvGetID, blockLvGetVolumeGroupID, remoteAddr)
	// case blockLvRemove.FullCommand():
	// 	cli.BlockLvRemove(blockLvRemoveID, blockLvRemoveVolumeGroupID, remoteAddr)
	// case blockPvCreate.FullCommand():
	// 	cli.BlockPvCreate(blockPvCreateDevice, remoteAddr)
	// case blockPvGet.FullCommand():
	// 	cli.BlockPvGet(blockPvGetDevice, remoteAddr)
	// case blockPvRemove.FullCommand():
	// 	cli.BlockPvRemove(blockPvRemoveDevice, remoteAddr)
	// case blockVgCreate.FullCommand():
	// 	cli.BlockVgCreate(blockVgCreateDevice, blockVgCreateID, remoteAddr)
	// case blockVgGet.FullCommand():
	// 	cli.BlockVgGet(blockVgGetID, remoteAddr)
	// case blockVgRemove.FullCommand():
	// 	cli.BlockVgRemove(blockVgRemoveID, remoteAddr)
	case managerInit.FullCommand():
		cli.ManagerInit(socket)
	case managerJoin.FullCommand():
		cli.ManagerJoin(managerJoinAddr, socket)
	case managerMembers.FullCommand():
		cli.ManagerMembers(socket)
	case managerRemove.FullCommand():
		cli.ManagerRemove(managerRemoveMemberID, socket)
	}
}
