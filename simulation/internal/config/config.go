package config

import (
	"github.com/metamogul/timestone/v2/simulation/config"
	"time"
)

type At struct {
	Tags []string
	Time time.Time
}

func (b At) GetTags() []string {
	return b.Tags
}

func Convert(before config.Before, eventTime time.Time) At {
	return At{
		Tags: before.Tags,
		Time: eventTime.Add(before.Interval),
	}
}
