package manager

import (
	"errors"
	"sync"

	"github.com/erkrnt/symphony/api"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

type Action interface {
	Execute(eventCtx EventContent) EventType
}

type CreateLogicalVolumeAction struct{}
type CreatePhysicalVolumeAction struct{}
type CreateVolumeGroupAction struct{}

type Events map[EventType]StateType
type EventContent interface{}
type EventType string

type ReviewLogicalVolumeAction struct{}
type ReviewPhysicalVolumeAction struct{}
type ReviewVolumeGroupAction struct{}

type State struct {
	Action Action
	Events Events
}
type States map[StateType]State
type StateCreateFailAction struct{}
type StateEventContext struct {
	Manager      *Manager
	ResourceID   uuid.UUID
	ResourceType api.ResourceType
}
type StateMachine struct {
	Current  StateType
	Previous StateType
	States   States
	mutex    sync.Mutex
}
type StateType string
type StateReviewCompleteAction struct{}
type StateReviewFailAction struct{}

const (
	Default StateType = ""

	NoOp EventType = "NoOp"
)

var (
	CreateCompleted                EventType = EventType(api.ResourceStatus_CREATE_COMPLETED.String())
	CreateFailed                   EventType = EventType(api.ResourceStatus_CREATE_FAILED.String())
	CreateInProgressLogicalVolume  EventType = EventType("CreateInProgressLogicalVolume")
	CreateInProgressPhysicalVolume EventType = EventType("CreateInProgressPhysicalVolume")
	CreateInProgressVolumeGroup    EventType = EventType("CreateInProgressVolumeGroup")
	ReviewCompleted                EventType = EventType(api.ResourceStatus_REVIEW_COMPLETED.String())
	ReviewFailed                   EventType = EventType(api.ResourceStatus_REVIEW_FAILED.String())
	ReviewInProgressLogicalVolume  EventType = EventType("ReviewInProgressLogicalVolume")
	ReviewInProgressPhysicalVolume EventType = EventType("ReviewInProgressPhysicalVolume")
	ReviewInProgressVolumeGroup    EventType = EventType("ReviewInProgressVolumeGroup")

	CreateComplete       StateType = StateType("CreateComplete")
	CreateFail           StateType = StateType("CreateFail")
	CreateLogicalVolume  StateType = StateType("CreateLogicalVolume")
	CreatePhysicalVolume StateType = StateType("CreatePhysicalVolume")
	CreateVolumeGroup    StateType = StateType("CreateVolumeGroup")
	ReviewComplete       StateType = StateType("ReviewComplete")
	ReviewFail           StateType = StateType("ReviewFail")
	ReviewLogicalVolume  StateType = StateType("ReviewLogicalVolume")
	ReviewPhysicalVolume StateType = StateType("ReviewPhysicalVolume")
	ReviewVolumeGroup    StateType = StateType("ReviewVolumeGroup")
)

func newState() *StateMachine {
	defaultState := State{
		Events: Events{
			ReviewInProgressLogicalVolume:  ReviewLogicalVolume,
			ReviewInProgressPhysicalVolume: ReviewPhysicalVolume,
			ReviewInProgressVolumeGroup:    ReviewVolumeGroup,
		},
	}

	createFailState := State{
		Action: &StateCreateFailAction{},
		Events: Events{},
	}

	createLogicalVolumeState := State{
		Action: &CreateLogicalVolumeAction{},
		Events: Events{
			CreateFailed: CreateFail,
		},
	}

	createPhysicalVolumeState := State{
		Action: &CreatePhysicalVolumeAction{},
		Events: Events{
			CreateFailed: CreateFail,
		},
	}

	createVolumeGroupState := State{
		Action: &CreateVolumeGroupAction{},
		Events: Events{
			CreateFailed: CreateFail,
		},
	}

	reviewCompleteState := State{
		Action: &StateReviewCompleteAction{},
		Events: Events{
			CreateFailed:                   CreateFail,
			CreateInProgressLogicalVolume:  CreateLogicalVolume,
			CreateInProgressPhysicalVolume: CreatePhysicalVolume,
			CreateInProgressVolumeGroup:    CreateVolumeGroup,
		},
	}

	reviewFailState := State{
		Action: &StateReviewFailAction{},
		Events: Events{},
	}

	reviewLogicalVolumeState := State{
		Action: &ReviewLogicalVolumeAction{},
		Events: Events{
			ReviewCompleted: ReviewComplete,
			ReviewFailed:    ReviewFail,
		},
	}

	reviewPhysicalVolumeState := State{
		Action: &ReviewPhysicalVolumeAction{},
		Events: Events{
			ReviewCompleted: ReviewComplete,
			ReviewFailed:    ReviewFail,
		},
	}

	reviewVolumeGroupState := State{
		Action: &ReviewVolumeGroupAction{},
		Events: Events{
			ReviewCompleted: ReviewComplete,
			ReviewFailed:    ReviewFail,
		},
	}

	state := &StateMachine{
		States: States{
			Default:              defaultState,
			CreateFail:           createFailState,
			CreateLogicalVolume:  createLogicalVolumeState,
			CreatePhysicalVolume: createPhysicalVolumeState,
			CreateVolumeGroup:    createVolumeGroupState,
			ReviewComplete:       reviewCompleteState,
			ReviewFail:           reviewFailState,
			ReviewLogicalVolume:  reviewLogicalVolumeState,
			ReviewPhysicalVolume: reviewPhysicalVolumeState,
			ReviewVolumeGroup:    reviewVolumeGroupState,
		},
	}

	return state
}

func saveResouceStatus(context *StateEventContext, status api.ResourceStatus) error {
	switch context.ResourceType {
	case api.ResourceType_LOGICAL_VOLUME:
		lv, err := context.Manager.logicalVolumeByID(context.ResourceID)

		if err != nil {
			return err
		}

		lv.Status = status

		saveErr := context.Manager.saveLogicalVolume(lv)

		if saveErr != nil {
			return saveErr
		}
	case api.ResourceType_PHYSICAL_VOLUME:
		pv, err := context.Manager.physicalVolumeByID(context.ResourceID)

		if err != nil {
			return err
		}

		pv.Status = status

		saveErr := context.Manager.savePhysicalVolume(pv)

		if saveErr != nil {
			return saveErr
		}
	case api.ResourceType_VOLUME_GROUP:
		vg, err := context.Manager.volumeGroupByID(context.ResourceID)

		if err != nil {
			return err
		}

		vg.Status = status

		saveErr := context.Manager.saveVolumeGroup(vg)

		if saveErr != nil {
			return saveErr
		}
	}

	return nil
}

func (a *StateCreateFailAction) Execute(eventCtx EventContent) EventType {
	context := eventCtx.(*StateEventContext)

	status := api.ResourceStatus_CREATE_FAILED

	saveErr := saveResouceStatus(context, status)

	if saveErr != nil {
		return NoOp
	}

	fields := logrus.Fields{
		"ResourceID":   context.ResourceID,
		"ResourceType": context.ResourceType.String(),
	}

	logrus.WithFields(fields).Error(status.String())

	return NoOp
}

func (a *StateReviewCompleteAction) Execute(eventCtx EventContent) EventType {
	context := eventCtx.(*StateEventContext)

	status := api.ResourceStatus_CREATE_IN_PROGRESS

	saveErr := saveResouceStatus(context, status)

	if saveErr != nil {
		return CreateFailed
	}

	fields := logrus.Fields{
		"ResourceID":   context.ResourceID,
		"ResourceType": context.ResourceType.String(),
	}

	logrus.WithFields(fields).Info(status.String())

	switch context.ResourceType {
	case api.ResourceType_LOGICAL_VOLUME:
		return CreateInProgressLogicalVolume
	case api.ResourceType_PHYSICAL_VOLUME:
		return CreateInProgressPhysicalVolume
	case api.ResourceType_VOLUME_GROUP:
		return CreateInProgressVolumeGroup
	}

	return CreateFailed
}

func (a *StateReviewFailAction) Execute(eventCtx EventContent) EventType {
	context := eventCtx.(*StateEventContext)

	status := api.ResourceStatus_REVIEW_FAILED

	saveErr := saveResouceStatus(context, status)

	if saveErr != nil {
		return NoOp
	}

	fields := logrus.Fields{
		"ID":   context.ResourceID,
		"Type": context.ResourceType.String(),
	}

	logrus.WithFields(fields).Error(status.String())

	return NoOp
}

func (a *CreateLogicalVolumeAction) Execute(eventCtx EventContent) EventType {
	context := eventCtx.(*StateEventContext)

	m := context.Manager

	lv, err := m.logicalVolumeByID(context.ResourceID)

	if err != nil {
		return CreateFailed
	}

	volumeGroupID, err := uuid.Parse(lv.VolumeGroupID)

	if err != nil {
		return CreateFailed
	}

	vg, err := m.volumeGroupByID(volumeGroupID)

	if err != nil {
		return CreateFailed
	}

	physicalVolumeID, err := uuid.Parse(vg.PhysicalVolumeID)

	if err != nil {
		return CreateFailed
	}

	pv, err := m.physicalVolumeByID(physicalVolumeID)

	if err != nil {
		return CreateFailed
	}

	serviceID, err := uuid.Parse(pv.ServiceID)

	if err != nil {
		return CreateFailed
	}

	agentService, err := m.agentServiceByID(serviceID)

	if agentService == nil {
		return CreateFailed
	}

	agentServiceHealth, err := m.agentServiceHealth(agentService.ID)

	if err != nil {
		return CreateFailed
	}

	if agentServiceHealth == "critical" {
		return CreateFailed
	}

	lvCreateErr := m.lvCreate(agentService, lv)

	if lvCreateErr != nil {
		return CreateFailed
	}

	status := api.ResourceStatus_CREATE_COMPLETED

	lv.Status = status

	saveErr := m.saveLogicalVolume(lv)

	if saveErr != nil {
		return CreateFailed
	}

	return CreateCompleted
}

func (a *CreatePhysicalVolumeAction) Execute(eventCtx EventContent) EventType {
	context := eventCtx.(*StateEventContext)

	m := context.Manager

	pv, err := m.physicalVolumeByID(context.ResourceID)

	if err != nil {
		return CreateFailed
	}

	if pv.Status != api.ResourceStatus_CREATE_IN_PROGRESS {
		return CreateFailed
	}

	pvServiceID, err := uuid.Parse(pv.ServiceID)

	if err != nil {
		return CreateFailed
	}

	agentService, err := m.agentServiceByID(pvServiceID)

	if err != nil {
		return CreateFailed
	}

	if agentService == nil {
		return CreateFailed
	}

	agentServiceHealth, err := m.agentServiceHealth(agentService.ID)

	if err != nil {
		return CreateFailed
	}

	if agentServiceHealth == "critical" {
		return CreateFailed
	}

	pvCreateErr := m.pvCreate(agentService, pv)

	if pvCreateErr != nil {
		return CreateFailed
	}

	status := api.ResourceStatus_CREATE_COMPLETED

	pv.Status = status

	saveErr := m.savePhysicalVolume(pv)

	if saveErr != nil {
		return CreateFailed
	}

	fields := logrus.Fields{
		"ResourceID":   context.ResourceID,
		"ResourceType": context.ResourceType.String(),
	}

	logrus.WithFields(fields).Error(status.String())

	return CreateCompleted
}

func (a *CreateVolumeGroupAction) Execute(eventCtx EventContent) EventType {
	context := eventCtx.(*StateEventContext)

	m := context.Manager

	vg, err := m.volumeGroupByID(context.ResourceID)

	if err != nil {
		return CreateFailed
	}

	physicalVolumeID, err := uuid.Parse(vg.PhysicalVolumeID)

	if err != nil {
		return CreateFailed
	}

	pv, err := m.physicalVolumeByID(physicalVolumeID)

	if err != nil {
		return CreateFailed
	}

	serviceID, err := uuid.Parse(pv.ServiceID)

	if err != nil {
		return CreateFailed
	}

	agentService, err := m.agentServiceByID(serviceID)

	if agentService == nil {
		return CreateFailed
	}

	agentServiceHealth, err := m.agentServiceHealth(agentService.ID)

	if err != nil {
		return CreateFailed
	}

	if agentServiceHealth == "critical" {
		return CreateFailed
	}

	vgCreateErr := m.vgCreate(agentService, vg)

	if vgCreateErr != nil {
		return CreateFailed
	}

	status := api.ResourceStatus_CREATE_COMPLETED

	vg.Status = status

	saveErr := m.saveVolumeGroup(vg)

	if saveErr != nil {
		return CreateFailed
	}

	return CreateCompleted
}

func (a *ReviewLogicalVolumeAction) Execute(eventCtx EventContent) EventType {
	context := eventCtx.(*StateEventContext)

	m := context.Manager

	lv, err := m.logicalVolumeByID(context.ResourceID)

	if err != nil {
		return ReviewFailed
	}

	if lv.Status != api.ResourceStatus_REVIEW_IN_PROGRESS {
		return ReviewFailed
	}

	volumeGroupID, err := uuid.Parse(lv.VolumeGroupID)

	if err != nil {
		return ReviewFailed
	}

	vg, err := m.volumeGroupByID(volumeGroupID)

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

	pv, err := m.physicalVolumeByID(physicalVolumeID)

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

	as, err := m.agentServiceByID(serviceID)

	if err != nil {
		return ReviewFailed
	}

	if as == nil {
		return ReviewFailed
	}

	status := api.ResourceStatus_REVIEW_COMPLETED

	lv.Status = status

	saveErr := m.saveLogicalVolume(lv)

	if saveErr != nil {
		return ReviewFailed
	}

	fields := logrus.Fields{
		"ResourceID":   context.ResourceID,
		"ResourceType": context.ResourceType.String(),
	}

	logrus.WithFields(fields).Info(status.String())

	return ReviewCompleted
}

func (a *ReviewPhysicalVolumeAction) Execute(eventCtx EventContent) EventType {
	context := eventCtx.(*StateEventContext)

	m := context.Manager

	pv, err := m.physicalVolumeByID(context.ResourceID)

	if err != nil {
		return ReviewFailed
	}

	if pv.Status != api.ResourceStatus_REVIEW_IN_PROGRESS {
		return ReviewFailed
	}

	pvServiceID, err := uuid.Parse(pv.ServiceID)

	if err != nil {
		return ReviewFailed
	}

	as, err := m.agentServiceByID(pvServiceID)

	if err != nil {
		return ReviewFailed
	}

	if as == nil {
		return ReviewFailed
	}

	status := api.ResourceStatus_REVIEW_COMPLETED

	pv.Status = status

	saveErr := m.savePhysicalVolume(pv)

	if saveErr != nil {
		return ReviewFailed
	}

	fields := logrus.Fields{
		"ResourceID":   context.ResourceID,
		"ResourceType": context.ResourceType.String(),
	}

	logrus.WithFields(fields).Info(status.String())

	return ReviewCompleted
}

func (a *ReviewVolumeGroupAction) Execute(eventCtx EventContent) EventType {
	context := eventCtx.(*StateEventContext)

	m := context.Manager

	vg, err := m.volumeGroupByID(context.ResourceID)

	if err != nil {
		return ReviewFailed
	}

	if vg.Status != api.ResourceStatus_REVIEW_IN_PROGRESS {
		return ReviewFailed
	}

	physicalVolumeID, err := uuid.Parse(vg.PhysicalVolumeID)

	if err != nil {
		return ReviewFailed
	}

	pv, err := m.physicalVolumeByID(physicalVolumeID)

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

	as, err := m.agentServiceByID(serviceID)

	if err != nil {
		return ReviewFailed
	}

	if as == nil {
		return ReviewFailed
	}

	status := api.ResourceStatus_REVIEW_COMPLETED

	vg.Status = status

	saveErr := m.saveVolumeGroup(vg)

	if saveErr != nil {
		return ReviewFailed
	}

	fields := logrus.Fields{
		"ResourceID":   context.ResourceID,
		"ResourceType": context.ResourceType.String(),
	}

	logrus.WithFields(fields).Info(status.String())

	return ReviewCompleted
}

func (s *StateMachine) getNextState(event EventType) (StateType, error) {
	if state, ok := s.States[s.Current]; ok {
		if state.Events != nil {
			if next, ok := state.Events[event]; ok {
				return next, nil
			}
		}
	}

	return Default, errors.New("invalid_event")
}

func (s *StateMachine) SendEvent(event EventType, eventCtx EventContent) {
	s.mutex.Lock()

	defer s.mutex.Unlock()

	for {
		nextState, err := s.getNextState(event)

		if err != nil {
			logrus.Error(err)

			return
		}

		state, ok := s.States[nextState]

		if !ok || state.Action == nil {
			// configuration error
			return
		}

		s.Previous = s.Current

		s.Current = nextState

		nextEvent := state.Action.Execute(eventCtx)

		if nextEvent == NoOp {
			return
		}

		event = nextEvent
	}
}
