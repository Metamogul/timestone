package simulation

import "time"

type ScheduleMode int

const (
	ScheduleModeUndefined ScheduleMode = iota
	// ScheduleModeSequential configures an action to execute in the go routine the
	// scheduler was started in. The scheduler will wait for all
	// previously started actions to finish before performing a serial
	// action.
	ScheduleModeSequential
	// ScheduleModeAsync configures an action to execute in its own go routine.
	// Before execution the scheduler will wait for all actions to complete that
	// have been configured for the action. Use this mode if actions implement
	// their own syncing logic.
	ScheduleModeAsync
)

type RecursiveMode int

const (
	// Default assumes that an action will not schedule more actions.
	RecursiveModeDefault = iota
	// RecursiveModeWaitForActions makes the scheduler wait for more actions
	// being scheduled from inside an action. Call DoneSchedulingNewActions()
	// on the context inside the action to signal the scheduler to continue
	// with its event loop. This has no effect for actions scheduled in
	// ScheduleModeSequential.
	RecursiveModeWaitForActions
)

const (
	// Represents the default priority for an event if non has been
	// set.
	EventPriorityDefault = iota
)

type EventConfiguration struct {
	// Assign a ScheduleMode to control how an action is scheduled to be
	// executed. This will override the schedulers default behaviour if
	// a default behaviour has been configured.
	ScheduleMode ScheduleMode
	// Per default, the scheduler assumes that an action will not schedule
	// more actions. If an action attemps to schedule more actions and this
	// action was scheduled to be executed asynchronously, this can lead to
	// a situation where the event loop stops because the newly scheduled events
	// arriving from the scheduling action's goroutine haven't made it to the
	// buffer yet.
	// To prevent this, assign the RecursiveModeWaitForActions to any action,
	// that is scheduled asynchronously and will schedule more actions.
	RecursiveMode RecursiveMode
	// Assign a Priority to define scheduling order in case of simultaneous
	// actions.
	Priority int
	// Delay the execution of an action until all WaitForActions have been
	// completed. This doesn't change the behaviour for actions scheduled
	// in ScheduleModeSequential which always waits for every other action
	// to complete.
	WaitForActions []string
}

type nameAndTimeKey struct {
	actionName         string
	unixMilliTimestamp int64
}

type eventConfigurations struct {
	configsByName        map[string]*EventConfiguration
	configsByNameAndTime map[nameAndTimeKey]*EventConfiguration

	defaultScheduleMode ScheduleMode
}

func newEventConfigurations() *eventConfigurations {
	return &eventConfigurations{
		configsByName:        make(map[string]*EventConfiguration),
		configsByNameAndTime: make(map[nameAndTimeKey]*EventConfiguration),
	}
}

func (e *eventConfigurations) set(actionName string, time *time.Time, config EventConfiguration) {
	if time != nil {
		e.configsByNameAndTime[nameAndTimeKey{actionName, time.UnixMilli()}] = &config
		return
	}

	e.configsByName[actionName] = &config
}

func (e *eventConfigurations) get(event *Event) *EventConfiguration {
	if config, hasConfig := e.configsByNameAndTime[nameAndTimeKey{event.Name(), event.Time.UnixMilli()}]; hasConfig {
		return config
	}

	if config, hasConfig := e.configsByName[event.Name()]; hasConfig {
		return config
	}

	return nil
}

func (e *eventConfigurations) getScheduleMode(event *Event) ScheduleMode {
	if config := e.get(event); config != nil && config.ScheduleMode != ScheduleModeUndefined {
		return config.ScheduleMode
	}

	return e.defaultScheduleMode
}

func (e *eventConfigurations) getRecursiveMode(event *Event) RecursiveMode {
	if config := e.get(event); config != nil {
		return config.RecursiveMode
	}

	return RecursiveModeDefault
}

func (e *eventConfigurations) getPriority(event *Event) int {
	if config := e.get(event); config != nil {
		return config.Priority
	}

	return EventPriorityDefault
}

func (e *eventConfigurations) getBlockingActions(event *Event) []string {
	if config := e.get(event); config != nil {
		return config.WaitForActions
	}

	return []string{}
}
