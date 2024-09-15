package clock

import (
	"time"
)

type Clock struct {
	now time.Time
}

func NewClock(now time.Time) *Clock {
	return &Clock{
		now: now,
	}
}

func (c *Clock) Now() time.Time {
	return c.now
}

func (c *Clock) Set(t time.Time) {
	if t.Before(c.now) {
		panic("time can't be in the past")
	}

	c.now = t
}
