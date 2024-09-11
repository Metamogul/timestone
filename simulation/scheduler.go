package simulation

import (
	"context"
	"fmt"
	"maps"
	"slices"
	"sync"
	"time"

	"github.com/metamogul/timestone"
)

type Scheduler struct {
	*clock

	eventQueue        *eventQueue
	eventGeneratorsMu sync.RWMutex

	eventConfigs *eventConfigurations

	finishedEventsWaitGroups *waitGroups
}

func NewScheduler(now time.Time) *Scheduler {
	eventConfigs := newEventConfigurations()

	return &Scheduler{
		clock:                    newClock(now),
		eventQueue:               newEventQueue(eventConfigs),
		eventConfigs:             eventConfigs,
		finishedEventsWaitGroups: newWaitGroups(),
	}
}

func (s *Scheduler) SetDefaultMode(mode ExecMode) {
	s.eventConfigs.defaultExecMode = mode
}

func (s *Scheduler) ConfigureEvent(actionName string, time *time.Time, config EventConfiguration) {
	s.eventConfigs.set(actionName, time, config)
}

func (s *Scheduler) ForwardOne() {
	s.eventGeneratorsMu.RLock()

	if s.eventQueue.Finished() {
		s.eventGeneratorsMu.RUnlock()
		return
	}

	nextEvent := s.eventQueue.Pop()
	s.eventGeneratorsMu.RUnlock()

	s.scheduleEvent(nextEvent)
}

func (s *Scheduler) WaitFor(actionNames ...string) {
	s.finishedEventsWaitGroups.waitFor(actionNames...)
}

func (s *Scheduler) Wait() {
	s.finishedEventsWaitGroups.wait()
}

func (s *Scheduler) Forward(interval time.Duration) {
	targetTime := s.clock.Now().Add(interval)

	for s.scheduleNextEvent(targetTime) {
	}

	s.finishedEventsWaitGroups.wait()
}

func (s *Scheduler) scheduleNextEvent(targetTime time.Time) (shouldContinue bool) {
	s.eventGeneratorsMu.RLock()

	if s.eventQueue.Finished() {
		s.clock.set(targetTime)
		s.eventGeneratorsMu.RUnlock()
		return false
	}

	if s.eventQueue.Peek().After(targetTime) {
		s.clock.set(targetTime)
		s.eventGeneratorsMu.RUnlock()
		return false
	}

	nextEvent := s.eventQueue.Pop()
	s.eventGeneratorsMu.RUnlock()

	s.scheduleEvent(nextEvent)

	return true
}

func (s *Scheduler) scheduleEvent(event *Event) {
	s.clock.set(event.Time)

	actionName := event.Name()
	execMode := s.eventConfigs.getExecMode(event)
	blockingActions := s.eventConfigs.getBlockingActions(event)
	wantedNewGenerators := s.eventConfigs.getWantedNewGenerators(event)

	for wantedActionName, wantedEventCount := range wantedNewGenerators {
		s.eventQueue.newGeneratorsWaitGroups.new(wantedActionName)
		s.eventQueue.newGeneratorsWaitGroups.add(wantedActionName, wantedEventCount)
	}

	switch execMode {

	case ExecModeAsync:
		s.finishedEventsWaitGroups.add(actionName, 1)
		go func() {
			s.finishedEventsWaitGroups.waitFor(blockingActions...)
			event.Perform(context.WithValue(event.Context, timestone.ActionContextClockKey, newClock(event.Time)))
			s.finishedEventsWaitGroups.done(actionName)
		}()

	case ExecModeSequential:
		s.finishedEventsWaitGroups.wait()
		event.Perform(context.WithValue(event.Context, timestone.ActionContextClockKey, newClock(event.Time)))

	default:
		panic(fmt.Sprintf("No schedule mode defined for action %s", event.Action.Name()))
	}

	wantedActionNames := slices.Collect(maps.Keys(wantedNewGenerators))
	s.eventQueue.newGeneratorsWaitGroups.waitFor(wantedActionNames...)
}

func (s *Scheduler) PerformNow(ctx context.Context, action timestone.Action) {
	s.AddEventGenerators(newSingleEventGenerator(ctx, action, s.now))
}

func (s *Scheduler) PerformAfter(ctx context.Context, action timestone.Action, interval time.Duration) {
	s.AddEventGenerators(newSingleEventGenerator(ctx, action, s.now.Add(interval)))
}

func (s *Scheduler) PerformRepeatedly(ctx context.Context, action timestone.Action, until *time.Time, interval time.Duration) {
	s.AddEventGenerators(newPeriodicEventGenerator(ctx, action, s.Now(), until, interval))
}

func (s *Scheduler) AddEventGenerators(generators ...EventGenerator) {
	s.eventGeneratorsMu.Lock()
	defer s.eventGeneratorsMu.Unlock()

	for _, generator := range generators {
		if generator.Finished() {
			continue
		}

		s.finishedEventsWaitGroups.new(generator.Peek().Name())
		s.eventQueue.add(generator)
	}
}
