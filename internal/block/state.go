package block

import (
	"net"

	"github.com/erkrnt/symphony/api"
	"github.com/erkrnt/symphony/internal/utils"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

type CreateCompleteAction struct {
	ManagerAddr *net.TCPAddr
}
type CreateEventContext struct {
	ResourceID   uuid.UUID
	ResourceType api.ResourceType
}
type CreateFailAction struct {
	ManagerAddr *net.TCPAddr
}
type CreateLogicalVolumeAction struct {
	ManagerAddr *net.TCPAddr
}
type CreatePhysicalVolumeAction struct {
	ManagerAddr *net.TCPAddr
}
type CreateVolumeGroupAction struct {
	ManagerAddr *net.TCPAddr
}

const (
	CreateInProgressLogicalVolume  utils.EventType = utils.EventType("CreateInProgressLogicalVolume")
	CreateInProgressPhysicalVolume utils.EventType = utils.EventType("CreateInProgressPhysicalVolume")
	CreateInProgressVolumeGroup    utils.EventType = utils.EventType("CreateInProgressVolumeGroup")

	CreateLogicalVolume  utils.StateType = utils.StateType("CreateLogicalVolume")
	CreatePhysicalVolume utils.StateType = utils.StateType("CreatePhysicalVolume")
	CreateVolumeGroup    utils.StateType = utils.StateType("CreateVolumeGroup")
)

func newState(managerAddr *net.TCPAddr) *utils.StateMachine {
	defaultState := utils.State{
		Events: utils.Events{
			CreateInProgressLogicalVolume:  CreateLogicalVolume,
			CreateInProgressPhysicalVolume: CreatePhysicalVolume,
			CreateInProgressVolumeGroup:    CreateVolumeGroup,
		},
	}

	createCompleteState := utils.State{
		Action: &CreateCompleteAction{
			ManagerAddr: managerAddr,
		},
		Events: utils.Events{
			CreateInProgressLogicalVolume:  CreateLogicalVolume,
			CreateInProgressPhysicalVolume: CreatePhysicalVolume,
			CreateInProgressVolumeGroup:    CreateVolumeGroup,
		},
	}

	createFailState := utils.State{
		Action: &CreateFailAction{
			ManagerAddr: managerAddr,
		},
		Events: utils.Events{
			CreateInProgressLogicalVolume:  CreateLogicalVolume,
			CreateInProgressPhysicalVolume: CreatePhysicalVolume,
			CreateInProgressVolumeGroup:    CreateVolumeGroup,
		},
	}

	createLogicalVolumeState := utils.State{
		Action: &CreateLogicalVolumeAction{
			ManagerAddr: managerAddr,
		},
		Events: utils.Events{
			utils.CreateCompleted: utils.CreateComplete,
			utils.CreateFailed:    utils.CreateFail,
		},
	}

	createPhysicalVolumeState := utils.State{
		Action: &CreatePhysicalVolumeAction{
			ManagerAddr: managerAddr,
		},
		Events: utils.Events{
			utils.CreateCompleted: utils.CreateComplete,
			utils.CreateFailed:    utils.CreateFail,
		},
	}

	createVolumeGroupState := utils.State{
		Action: &CreateVolumeGroupAction{
			ManagerAddr: managerAddr,
		},
		Events: utils.Events{
			utils.CreateCompleted: utils.CreateComplete,
			utils.CreateFailed:    utils.CreateFail,
		},
	}

	state := &utils.StateMachine{
		States: utils.States{
			utils.Default:        defaultState,
			utils.CreateComplete: createCompleteState,
			utils.CreateFail:     createFailState,
			CreateLogicalVolume:  createLogicalVolumeState,
			CreatePhysicalVolume: createPhysicalVolumeState,
			CreateVolumeGroup:    createVolumeGroupState,
		},
	}

	return state
}

func saveResouceStatus(context *CreateEventContext, managerAddr *net.TCPAddr, status api.ResourceStatus) error {
	switch context.ResourceType {
	case api.ResourceType_LOGICAL_VOLUME:
		_, updateErr := updateLogicalVolume(managerAddr, context.ResourceID, status)

		if updateErr != nil {
			return updateErr
		}
	case api.ResourceType_PHYSICAL_VOLUME:
		_, updateErr := updatePhysicalVolume(managerAddr, context.ResourceID, status)

		if updateErr != nil {
			return updateErr
		}
	case api.ResourceType_VOLUME_GROUP:
		_, updateErr := updateVolumeGroup(managerAddr, context.ResourceID, status)

		if updateErr != nil {
			return updateErr
		}
	}

	return nil
}

func (action *CreateCompleteAction) Execute(eventCtx utils.EventContent) utils.EventType {
	managerAddr := action.ManagerAddr

	context := eventCtx.(*CreateEventContext)

	status := api.ResourceStatus_CREATE_COMPLETED

	saveErr := saveResouceStatus(context, managerAddr, status)

	if saveErr != nil {
		return utils.CreateFailed
	}

	fields := logrus.Fields{
		"ResourceID":   context.ResourceID,
		"ResourceType": context.ResourceType.String(),
	}

	logrus.WithFields(fields).Info(status.String())

	return utils.NoOp
}

func (action *CreateFailAction) Execute(eventCtx utils.EventContent) utils.EventType {
	managerAddr := action.ManagerAddr

	context := eventCtx.(*CreateEventContext)

	status := api.ResourceStatus_CREATE_FAILED

	saveErr := saveResouceStatus(context, managerAddr, status)

	if saveErr != nil {
		return utils.NoOp
	}

	fields := logrus.Fields{
		"ResourceID":   context.ResourceID,
		"ResourceType": context.ResourceType.String(),
	}

	logrus.WithFields(fields).Error(status.String())

	return utils.NoOp
}

func (action *CreateLogicalVolumeAction) Execute(eventCtx utils.EventContent) utils.EventType {
	context := eventCtx.(*CreateEventContext)

	fields := logrus.Fields{
		"ResourceID":   context.ResourceID,
		"ResourceType": context.ResourceType.String(),
	}

	status := api.ResourceStatus_CREATE_IN_PROGRESS

	logrus.WithFields(fields).Info(status.String())

	lv, err := logicalVolume(action.ManagerAddr, context.ResourceID)

	if err != nil {
		return utils.CreateFailed
	}

	vgID, err := uuid.Parse(lv.VolumeGroupID)

	if err != nil {
		return utils.CreateFailed
	}

	_, newLvErr := newLv(vgID, context.ResourceID, lv.Size)

	if newLvErr != nil {
		return utils.CreateFailed
	}

	return utils.CreateCompleted
}

func (action *CreatePhysicalVolumeAction) Execute(eventCtx utils.EventContent) utils.EventType {
	context := eventCtx.(*CreateEventContext)

	fields := logrus.Fields{
		"ResourceID":   context.ResourceID,
		"ResourceType": context.ResourceType.String(),
	}

	status := api.ResourceStatus_CREATE_IN_PROGRESS

	logrus.WithFields(fields).Info(status.String())

	pv, err := physicalVolume(action.ManagerAddr, context.ResourceID)

	if err != nil {
		return utils.CreateFailed
	}

	_, newPvErr := newPv(pv.DeviceName)

	if newPvErr != nil {
		return utils.CreateFailed
	}

	return utils.CreateCompleted
}

func (action *CreateVolumeGroupAction) Execute(eventCtx utils.EventContent) utils.EventType {
	context := eventCtx.(*CreateEventContext)

	fields := logrus.Fields{
		"ResourceID":   context.ResourceID,
		"ResourceType": context.ResourceType.String(),
	}

	status := api.ResourceStatus_CREATE_IN_PROGRESS

	logrus.WithFields(fields).Info(status.String())

	vg, err := volumeGroup(action.ManagerAddr, context.ResourceID)

	if err != nil {
		return utils.CreateFailed
	}

	physicalVolumeID, err := uuid.Parse(vg.PhysicalVolumeID)

	if err != nil {
		return utils.CreateFailed
	}

	pv, err := physicalVolume(action.ManagerAddr, physicalVolumeID)

	_, newVgErr := newVg(pv.DeviceName, context.ResourceID)

	if newVgErr != nil {
		return utils.CreateFailed
	}

	return utils.CreateCompleted
}
