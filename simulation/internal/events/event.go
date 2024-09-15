//go:generate go run github.com/vektra/mockery/v2@v2.43.2
package events

import (
	"context"
	"time"

	"github.com/metamogul/timestone"
)

const DefaultTag = "<default>"

type Event struct {
	timestone.Action
	time.Time

	context.Context

	tags []string
}

func NewEvent(ctx context.Context, action timestone.Action, time time.Time, tags []string) *Event {
	if action == nil {
		panic("action can't be nil")
	}

	if len(tags) == 0 {
		tags = []string{DefaultTag}
	}

	return &Event{
		Action:  action,
		Time:    time,
		Context: ctx,
		tags:    tags,
	}
}

func (e *Event) Tags() []string {
	return e.tags
}
