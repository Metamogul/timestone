package simulation

import (
	"context"
	"fmt"
	"github.com/metamogul/timestone/v2/simulation/config"
	"github.com/metamogul/timestone/v2/simulation/internal/clock"
	"github.com/metamogul/timestone/v2/simulation/internal/events"
	"github.com/metamogul/timestone/v2/simulation/internal/waitgroups"
	"slices"
	"sync"
	"testing"
	"time"

	"github.com/metamogul/timestone/v2"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestNewScheduler(t *testing.T) {
	t.Parallel()

	now := time.Now()

	newEventScheduler := NewScheduler(now)

	require.NotNil(t, newEventScheduler)
	require.IsType(t, &Scheduler{}, newEventScheduler)

	require.NotNil(t, newEventScheduler.clock)
	require.Equal(t, now, newEventScheduler.clock.Now())

	require.NotNil(t, newEventScheduler.eventQueue)
	require.NotNil(t, newEventScheduler.eventConfigs)
	require.NotNil(t, newEventScheduler.eventWaitGroups)
}

func TestScheduler_Now(t *testing.T) {
	t.Parallel()

	now := time.Now()
	s := NewScheduler(now)

	require.Equal(t, now, s.Now())
}

func TestScheduler_ConfigureEvent(t *testing.T) {
	t.Parallel()

	now := time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC)

	NewScheduler(now).ConfigureEvents(config.Config{Tags: []string{"test"}})
}

func TestScheduler_ForwardOne(t *testing.T) {
	t.Parallel()

	now := time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC)

	mu := sync.Mutex{}
	eventTimes := make([]time.Time, 0)

	longRunningAction1 := timestone.NewMockAction(t)
	longRunningAction1.EXPECT().
		Perform(mock.Anything).
		Run(func(ctx context.Context) {
			time.Sleep(100 * time.Millisecond)

			mu.Lock()
			eventTimes = append(eventTimes, ctx.Value(timestone.ActionContextClockKey).(timestone.Clock).Now())
			mu.Unlock()
		}).
		Once()

	longRunningAction2 := timestone.NewMockAction(t)
	longRunningAction2.EXPECT().
		Perform(mock.Anything).
		Run(func(ctx context.Context) {
			time.Sleep(50 * time.Millisecond)

			mu.Lock()
			eventTimes = append(eventTimes, ctx.Value(timestone.ActionContextClockKey).(timestone.Clock).Now())
			mu.Unlock()
		}).
		Once()

	s := NewScheduler(now)
	s.PerformAfter(context.Background(), longRunningAction1, 1*time.Second, "longRunningAction1")
	s.PerformAfter(context.Background(), longRunningAction2, 2*time.Second, "longRunningAction2")

	s.ForwardOne()
	s.WaitFor(config.All{Tags: []string{"longRunningAction1"}})
	require.Len(t, eventTimes, 1)
	require.Equal(t, now.Add(1*time.Second), eventTimes[0])
	require.Equal(t, now.Add(1*time.Second), s.clock.Now())

	s.ForwardOne()
	s.WaitFor(config.All{Tags: []string{"longRunningAction2"}})
	require.Len(t, eventTimes, 2)
	require.Equal(t, now.Add(2*time.Second), eventTimes[1])
	require.Equal(t, now.Add(2*time.Second), s.clock.Now())
}

