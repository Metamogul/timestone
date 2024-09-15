package events

import (
	"context"
	"github.com/metamogul/timestone/simulation/event"
	"github.com/metamogul/timestone/simulation/internal/tags"
	"testing"
	"time"

	"github.com/metamogul/timestone"
	"github.com/stretchr/testify/require"
)

func Test_NewConfigs(t *testing.T) {
	t.Parallel()

	newEventConfigurations := NewConfigs()

	require.NotNil(t, newEventConfigurations)
	require.NotNil(t, newEventConfigurations.configsByTags)
	require.Empty(t, newEventConfigurations.configsByTags.All())
	require.NotNil(t, newEventConfigurations.configsByTagsAndTime)
	require.Empty(t, newEventConfigurations.configsByTagsAndTime)
}

func Test_Configs_Add(t *testing.T) {
	t.Parallel()

	e := NewConfigs()

	require.Panics(t, func() {
		e.Set(event.Config{}, nil)
	})

	e.Set(event.Config{}, nil, "test1", "test2")
	require.Len(t, e.configsByTags.All(), 1)
	require.Len(t, e.configsByTagsAndTime, 0)

	e.Set(event.Config{}, nil, "test1", "test2")
	require.Len(t, e.configsByTags.All(), 1)
	require.Len(t, e.configsByTagsAndTime, 0)

	e.Set(event.Config{}, &time.Time{}, "test1", "test2")
	require.Len(t, e.configsByTags.All(), 1)
	require.Len(t, e.configsByTagsAndTime, 1)
	require.Len(t, e.configsByTagsAndTime[time.Time{}.UnixMilli()].All(), 1)

	e.Set(event.Config{}, &time.Time{}, "test1", "test2")
	require.Len(t, e.configsByTags.All(), 1)
	require.Len(t, e.configsByTagsAndTime, 1)
	require.Len(t, e.configsByTagsAndTime[time.Time{}.UnixMilli()].All(), 1)
}

func Test_Configs_Priority(t *testing.T) {
	t.Parallel()

	type insertConfigArgs struct {
		config event.Config
		time   *time.Time
		tags   []string
	}

	testcases := []struct {
		name          string
		insertConfigs []insertConfigArgs
		wantPriority  int
	}{
		{
			name: "valid config",
			insertConfigs: []insertConfigArgs{
				{config: event.Config{Priority: 1}, time: nil, tags: []string{"test1", "test2"}},
			},
			wantPriority: 1,
		},
		{
			name:          "no config for event",
			insertConfigs: []insertConfigArgs{},
			wantPriority:  EventPriorityDefault,
		},
	}

	for _, tt := range testcases {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			e := NewConfigs()
			for _, configArgs := range tt.insertConfigs {
				e.Set(configArgs.config, configArgs.time, configArgs.tags...)
			}

			mockEvent := NewEvent(
				context.Background(),
				timestone.NewMockAction(t),
				time.Time{},
				[]string{"test1", "test2"},
			)

			priority := e.Priority(mockEvent)
			require.Equal(t, tt.wantPriority, priority)
		})
	}
}

func Test_Configs_BlockingEvents(t *testing.T) {
	t.Parallel()

	type insertConfigArgs struct {
		config event.Config
		time   *time.Time
		tags   []string
	}

	testcases := []struct {
		name               string
		insertConfigs      []insertConfigArgs
		wantBlockingEvents []*event.Key
	}{
		{
			name: "valid config",
			insertConfigs: []insertConfigArgs{
				{
					config: event.Config{WaitForEvents: []*event.Key{{Tags: []string{"test1", "test2"}}}},
					time:   nil,
					tags:   []string{"test1", "test2"},
				},
			},
			wantBlockingEvents: []*event.Key{{Tags: []string{"test1", "test2"}}},
		},
		{
			name:               "no config for event",
			insertConfigs:      []insertConfigArgs{},
			wantBlockingEvents: nil,
		},
	}

	for _, tt := range testcases {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			e := NewConfigs()
			for _, configArgs := range tt.insertConfigs {
				e.Set(configArgs.config, configArgs.time, configArgs.tags...)
			}

			mockEvent := NewEvent(
				context.Background(),
				timestone.NewMockAction(t),
				time.Time{},
				[]string{"test1", "test2"},
			)

			blockingEvents := e.BlockingEvents(mockEvent)
			require.Equal(t, tt.wantBlockingEvents, blockingEvents)
		})
	}
}

