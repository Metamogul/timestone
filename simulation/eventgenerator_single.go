package simulation

import (
	"context"
	"time"

	"github.com/metamogul/timestone"
)

type singleEventGenerator struct {
	*Event
	ctx context.Context
}

func newSingleEventGenerator(ctx context.Context, action timestone.Action, time time.Time) *singleEventGenerator {
	return &singleEventGenerator{
		Event: NewEvent(ctx, action, time),
		ctx:   ctx,
	}
}

func (s *singleEventGenerator) Pop() *Event {
	if s.Finished() {
		panic(ErrEventGeneratorFinished)
	}

	defer func() { s.Event = nil }()

	return s.Event
}

func (s *singleEventGenerator) Peek() Event {
	if s.Finished() {
		panic(ErrEventGeneratorFinished)
	}

	return *s.Event
}

func (s *singleEventGenerator) Finished() bool {
	return s.Event == nil || s.ctx.Err() != nil
}
