package simulation

import (
	"context"
	"github.com/metamogul/timestone/v2/simulation/config"

	"github.com/metamogul/timestone/v2/simulation/internal/clock"
	"github.com/metamogul/timestone/v2/simulation/internal/events"
	"github.com/metamogul/timestone/v2/simulation/internal/waitgroups"
	"sync"
	"time"

	"github.com/metamogul/timestone/v2"
)

type Scheduler struct {
	clock *clock.Clock

	eventQueue        *events.Queue
	eventGeneratorsMu sync.RWMutex

	eventConfigs *events.Configs

	eventWaitGroups *waitgroups.EventWaitGroups
}

// NewScheduler will return a newMatching Scheduler instance, with its
// clock initialized to return now.
func NewScheduler(now time.Time) *Scheduler {
	eventConfigs := events.NewConfigs()

	return &Scheduler{
		clock:           clock.NewClock(now),
		eventQueue:      events.NewQueue(eventConfigs),
		eventConfigs:    eventConfigs,
		eventWaitGroups: waitgroups.NewEventWaitGroups(),
	}
}

func (s *Scheduler) Now() time.Time {
	return s.clock.Now()
}

// ConfigureEvents provides an config.Config to configure one ore multiple events.
func (s *Scheduler) ConfigureEvents(configs ...config.Config) {
	for _, configuration := range configs {
		s.eventConfigs.Set(configuration)
	}
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

	s.execEvent(nextEvent)
}

// WaitFor is to be used after ForwardOne and blocks until all scheduled
// events embedding actions with the specified actionNames have finished.
func (s *Scheduler) WaitFor(events ...config.Event) {
	s.eventWaitGroups.WaitFor(events)
}

// Wait is to be used after ForwardOne and blocks until all scheduled
// events have finished.
func (s *Scheduler) Wait() {
	s.eventWaitGroups.Wait()
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
// with the Config.Priority passed through ConfigureEvents.
//
// Event s configured via their Config.WaitFor will only start
// execution once the specified events have finished.
//
// Event s configured via Config.Adds will block the run
// loop until the specified Generator instances have been passed to the
// Scheduler, either via one of the Perform... methods or via AddEventGenerators.
func (s *Scheduler) Forward(interval time.Duration) {
	targetTime := s.clock.Now().Add(interval)

	for s.execNextEvent(targetTime) {
	}

	s.eventWaitGroups.Wait()
}

func (s *Scheduler) execNextEvent(targetTime time.Time) (shouldContinue bool) {
	s.eventGeneratorsMu.RLock()

	if s.eventQueue.Finished() {
		s.clock.Set(targetTime)
		s.eventGeneratorsMu.RUnlock()
		return false
	}

	if s.eventQueue.Peek().After(targetTime) {
		s.clock.Set(targetTime)
		s.eventGeneratorsMu.RUnlock()
		return false
	}

	nextEvent := s.eventQueue.Pop()
	s.eventGeneratorsMu.RUnlock()

	s.execEvent(nextEvent)

	return true
}

func (s *Scheduler) execEvent(eventToExec *events.Event) {
	s.clock.Set(eventToExec.Time)

	blockingEvents := s.eventConfigs.BlockingEvents(eventToExec)
	expectedGenerators := s.eventConfigs.ExpectedGenerators(eventToExec)

	s.eventQueue.ExpectGenerators(expectedGenerators)

	eventWaitGroup := s.eventWaitGroups.New(eventToExec.Time, eventToExec.Tags())
	go func() {
		s.eventWaitGroups.WaitFor(blockingEvents)
		eventToExec.Perform(context.WithValue(eventToExec.Context, timestone.ActionContextClockKey, clock.NewClock(eventToExec.Time)))
		eventWaitGroup.Done()
	}()

	s.eventQueue.WaitForExpectedGenerators(expectedGenerators)
}

// PerformNow schedules action to be executed immediately, that is
// at the current time of the Scheduler's clock. It adds a newMatching Event
// generator which materializes a corresponding event to the Scheduler's
// event queue.
func (s *Scheduler) PerformNow(ctx context.Context, action timestone.Action, tags ...string) {
	s.AddEventGenerators(events.NewOnceGenerator(ctx, action, s.clock.Now(), tags))
}

// PerformAfter schedules an action to be run once after a delay
// of duration. It adds a newMatching Event  generator which materializes a
// corresponding event to the Scheduler's event queue.
func (s *Scheduler) PerformAfter(ctx context.Context, action timestone.Action, interval time.Duration, tags ...string) {
	s.AddEventGenerators(events.NewOnceGenerator(ctx, action, s.clock.Now().Add(interval), tags))
}

// PerformRepeatedly schedules an action to be run every interval
// after an initial delay of interval. If until is provided, the last
// event will be run before or at until. It adds a newMatching Event
// generator which materializes corresponding events to the Scheduler's
// event queue.
func (s *Scheduler) PerformRepeatedly(ctx context.Context, action timestone.Action, until *time.Time, interval time.Duration, tags ...string) {
	s.AddEventGenerators(events.NewPeriodicGenerator(ctx, action, s.clock.Now(), until, interval, tags))
}

// AddEventGenerators is used by the Perform... methods of the Scheduler.
// It can be used to pass a custom event generator if Timestone is used
// to run event-based simulations.
func (s *Scheduler) AddEventGenerators(generators ...events.Generator) {
	s.eventGeneratorsMu.Lock()
	defer s.eventGeneratorsMu.Unlock()

	for _, generator := range generators {
		if generator.Finished() {
			continue
		}

		s.eventQueue.Add(generator)
	}
}
