package main

import (
	"github.com/erkrnt/symphony/internal/cli"
	"gopkg.in/alecthomas/kingpin.v2"
)

var (
	block                           = kingpin.Command("block", "Block service commands.")
	blockInit                       = block.Command("init", "Joins an existing initialized cluster.")
	blockInitAddr                   = blockInit.Arg("addr", "Manager address of existing initialized cluster.").Required().String()
	manager                         = kingpin.Command("manager", "Manager service commands.")
	managerEndpoint                 = manager.Flag("endpoint", "Endpoint for manager service.").Default("127.0.0.1:15760").String()
	managerInit                     = manager.Command("init", "Initializes new cluster.")
	managerInitAddr                 = managerInit.Arg("addr", "Initializes in an existing cluster.").String()
	managerLv                       = manager.Command("lv", "Block service logical volume commands.")
	managerLvCreate                 = managerLv.Command("create", "Creates a logical volume.")
	managerLvCreateServiceID        = managerLvCreate.Arg("service-id", "Set logical volume service id.").Required().String()
	managerLvCreateVolumeGroupID    = managerLvCreate.Arg("volume-group-id", "Set logical volume group id.").Required().String()
	managerLvCreateSize             = managerLvCreate.Arg("size", "Set logical volume size.").Required().Int64()
	managerLvGet                    = managerLv.Command("get", "Gets a logical volume.")
	managerLvGetID                  = managerLvGet.Arg("id", "Sets logical volume id.").Required().String()
	managerLvGetVolumeGroupID       = managerLvGet.Arg("volume-group-id", "Set logical volume group id.").Required().String()
	managerLvRemove                 = managerLv.Command("remove", "Removes a logical volume.")
	managerLvRemoveID               = managerLvRemove.Arg("id", "Sets logical volume id.").Required().String()
	managerLvRemoveVolumeGroupID    = managerLvRemove.Arg("volume-group-id", "Set logical volume group id.").Required().String()
	managerPv                       = manager.Command("pv", "manager service physical volume commands.")
	managerPvCreate                 = managerPv.Command("create", "Creates a physical volume.")
	managerPvCreateDeviceName       = managerPvCreate.Arg("device-name", "Specific physical volume to create.").Required().String()
	managerPvCreateServiceID        = managerPvCreate.Arg("service-id", "Specific service to create physical volume.").Required().String()
	managerPvGet                    = managerPv.Command("get", "Gets a physical volume's metadata.")
	managerPvGetID                  = managerPvGet.Arg("id", "Specific physical volume to retrieve.").Required().String()
	managerPvRemove                 = managerPv.Command("remove", "Remove a physical volume.")
	managerPvRemoveID               = managerPvRemove.Arg("id", "Specific physical volume to remove.").Required().String()
	managerRemove                   = manager.Command("remove", "Removes a service from the cluster.")
	managerRemoveMemberID           = managerRemove.Arg("service-id", "Service to be removed from the cluster.").Required().String()
	managerVg                       = manager.Command("vg", "manager service volume group commands.")
	managerVgCreate                 = managerVg.Command("create", "Creates a volume group.")
	managerVgCreatePhysicalVolumeID = managerVgCreate.Arg("physical-volume-id", "Sets physical volume for group.").Required().String()
	managerVgGet                    = managerVg.Command("get", "Gets a volume group.")
	managerVgGetID                  = managerVgGet.Arg("id", "Sets volume group id.").Required().String()
	managerVgRemove                 = managerVg.Command("remove", "Removes a volume group.")
	managerVgRemoveID               = managerVgRemove.Arg("id", "Sets volume group id.").Required().String()
	socket                          = kingpin.Flag("socket", "Sets the socket connection for the client.").Default("./control.sock").String()
)

func main() {
	switch kingpin.Parse() {
	case blockInit.FullCommand():
		cli.BlockInit(blockInitAddr, socket)
	case managerInit.FullCommand():
		cli.ManagerInit(managerInitAddr, socket)
	// case managerLvCreate.FullCommand():
	// 	cli.ManagerLvCreate(managerEndpoint, managerLvCreateServiceID, managerLvCreateSize, managerLvCreateVolumeGroupID)
	// case managerLvGet.FullCommand():
	// 	cli.ManagerLvGet(managerEndpoint, managerLvGetID)
	// case managerLvRemove.FullCommand():
	// 	cli.ManagerLvRemove(managerEndpoint, managerLvRemoveID)
	case managerPvCreate.FullCommand():
		cli.ManagerPvCreate(managerPvCreateDeviceName, managerEndpoint, managerPvCreateServiceID)
	case managerPvGet.FullCommand():
		cli.ManagerPvGet(managerEndpoint, managerPvGetID)
	case managerPvRemove.FullCommand():
		cli.ManagerPvRemove(managerEndpoint, managerPvRemoveID)
	case managerRemove.FullCommand():
		cli.ManagerRemove(managerRemoveMemberID, socket)
	case managerVgCreate.FullCommand():
		cli.ManagerVgCreate(managerEndpoint, managerVgCreatePhysicalVolumeID)
	case managerVgGet.FullCommand():
		cli.ManagerVgGet(managerEndpoint, managerVgGetID)
		// case managerVgRemove.FullCommand():
		// 	cli.ManagerVgRemove(managerVgRemoveID, managerEndpoint)
	}
}
