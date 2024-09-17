package events

import (
	"context"
	"time"

	"github.com/metamogul/timestone/v2"
)

type OnceGenerator struct {
	event *Event
	ctx   context.Context
}

func NewOnceGenerator(ctx context.Context, action timestone.Action, time time.Time, tags []string) *OnceGenerator {
	return &OnceGenerator{
		event: NewEvent(ctx, action, time, tags),
		ctx:   ctx,
	}
}

func (o *OnceGenerator) Pop() *Event {
	if o.Finished() {
		panic(ErrGeneratorFinished)
	}

	defer func() { o.event = nil }()

	return o.event
}

func (o *OnceGenerator) Peek() Event {
	if o.Finished() {
		panic(ErrGeneratorFinished)
	}

	return *o.event
}

func (o *OnceGenerator) Finished() bool {
	return o.event == nil || o.ctx.Err() != nil
}
