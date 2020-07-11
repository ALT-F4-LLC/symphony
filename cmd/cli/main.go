package main

import (
	"github.com/erkrnt/symphony/internal/cli"
	"gopkg.in/alecthomas/kingpin.v2"
)

var (
	block = kingpin.Command("block", "Block service commands.")

	blockInit     = block.Command("init", "Joins an existing initialized cluster.")
	blockInitAddr = blockInit.Arg("addr", "Manager address of existing initialized cluster.").Required().String()

	manager = kingpin.Command("manager", "Manager service commands.")

	managerLv                    = manager.Command("lv", "Block service logical volume commands.")
	managerLvCreate              = managerLv.Command("create", "Creates a logical volume.")
	managerLvCreateSize          = managerLvCreate.Arg("size", "Sets logical volume size.").Required().Int64()
	managerLvCreateVolumeGroupID = managerLvCreate.Arg("volume-group-id", "Sets logical volume group id.").Required().String()
	managerLvGet                 = managerLv.Command("get", "Gets a logical volume.")
	managerLvGetID               = managerLvGet.Arg("id", "Sets logical volume id.").Required().String()
	managerLvList                = managerLv.Command("list", "Gets all logical volumes.")
	managerLvRemove              = managerLv.Command("remove", "Removes a logical volume.")
	managerLvRemoveID            = managerLvRemove.Arg("id", "Sets logical volume id.").Required().String()

	managerPv                 = manager.Command("pv", "manager service physical volume commands.")
	managerPvCreate           = managerPv.Command("create", "Creates a physical volume.")
	managerPvCreateDeviceName = managerPvCreate.Arg("device-name", "Specific physical volume to create.").Required().String()
	managerPvCreateServiceID  = managerPvCreate.Arg("service-id", "Specific service to create physical volume.").Required().String()
	managerPvGet              = managerPv.Command("get", "Gets a physical volume's metadata.")
	managerPvGetID            = managerPvGet.Arg("id", "Specific physical volume to retrieve.").Required().String()
	managerPvList             = managerPv.Command("list", "Gets all physical volumes.")
	managerPvRemove           = managerPv.Command("remove", "Remove a physical volume.")
	managerPvRemoveID         = managerPvRemove.Arg("id", "Specific physical volume to remove.").Required().String()

	managerService               = manager.Command("service", "Block service logical volume commands.")
	managerServiceInit           = managerService.Command("init", "Initializes new cluster.")
	managerServiceList           = managerService.Command("list", "Gets all cloud services.")
	managerServiceRemove         = managerService.Command("remove", "Removes a service from the cluster.")
	managerServiceRemoveMemberID = managerServiceRemove.Arg("service-id", "Service to be removed from the cluster.").Required().String()

	managerVg                       = manager.Command("vg", "manager service volume group commands.")
	managerVgCreate                 = managerVg.Command("create", "Creates a volume group.")
	managerVgCreatePhysicalVolumeID = managerVgCreate.Arg("physical-volume-id", "Sets physical volume for group.").Required().String()
	managerVgGet                    = managerVg.Command("get", "Gets a volume group.")
	managerVgGetID                  = managerVgGet.Arg("id", "Sets volume group id.").Required().String()
	managerVgList                   = managerVg.Command("list", "Gets all volume groups.")
	managerVgRemove                 = managerVg.Command("remove", "Removes a volume group.")
	managerVgRemoveID               = managerVgRemove.Arg("id", "Sets volume group id.").Required().String()

	socket = kingpin.Flag("socket", "Sets the socket connection for the client.").Default("./control.sock").String()
)

func main() {
	switch kingpin.Parse() {
	case blockInit.FullCommand():
		cli.BlockServiceInit(blockInitAddr, socket)

	case managerLvCreate.FullCommand():
		cli.ManagerNewLogicalVolume(managerLvCreateSize, managerLvCreateVolumeGroupID, socket)
	case managerLvGet.FullCommand():
		cli.ManagerGetLogicalVolume(managerLvGetID, socket)
	case managerLvList.FullCommand():
		cli.ManagerListLogicalVolumes(socket)
	case managerLvRemove.FullCommand():
		cli.ManagerRemoveLogicalVolume(managerLvRemoveID, socket)

	case managerPvCreate.FullCommand():
		cli.ManagerNewPhysicalVolume(managerPvCreateDeviceName, managerPvCreateServiceID, socket)
	case managerPvGet.FullCommand():
		cli.ManagerGetPhysicalVolume(managerPvGetID, socket)
	case managerPvList.FullCommand():
		cli.ManagerListPhysicalVolumes(socket)
	case managerPvRemove.FullCommand():
		cli.ManagerRemovePhysicalVolume(managerPvRemoveID, socket)

	case managerServiceInit.FullCommand():
		cli.ManagerServiceInit(socket)
	case managerServiceList.FullCommand():
		cli.ManagerServiceList(socket)
	case managerServiceRemove.FullCommand():
		cli.ManagerRemoveService(managerServiceRemoveMemberID, socket)

	case managerVgCreate.FullCommand():
		cli.ManagerNewVolumeGroup(managerVgCreatePhysicalVolumeID, socket)
	case managerVgGet.FullCommand():
		cli.ManagerGetVolumeGroup(managerVgGetID, socket)
	case managerVgList.FullCommand():
		cli.ManagerListVolumeGroups(socket)
	case managerVgRemove.FullCommand():
		cli.ManagerRemoveVolumeGroup(managerVgRemoveID, socket)
	}
}
