package utils

import (
	"errors"
	"sync"

	"github.com/sirupsen/logrus"
)

type Action interface {
	Execute(eventCtx EventContent) EventType
}

type Events map[EventType]StateType
type EventContent interface{}
type EventType string

type State struct {
	Action Action
	Events Events
}
type States map[StateType]State
type StateCreateFailAction struct{}

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
	CreateCompleted EventType = EventType("CreateCompleted")
	CreateFailed    EventType = EventType("CreateFailed")

	CreateComplete StateType = StateType("CreateComplete")
	CreateFail     StateType = StateType("CreateFail")
)

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

func (s *StateMachine) Send(event EventType, eventCtx EventContent) {
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
