package main

import (
	"github.com/erkrnt/symphony/internal/cli"
	"gopkg.in/alecthomas/kingpin.v2"
)

var (
	apiserverAddr = kingpin.Flag("apiserver-addr", "Sets the apiserver address. Defaults to 127.0.0.1:15760").Default("127.0.0.1:15760").String()
	socketPath    = kingpin.Flag("socket-path", "Sets the socket address. Defaults to ./control.sock").Default("./control.sock").String()

	block                         = kingpin.Command("block", "Block management commands.")
	blockLv                       = block.Command("lv", "Block service logical volume commands.")
	blockLvCreate                 = block.Command("create", "Creates a logical volume.")
	blockLvCreateSize             = blockLvCreate.Arg("size", "Sets logical volume size.").Required().Int64()
	blockLvCreateVolumeGroupID    = blockLvCreate.Arg("volume-group-id", "Sets logical volume group id.").Required().String()
	blockLvGet                    = blockLv.Command("get", "Gets a logical volume.")
	blockLvGetID                  = blockLvGet.Arg("id", "Sets logical volume id.").Required().String()
	blockLvList                   = blockLv.Command("list", "Gets all logical volumes.")
	blockLvRemove                 = blockLv.Command("remove", "Removes a logical volume.")
	blockLvRemoveID               = blockLvRemove.Arg("id", "Sets logical volume id.").Required().String()
	blockPv                       = block.Command("pv", "manager service physical volume commands.")
	blockPvCreate                 = blockPv.Command("create", "Creates a physical volume.")
	blockPvCreateDeviceName       = blockPvCreate.Arg("device-name", "Specific physical volume to create.").Required().String()
	blockPvCreateServiceID        = blockPvCreate.Arg("service-id", "Specific service to create physical volume.").Required().String()
	blockPvGet                    = blockPv.Command("get", "Gets a physical volume's metadata.")
	blockPvGetID                  = blockPvGet.Arg("id", "Specific physical volume to retrieve.").Required().String()
	blockPvList                   = blockPv.Command("list", "Gets all physical volumes.")
	blockPvRemove                 = blockPv.Command("remove", "Remove a physical volume.")
	blockPvRemoveID               = blockPvRemove.Arg("id", "Specific physical volume to remove.").Required().String()
	blockVg                       = block.Command("vg", "manager service volume group commands.")
	blockVgCreate                 = blockVg.Command("create", "Creates a volume group.")
	blockVgCreatePhysicalVolumeID = blockVgCreate.Arg("physical-volume-id", "Sets physical volume for group.").Required().String()
	blockVgGet                    = blockVg.Command("get", "Gets a volume group.")
	blockVgGetID                  = blockVgGet.Arg("id", "Sets volume group id.").Required().String()
	blockVgList                   = blockVg.Command("list", "Gets all volume groups.")
	blockVgRemove                 = blockVg.Command("remove", "Removes a volume group.")
	blockVgRemoveID               = blockVgRemove.Arg("id", "Sets volume group id.").Required().String()

	service               = kingpin.Command("service", "Service management commands.")
	serviceNew            = service.Command("new", "Initializes service to an existing cluster.")
	serviceRemove         = service.Command("remove", "Removes a service from the cluster.")
	serviceRemoveMemberID = serviceRemove.Arg("service-id", "Service to be removed from the cluster.").Required().String()
)

func main() {
	switch kingpin.Parse() {
	/* case blockLvCreate.FullCommand():*/
	//cli.ManagerNewLogicalVolume(blockLvCreateSize, blockLvCreateVolumeGroupID)
	//case blockLvGet.FullCommand():
	//cli.ManagerGetLogicalVolume(blockLvGetID)
	//case blockLvList.FullCommand():
	//cli.ManagerListLogicalVolumes()
	//case blockLvRemove.FullCommand():
	//cli.ManagerRemoveLogicalVolume(blockLvRemoveID)
	//case blockPvCreate.FullCommand():
	//cli.ManagerNewPhysicalVolume(blockPvCreateDeviceName, blockPvCreateServiceID)
	//case blockPvGet.FullCommand():
	//cli.ManagerGetPhysicalVolume(blockPvGetID)
	//case blockPvList.FullCommand():
	//cli.ManagerListPhysicalVolumes()
	//case blockPvRemove.FullCommand():
	//cli.ManagerRemovePhysicalVolume(blockPvRemoveID)
	//case blockVgCreate.FullCommand():
	//cli.ManagerNewVolumeGroup(blockVgCreatePhysicalVolumeID)
	//case blockVgGet.FullCommand():
	//cli.ManagerGetVolumeGroup(blockVgGetID)
	//case blockVgList.FullCommand():
	//cli.ManagerListVolumeGroups()
	//case blockVgRemove.FullCommand():
	//cli.ManagerRemoveVolumeGroup(blockVgRemoveID)

	case serviceNew.FullCommand():
		cli.ServiceNew(cli.ServiceNewOptions{
			SocketPath: socketPath,
		})

		/* case serviceList.FullCommand():*/
		//cli.ManagerServiceList()
		//case serviceRemove.FullCommand():
		//cli.ManagerRemoveService(serviceRemoveMemberID)
	}
}
