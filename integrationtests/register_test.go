package integrationtests

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/metamogul/timestone"
	"github.com/metamogul/timestone/simulation"
	"github.com/stretchr/testify/require"
)

type register struct {
	counter int
}

type Action func(timestone.ActionContext)

func (a Action) Perform(ctx timestone.ActionContext) { a(ctx) }

func (a Action) Name() string { return "" }

func (r *register) incrementAfterOneMinute(scheduler timestone.Scheduler) {
	scheduler.PerformAfter(
		context.Background(),
		Action(func(timestone.ActionContext) {
			// Simulate execution time
			time.Sleep(100 * time.Millisecond)

			r.counter++
		}),
		time.Minute,
	)
}

func (r *register) incrementEveryMinute(scheduler timestone.Scheduler) {
	mu := sync.Mutex{}

	scheduler.PerformRepeatedly(
		context.Background(),
		Action(func(timestone.ActionContext) {
			mu.Lock()

			// Simulate execution time
			time.Sleep(10 * time.Millisecond)

			r.counter++

			mu.Unlock()
		}),
		nil,
		time.Minute,
	)
}

func Test_incrementAfterOneMinute(t *testing.T) {
	t.Parallel()

	now := time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC)

	scheduler := simulation.NewScheduler(now)
	scheduler.SetDefaultMode(simulation.ScheduleModeAsync)

	r := &register{}

	r.incrementAfterOneMinute(scheduler)

	scheduler.Forward(time.Minute * 60)
	require.Equal(t, 1, r.counter)
}

func Test_incrementEveryMinute(t *testing.T) {
	t.Parallel()

	now := time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC)

	scheduler := simulation.NewScheduler(now)
	scheduler.SetDefaultMode(simulation.ScheduleModeAsync)

	r := &register{}

	r.incrementEveryMinute(scheduler)

	scheduler.Forward(time.Minute * 60)
	require.Equal(t, 60, r.counter)
}
