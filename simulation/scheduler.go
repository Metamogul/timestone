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

// NewScheduler will return a new Scheduler instance, with its
// clock initialized to return now.
func NewScheduler(now time.Time) *Scheduler {
	eventConfigs := newEventConfigurations()

	return &Scheduler{
		clock:                    newClock(now),
		eventQueue:               newEventQueue(eventConfigs),
		eventConfigs:             eventConfigs,
		finishedEventsWaitGroups: newWaitGroups(),
	}
}

// SetDefaultMode will set a default ExecMode that is applied
// to all events at execution time that don't have their ExecMode
// individually configured via ConfigureEvent. If no default ExecMode
// is provided, the Scheduler will panic once it reaches an Event that
// doesn't have an EventConfiguration with a ScheduleMode set to something
// else than ExecModeUndefined.
func (s *Scheduler) SetDefaultMode(mode ExecMode) {
	s.eventConfigs.defaultExecMode = mode
}

// ConfigureEvent provides an EventConfiguration for either a single event,
// identified by the name of its embedded action and the time of its occurrence,
// or for every event matching the actionName if no time is provided.
func (s *Scheduler) ConfigureEvent(actionName string, time *time.Time, config EventConfiguration) {
	s.eventConfigs.set(actionName, time, config)
}

// ForwardOne executes just the next event that is scheduled on the
// event queue of the Scheduler, and sets the timestone.Clock of the Scheduler
// to the time of the event.
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

// WaitFor is to be used after ForwardOne and blocks until all scheduled
// events embedding actions with the specified actionNames have finished.
func (s *Scheduler) WaitFor(actionNames ...string) {
	s.finishedEventsWaitGroups.waitFor(actionNames...)
}

// Wait is to be used after ForwardOne and blocks until all scheduled
// events have finished.
func (s *Scheduler) Wait() {
	s.finishedEventsWaitGroups.wait()
}

// Forward will forward the Scheduler.Clock while running all events to
// occur until Scheduler.Clock.Now() + interval. Each action will receive in
// its context.Context a timestone.Clock set to return the respective execution
// time for Now().
//
// Depending on their individual configuration, Event s will either be run
// sequentially, waiting for all preciously started Event s to finish, or
// asynchronously.
//
// Event s will be materialized and executed from the schedulers event queue in
// temporal order. In case of simultaneousness, the exection order can be changed
// with the EventConfiguration.Priority passed through ConfigureEvent.
//
// Event s configured via their EventConfiguration.WaitForActions will only start
// execution once the specified events have finished.
//
// Event s configured via EventConfiguration.WantsNewGenerators will block the run
// loop until the specified EventGenerator instances have been passed to the
// Scheduler, either via one of the Perform... methods or via AddEventGenerators.
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

// PerformNow schedules action to be executed immediately, that is
// at the current time of the Scheduler's clock. It adds a new Event
// generator which materializes a corresponding event to the Scheduler's
// event queue.
func (s *Scheduler) PerformNow(ctx context.Context, action timestone.Action) {
	s.AddEventGenerators(newSingleEventGenerator(ctx, action, s.now))
}

// PerformAfter schedules an action to be run once after a delay
// of duration. It adds a new Event  generator which materializes a
// corresponding event to the Scheduler's event queue.
func (s *Scheduler) PerformAfter(ctx context.Context, action timestone.Action, interval time.Duration) {
	s.AddEventGenerators(newSingleEventGenerator(ctx, action, s.now.Add(interval)))
}

// PerformRepeatedly schedules an action to be run every interval
// after an initial delay of interval. If until is provided, the last
// event will be run before or at until. It adds a new Event
// generator which materializes corresponding events to the Scheduler's
// event queue.
func (s *Scheduler) PerformRepeatedly(ctx context.Context, action timestone.Action, until *time.Time, interval time.Duration) {
	s.AddEventGenerators(newPeriodicEventGenerator(ctx, action, s.Now(), until, interval))
}

// AddEventGenerators is used by the Perform... methods of the Scheduler.
// It can be used to pass a custom event generator if Timestone is used
// to run event-based simulations.
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
