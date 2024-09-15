package events

import (
	"github.com/metamogul/timestone/simulation/event"
	"github.com/metamogul/timestone/simulation/internal/tags"
	"time"
)

const (
	// EventPriorityDefault represents the default priority for an event if
	// none has been Set.
	EventPriorityDefault = iota
)

type Configs struct {
	configsByTags        *tags.TaggedStore[*event.Config]
	configsByTagsAndTime map[int64]*tags.TaggedStore[*event.Config]
}

func NewConfigs() *Configs {
	return &Configs{
		configsByTags:        tags.NewTaggedStore[*event.Config](),
		configsByTagsAndTime: make(map[int64]*tags.TaggedStore[*event.Config]),
	}
}

func (c *Configs) Set(config event.Config, time *time.Time, tags ...string) {
	if time != nil {
		c.configsByTagsForTime(*time).Set(&config, tags)
		return
	}

	c.configsByTags.Set(&config, tags)
}

func (c *Configs) Priority(event *Event) int {
	if config := c.get(event); config != nil {
		return config.Priority
	}

	return EventPriorityDefault
}

func (c *Configs) BlockingEvents(event *Event) []*event.Key {
	if config := c.get(event); config != nil {
		return config.WaitForEvents
	}

	return nil
}

func (c *Configs) ExpectedGenerators(event *Event) []*event.GeneratorExpectation {
	if config := c.get(event); config != nil {
		return config.AddsGenerators
	}

	return nil
}

func (c *Configs) configsByTagsForTime(time time.Time) *tags.TaggedStore[*event.Config] {
	result, exists := c.configsByTagsAndTime[time.UnixMilli()]

	if !exists {
		result = tags.NewTaggedStore[*event.Config]()
		c.configsByTagsAndTime[time.UnixMilli()] = result
	}

	return result
}

func (c *Configs) get(event *Event) *event.Config {
	if config := c.configsByTagsForTime(event.Time).Matching(event.tags); config != nil {
		return config
	}

	if config := c.configsByTags.Matching(event.tags); config != nil {
		return config
	}

	return nil
}
