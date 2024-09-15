package waitgroups

import (
	"fmt"
	"github.com/metamogul/timestone/simulation/event"
	"github.com/metamogul/timestone/simulation/internal/tags"
	"sync"
	"time"
)

type EventWaitGroups struct {
	waitGroups *tags.TaggedStore[map[int64]*sync.WaitGroup]

	mu sync.RWMutex
}

func NewEventWaitGroups() *EventWaitGroups {
	return &EventWaitGroups{
		waitGroups: tags.NewTaggedStore[map[int64]*sync.WaitGroup](),
	}
}

func (e *EventWaitGroups) New(time time.Time, tags []string) *sync.WaitGroup {
	e.mu.Lock()
	defer e.mu.Unlock()

	setForTags := e.waitGroups.Matching(tags)
	if setForTags == nil {
		setForTags = make(map[int64]*sync.WaitGroup)
		e.waitGroups.Set(setForTags, tags)
	}

	timeUnixMilli := time.UnixMilli()
	waitGroupForTagsAndTime, exists := setForTags[timeUnixMilli]
	if !exists {
		waitGroupForTagsAndTime = new(sync.WaitGroup)
		setForTags[timeUnixMilli] = waitGroupForTagsAndTime
	}

	waitGroupForTagsAndTime.Add(1)

	return waitGroupForTagsAndTime
}

func (e *EventWaitGroups) WaitFor(events []*event.Key) {
	// To understand why this implementation has been chosen,
	// consider an action with tag "action2" adding more actions tagged
	// "action2.1", with an "action1" previously called that has been
	// configured to  Wait for "action2" as well as all "action2.1" created
	// by it.
	// At the time of calling WaitFor, no "action2.1" exists yet, and
	// in consequence also no WaitGroup for this name. Therefore we first
	// Wait for "action2" (or all other actions that already have a
	// corresponding WaitGroup) to give it a chance to spawn
	// the missing WaitGroups and avoid a panic.

	for len(events) > 0 {
		var remainingEvents []*event.Key

		e.mu.RLock()
		for _, eventKey := range events {
			foundAllWaitGroups := e.waitFor(eventKey)
			if !foundAllWaitGroups {
				remainingEvents = append(remainingEvents, eventKey)
			}
		}
		e.mu.RUnlock()

		if len(remainingEvents) == len(events) {
			var remainingEventsValues []event.Key
			for _, eventKey := range remainingEvents {
				remainingEventsValues = append(remainingEventsValues, *eventKey)
			}
			panic(fmt.Sprintf("Wait group(s) for %v do not exist", remainingEventsValues))
		}

		events = remainingEvents
	}
}

func (e *EventWaitGroups) waitFor(event *event.Key) (success bool) {
	waitGroupSetsForTagsByTime := e.waitGroups.Containing(event.Tags)
	if len(waitGroupSetsForTagsByTime) == 0 {
		return false
	}

	if event.Time != nil {
		// Wait for events at time
		timeUnixMilli := event.Time.UnixMilli()
		for _, waitGroupsForTagsByTime := range waitGroupSetsForTagsByTime {
			wg, exists := waitGroupsForTagsByTime[timeUnixMilli]
			if !exists {
				return false
			}

			e.mu.RUnlock() // Unlock before waiting to avoid deadlocks
			wg.Wait()
			e.mu.RLock() // Reacquire the lock after waiting
		}
	} else {
		// Wait for all events containing tags
		for _, waitGroupsForTagsByTime := range waitGroupSetsForTagsByTime {
			e.mu.RUnlock() // Unlock before waiting to avoid deadlocks
			for _, wg := range waitGroupsForTagsByTime {
				wg.Wait()
			}
			e.mu.RLock() // Reacquire the lock after waiting
		}
	}

	return true
}

func (e *EventWaitGroups) Wait() {
	e.mu.RLock()
	allWaitGroupsByTime := e.waitGroups.All()
	e.mu.RUnlock()

	for _, waitGroupsByTime := range allWaitGroupsByTime {
		for _, wg := range waitGroupsByTime {
			wg.Wait()
		}
	}
}