func TestScheduler_ForwardOne_Recursive(t *testing.T) {
	t.Parallel()

	now := time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC)

	mu := sync.Mutex{}
	eventTimes := make([]time.Time, 0)

	s := NewScheduler(now)

	innerAction := timestone.NewMockAction(t)
	innerAction.EXPECT().
		Perform(mock.Anything).
		Run(func(ctx context.Context) {
			mu.Lock()
			eventTimes = append(
				eventTimes,
				ctx.Value(timestone.ActionContextClockKey).(timestone.Clock).Now(),
			)
			mu.Unlock()
		}).
		Once()

	outerAction := timestone.NewMockAction(t)
	outerAction.EXPECT().
		Perform(mock.Anything).
		Run(func(ctx context.Context) {
			s.PerformAfter(ctx, innerAction, time.Second, "innerAction")

			mu.Lock()
			eventTimes = append(
				eventTimes,
				ctx.Value(timestone.ActionContextClockKey).(timestone.Clock).Now(),
			)
			mu.Unlock()
		}).
		Once()

	s.PerformAfter(context.Background(), outerAction, 1*time.Second, "outerAction")
	s.ConfigureEvents(config.Config{
		Tags: []string{"outerAction"},
		Adds: []*config.Generator{{Tags: []string{"innerAction"}, Count: 1}},
	})

	s.ForwardOne()
	s.WaitFor(config.All{Tags: []string{"outerAction"}})
	require.Len(t, eventTimes, 1)
	require.Equal(t, now.Add(1*time.Second), eventTimes[0])
	require.Equal(t, now.Add(1*time.Second), s.clock.Now())

	s.ForwardOne()
	s.WaitFor(config.All{Tags: []string{"innerAction"}})
	require.Len(t, eventTimes, 2)
	require.Equal(t, now.Add(2*time.Second), eventTimes[1])
	require.Equal(t, now.Add(2*time.Second), s.clock.Now())
}

func TestScheduler_WaitFor(t *testing.T) {
	t.Parallel()

	now := time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC)

	s := NewScheduler(now)

	wg1 := s.eventWaitGroups.New(now, []string{"test1", "group"})
	wg2 := s.eventWaitGroups.New(now, []string{"test2", "group"})
	go func() {
		wg1.Done()
		wg2.Done()
	}()

	s.eventWaitGroups.WaitFor([]config.Event{config.All{Tags: []string{"group"}}})
}

func TestScheduler_Wait(t *testing.T) {
	t.Parallel()

	now := time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC)

	s := NewScheduler(now)

	for i := range 5 {
		testName := fmt.Sprintf("test%d", i)
		wg := s.eventWaitGroups.New(now, []string{testName})
		go func() { wg.Done() }()
	}

	s.eventWaitGroups.Wait()
}

func TestScheduler_Forward(t *testing.T) {
	t.Parallel()

	now := time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC)

	const fooActionSimulateLoad = 500 * time.Millisecond
	const barActionSimulateLoad = 50 * time.Millisecond

	testcases := []struct {
		name                   string
		fooActionScheduleAfter time.Duration
		barActionScheduleAfter time.Duration
		configureScheduler     func(s *Scheduler)
		wantResult             string
		wantTimes              []time.Time
	}{
		{
			name:                   "no syncing",
			fooActionScheduleAfter: 1 * time.Millisecond,
			barActionScheduleAfter: 2 * time.Millisecond,
			wantResult:             "barfoo",
			wantTimes: []time.Time{
				now.Add(2 * time.Millisecond),
				now.Add(1 * time.Millisecond),
			},
		},
		{
			name:                   "wait for actions",
			fooActionScheduleAfter: 1 * time.Millisecond,
			barActionScheduleAfter: 2 * time.Millisecond,
			configureScheduler: func(s *Scheduler) {
				s.ConfigureEvents(config.Config{
					Tags:    []string{"barAction"},
					WaitFor: []config.Event{config.All{Tags: []string{"fooAction"}}},
				})
			},
			wantResult: "foobar",
			wantTimes: []time.Time{
				now.Add(1 * time.Millisecond),
				now.Add(2 * time.Millisecond),
			},
		},
	}

	for _, tt := range testcases {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result := ""
			executionTimes := make([]time.Time, 0)

			fooAction := timestone.NewMockAction(t)
			fooAction.EXPECT().
				Perform(mock.Anything).
				Run(func(ctx context.Context) {
					time.Sleep(fooActionSimulateLoad)
					result += "foo"
					executionTimes = append(
						executionTimes, ctx.Value(timestone.ActionContextClockKey).(timestone.Clock).Now(),
					)
				}).
				Once()

			barAction := timestone.NewMockAction(t)
			barAction.EXPECT().
				Perform(mock.Anything).
				Run(func(ctx context.Context) {
					time.Sleep(barActionSimulateLoad)
					result += "bar"
					executionTimes = append(
						executionTimes, ctx.Value(timestone.ActionContextClockKey).(timestone.Clock).Now(),
					)
				}).
				Once()

			s := NewScheduler(now)
			s.PerformAfter(context.Background(), fooAction, tt.fooActionScheduleAfter, "fooAction")
			s.PerformAfter(context.Background(), barAction, tt.barActionScheduleAfter, "barAction")

			if tt.configureScheduler != nil {
				tt.configureScheduler(s)
			}

			s.Forward(tt.barActionScheduleAfter + tt.fooActionScheduleAfter)

			require.Equal(t, tt.wantResult, result)
			require.Equal(t, tt.wantTimes, executionTimes)
		})
	}
}

