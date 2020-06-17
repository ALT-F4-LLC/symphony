package main

import (
	"github.com/erkrnt/symphony/internal/cli"
	"gopkg.in/alecthomas/kingpin.v2"
)

var (
	block                      = kingpin.Command("block", "Block service commands.")
	blockEndpoint              = block.Flag("endpoint", "Endpoint for block service.").Default("127.0.0.1:15760").String()
	blockInit                  = block.Command("init", "Joins an existing initialized cluster.")
	blockInitAddr              = blockInit.Arg("addr", "Manager address of existing initialized cluster.").Required().String()
	blockLv                    = block.Command("lv", "Block service logical volume commands.")
	blockLvCreate              = blockLv.Command("create", "Creates a logical volume.")
	blockLvCreateID            = blockLvCreate.Arg("id", "Set logical volume id.").Required().String()
	blockLvCreateVolumeGroupID = blockLvCreate.Arg("volume-group-id", "Set logical volume group id.").Required().String()
	blockLvCreateSize          = blockLvCreate.Arg("size", "Set logical volume size.").Required().String()
	blockLvGet                 = blockLv.Command("get", "Gets a logical volume.")
	blockLvGetID               = blockLvGet.Arg("id", "Sets logical volume id.").Required().String()
	blockLvGetVolumeGroupID    = blockLvGet.Arg("volume-group-id", "Set logical volume group id.").Required().String()
	blockLvRemove              = blockLv.Command("remove", "Removes a logical volume.")
	blockLvRemoveID            = blockLvRemove.Arg("id", "Sets logical volume id.").Required().String()
	blockLvRemoveVolumeGroupID = blockLvRemove.Arg("volume-group-id", "Set logical volume group id.").Required().String()
	blockPv                    = block.Command("pv", "Block service physical volume commands.")
	blockPvCreate              = blockPv.Command("create", "Creates a physical volume.")
	blockPvCreateDevice        = blockPvCreate.Arg("device", "Specific physical volume to create.").Required().String()
	blockPvGet                 = blockPv.Command("get", "Gets a physical volume's metadata.")
	blockPvGetDevice           = blockPvGet.Arg("device", "Specific physical volume to retrieve.").Required().String()
	blockPvRemove              = blockPv.Command("remove", "Remove a physical volume.")
	blockPvRemoveDevice        = blockPvRemove.Arg("device", "Specific physical volume to remove.").Required().String()
	blockVg                    = block.Command("vg", "Block service volume group commands.")
	blockVgCreate              = blockVg.Command("create", "Creates a volume group.")
	blockVgCreateDevice        = blockVgCreate.Arg("device", "Set volume group device.").Required().String()
	blockVgCreateID            = blockVgCreate.Arg("id", "Set volume group id.").Required().String()
	blockVgGet                 = blockVg.Command("get", "Gets a volume group.")
	blockVgGetID               = blockVgGet.Arg("id", "Sets volume group id.").Required().String()
	blockVgRemove              = blockVg.Command("remove", "Removes a volume group.")
	blockVgRemoveID            = blockVgRemove.Arg("id", "Sets volume group id.").Required().String()
	manager                    = kingpin.Command("manager", "Manager service commands.")
	managerInit                = manager.Command("init", "Initializes new cluster.")
	managerInitAddr            = managerInit.Arg("addr", "Initializes in an existing cluster.").String()
	managerRemove              = manager.Command("remove", "Removes a service from the cluster.")
	managerRemoveMemberID      = managerRemove.Arg("service-id", "Service to be removed from the cluster.").Required().String()
	socket                     = kingpin.Flag("socket", "Sets the socket connection for the client.").Default("./control.sock").String()
)

func main() {
	switch kingpin.Parse() {
	case blockInit.FullCommand():
		cli.BlockInit(blockInitAddr, socket)
	case blockLvCreate.FullCommand():
		cli.BlockLvCreate(blockLvCreateID, blockLvCreateVolumeGroupID, blockLvCreateSize, blockEndpoint)
	case blockLvGet.FullCommand():
		cli.BlockLvGet(blockLvGetID, blockLvGetVolumeGroupID, blockEndpoint)
	case blockLvRemove.FullCommand():
		cli.BlockLvRemove(blockLvRemoveID, blockLvRemoveVolumeGroupID, blockEndpoint)
	case blockPvCreate.FullCommand():
		cli.BlockPvCreate(blockPvCreateDevice, blockEndpoint)
	case blockPvGet.FullCommand():
		cli.BlockPvGet(blockPvGetDevice, blockEndpoint)
	case blockPvRemove.FullCommand():
		cli.BlockPvRemove(blockPvRemoveDevice, blockEndpoint)
	case blockVgCreate.FullCommand():
		cli.BlockVgCreate(blockVgCreateDevice, blockVgCreateID, blockEndpoint)
	case blockVgGet.FullCommand():
		cli.BlockVgGet(blockVgGetID, blockEndpoint)
	case blockVgRemove.FullCommand():
		cli.BlockVgRemove(blockVgRemoveID, blockEndpoint)
	case managerInit.FullCommand():
		cli.ManagerInit(managerInitAddr, socket)
	case managerRemove.FullCommand():
		cli.ManagerRemove(managerRemoveMemberID, socket)
	}
}
