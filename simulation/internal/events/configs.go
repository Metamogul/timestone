package events

import (
	"github.com/metamogul/timestone/v2/simulation/config"
	configinternal "github.com/metamogul/timestone/v2/simulation/internal/config"
	"github.com/metamogul/timestone/v2/simulation/internal/data"
	"time"
)

const (
	// EventPriorityDefault represents the default priority for an event if
	// none has been Set.
	EventPriorityDefault = iota
)

type Configs struct {
	configsByTags        *data.TaggedStore[*config.Config]
	configsByTagsAndTime map[int64]*data.TaggedStore[*config.Config]
}

func NewConfigs() *Configs {
	return &Configs{
		configsByTags:        data.NewTaggedStore[*config.Config](),
		configsByTagsAndTime: make(map[int64]*data.TaggedStore[*config.Config]),
	}
}

func (c *Configs) Set(config config.Config) {
	if !config.Time.IsZero() {
		c.configsByTagsForTime(config.Time).Set(&config, config.Tags)
		return
	}

	c.configsByTags.Set(&config, config.Tags)
}

func (c *Configs) Priority(event *Event) int {
	if configuration := c.get(event); configuration != nil {
		return configuration.Priority
	}

	return EventPriorityDefault
}

func (c *Configs) BlockingEvents(event *Event) []config.Event {
	if configuration := c.get(event); configuration != nil {

		blockingEvents := configuration.WaitFor

		result := make([]config.Event, len(blockingEvents))
		for i, blockingEvent := range blockingEvents {
			switch blockingEvent := blockingEvent.(type) {
			case config.Before:
				result[i] = configinternal.Convert(blockingEvent, event.Time)
			default:
				result[i] = blockingEvent
			}
		}

		return result
	}

	return nil
}

func (c *Configs) ExpectedGenerators(event *Event) []*config.Generator {
	if configuration := c.get(event); configuration != nil {
		return configuration.Adds
	}

	return nil
}

func (c *Configs) configsByTagsForTime(time time.Time) *data.TaggedStore[*config.Config] {
	result, exists := c.configsByTagsAndTime[time.UnixMilli()]

	if !exists {
		result = data.NewTaggedStore[*config.Config]()
		c.configsByTagsAndTime[time.UnixMilli()] = result
	}

	return result
}

func (c *Configs) get(event *Event) *config.Config {
	if configuration := c.configsByTagsForTime(event.Time).Matching(event.tags); configuration != nil {
		return configuration
	}

	if configuration := c.configsByTags.Matching(event.tags); configuration != nil {
		return configuration
	}

	return nil
}