func TestScheduler_Forward_Recursive(t *testing.T) {
	t.Parallel()

	now := time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC)

	executionTimes := make([]time.Time, 0)

	s := NewScheduler(now)

	innerAction := timestone.NewMockAction(t)
	innerAction.EXPECT().
		Perform(mock.Anything).
		Run(func(ctx context.Context) {
			executionTimes = append(
				executionTimes,
				ctx.Value(timestone.ActionContextClockKey).(timestone.Clock).Now(),
			)
		}).
		Once()

	outerAction := timestone.NewMockAction(t)
	outerAction.EXPECT().
		Perform(mock.Anything).
		Run(func(ctx context.Context) {
			s.PerformAfter(context.Background(), innerAction, time.Second, "innerAction")
			executionTimes = append(
				executionTimes,
				ctx.Value(timestone.ActionContextClockKey).(timestone.Clock).Now(),
			)
		}).
		Once()

	s.ConfigureEvents(config.Config{
		Tags: []string{"outerAction"},
		Adds: []*config.Generator{{[]string{"innerAction"}, 1}},
	})

	s.PerformAfter(context.Background(), outerAction, time.Second, "outerAction")

	s.Forward(3 * time.Second)

	sorted := slices.IsSortedFunc(executionTimes, func(a, b time.Time) int {
		return a.Compare(b)
	})
	require.True(t, sorted)

}

func TestScheduler_execNextEvent(t *testing.T) {
	t.Parallel()

	now := time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC)
	targetTime := now.Add(time.Minute)

	tests := []struct {
		name               string
		eventGenerators    func() []events.Generator
		wantShouldContinue bool
	}{
		{
			name:            "all event generators finished",
			eventGenerators: func() []events.Generator { return nil },
		},
		{
			name: "next event after target time",
			eventGenerators: func() []events.Generator {
				mockAction := timestone.NewMockAction(t)
				return []events.Generator{events.NewOnceGenerator(context.Background(), mockAction, now.Add(1*time.Hour), []string{"test"})}
			},
		},
		{
			name: "event dispatched successfully",
			eventGenerators: func() []events.Generator {
				mockAction := timestone.NewMockAction(t)
				mockAction.EXPECT().
					Perform(mock.Anything).
					Once()
				return []events.Generator{events.NewOnceGenerator(context.Background(), mockAction, now.Add(1*time.Second), []string{"test"})}
			},
			wantShouldContinue: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			eventConfigs := events.NewConfigs()
			eventQueue := events.NewQueue(eventConfigs)
			for _, generator := range tt.eventGenerators() {
				eventQueue.Add(generator)
			}

			s := &Scheduler{
				clock:           clock.NewClock(now),
				eventQueue:      eventQueue,
				eventConfigs:    eventConfigs,
				eventWaitGroups: waitgroups.NewEventWaitGroups(),
			}

			if gotShouldContinue := s.execNextEvent(targetTime); gotShouldContinue != tt.wantShouldContinue {
				t.Errorf("performNextEvent() = %v, want %v", gotShouldContinue, tt.wantShouldContinue)
			}
			s.eventWaitGroups.Wait()

			if tt.wantShouldContinue == true {
				require.Equal(t, now.Add(time.Second), s.clock.Now())
			} else {
				require.Equal(t, targetTime, s.clock.Now())
			}
		})
	}
}

