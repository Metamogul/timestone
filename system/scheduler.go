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

func (s *Scheduler) PerformNow(ctx context.Context, action timestone.Action) {
	go func() {
		select {
		case <-ctx.Done():
			return
		default:
			action.Perform(newActionContext(ctx, s.Clock))
		}
	}()
}

func (s *Scheduler) PerformAfter(ctx context.Context, action timestone.Action, duration time.Duration) {
	go func() {
		select {
		case <-time.After(duration):
			action.Perform(newActionContext(ctx, s.Clock))
		case <-ctx.Done():
			return
		}
	}()
}

func (s *Scheduler) PerformRepeatedly(ctx context.Context, action timestone.Action, until *time.Time, interval time.Duration) {
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
				action.Perform(newActionContext(ctx, s.Clock))
			case <-timer.C:
				return
			case <-ctx.Done():
				return
			}
		}
	}()
}
