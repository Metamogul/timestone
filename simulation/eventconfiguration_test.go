package simulation

import (
	"context"
	"testing"
	"time"

	"github.com/metamogul/timestone"
	"github.com/stretchr/testify/require"
)

func Test_newEventConfigurations(t *testing.T) {
	t.Parallel()

	newEventConfigurations := newEventConfigurations()

	require.NotNil(t, newEventConfigurations)
	require.NotNil(t, newEventConfigurations.configsByName)
	require.Empty(t, newEventConfigurations.configsByName)
	require.NotNil(t, newEventConfigurations.configsByNameAndTime)
	require.Empty(t, newEventConfigurations.configsByNameAndTime)
}

func Test_eventConfigurations_add(t *testing.T) {
	t.Parallel()

	e := newEventConfigurations()

	e.set("test", nil, EventConfiguration{})
	require.Len(t, e.configsByName, 1)
	require.Len(t, e.configsByNameAndTime, 0)

	e.set("test", nil, EventConfiguration{})
	require.Len(t, e.configsByName, 1)
	require.Len(t, e.configsByNameAndTime, 0)

	e.set("test", &time.Time{}, EventConfiguration{})
	require.Len(t, e.configsByName, 1)
	require.Len(t, e.configsByNameAndTime, 1)

	e.set("test", &time.Time{}, EventConfiguration{})
	require.Len(t, e.configsByName, 1)
	require.Len(t, e.configsByNameAndTime, 1)
}

func Test_eventConfigurations_get(t *testing.T) {
	t.Parallel()

	testcases := []struct {
		name                 string
		configsByName        map[string]*EventConfiguration
		configsByNameAndTime map[nameAndTimeKey]*EventConfiguration
		nameCallCount        int
		wantConfiguration    *EventConfiguration
	}{
		{
			name: "both configs exist",
			configsByName: map[string]*EventConfiguration{
				"test": {Priority: 10},
			},
			configsByNameAndTime: map[nameAndTimeKey]*EventConfiguration{
				{"test", time.Time{}.UnixMilli()}: {Priority: 20},
			},
			nameCallCount:     1,
			wantConfiguration: &EventConfiguration{Priority: 20},
		},
		{
			name: "config for name exists",
			configsByName: map[string]*EventConfiguration{
				"test": {Priority: 10},
			},
			configsByNameAndTime: map[nameAndTimeKey]*EventConfiguration{},
			nameCallCount:        2,
			wantConfiguration:    &EventConfiguration{Priority: 10},
		},
		{
			name:          "config for name + time exists",
			configsByName: map[string]*EventConfiguration{},
			configsByNameAndTime: map[nameAndTimeKey]*EventConfiguration{
				{"test", time.Time{}.UnixMilli()}: {Priority: 20},
			},
			nameCallCount:     1,
			wantConfiguration: &EventConfiguration{Priority: 20},
		},
		{
			name:                 "no config exists",
			configsByName:        map[string]*EventConfiguration{},
			configsByNameAndTime: map[nameAndTimeKey]*EventConfiguration{},
			nameCallCount:        2,
			wantConfiguration:    nil,
		},
	}

	for _, tt := range testcases {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			e := newEventConfigurations()
			e.configsByName = tt.configsByName
			e.configsByNameAndTime = tt.configsByNameAndTime

			mockAction := timestone.NewMockAction(t)
			mockAction.EXPECT().
				Name().
				Return("test").
				Times(tt.nameCallCount)
			mockEvent := NewEvent(context.Background(), mockAction, time.Time{})

			config := e.get(mockEvent)
			require.Equal(t, tt.wantConfiguration, config)
		})
	}
}

