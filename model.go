package timestone

import (
	"context"
	"time"
)

// Clock provides access to the current time and should be used inside
// actions instead of calling time.Now(). It is available as value in
// the context.Context inside an action under the key
// ActionContextClockKey
type Clock interface {
	// Now returns the current time.
	Now() time.Time
}

// ActionContextClockKey provides access to a Clock as value in the
// context.Context inside an Action.
const ActionContextClockKey = "timestone.ActionContextClock"

// An Action is a function to be scheduled by a Scheduler instance.
// It is identified by a name, e.g. for other Action s to wait for it.
type Action interface {
	// Perform executes the action. A clock is passed inside ctx at the
	// ActionContextClockKey.
	Perform(ctx context.Context)
}

// SimpleAction provides a reference implementation for Action that
// covers most use cases.
type SimpleAction func(context.Context)

// Perform implements Action and performs the func aliased by SimpleAction.
func (s SimpleAction) Perform(ctx context.Context) {
	s(ctx)
}

// Scheduler encapsulates the scheduling of Action s and should replace
// every use of goroutines to enable deterministic unit tests.
//
// The system.Scheduler implementation will use goroutines for scheduling
// using well established concurrency patterns. It is intended to be
// passed as the actual production dependency to all components that
// need to perform asynchronous Action s.
//
// The simulation.Scheduler implementation uses a configurable run loop
// instead. It is intended for use in unit tests, where you can use the
// simulation.Scheduler.ConfigureEvents method to provide various options
// that help the Scheduler to establish a deterministic and repeatable
// execution order of actions.
type Scheduler interface {
	// Clock embeds a clock that represents the current point in time
	// as events are being executed.
	Clock
	// PerformNow schedules action to be executed immediately, that is
	// at the current time of the Scheduler's clock.
	PerformNow(ctx context.Context, action Action, tags ...string)
	// PerformAfter schedules an action to be run once after a delay
	// of duration.
	PerformAfter(ctx context.Context, action Action, duration time.Duration, tags ...string)
	// PerformRepeatedly schedules an action to be run every interval
	// after an initial delay of interval. If until is provided, the last
	// event will be run before or at until.
	PerformRepeatedly(ctx context.Context, action Action, until *time.Time, interval time.Duration, tags ...string)
}
