package simulation

import (
	"context"
	"fmt"
	"slices"
	"sync"
	"testing"
	"time"

	"github.com/metamogul/timestone"
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
	require.Equal(t, now, newEventScheduler.Now())

	require.NotNil(t, newEventScheduler.eventQueue)
	require.NotNil(t, newEventScheduler.eventConfigs)
	require.NotNil(t, newEventScheduler.finishedEventsWaitGroups)
}

func TestScheduler_SetDefaultMode(t *testing.T) {
	t.Parallel()

	now := time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC)

	NewScheduler(now).SetDefaultMode(ExecModeAsync)
}

func TestScheduler_ConfigureEvent(t *testing.T) {
	t.Parallel()

	now := time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC)

	NewScheduler(now).ConfigureEvent("test", nil, EventConfiguration{})
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
	longRunningAction1.EXPECT().
		Name().
		Return("longRunningAction1").
		Maybe()

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
	longRunningAction2.EXPECT().
		Name().
		Return("longRunningAction2").
		Maybe()

	s := NewScheduler(now)
	s.SetDefaultMode(ExecModeAsync)
	s.PerformAfter(context.Background(), longRunningAction1, 1*time.Second)
	s.PerformAfter(context.Background(), longRunningAction2, 2*time.Second)

	s.ForwardOne()
	s.WaitFor("longRunningAction1")
	require.Len(t, eventTimes, 1)
	require.Equal(t, now.Add(1*time.Second), eventTimes[0])
	require.Equal(t, now.Add(1*time.Second), s.Now())

	s.ForwardOne()
	s.WaitFor("longRunningAction2")
	require.Len(t, eventTimes, 2)
	require.Equal(t, now.Add(2*time.Second), eventTimes[1])
	require.Equal(t, now.Add(2*time.Second), s.Now())
}

func TestScheduler_ForwardOne_Recursive(t *testing.T) {
	t.Parallel()

	now := time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC)

	testcases := []struct {
		name     string
		execMode ExecMode
	}{
		{
			name:     "sequential",
			execMode: ExecModeSequential,
		},
		{
			name:     "async",
			execMode: ExecModeAsync,
		},
	}

	for _, tt := range testcases {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

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
			innerAction.EXPECT().
				Name().
				Return("innerAction").
				Maybe()

			outerAction := timestone.NewMockAction(t)
			outerAction.EXPECT().
				Perform(mock.Anything).
				Run(func(ctx context.Context) {
					s.PerformAfter(ctx, innerAction, time.Second)

					mu.Lock()
					eventTimes = append(
						eventTimes,
						ctx.Value(timestone.ActionContextClockKey).(timestone.Clock).Now(),
					)
					mu.Unlock()
				}).
				Once()
			outerAction.EXPECT().
				Name().
				Return("outerAction").
				Maybe()

			s.SetDefaultMode(tt.execMode)
			s.PerformAfter(context.Background(), outerAction, 1*time.Second)
			s.ConfigureEvent("outerAction", nil, EventConfiguration{
				WantsNewGenerators: map[string]int{"innerAction": 1},
			})

			s.ForwardOne()
			s.WaitFor("outerAction")
			require.Len(t, eventTimes, 1)
			require.Equal(t, now.Add(1*time.Second), eventTimes[0])
			require.Equal(t, now.Add(1*time.Second), s.Now())

			s.ForwardOne()
			s.WaitFor("innerAction")
			require.Len(t, eventTimes, 2)
			require.Equal(t, now.Add(2*time.Second), eventTimes[1])
			require.Equal(t, now.Add(2*time.Second), s.Now())
		})
	}
}

func TestScheduler_WaitFor(t *testing.T) {
	t.Parallel()

	now := time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC)

	s := NewScheduler(now)
	s.finishedEventsWaitGroups.new("test1")
	s.finishedEventsWaitGroups.new("test2")
	s.finishedEventsWaitGroups.add("test1", 1)
	s.finishedEventsWaitGroups.add("test2", 1)
	go func() {
		s.finishedEventsWaitGroups.done("test1")
		s.finishedEventsWaitGroups.done("test2")
	}()

	s.finishedEventsWaitGroups.waitFor("test1", "test2")
}