func Test_Configs_ExpectedGenerators(t *testing.T) {
	t.Parallel()

	type insertConfigArgs struct {
		config event.Config
		time   *time.Time
		tags   []string
	}

	testcases := []struct {
		name                   string
		insertConfigs          []insertConfigArgs
		wantExpectedGenerators []*event.GeneratorExpectation
	}{
		{
			name: "valid config",
			insertConfigs: []insertConfigArgs{
				{
					config: event.Config{
						AddsGenerators: []*event.GeneratorExpectation{{[]string{"testWanted"}, 1}},
					},
					time: nil,
					tags: []string{"test1", "test2"},
				},
			},
			wantExpectedGenerators: []*event.GeneratorExpectation{{[]string{"testWanted"}, 1}},
		},
		{
			name:                   "no config for event",
			insertConfigs:          []insertConfigArgs{},
			wantExpectedGenerators: nil,
		},
	}

	for _, tt := range testcases {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			e := NewConfigs()
			for _, configArgs := range tt.insertConfigs {
				e.Set(configArgs.config, configArgs.time, configArgs.tags...)
			}

			mockEvent := NewEvent(
				context.Background(),
				timestone.NewMockAction(t),
				time.Time{},
				[]string{"test1", "test2"},
			)

			wantedNewGenerators := e.ExpectedGenerators(mockEvent)
			require.Equal(t, tt.wantExpectedGenerators, wantedNewGenerators)
		})
	}
}

func Test_Configs_configsByTagsForTime(t *testing.T) {
	t.Parallel()

	testcases := []struct {
		name                 string
		configsByTagsAndTime map[int64]*tags.TaggedStore[*event.Config]
	}{
		{
			name: "entry exists",
			configsByTagsAndTime: map[int64]*tags.TaggedStore[*event.Config]{
				0: tags.NewTaggedStore[*event.Config](),
			},
		},
		{
			name:                 "entry does not exist",
			configsByTagsAndTime: make(map[int64]*tags.TaggedStore[*event.Config]),
		},
	}

	for _, tt := range testcases {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			e := NewConfigs()
			e.configsByTagsAndTime = tt.configsByTagsAndTime

			result := e.configsByTagsForTime(time.Time{})
			require.Equal(t, tags.NewTaggedStore[*event.Config](), result)
		})
	}
}

func Test_Configs_get(t *testing.T) {
	t.Parallel()

	type insertConfigArgs struct {
		config event.Config
		time   *time.Time
		tags   []string
	}

	testcases := []struct {
		name              string
		insertConfigs     []insertConfigArgs
		nameCallCount     int
		wantConfiguration *event.Config
	}{
		{
			name: "both configs exist",
			insertConfigs: []insertConfigArgs{
				{config: event.Config{Priority: 20}, time: &time.Time{}, tags: []string{"test1", "test2"}},
				{config: event.Config{Priority: 10}, time: nil, tags: []string{"test1", "test2"}},
			},
			nameCallCount:     1,
			wantConfiguration: &event.Config{Priority: 20},
		},
		{
			name: "config for name exists",
			insertConfigs: []insertConfigArgs{
				{config: event.Config{Priority: 10}, time: nil, tags: []string{"test1", "test2"}},
			},
			nameCallCount:     2,
			wantConfiguration: &event.Config{Priority: 10},
		},
		{
			name: "config for name + time exists",
			insertConfigs: []insertConfigArgs{
				{config: event.Config{Priority: 20}, time: &time.Time{}, tags: []string{"test1", "test2"}},
			},
			nameCallCount:     1,
			wantConfiguration: &event.Config{Priority: 20},
		},
		{
			name:              "no config exists",
			insertConfigs:     []insertConfigArgs{},
			nameCallCount:     2,
			wantConfiguration: nil,
		},
	}

	for _, tt := range testcases {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			e := NewConfigs()
			for _, configArgs := range tt.insertConfigs {
				e.Set(configArgs.config, configArgs.time, configArgs.tags...)
			}

			mockEvent := NewEvent(
				context.Background(),
				timestone.NewMockAction(t),
				time.Time{},
				[]string{"test1", "test2"},
			)

			config := e.get(mockEvent)
			require.Equal(t, tt.wantConfiguration, config)
		})
	}
}
