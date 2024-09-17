package config

import "time"

// Event is used to target one or multiple events.
type Event interface {
	GetTags() []string
}

// All is a Event targeting all events with Tags.
type All struct {
	// Tags to address events. An event will match if it has been at least
	// tagged with all entries in Tags.
	Tags []string
}

func (a All) GetTags() []string { return a.Tags }

// At is a key that will target at Time by Tags.
type At struct {
	// Time is used to match only actions at the specific time. If you want
	// to match all actions with the given Tags, pass nil for Time.
	//
	// If no matching event to wait for is found the scheduler will panic.
	Time time.Time
	// Tags to address events. An event will match if it has been at least
	// tagged with all entries in Tags.
	Tags []string
}

func (a At) GetTags() []string { return a.Tags }

type Before struct {
	// Before will match an event relative to the event that is configured.
	//
	// Unlike At, where a missing match will result in a panic, a missing
	// match from the Before Event will be silently ignored
	Interval time.Duration
	// Tags to address events. An event will match if it has been at least
	// tagged with all entries in Tags.
	Tags []string
}

func (r Before) GetTags() []string { return r.Tags }

// Generator represents an expectation  for a number of
// event generators to be added to a simulation.Scheduler. It will block
// the simulation.Scheduler until the expectation has been fulfilled.
type Generator struct {
	// Tags are used to identify the expected generators.
	Tags []string
	// How many generators are expected to be added.
	Count int
}

// Config is used to provide settings for events when
// being scheduled and executed in the simulation.Scheduler.
type Config struct {
	// Tags to address events to configure. An event will match if it has
	// been at least tagged with all entries in Tags.
	Tags []string
	// Time is optional. If set, the Config will match specifically events
	// at the given Time.
	Time time.Time
	// Assign a Priority to define scheduling order in case of simultaneous
	// actions.
	Priority int
	// Delay the start of the execution of an action until the execution of
	// all WaitFor has been completed. This doesn't change the
	// behaviour for actions scheduled in ExecModeSequential which always
	// waits for every other scheduled action to complete execution. You can
	// set multiple groups of tags to target the respective actions.
	WaitFor []Event
	// Signal the Scheduler that the configured event is supposed to set more
	// event generators to the schedulers queue. The key of the map is the
	// name of the events spawned by the generator, while the value is the
	// number of corresponding newMatching event generators the Scheduler will
	// expect to hold before continuing.
	Adds []*Generator
}
