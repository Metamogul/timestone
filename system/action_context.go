package system

import (
	"context"

	"github.com/metamogul/timestone"
)

type actionContext struct {
	context.Context
	clock timestone.Clock
}

func newActionContext(ctx context.Context, clock timestone.Clock) *actionContext {
	return &actionContext{
		Context: ctx,
		clock:   clock,
	}
}

func (a *actionContext) Clock() timestone.Clock {
	return a.clock
}

func (a *actionContext) DoneSchedulingNewActions() { /*Noop*/ }

func (a *actionContext) Value(key any) any {
	switch key {
	case timestone.ActionContextClockKey:
		return a.clock
	default:
		return a.Context.Value(key)
	}
}
