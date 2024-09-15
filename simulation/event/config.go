package event

import "time"

// Key is used to target one or multiple events.
type Key struct {
	// Time is used to match only actions at the specific time. If you want
	// to match all actions with the given Tags, pass nil for Time.
	Time *time.Time
	// Tags to address events. An event will match if it has been at least
	// tagged with all entries in Tags.
	Tags []string
}

// GeneratorExpectation represents an expectation  for a number of
// event generators to be added to a simulation.Scheduler. It will block
// the simulation.Scheduler until it has been fulfilled.
type GeneratorExpectation struct {
	// Tags are used to identify the expected generators.
	Tags []string
	// How many generators are expected to be added.
	Count int
}

// Config is used to provide settings for events when
// being scheduled and executed in the simulation.Scheduler.
type Config struct {
	// Assign a Priority to define scheduling order in case of simultaneous
	// actions.
	Priority int
	// Delay the start of the execution of an action until the execution of
	// all WaitForEvents has been completed. This doesn't change the
	// behaviour for actions scheduled in ExecModeSequential which always
	// waits for every other scheduled action to complete execution. You can
	// set multiple groups of tags to target the respective actions.
	WaitForEvents []*Key
	// Signal the Scheduler that the configured event is supposed to set more
	// event generators to the schedulers queue. The key of the map is the
	// name of the events spawned by the generator, while the value is the
	// number of corresponding newMatching event generators the Scheduler will
	// expect to hold before continuing.
	AddsGenerators []*GeneratorExpectation
}
