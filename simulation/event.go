//go:generate go run github.com/vektra/mockery/v2@v2.43.2
package simulation

import (
	"context"
	"time"

	"github.com/metamogul/timestone"
)

type Event struct {
	timestone.Action
	time.Time
	context.Context
}

func NewEvent(ctx context.Context, action timestone.Action, time time.Time) *Event {
	if action == nil {
		panic("action can't be nil")
	}

	return &Event{
		Action:  action,
		Time:    time,
		Context: ctx,
	}
}