func TestScheduler_Wait(t *testing.T) {
	t.Parallel()

	now := time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC)

	s := NewScheduler(now)

	for i := range 5 {
		testName := fmt.Sprintf("test%d", i)
		s.finishedEventsWaitGroups.new(testName)
		s.finishedEventsWaitGroups.add(testName, 1)
		go func() { s.finishedEventsWaitGroups.done(testName) }()
	}

	s.finishedEventsWaitGroups.wait()
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
			name:                   "all async, no syncing",
			fooActionScheduleAfter: 1 * time.Millisecond,
			barActionScheduleAfter: 2 * time.Millisecond,
			configureScheduler: func(s *Scheduler) {
				s.SetDefaultMode(ExecModeAsync)
			},
			wantResult: "barfoo",
			wantTimes: []time.Time{
				now.Add(2 * time.Millisecond),
				now.Add(1 * time.Millisecond),
			},
		},
		{
			name:                   "all async, wait for actions",
			fooActionScheduleAfter: 1 * time.Millisecond,
			barActionScheduleAfter: 2 * time.Millisecond,
			configureScheduler: func(s *Scheduler) {
				s.SetDefaultMode(ExecModeAsync)
				s.ConfigureEvent("barAction", nil, EventConfiguration{WaitForActions: []string{"fooAction"}})
			},
			wantResult: "foobar",
			wantTimes: []time.Time{
				now.Add(1 * time.Millisecond),
				now.Add(2 * time.Millisecond),
			},
		},
		{
			name:                   "mixed schedule mode",
			fooActionScheduleAfter: 1 * time.Millisecond,
			barActionScheduleAfter: 2 * time.Millisecond,
			configureScheduler: func(s *Scheduler) {
				s.SetDefaultMode(ExecModeAsync)
				s.ConfigureEvent("barAction", nil, EventConfiguration{ExecMode: ExecModeSequential})
			},
			wantResult: "foobar",
			wantTimes: []time.Time{
				now.Add(1 * time.Millisecond),
				now.Add(2 * time.Millisecond),
			},
		},
		{
			name:                   "all sequantial",
			fooActionScheduleAfter: 1 * time.Millisecond,
			barActionScheduleAfter: 2 * time.Millisecond,
			configureScheduler: func(s *Scheduler) {
				s.SetDefaultMode(ExecModeSequential)
			},
			wantResult: "foobar",
			wantTimes: []time.Time{
				now.Add(1 * time.Millisecond),
				now.Add(2 * time.Millisecond),
			},
		},
		{
			name:                   "all sequantial, resort simultaneous events",
			fooActionScheduleAfter: 1 * time.Millisecond,
			barActionScheduleAfter: 1 * time.Millisecond,
			configureScheduler: func(s *Scheduler) {
				s.SetDefaultMode(ExecModeSequential)
				s.ConfigureEvent("fooAction", nil, EventConfiguration{Priority: 2})
				s.ConfigureEvent("barAction", nil, EventConfiguration{Priority: 1})
			},
			wantResult: "foobar",
			wantTimes: []time.Time{
				now.Add(1 * time.Millisecond),
				now.Add(1 * time.Millisecond),
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
			fooAction.EXPECT().
				Name().
				Return("fooAction").
				Maybe()

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
			barAction.EXPECT().
				Name().
				Return("barAction").
				Maybe()

			s := NewScheduler(now)
			s.PerformAfter(context.Background(), fooAction, tt.fooActionScheduleAfter)
			s.PerformAfter(context.Background(), barAction, tt.barActionScheduleAfter)

			tt.configureScheduler(s)

			s.Forward(tt.barActionScheduleAfter + tt.fooActionScheduleAfter)

			require.Equal(t, tt.wantResult, result)
			require.Equal(t, tt.wantTimes, executionTimes)
		})
	}
}

func TestScheduler_Forward_Recursive(t *testing.T) {
	t.Parallel()

	now := time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC)

	testcases := []struct {
		name     string
		execMode ExecMode
	}{
		{
			name:     "sequential",
			execMode: ExecModeSequential,
		},
		{
			name:     "async",
			execMode: ExecModeAsync,
		},
	}

	for _, tt := range testcases {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

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
			innerAction.EXPECT().
				Name().
				Return("innerAction").
				Maybe()

			outerAction := timestone.NewMockAction(t)
			outerAction.EXPECT().
				Perform(mock.Anything).
				Run(func(ctx context.Context) {
					s.PerformAfter(context.Background(), innerAction, time.Second)
					executionTimes = append(
						executionTimes,
						ctx.Value(timestone.ActionContextClockKey).(timestone.Clock).Now(),
					)
				}).
				Once()
			outerAction.EXPECT().
				Name().
				Return("outerAction").
				Maybe()

			s.SetDefaultMode(tt.execMode)
			s.ConfigureEvent("outerAction", nil, EventConfiguration{
				WantsNewGenerators: map[string]int{"innerAction": 1},
			})

			s.PerformAfter(context.Background(), outerAction, time.Second)

			s.Forward(3 * time.Second)

			sorted := slices.IsSortedFunc(executionTimes, func(a, b time.Time) int {
				return a.Compare(b)
			})
			require.True(t, sorted)
		})
	}
}

