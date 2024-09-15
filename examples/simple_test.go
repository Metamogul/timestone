package examples

import (
	"context"
	"github.com/metamogul/timestone/internal"
	"github.com/metamogul/timestone/simulation/event"
	"testing"
	"time"

	"github.com/metamogul/timestone"
	"github.com/metamogul/timestone/simulation"
	"github.com/stretchr/testify/require"
)

type writer struct {
	result    string
	scheduler timestone.Scheduler
}

func (w *writer) writeOne(context.Context) {
	w.result += "one "
}

func (w *writer) writeTwo(context.Context) {
	w.result += "two "
}

func (w *writer) run(ctx context.Context, repetitionInterval time.Duration) {
	w.scheduler.PerformRepeatedly(
		ctx, timestone.SimpleAction(w.writeOne), nil, repetitionInterval, "writeOne",
	)
	w.scheduler.PerformRepeatedly(
		ctx, timestone.SimpleAction(w.writeTwo), nil, repetitionInterval, "writeTwo",
	)
}

func TestNoRaceWriting(t *testing.T) {
	t.Parallel()

	now := time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC)
	writeInterval := time.Minute

	testcases := []struct {
		name               string
		configureScheduler func(*simulation.Scheduler)
		expectedResult     string
	}{
		{
			name: "one two one two",
			configureScheduler: func(s *simulation.Scheduler) {
				s.ConfigureEvent(
					event.Config{Priority: 1},
					internal.Ptr(now.Add(writeInterval)),
					"writeOne",
				)
				s.ConfigureEvent(
					event.Config{
						Priority: 2,
						WaitForEvents: []*event.Key{
							{
								Tags: []string{"writeOne"},
								Time: internal.Ptr(now.Add(writeInterval)),
							},
						},
					},
					internal.Ptr(now.Add(writeInterval)),
					"writeTwo",
				)
				s.ConfigureEvent(
					event.Config{
						Priority: 3,
						WaitForEvents: []*event.Key{
							{
								Tags: []string{"writeTwo"},
								Time: internal.Ptr(now.Add(writeInterval)),
							},
						},
					},
					internal.Ptr(now.Add(writeInterval*2)),
					"writeOne",
				)
				s.ConfigureEvent(
					event.Config{
						Priority: 4,
						WaitForEvents: []*event.Key{
							{
								Tags: []string{"writeOne"},
								Time: internal.Ptr(now.Add(writeInterval * 2)),
							},
						},
					},
					internal.Ptr(now.Add(writeInterval*2)),
					"writeTwo",
				)
			},
			expectedResult: "one two one two ",
		},
		{
			name: "one two two one",
			configureScheduler: func(s *simulation.Scheduler) {
				s.ConfigureEvent(
					event.Config{Priority: 1},
					internal.Ptr(now.Add(writeInterval)),
					"writeOne",
				)
				s.ConfigureEvent(
					event.Config{
						Priority: 2,
						WaitForEvents: []*event.Key{
							{
								Tags: []string{"writeOne"},
								Time: internal.Ptr(now.Add(writeInterval)),
							},
						},
					},
					internal.Ptr(now.Add(writeInterval)),
					"writeTwo",
				)
				s.ConfigureEvent(
					event.Config{
						Priority: 3,
						WaitForEvents: []*event.Key{
							{
								Tags: []string{"writeTwo"},
								Time: internal.Ptr(now.Add(writeInterval)),
							},
						},
					},
					internal.Ptr(now.Add(writeInterval*2)),
					"writeTwo",
				)
				s.ConfigureEvent(
					event.Config{
						Priority: 4,
						WaitForEvents: []*event.Key{
							{
								Tags: []string{"writeTwo"},
								Time: internal.Ptr(now.Add(writeInterval * 2)),
							},
						},
					},
					internal.Ptr(now.Add(writeInterval*2)),
					"writeOne",
				)
			},
			expectedResult: "one two two one ",
		},
	}

	for _, tt := range testcases {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			scheduler := simulation.NewScheduler(now)
			tt.configureScheduler(scheduler)

			w := writer{scheduler: scheduler}
			w.run(context.Background(), writeInterval)

			scheduler.Forward(2 * writeInterval)

			require.Equal(t, tt.expectedResult, w.result)
		})
	}
}
