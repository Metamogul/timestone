package simulation

import (
	"time"
)

type clock struct {
	now time.Time
}

func newClock(now time.Time) *clock {
	return &clock{
		now: now,
	}
}

func (c *clock) Now() time.Time {
	return c.now
}

func (c *clock) set(t time.Time) {
	if t.Before(c.now) {
		panic("time can't be in the past")
	}

	c.now = t
}
