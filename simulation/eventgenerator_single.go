package simulation

import (
	"context"
	"time"

	"github.com/metamogul/timestone"
)

type singleEventGenerator struct {
	event *Event
	ctx   context.Context
}

func newSingleEventGenerator(ctx context.Context, action timestone.Action, time time.Time) *singleEventGenerator {
	return &singleEventGenerator{
		event: NewEvent(ctx, action, time),
		ctx:   ctx,
	}
}

func (s *singleEventGenerator) Pop() *Event {
	if s.Finished() {
		panic(ErrEventGeneratorFinished)
	}

	defer func() { s.event = nil }()

	return s.event
}

func (s *singleEventGenerator) Peek() Event {
	if s.Finished() {
		panic(ErrEventGeneratorFinished)
	}

	return *s.event
}

func (s *singleEventGenerator) Finished() bool {
	return s.event == nil || s.ctx.Err() != nil
}
