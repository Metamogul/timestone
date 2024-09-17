package events

import (
	"context"
	"time"

	"github.com/metamogul/timestone/v2"
)

type PeriodicGenerator struct {
	action   timestone.Action
	from     time.Time
	to       *time.Time
	interval time.Duration

	tags []string

	nextEvent *Event

	ctx context.Context
}

func NewPeriodicGenerator(
	ctx context.Context,
	action timestone.Action,
	from time.Time,
	to *time.Time,
	interval time.Duration,
	tags []string,
) *PeriodicGenerator {
	if action == nil {
		panic("Action can't be nil")
	}

	if to != nil && !to.After(from) {
		panic("to must be after from")
	}

	if interval == 0 {
		panic("interval must be greater than zero")
	}

	if to != nil && interval >= to.Sub(from) {
		panic("interval must be shorter than timespan given by from and to")
	}

	firstEvent := NewEvent(ctx, action, from.Add(interval), tags)

	return &PeriodicGenerator{
		action:   action,
		from:     from,
		to:       to,
		interval: interval,

		tags: tags,

		nextEvent: firstEvent,

		ctx: ctx,
	}
}

func (p *PeriodicGenerator) Pop() *Event {
	if p.Finished() {
		panic(ErrGeneratorFinished)
	}

	defer func() { p.nextEvent = NewEvent(p.ctx, p.action, p.nextEvent.Time.Add(p.interval), p.tags) }()

	return p.nextEvent
}

func (p *PeriodicGenerator) Peek() Event {
	if p.Finished() {
		panic(ErrGeneratorFinished)
	}

	return *p.nextEvent
}

func (p *PeriodicGenerator) Finished() bool {
	if p.ctx.Err() != nil {
		return true
	}

	if p.to == nil {
		return false
	}

	return p.nextEvent.Add(p.interval).After(*p.to)
}
