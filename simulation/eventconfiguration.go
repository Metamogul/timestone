package simulation

import "time"

type ExecMode int

const (
	ExecModeUndefined ExecMode = iota
	// ExecModeSequential configures an action to execute in the go routine the
	// scheduler was started in. The scheduler will wait for all
	// previously started actions to finish before performing a serial
	// action.
	ExecModeSequential
	// ExecModeAsync configures an action to execute in its own go routine.
	// Before execution the scheduler will wait for all actions to complete that
	// have been configured for the action. Use this mode if actions implement
	// their own syncing logic.
	ExecModeAsync
)

const (
	// Represents the default priority for an event if non has been
	// set.
	EventPriorityDefault = iota
)

type EventConfiguration struct {
	// Assign a ExecMode to control how an action is executed. This will
	// override the schedulers default behaviour if a default behaviour
	// has been configured.
	ExecMode ExecMode
	// Assign a Priority to define scheduling order in case of simultaneous
	// actions.
	Priority int
	// Delay the start of the execution of an action until the execution of
	// all WaitForActions has been completed. This doesn't change the
	// behaviour for actions scheduled in ExecModeSequential which always
	// waits for every other scheduled action to complete execution.
	WaitForActions []string
	// Signal the Scheduler that the configured event is supposed to add more
	// event generators to the schedulers queue. The key of the map is the
	// name of the events spawned by the generator, while the value is the
	// number of corresponding new event generators the Scheduler will
	// expect to hold before continuing.
	WantsNewGenerators map[string]int
}

type nameAndTimeKey struct {
	actionName         string
	unixMilliTimestamp int64
}

type eventConfigurations struct {
	configsByName        map[string]*EventConfiguration
	configsByNameAndTime map[nameAndTimeKey]*EventConfiguration

	defaultExecMode ExecMode
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

func (e *eventConfigurations) getExecMode(event *Event) ExecMode {
	if config := e.get(event); config != nil && config.ExecMode != ExecModeUndefined {
		return config.ExecMode
	}

	return e.defaultExecMode
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

func (e *eventConfigurations) getWantedNewGenerators(event *Event) map[string]int {
	if config := e.get(event); config != nil {
		return config.WantsNewGenerators
	}

	return make(map[string]int)
}
