package events

import (
	"context"
	"github.com/metamogul/timestone/simulation/config"
	configinternal "github.com/metamogul/timestone/simulation/internal/config"
	"github.com/metamogul/timestone/simulation/internal/data"
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
		e.Set(config.Config{})
	})

	e.Set(config.Config{Tags: []string{"test1", "test2"}})
	require.Len(t, e.configsByTags.All(), 1)
	require.Len(t, e.configsByTagsAndTime, 0)

	e.Set(config.Config{Tags: []string{"test1", "test2"}})
	require.Len(t, e.configsByTags.All(), 1)
	require.Len(t, e.configsByTagsAndTime, 0)

	now := time.Now()

	e.Set(config.Config{Time: now, Tags: []string{"test1", "test2"}})
	require.Len(t, e.configsByTags.All(), 1)
	require.Len(t, e.configsByTagsAndTime, 1)
	require.Len(t, e.configsByTagsAndTime[now.UnixMilli()].All(), 1)

	e.Set(config.Config{Time: now, Tags: []string{"test1", "test2"}})
	require.Len(t, e.configsByTags.All(), 1)
	require.Len(t, e.configsByTagsAndTime, 1)
	require.Len(t, e.configsByTagsAndTime[now.UnixMilli()].All(), 1)
}

func Test_Configs_Priority(t *testing.T) {
	t.Parallel()

	testcases := []struct {
		name          string
		insertConfigs []config.Config
		wantPriority  int
	}{
		{
			name:          "valid config",
			insertConfigs: []config.Config{{Tags: []string{"test1", "test2"}, Priority: 1}},
			wantPriority:  1,
		},
		{
			name:          "no config for event",
			insertConfigs: []config.Config{},
			wantPriority:  EventPriorityDefault,
		},
	}

	for _, tt := range testcases {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			e := NewConfigs()
			for _, configToInsert := range tt.insertConfigs {
				e.Set(configToInsert)
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

	testcases := []struct {
		name               string
		insertConfigs      []config.Config
		wantBlockingEvents []config.Event
	}{
		{
			name: "transformed event keys",
			insertConfigs: []config.Config{
				{
					Tags:    []string{"test1", "test2"},
					WaitFor: []config.Event{config.Before{Interval: -1, Tags: []string{"test1", "test2"}}},
				},
			},
			wantBlockingEvents: []config.Event{configinternal.At{Time: time.Time{}.Add(-1), Tags: []string{"test1", "test2"}}},
		},
		{
			name: "valid config",
			insertConfigs: []config.Config{
				{
					Tags:    []string{"test1", "test2"},
					WaitFor: []config.Event{config.All{Tags: []string{"test1", "test2"}}},
				},
			},
			wantBlockingEvents: []config.Event{config.All{Tags: []string{"test1", "test2"}}},
		},
		{
			name:               "no config for event",
			insertConfigs:      []config.Config{},
			wantBlockingEvents: nil,
		},
	}

	for _, tt := range testcases {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			e := NewConfigs()
			for _, configToInsert := range tt.insertConfigs {
				e.Set(configToInsert)
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
	testcases := []struct {
		name                   string
		insertConfigs          []config.Config
		wantExpectedGenerators []*config.Generator
	}{
		{
			name: "valid config",
			insertConfigs: []config.Config{
				{
					Tags: []string{"test1", "test2"},
					Adds: []*config.Generator{{[]string{"testWanted"}, 1}},
				},
			},
			wantExpectedGenerators: []*config.Generator{{[]string{"testWanted"}, 1}},
		},
		{
			name:                   "no config for event",
			insertConfigs:          []config.Config{},
			wantExpectedGenerators: nil,
		},
	}

	for _, tt := range testcases {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			e := NewConfigs()
			for _, configToInsert := range tt.insertConfigs {
				e.Set(configToInsert)
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
		configsByTagsAndTime map[int64]*data.TaggedStore[*config.Config]
	}{
		{
			name: "entry exists",
			configsByTagsAndTime: map[int64]*data.TaggedStore[*config.Config]{
				0: data.NewTaggedStore[*config.Config](),
			},
		},
		{
			name:                 "entry does not exist",
			configsByTagsAndTime: make(map[int64]*data.TaggedStore[*config.Config]),
		},
	}

	for _, tt := range testcases {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			e := NewConfigs()
			e.configsByTagsAndTime = tt.configsByTagsAndTime

			result := e.configsByTagsForTime(time.Time{})
			require.Equal(t, data.NewTaggedStore[*config.Config](), result)
		})
	}
}

func Test_Configs_get(t *testing.T) {
	t.Parallel()

	testcases := []struct {
		name              string
		insertConfigs     []config.Config
		wantConfiguration *config.Config
	}{
		{
			name: "both configs exist",
			insertConfigs: []config.Config{
				{Tags: []string{"test1", "test2"}, Time: time.Time{}.Add(1), Priority: 20},
				{Tags: []string{"test1", "test2"}, Priority: 10},
			},
			wantConfiguration: &config.Config{Tags: []string{"test1", "test2"}, Time: time.Time{}.Add(1), Priority: 20},
		},
		{
			name: "config for name exists",
			insertConfigs: []config.Config{
				{Tags: []string{"test1", "test2"}, Priority: 10},
			},
			wantConfiguration: &config.Config{Tags: []string{"test1", "test2"}, Priority: 10},
		},
		{
			name: "config for name + time exists",
			insertConfigs: []config.Config{
				{Tags: []string{"test1", "test2"}, Time: time.Time{}.Add(1), Priority: 20},
			},
			wantConfiguration: &config.Config{Tags: []string{"test1", "test2"}, Time: time.Time{}.Add(1), Priority: 20},
		},
		{
			name:              "no config exists",
			insertConfigs:     []config.Config{},
			wantConfiguration: nil,
		},
	}

	for _, tt := range testcases {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			e := NewConfigs()
			for _, configToInsert := range tt.insertConfigs {
				e.Set(configToInsert)
			}

			mockEvent := NewEvent(
				context.Background(),
				timestone.NewMockAction(t),
				time.Time{},
				[]string{"test1", "test2"},
			)

			gotConfig := e.get(mockEvent)
			require.Equal(t, tt.wantConfiguration, gotConfig)
		})
	}
}
