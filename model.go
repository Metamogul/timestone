package timestone

import (
	"context"
	"time"
)

type Clock interface {
	Now() time.Time
}

const ActionContextClockKey = "actionContextClock"

type ActionContext interface {
	context.Context
	Clock() Clock
	DoneSchedulingNewActions()
}

type Action interface {
	Perform(ActionContext)
	Name() string
}

type SimpleAction struct {
	action func(ActionContext)
	name   string
}

func NewSimpleAction(action func(ActionContext), name string) *SimpleAction {
	return &SimpleAction{action, name}
}

func (s *SimpleAction) Perform(ctx ActionContext) {
	s.action(ctx)
}

func (s *SimpleAction) Name() string {
	return s.name
}

type Scheduler interface {
	Clock
	PerformNow(ctx context.Context, action Action)
	PerformAfter(ctx context.Context, action Action, duration time.Duration)
	PerformRepeatedly(ctx context.Context, action Action, until *time.Time, interval time.Duration)
}