func Test_eventConfigurations_getScheduleMode(t *testing.T) {
	t.Parallel()

	defaultMode := ScheduleModeSequential

	testcases := []struct {
		name          string
		configsByName map[string]*EventConfiguration
		wantMode      ScheduleMode
	}{
		{
			name: "valid config",
			configsByName: map[string]*EventConfiguration{
				"test": {ScheduleMode: ScheduleModeAsync},
			},
			wantMode: ScheduleModeAsync,
		},
		{
			name: "config with undefined schedule mode",
			configsByName: map[string]*EventConfiguration{
				"test": {},
			},
			wantMode: defaultMode,
		},
		{
			name:          "no config for event",
			configsByName: map[string]*EventConfiguration{},
			wantMode:      defaultMode,
		},
	}

	for _, tt := range testcases {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			e := newEventConfigurations()
			e.defaultScheduleMode = defaultMode
			e.configsByName = tt.configsByName

			mockAction := timestone.NewMockAction(t)
			mockAction.EXPECT().
				Name().
				Return("test").
				Twice()
			mockEvent := NewEvent(context.Background(), mockAction, time.Time{})

			// Valid ScheduleModeAsync has been provided in event config
			mode := e.getScheduleMode(mockEvent)
			require.Equal(t, tt.wantMode, mode)
		})
	}
}

func Test_eventConfigurations_getRecursiveMode(t *testing.T) {
	t.Parallel()

	testcases := []struct {
		name          string
		configsByName map[string]*EventConfiguration
		wantMode      RecursiveMode
	}{
		{
			name: "valid config",
			configsByName: map[string]*EventConfiguration{
				"test": {RecursiveMode: RecursiveModeWaitForActions},
			},
			wantMode: RecursiveModeWaitForActions,
		},
		{
			name:          "no config for event",
			configsByName: map[string]*EventConfiguration{},
			wantMode:      RecursiveModeDefault,
		},
	}

	for _, tt := range testcases {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			e := newEventConfigurations()
			e.configsByName = tt.configsByName

			mockAction := timestone.NewMockAction(t)
			mockAction.EXPECT().
				Name().
				Return("test").
				Twice()
			mockEvent := NewEvent(context.Background(), mockAction, time.Time{})

			// Valid ScheduleModeAsync has been provided in event config
			mode := e.getRecursiveMode(mockEvent)
			require.Equal(t, tt.wantMode, mode)
		})
	}
}

func Test_eventConfigurations_getPriority(t *testing.T) {
	t.Parallel()

	testcases := []struct {
		name          string
		configsByName map[string]*EventConfiguration
		wantPriority  int
	}{
		{
			name: "valid config",
			configsByName: map[string]*EventConfiguration{
				"test": {Priority: 1},
			},
			wantPriority: 1,
		},
		{
			name:          "no config for event",
			configsByName: map[string]*EventConfiguration{},
			wantPriority:  EventPriorityDefault,
		},
	}

	for _, tt := range testcases {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			e := newEventConfigurations()
			e.configsByName = tt.configsByName

			mockAction := timestone.NewMockAction(t)
			mockAction.EXPECT().
				Name().
				Return("test").
				Twice()
			mockEvent := NewEvent(context.Background(), mockAction, time.Time{})

			// Valid ScheduleModeAsync has been provided in event config
			priority := e.getPriority(mockEvent)
			require.Equal(t, tt.wantPriority, priority)
		})
	}
}

func Test_eventConfigurations_getBlockingActions(t *testing.T) {
	t.Parallel()

	testcases := []struct {
		name                string
		configsByName       map[string]*EventConfiguration
		wantBlockingActions []string
	}{
		{
			name: "valid config",
			configsByName: map[string]*EventConfiguration{
				"test": {WaitForActions: []string{"test1"}},
			},
			wantBlockingActions: []string{"test1"},
		},
		{
			name:                "no config for event",
			configsByName:       map[string]*EventConfiguration{},
			wantBlockingActions: []string{},
		},
	}

	for _, tt := range testcases {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			e := newEventConfigurations()
			e.configsByName = tt.configsByName

			mockAction := timestone.NewMockAction(t)
			mockAction.EXPECT().
				Name().
				Return("test").
				Twice()
			mockEvent := NewEvent(context.Background(), mockAction, time.Time{})

			// Valid ScheduleModeAsync has been provided in event config
			blockingActions := e.getBlockingActions(mockEvent)
			require.Equal(t, tt.wantBlockingActions, blockingActions)
		})
	}
}
