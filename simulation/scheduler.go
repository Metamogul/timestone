package simulation

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/metamogul/timestone"
)

type Scheduler struct {
	*clock

	eventGenerators   *eventCombinator
	eventGeneratorsMu sync.RWMutex

	eventConfigs *eventConfigurations

	actionWaitGroups *waitGroups
}

func NewScheduler(now time.Time) *Scheduler {
	eventConfigs := newEventConfigurations()

	return &Scheduler{
		clock:            newClock(now),
		eventGenerators:  newEventCombinator(eventConfigs),
		eventConfigs:     eventConfigs,
		actionWaitGroups: newWaitGroups(),
	}
}

func (s *Scheduler) SetDefaultMode(mode ScheduleMode) {
	s.eventConfigs.defaultScheduleMode = mode
}

func (s *Scheduler) ConfigureEvent(actionName string, time *time.Time, config EventConfiguration) {
	s.eventConfigs.set(actionName, time, config)
}

func (s *Scheduler) ForwardOne() {
	s.eventGeneratorsMu.RLock()

	if s.eventGenerators.Finished() {
		s.eventGeneratorsMu.RUnlock()
		return
	}

	nextEvent := s.eventGenerators.Pop()
	s.eventGeneratorsMu.RUnlock()

	s.scheduleEvent(nextEvent)
}

func (s *Scheduler) WaitFor(actionNames ...string) {
	s.actionWaitGroups.waitFor(actionNames...)
}

func (s *Scheduler) Wait() {
	s.actionWaitGroups.wait()
}

func (s *Scheduler) Forward(interval time.Duration) {
	targetTime := s.clock.Now().Add(interval)

	for s.scheduleNextEvent(targetTime) {
	}

	s.actionWaitGroups.wait()
}

func (s *Scheduler) scheduleNextEvent(targetTime time.Time) (shouldContinue bool) {
	s.eventGeneratorsMu.RLock()

	if s.eventGenerators.Finished() {
		s.clock.set(targetTime)
		s.eventGeneratorsMu.RUnlock()
		return false
	}

	if s.eventGenerators.Peek().After(targetTime) {
		s.clock.set(targetTime)
		s.eventGeneratorsMu.RUnlock()
		return false
	}

	nextEvent := s.eventGenerators.Pop()
	s.eventGeneratorsMu.RUnlock()

	s.scheduleEvent(nextEvent)

	return true
}

func (s *Scheduler) scheduleEvent(event *Event) {
	s.clock.set(event.Time)

	actionName := event.Name()
	recursiveMode := s.eventConfigs.getRecursiveMode(event)
	scheduleMode := s.eventConfigs.getScheduleMode(event)
	blockingActions := s.eventConfigs.getBlockingActions(event)

	if scheduleMode == ScheduleModeAsync && recursiveMode == RecursiveModeWaitForActions {
		recursiveSchedulingBlocker := new(sync.WaitGroup)
		recursiveSchedulingBlocker.Add(1)

		s.actionWaitGroups.add(actionName, 1)
		go func() {
			defer s.actionWaitGroups.done(actionName)
			s.actionWaitGroups.waitFor(blockingActions...)
			event.Perform(newActionContext(event.Context, newClock(event.Time), recursiveSchedulingBlocker))
		}()

		recursiveSchedulingBlocker.Wait()
		return
	}

	if scheduleMode == ScheduleModeAsync {
		s.actionWaitGroups.add(actionName, 1)
		go func() {
			defer s.actionWaitGroups.done(actionName)
			s.actionWaitGroups.waitFor(blockingActions...)
			event.Perform(newActionContext(event.Context, newClock(event.Time), nil))
		}()
		return
	}

	if scheduleMode == ScheduleModeSequential {
		s.actionWaitGroups.wait()
		event.Perform(newActionContext(event.Context, newClock(event.Time), nil))
		return
	}

	panic(fmt.Sprintf("No schedule mode defined for action %s", event.Action.Name()))
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
		if !generator.Finished() {
			s.actionWaitGroups.new(generator.Peek().Name())
		}

		s.eventGenerators.add(generator)
	}
}