func TestScheduler_scheduleNextEvent(t *testing.T) {
	t.Parallel()

	now := time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC)
	targetTime := now.Add(time.Minute)

	tests := []struct {
		name               string
		eventGenerators    func() []EventGenerator
		wantShouldContinue bool
	}{
		{
			name:            "all event generators finished",
			eventGenerators: func() []EventGenerator { return nil },
		},
		{
			name: "next event after target time",
			eventGenerators: func() []EventGenerator {
				mockAction := timestone.NewMockAction(t)
				mockAction.EXPECT().
					Name().
					Return("test").
					Maybe()
				return []EventGenerator{newSingleEventGenerator(context.Background(), mockAction, now.Add(1*time.Hour))}
			},
		},
		{
			name: "event dispatched successfully",
			eventGenerators: func() []EventGenerator {
				mockAction := timestone.NewMockAction(t)
				mockAction.EXPECT().
					Perform(mock.Anything).
					Once()
				mockAction.EXPECT().
					Name().
					Return("test").
					Maybe()
				return []EventGenerator{newSingleEventGenerator(context.Background(), mockAction, now.Add(1*time.Second))}
			},
			wantShouldContinue: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			eventConfigs := newEventConfigurations()
			eventQueue := newEventQueue(eventConfigs)
			for _, generator := range tt.eventGenerators() {
				eventQueue.add(generator)
			}

			s := &Scheduler{
				clock:                    newClock(now),
				eventQueue:               eventQueue,
				eventConfigs:             eventConfigs,
				finishedEventsWaitGroups: newWaitGroups(),
			}
			s.SetDefaultMode(ExecModeAsync)
			s.finishedEventsWaitGroups.new("test")

			if gotShouldContinue := s.scheduleNextEvent(targetTime); gotShouldContinue != tt.wantShouldContinue {
				t.Errorf("performNextEvent() = %v, want %v", gotShouldContinue, tt.wantShouldContinue)
			}
			s.finishedEventsWaitGroups.wait()

			if tt.wantShouldContinue == true {
				require.Equal(t, now.Add(time.Second), s.Now())
			} else {
				require.Equal(t, targetTime, s.Now())
			}
		})
	}
}

func TestScheduler_scheduleEvent(t *testing.T) {
	t.Parallel()

	now := time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC)

	t.Run("async, wants new generators", func(t *testing.T) {
		t.Parallel()

		s := NewScheduler(now)

		mockAction := timestone.NewMockAction(t)
		mockAction.EXPECT().
			Name().
			Return("test").
			Maybe()
		mockAction.EXPECT().
			Perform(mock.Anything).
			Run(func(context.Context) {
				mockAction := timestone.NewMockAction(t)
				mockAction.EXPECT().
					Name().
					Return("scheduledByTest").
					Maybe()
				s.PerformNow(context.Background(), mockAction)
			}).
			Once()

		event := NewEvent(context.Background(), mockAction, now.Add(time.Minute))
		eventConfig := EventConfiguration{
			ExecMode:           ExecModeAsync,
			WantsNewGenerators: map[string]int{"scheduledByTest": 1},
		}

		s.eventConfigs.set("test", nil, eventConfig)
		s.finishedEventsWaitGroups.new("test")

		s.scheduleEvent(event)
		s.Wait()

		require.Equal(t, now.Add(time.Minute), s.clock.Now())
	})

	t.Run("async, no new generators", func(t *testing.T) {
		t.Parallel()

		mockAction := timestone.NewMockAction(t)
		mockAction.EXPECT().
			Name().
			Return("test").
			Maybe()
		mockAction.EXPECT().
			Perform(mock.Anything).
			Once()

		event := NewEvent(context.Background(), mockAction, now.Add(time.Minute))
		eventConfig := EventConfiguration{
			ExecMode: ExecModeAsync,
		}

		s := NewScheduler(now)
		s.eventConfigs.set("test", nil, eventConfig)
		s.finishedEventsWaitGroups.new("test")

		s.scheduleEvent(event)
		s.Wait()

		require.Equal(t, now.Add(time.Minute), s.clock.Now())
	})

	t.Run("sequential, wants new generators", func(t *testing.T) {
		t.Parallel()

		s := NewScheduler(now)

		mockAction := timestone.NewMockAction(t)
		mockAction.EXPECT().
			Name().
			Return("test").
			Maybe()
		mockAction.EXPECT().
			Perform(mock.Anything).
			Run(func(context.Context) {
				mockAction := timestone.NewMockAction(t)
				mockAction.EXPECT().
					Name().
					Return("scheduledByTest").
					Maybe()
				s.PerformNow(context.Background(), mockAction)
			}).
			Once()

		event := NewEvent(context.Background(), mockAction, now.Add(time.Minute))
		eventConfig := EventConfiguration{
			ExecMode:           ExecModeSequential,
			WantsNewGenerators: map[string]int{"scheduledByTest": 1},
		}

		s.eventConfigs.set("test", nil, eventConfig)
		s.finishedEventsWaitGroups.new("test")

		s.scheduleEvent(event)
		s.Wait()

		require.Equal(t, now.Add(time.Minute), s.clock.Now())
	})

	t.Run("sequential, no new generators", func(t *testing.T) {
		t.Parallel()

		mockAction := timestone.NewMockAction(t)
		mockAction.EXPECT().
			Name().
			Return("test").
			Maybe()
		mockAction.EXPECT().
			Perform(mock.Anything).
			Once()

		event := NewEvent(context.Background(), mockAction, now.Add(time.Minute))
		eventConfig := EventConfiguration{
			ExecMode: ExecModeSequential,
		}

		s := NewScheduler(now)
		s.eventConfigs.set("test", nil, eventConfig)
		s.finishedEventsWaitGroups.new("test")

		s.scheduleEvent(event)
		s.Wait()

		require.Equal(t, now.Add(time.Minute), s.clock.Now())
	})
}

