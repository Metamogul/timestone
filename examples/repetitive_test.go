package examples

import (
	"context"
	"fmt"
	c "github.com/metamogul/timestone/v2/simulation/config"
	"math/rand/v2"
	"sync"
	"testing"
	"time"

	"github.com/metamogul/timestone/v2"
	"github.com/metamogul/timestone/v2/simulation"
	"github.com/stretchr/testify/require"
)

const simulateWriteLoadMilliseconds = 30

type writer struct {
	result    string
	scheduler timestone.Scheduler

	countWriteOne int
	countWriteTwo int

	mu sync.Mutex
}

func (w *writer) writeOne(context.Context) {
	w.mu.Lock()
	defer w.mu.Unlock()
	time.Sleep(time.Duration(rand.Int64N(simulateWriteLoadMilliseconds)) * time.Millisecond)

	w.result += fmt.Sprintf("one%d ", w.countWriteOne)
	w.countWriteOne++
}

func (w *writer) writeTwo(context.Context) {
	w.mu.Lock()
	defer w.mu.Unlock()

	w.result += fmt.Sprintf("two%d ", w.countWriteTwo)
	w.countWriteTwo++
}

func (w *writer) run(ctx context.Context, writeInterval time.Duration) {
	w.scheduler.PerformRepeatedly(
		ctx, timestone.SimpleAction(w.writeOne), nil, writeInterval, "writeOne",
	)
	w.scheduler.PerformRepeatedly(
		ctx, timestone.SimpleAction(w.writeTwo), nil, writeInterval, "writeTwo",
	)
}

func TestNoRaceWriting_AutomaticOrder(t *testing.T) {
	t.Parallel()

	now := time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC)
	writeInterval := time.Minute

	scheduler := simulation.NewScheduler(now)
	scheduler.ConfigureEvents(
		c.Config{
			Tags:     []string{"writeOne"},
			Priority: 1,
			WaitFor: []c.Event{c.Before{
				Interval: -writeInterval,
				Tags:     []string{"writeTwo"},
			}},
		},
	)
	scheduler.ConfigureEvents(
		c.Config{
			Tags:     []string{"writeTwo"},
			Priority: 2,
			WaitFor: []c.Event{c.Before{
				Interval: 0,
				Tags:     []string{"writeOne"},
			}},
		},
	)

	w := writer{scheduler: scheduler}
	w.run(context.Background(), writeInterval)

	scheduler.Forward(6 * writeInterval)

	require.Equal(t, "one0 two0 one1 two1 one2 two2 one3 two3 one4 two4 one5 two5 ", w.result)
}

