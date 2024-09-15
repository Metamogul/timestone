package system

import (
	"context"
	"time"

	"github.com/metamogul/timestone"
)

type Clock struct{}

func (c Clock) Now() time.Time {
	return time.Now()
}

type Scheduler struct {
	Clock
}

func (s *Scheduler) PerformNow(ctx context.Context, action timestone.Action, _ ...string) {
	go func() {
		select {
		case <-ctx.Done():
			return
		default:
			action.Perform(context.WithValue(ctx, timestone.ActionContextClockKey, s.Clock))
		}
	}()
}

func (s *Scheduler) PerformAfter(ctx context.Context, action timestone.Action, duration time.Duration, _ ...string) {
	go func() {
		select {
		case <-time.After(duration):
			action.Perform(context.WithValue(ctx, timestone.ActionContextClockKey, s.Clock))
		case <-ctx.Done():
			return
		}
	}()
}

func (s *Scheduler) PerformRepeatedly(ctx context.Context, action timestone.Action, until *time.Time, interval time.Duration, _ ...string) {
	ticker := time.NewTicker(interval)

	var timer *time.Timer
	if until != nil {
		timer = time.NewTimer(until.Sub(s.Now()))
	} else {
		timer = &time.Timer{}
	}

	go func() {
		for {
			select {
			case <-ticker.C:
				action.Perform(context.WithValue(ctx, timestone.ActionContextClockKey, s.Clock))
			case <-timer.C:
				return
			case <-ctx.Done():
				return
			}
		}
	}()
}
