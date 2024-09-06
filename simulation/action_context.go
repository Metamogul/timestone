package simulation

import (
	"context"
	"sync"

	"github.com/metamogul/timestone"
)

const ActionContextEventLoopBlockerKey = "actionContextEventLoopBlocker"

type actionContext struct {
	context.Context

	clock            timestone.Clock
	eventLoopBlocker *sync.WaitGroup
}

func newActionContext(ctx context.Context, clock timestone.Clock, eventLoopBlocker *sync.WaitGroup) *actionContext {
	return &actionContext{
		Context: ctx,

		clock:            clock,
		eventLoopBlocker: eventLoopBlocker,
	}
}

func (a *actionContext) Clock() timestone.Clock {
	return a.clock
}

func (a *actionContext) DoneSchedulingNewActions() {
	if a.eventLoopBlocker == nil {
		return
	}

	a.eventLoopBlocker.Done()
}

func (a *actionContext) Value(key any) any {
	switch key {
	case timestone.ActionContextClockKey:
		return a.clock
	case ActionContextEventLoopBlockerKey:
		return a.eventLoopBlocker
	default:
		return a.Context.Value(key)
	}
}