func TestScheduler_PerformNow(t *testing.T) {
	t.Parallel()

	now := time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC)

	s := NewScheduler(now)

	mockAction := timestone.NewMockAction(t)
	mockAction.EXPECT().
		Name().
		Return("mockAction").
		Maybe()

	s.PerformNow(context.Background(), mockAction)

	require.Len(t, s.finishedEventsWaitGroups.waitGroups, 1)
	require.NotNil(t, s.finishedEventsWaitGroups.waitGroups["mockAction"])
	require.Len(t, s.eventQueue.activeGenerators, 1)
	require.IsType(t, &singleEventGenerator{}, s.eventQueue.activeGenerators[0])
}

func TestScheduler_PerformAfter(t *testing.T) {
	t.Parallel()

	now := time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC)

	s := NewScheduler(now)

	mockAction := timestone.NewMockAction(t)
	mockAction.EXPECT().
		Name().
		Return("mockAction").
		Maybe()

	s.PerformAfter(context.Background(), mockAction, time.Second)

	require.Len(t, s.finishedEventsWaitGroups.waitGroups, 1)
	require.NotNil(t, s.finishedEventsWaitGroups.waitGroups["mockAction"])
	require.Len(t, s.eventQueue.activeGenerators, 1)
	require.IsType(t, &singleEventGenerator{}, s.eventQueue.activeGenerators[0])
}

func TestScheduler_PerformRepeatedly(t *testing.T) {
	t.Parallel()

	now := time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC)

	s := NewScheduler(now)

	mockAction := timestone.NewMockAction(t)
	mockAction.EXPECT().
		Name().
		Return("mockAction").
		Maybe()

	s.PerformRepeatedly(context.Background(), mockAction, nil, time.Second)

	require.Len(t, s.finishedEventsWaitGroups.waitGroups, 1)
	require.NotNil(t, s.finishedEventsWaitGroups.waitGroups["mockAction"])
	require.Len(t, s.eventQueue.activeGenerators, 1)
	require.IsType(t, &periodicEventGenerator{}, s.eventQueue.activeGenerators[0])
}

func TestScheduler_AddEventGenerators(t *testing.T) {
	t.Parallel()

	mockAction := timestone.NewMockAction(t)
	mockAction.EXPECT().
		Name().
		Return("mockAction").
		Maybe()

	mockEvent := NewEvent(context.Background(), mockAction, time.Time{})

	mockEventGenerator1 := NewMockEventGenerator(t)
	mockEventGenerator1.EXPECT().
		Peek().
		Return(*mockEvent).
		Maybe()
	mockEventGenerator1.EXPECT().
		Finished().
		Return(false).
		Maybe()

	mockEventGenerator2 := NewMockEventGenerator(t)
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

	require.Len(t, s.finishedEventsWaitGroups.waitGroups, 1)
	require.NotNil(t, s.finishedEventsWaitGroups.waitGroups["mockAction"])
	require.Len(t, s.eventQueue.activeGenerators, 2)
}