func TestScheduler_execEvent(t *testing.T) {
	t.Parallel()

	now := time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC)

	t.Run("wants newMatching generators", func(t *testing.T) {
		t.Parallel()

		s := NewScheduler(now)

		mockAction := timestone.NewMockAction(t)
		mockAction.EXPECT().
			Perform(mock.Anything).
			Run(func(context.Context) {
				s.PerformNow(context.Background(), timestone.NewMockAction(t), "scheduledByTest")
			}).
			Once()

		eventToExec := events.NewEvent(context.Background(), mockAction, now.Add(time.Minute), []string{"test"})
		eventConfig := config.Config{
			Tags: []string{"test"},
			Adds: []*config.Generator{{[]string{"scheduledByTest"}, 1}},
		}

		s.eventConfigs.Set(eventConfig)

		s.execEvent(eventToExec)
		s.WaitFor(config.All{Tags: []string{"test"}})

		require.Equal(t, now.Add(time.Minute), s.clock.Now())
	})

	t.Run("no newMatching generators", func(t *testing.T) {
		t.Parallel()

		mockAction := timestone.NewMockAction(t)
		mockAction.EXPECT().
			Perform(mock.Anything).
			Once()

		eventToExec := events.NewEvent(context.Background(), mockAction, now.Add(time.Minute), []string{"test"})

		s := NewScheduler(now)

		s.execEvent(eventToExec)
		s.WaitFor(config.All{Tags: []string{"test"}})

		require.Equal(t, now.Add(time.Minute), s.clock.Now())
	})
}

func TestScheduler_PerformNow(t *testing.T) {
	t.Parallel()

	now := time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC)

	s := NewScheduler(now)

	s.PerformNow(context.Background(), timestone.NewMockAction(t), "mockAction")

	require.False(t, s.eventQueue.Finished())
}

func TestScheduler_PerformAfter(t *testing.T) {
	t.Parallel()

	now := time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC)

	s := NewScheduler(now)

	s.PerformAfter(context.Background(), timestone.NewMockAction(t), time.Second, "mockAction")

	require.False(t, s.eventQueue.Finished())
}

func TestScheduler_PerformRepeatedly(t *testing.T) {
	t.Parallel()

	now := time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC)

	s := NewScheduler(now)

	s.PerformRepeatedly(context.Background(), timestone.NewMockAction(t), nil, time.Second, "mockAction")

	require.False(t, s.eventQueue.Finished())
}

func TestScheduler_AddEventGenerators(t *testing.T) {
	t.Parallel()

	mockEvent := events.NewEvent(context.Background(), timestone.NewMockAction(t), time.Time{}, []string{"mockAction"})

	mockEventGenerator1 := events.NewMockGenerator(t)
	mockEventGenerator1.EXPECT().
		Peek().
		Return(*mockEvent).
		Maybe()
	mockEventGenerator1.EXPECT().
		Finished().
		Return(false).
		Maybe()

	mockEventGenerator2 := events.NewMockGenerator(t)
	mockEventGenerator2.EXPECT().
		Peek().
		Return(*mockEvent).
		Maybe()
	mockEventGenerator2.EXPECT().
		Finished().
		Return(false).
		Maybe()

	now := time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC)

	s := NewScheduler(now)

	s.AddEventGenerators(mockEventGenerator1, mockEventGenerator2)

	require.False(t, s.eventQueue.Finished())
}
