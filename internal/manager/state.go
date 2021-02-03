package manager

import (
	"encoding/json"

	"github.com/erkrnt/symphony/api"
	"github.com/erkrnt/symphony/internal/utils"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

type ReviewCompleteAction struct {
	ConsulAddr string
}
type ReviewEventContext struct {
	ResourceID   uuid.UUID
	ResourceType api.ResourceType
}
type ReviewFailAction struct {
	ConsulAddr string
}
type ReviewLogicalVolumeAction struct {
	ConsulAddr string
}
type ReviewPhysicalVolumeAction struct {
	ConsulAddr string
}
type ReviewVolumeGroupAction struct {
	ConsulAddr string
}

var (
	ReviewCompleted                utils.EventType = utils.EventType("ReviewCompleted")
	ReviewFailed                   utils.EventType = utils.EventType("ReviewFailed")
	ReviewInProgressLogicalVolume  utils.EventType = utils.EventType("ReviewInProgressLogicalVolume")
	ReviewInProgressPhysicalVolume utils.EventType = utils.EventType("ReviewInProgressPhysicalVolume")
	ReviewInProgressVolumeGroup    utils.EventType = utils.EventType("ReviewInProgressVolumeGroup")

	ReviewComplete       utils.StateType = utils.StateType("ReviewComplete")
	ReviewFail           utils.StateType = utils.StateType("ReviewFail")
	ReviewLogicalVolume  utils.StateType = utils.StateType("ReviewLogicalVolume")
	ReviewPhysicalVolume utils.StateType = utils.StateType("ReviewPhysicalVolume")
	ReviewVolumeGroup    utils.StateType = utils.StateType("ReviewVolumeGroup")
)

func newState(consulAddr string) *utils.StateMachine {
	defaultState := utils.State{
		Events: utils.Events{
			ReviewInProgressLogicalVolume:  ReviewLogicalVolume,
			ReviewInProgressPhysicalVolume: ReviewPhysicalVolume,
			ReviewInProgressVolumeGroup:    ReviewVolumeGroup,
		},
	}

	reviewCompleteState := utils.State{
		Action: &ReviewCompleteAction{
			ConsulAddr: consulAddr,
		},
		Events: utils.Events{
			ReviewInProgressLogicalVolume:  ReviewLogicalVolume,
			ReviewInProgressPhysicalVolume: ReviewPhysicalVolume,
			ReviewInProgressVolumeGroup:    ReviewVolumeGroup,
		},
	}

	reviewFailState := utils.State{
		Action: &ReviewFailAction{
			ConsulAddr: consulAddr,
		},
		Events: utils.Events{},
	}

	reviewLogicalVolumeState := utils.State{
		Action: &ReviewLogicalVolumeAction{
			ConsulAddr: consulAddr,
		},
		Events: utils.Events{
			ReviewCompleted: ReviewComplete,
			ReviewFailed:    ReviewFail,
		},
	}

	reviewPhysicalVolumeState := utils.State{
		Action: &ReviewPhysicalVolumeAction{
			ConsulAddr: consulAddr,
		},
		Events: utils.Events{
			ReviewCompleted: ReviewComplete,
			ReviewFailed:    ReviewFail,
		},
	}

	reviewVolumeGroupState := utils.State{
		Action: &ReviewVolumeGroupAction{
			ConsulAddr: consulAddr,
		},
		Events: utils.Events{
			ReviewCompleted: ReviewComplete,
			ReviewFailed:    ReviewFail,
		},
	}

	state := &utils.StateMachine{
		States: utils.States{
			utils.Default:        defaultState,
			ReviewComplete:       reviewCompleteState,
			ReviewFail:           reviewFailState,
			ReviewLogicalVolume:  reviewLogicalVolumeState,
			ReviewPhysicalVolume: reviewPhysicalVolumeState,
			ReviewVolumeGroup:    reviewVolumeGroupState,
		},
	}

	return state
}

func saveResouceStatus(consulAddr string, context *ReviewEventContext, status api.ResourceStatus) error {
	switch context.ResourceType {
	case api.ResourceType_LOGICAL_VOLUME:
		lv, err := logicalVolumeByID(consulAddr, context.ResourceID)

		if err != nil {
			return err
		}

		lv.Status = status

		key := utils.KvKey(context.ResourceID, api.ResourceType_LOGICAL_VOLUME)

		value, err := json.Marshal(lv)

		if err != nil {
			return err
		}

		saveErr := putKVPair(consulAddr, key, value)

		if saveErr != nil {
			return saveErr
		}
	case api.ResourceType_PHYSICAL_VOLUME:
		pv, err := physicalVolumeByID(consulAddr, context.ResourceID)

		if err != nil {
			return err
		}

		pv.Status = status

		key := utils.KvKey(context.ResourceID, api.ResourceType_PHYSICAL_VOLUME)

		value, err := json.Marshal(pv)

		if err != nil {
			return err
		}

		saveErr := putKVPair(consulAddr, key, value)

		if saveErr != nil {
			return saveErr
		}
	case api.ResourceType_VOLUME_GROUP:
		vg, err := volumeGroupByID(consulAddr, context.ResourceID)

		if err != nil {
			return err
		}

		vg.Status = status

		key := utils.KvKey(context.ResourceID, api.ResourceType_VOLUME_GROUP)

		value, err := json.Marshal(vg)

		if err != nil {
			return err
		}

		saveErr := putKVPair(consulAddr, key, value)

		if saveErr != nil {
			return saveErr
		}
	}

	return nil
}

func (action *ReviewCompleteAction) Execute(eventCtx utils.EventContent) utils.EventType {
	consulAddr := action.ConsulAddr

	context := eventCtx.(*ReviewEventContext)

	status := api.ResourceStatus_REVIEW_COMPLETED

	saveErr := saveResouceStatus(consulAddr, context, status)

	if saveErr != nil {
		return ReviewFailed
	}

	fields := logrus.Fields{
		"ResourceID":   context.ResourceID,
		"ResourceType": context.ResourceType.String(),
	}

	logrus.WithFields(fields).Info(status.String())

	return utils.NoOp
}

func (action *ReviewFailAction) Execute(eventCtx utils.EventContent) utils.EventType {
	consulAddr := action.ConsulAddr

	context := eventCtx.(*ReviewEventContext)

	status := api.ResourceStatus_REVIEW_FAILED

	saveErr := saveResouceStatus(consulAddr, context, status)

	if saveErr != nil {
		return utils.NoOp
	}

	fields := logrus.Fields{
		"ID":   context.ResourceID,
		"Type": context.ResourceType.String(),
	}

	logrus.WithFields(fields).Error(status.String())

	return utils.NoOp
}

func (action *ReviewLogicalVolumeAction) Execute(eventCtx utils.EventContent) utils.EventType {
	consulAddr := action.ConsulAddr

	context := eventCtx.(*ReviewEventContext)

	fields := logrus.Fields{
		"ResourceID":   context.ResourceID,
		"ResourceType": context.ResourceType.String(),
	}

	status := api.ResourceStatus_REVIEW_IN_PROGRESS

	logrus.WithFields(fields).Info(status.String())

	lv, err := logicalVolumeByID(consulAddr, context.ResourceID)

	if err != nil {
		return ReviewFailed
	}

	if lv.Status != status {
		return ReviewFailed
	}

	volumeGroupID, err := uuid.Parse(lv.VolumeGroupID)

	if err != nil {
		return ReviewFailed
	}

	vg, err := volumeGroupByID(consulAddr, volumeGroupID)

	if err != nil {
		return ReviewFailed
	}

	if vg.Status != api.ResourceStatus_CREATE_COMPLETED {
		return ReviewFailed
	}

	physicalVolumeID, err := uuid.Parse(vg.PhysicalVolumeID)

	if err != nil {
		return ReviewFailed
	}

	pv, err := physicalVolumeByID(consulAddr, physicalVolumeID)

	if err != nil {
		return ReviewFailed
	}

	if pv.Status != api.ResourceStatus_CREATE_COMPLETED {
		return ReviewFailed
	}

	serviceID, err := uuid.Parse(pv.ServiceID)

	if err != nil {
		return ReviewFailed
	}

	as, err := agentServiceByID(consulAddr, serviceID)

	if err != nil {
		return ReviewFailed
	}

	if as == nil {
		return ReviewFailed
	}

	return ReviewCompleted
}

func (action *ReviewPhysicalVolumeAction) Execute(eventCtx utils.EventContent) utils.EventType {
	consulAddr := action.ConsulAddr

	context := eventCtx.(*ReviewEventContext)

	fields := logrus.Fields{
		"ResourceID":   context.ResourceID,
		"ResourceType": context.ResourceType.String(),
	}

	status := api.ResourceStatus_REVIEW_IN_PROGRESS

	logrus.WithFields(fields).Info(status.String())

	pv, err := physicalVolumeByID(consulAddr, context.ResourceID)

	if err != nil {
		return ReviewFailed
	}

	if pv.Status != status {
		return ReviewFailed
	}

	pvServiceID, err := uuid.Parse(pv.ServiceID)

	if err != nil {
		return ReviewFailed
	}

	as, err := agentServiceByID(consulAddr, pvServiceID)

	if err != nil {
		return ReviewFailed
	}

	if as == nil {
		return ReviewFailed
	}

	return ReviewCompleted
}

func (action *ReviewVolumeGroupAction) Execute(eventCtx utils.EventContent) utils.EventType {
	consulAddr := action.ConsulAddr

	context := eventCtx.(*ReviewEventContext)

	fields := logrus.Fields{
		"ResourceID":   context.ResourceID,
		"ResourceType": context.ResourceType.String(),
	}

	status := api.ResourceStatus_REVIEW_IN_PROGRESS

	logrus.WithFields(fields).Info(status.String())

	vg, err := volumeGroupByID(consulAddr, context.ResourceID)

	if err != nil {
		return ReviewFailed
	}

	if vg.Status != status {
		return ReviewFailed
	}

	physicalVolumeID, err := uuid.Parse(vg.PhysicalVolumeID)

	if err != nil {
		return ReviewFailed
	}

	pv, err := physicalVolumeByID(consulAddr, physicalVolumeID)

	if err != nil {
		return ReviewFailed
	}

	if pv.Status != api.ResourceStatus_CREATE_COMPLETED {
		return ReviewFailed
	}

	serviceID, err := uuid.Parse(pv.ServiceID)

	if err != nil {
		return ReviewFailed
	}

	as, err := agentServiceByID(consulAddr, serviceID)

	if err != nil {
		return ReviewFailed
	}

	if as == nil {
		return ReviewFailed
	}

	return ReviewCompleted
}