func TestNoRaceWriting_ManualOrder(t *testing.T) {
	t.Parallel()

	now := time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC)
	writeInterval := time.Minute

	scheduler := simulation.NewScheduler(now)

	scheduler.ConfigureEvents(c.Config{
		Tags:     []string{"writeOne"},
		Time:     now.Add(writeInterval),
		Priority: 1,
	})
	scheduler.ConfigureEvents(c.Config{
		Tags:     []string{"writeTwo"},
		Time:     now.Add(writeInterval),
		Priority: 2,
		WaitFor: []c.Event{c.At{
			Tags: []string{"writeOne"},
			Time: now.Add(writeInterval),
		}},
	})

	scheduler.ConfigureEvents(c.Config{
		Tags:     []string{"writeTwo"},
		Time:     now.Add(writeInterval * 2),
		Priority: 1,
		WaitFor: []c.Event{c.At{
			Tags: []string{"writeTwo"},
			Time: now.Add(writeInterval),
		}},
	})
	scheduler.ConfigureEvents(c.Config{
		Tags:     []string{"writeOne"},
		Time:     now.Add(writeInterval * 2),
		Priority: 2,
		WaitFor: []c.Event{c.At{
			Tags: []string{"writeTwo"},
			Time: now.Add(writeInterval * 2),
		}},
	})

	scheduler.ConfigureEvents(c.Config{
		Tags:     []string{"writeOne"},
		Time:     now.Add(writeInterval * 3),
		Priority: 1,
		WaitFor: []c.Event{c.At{
			Tags: []string{"writeOne"},
			Time: now.Add(writeInterval * 2),
		}},
	})
	scheduler.ConfigureEvents(c.Config{
		Tags:     []string{"writeTwo"},
		Time:     now.Add(writeInterval * 3),
		Priority: 2,
		WaitFor: []c.Event{c.At{
			Tags: []string{"writeOne"},
			Time: now.Add(writeInterval * 3),
		}},
	})

	scheduler.ConfigureEvents(c.Config{
		Tags:     []string{"writeOne"},
		Time:     now.Add(writeInterval * 4),
		Priority: 1,
		WaitFor: []c.Event{c.At{
			Tags: []string{"writeTwo"},
			Time: now.Add(writeInterval * 3),
		}},
	})
	scheduler.ConfigureEvents(c.Config{
		Tags:     []string{"writeTwo"},
		Time:     now.Add(writeInterval * 4),
		Priority: 2,
		WaitFor: []c.Event{c.At{
			Tags: []string{"writeOne"},
			Time: now.Add(writeInterval * 4),
		}},
	})

	scheduler.ConfigureEvents(c.Config{
		Tags:     []string{"writeOne"},
		Time:     now.Add(writeInterval * 5),
		Priority: 1,
		WaitFor: []c.Event{c.At{
			Tags: []string{"writeTwo"},
			Time: now.Add(writeInterval * 4),
		}},
	})
	scheduler.ConfigureEvents(c.Config{
		Tags:     []string{"writeTwo"},
		Time:     now.Add(writeInterval * 5),
		Priority: 2,
		WaitFor: []c.Event{c.At{
			Tags: []string{"writeOne"},
			Time: now.Add(writeInterval * 5),
		}},
	})

	scheduler.ConfigureEvents(c.Config{
		Tags:     []string{"writeOne"},
		Time:     now.Add(writeInterval * 6),
		Priority: 1,
		WaitFor: []c.Event{c.At{
			Tags: []string{"writeTwo"},
			Time: now.Add(writeInterval * 5),
		}},
	})
	scheduler.ConfigureEvents(c.Config{
		Tags:     []string{"writeTwo"},
		Time:     now.Add(writeInterval * 6),
		Priority: 2,
		WaitFor: []c.Event{c.At{
			Tags: []string{"writeOne"},
			Time: now.Add(writeInterval * 6),
		}},
	})

	w := writer{scheduler: scheduler}
	w.run(context.Background(), writeInterval)

	scheduler.Forward(6 * writeInterval)

	require.Equal(t, "one0 two0 two1 one1 one2 two2 one3 two3 one4 two4 one5 two5 ", w.result)
}

type timeWriter struct {
	scheduler timestone.Scheduler
	mu        sync.Mutex
}

func (w *timeWriter) writeTime(ctx context.Context) {
	w.mu.Lock()
	defer w.mu.Unlock()

	now := (ctx.Value(timestone.ActionContextClockKey)).(timestone.Clock).Now()
	time.Sleep(time.Duration(rand.Int64N(simulateWriteLoadMilliseconds)) * time.Millisecond)

	fmt.Printf("%v\n", now)
}

func (w *timeWriter) run(ctx context.Context, writeInterval time.Duration) {
	w.scheduler.PerformRepeatedly(
		ctx, timestone.SimpleAction(w.writeTime), nil, writeInterval, "writeTime",
	)
}

func ExampleNoRaceSelfWait() {
	now := time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC)
	writeInterval := time.Minute

	scheduler := simulation.NewScheduler(now)
	scheduler.ConfigureEvents(
		c.Config{
			Tags: []string{"writeTime"},
			WaitFor: []c.Event{c.Before{
				Interval: -writeInterval,
				Tags:     []string{"writeTime"},
			}},
		},
	)

	w := timeWriter{scheduler: scheduler}
	w.run(context.Background(), writeInterval)

	scheduler.Forward(6 * writeInterval)

	// Output:
	// 2024-01-01 12:01:00 +0000 UTC
	// 2024-01-01 12:02:00 +0000 UTC
	// 2024-01-01 12:03:00 +0000 UTC
	// 2024-01-01 12:04:00 +0000 UTC
	// 2024-01-01 12:05:00 +0000 UTC
	// 2024-01-01 12:06:00 +0000 UTC
}
