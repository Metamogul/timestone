package system

import (
	"context"
	"github.com/metamogul/timestone/internal"
	"sync"
	"testing"
	"time"

	"github.com/metamogul/timestone"
)

func TestScheduler_PerformNow(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	clock := Clock{}

	wg := &sync.WaitGroup{}

	mockAction := timestone.NewMockAction(t)
	mockAction.EXPECT().
		Perform(context.WithValue(ctx, timestone.ActionContextClockKey, clock)).
		Run(func(context.Context) { wg.Done() }).
		Once()

	s := &Scheduler{Clock: clock}
	wg.Add(1)
	s.PerformNow(ctx, mockAction)
	wg.Wait()
}

func TestScheduler_PerformNow_cancelled(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	clock := Clock{}

	s := &Scheduler{Clock: clock}
	s.PerformNow(ctx, timestone.NewMockAction(t))
	time.Sleep(2 * time.Millisecond)
}

func TestScheduler_PerformAfter(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	clock := Clock{}

	wg := &sync.WaitGroup{}

	mockAction := timestone.NewMockAction(t)
	mockAction.EXPECT().
		Perform(context.WithValue(ctx, timestone.ActionContextClockKey, clock)).
		Run(func(context.Context) { wg.Done() }).
		Once()

	s := &Scheduler{Clock: clock}
	wg.Add(1)
	s.PerformAfter(ctx, mockAction, time.Millisecond)
	wg.Wait()
}

func TestScheduler_PerformAfter_cancelled(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	clock := Clock{}

	s := &Scheduler{Clock: clock}
	s.PerformAfter(ctx, timestone.NewMockAction(t), time.Millisecond)
	time.Sleep(2 * time.Millisecond)
}

func TestScheduler_PerformRepeatedly_until(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	clock := Clock{}

	wg := &sync.WaitGroup{}

	mockAction := timestone.NewMockAction(t)
	mockAction.EXPECT().
		Perform(context.WithValue(ctx, timestone.ActionContextClockKey, clock)).
		Run(func(context.Context) { wg.Done() }).
		Twice()

	s := &Scheduler{Clock: Clock{}}
	wg.Add(2)
	s.PerformRepeatedly(ctx, mockAction, internal.Ptr(clock.Now().Add(3*time.Millisecond)), time.Millisecond)
	wg.Wait()
}

func TestScheduler_PerformRepeatedly_indefinitely(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	clock := Clock{}

	mockAction := timestone.NewMockAction(t)
	mockAction.EXPECT().
		Perform(context.WithValue(ctx, timestone.ActionContextClockKey, clock)).
		Twice()

	s := &Scheduler{Clock: Clock{}}
	s.PerformRepeatedly(ctx, mockAction, nil, time.Millisecond)
	time.Sleep(3 * time.Millisecond)
}

func TestScheduler_PerformRepeatedly_cancelled(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	clock := Clock{}

	s := &Scheduler{Clock: Clock{}}
	s.PerformRepeatedly(ctx, timestone.NewMockAction(t), internal.Ptr(clock.Now().Add(3*time.Millisecond)), time.Millisecond)
	time.Sleep(2 * time.Millisecond)
}
