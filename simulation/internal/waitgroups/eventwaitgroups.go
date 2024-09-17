package waitgroups

import (
	"fmt"
	"github.com/metamogul/timestone/v2/simulation/config"
	configinternal "github.com/metamogul/timestone/v2/simulation/internal/config"
	"github.com/metamogul/timestone/v2/simulation/internal/data"
	"sync"
	"time"
)

type EventWaitGroups struct {
	waitGroups *data.TaggedStore[map[int64]*sync.WaitGroup]

	mu sync.RWMutex
}

func NewEventWaitGroups() *EventWaitGroups {
	return &EventWaitGroups{
		waitGroups: data.NewTaggedStore[map[int64]*sync.WaitGroup](),
	}
}

func (e *EventWaitGroups) New(time time.Time, tags []string) *sync.WaitGroup {
	e.mu.Lock()
	defer e.mu.Unlock()

	waitGroupsForTags := e.waitGroups.Matching(tags)
	if waitGroupsForTags == nil {
		waitGroupsForTags = make(map[int64]*sync.WaitGroup)
		e.waitGroups.Set(waitGroupsForTags, tags)
	}

	timeUnixMilli := time.UnixMilli()
	waitGroupForTagsAndTime, exists := waitGroupsForTags[timeUnixMilli]
	if !exists {
		waitGroupForTagsAndTime = new(sync.WaitGroup)
		waitGroupsForTags[timeUnixMilli] = waitGroupForTagsAndTime
	}

	waitGroupForTagsAndTime.Add(1)

	return waitGroupForTagsAndTime
}

func (e *EventWaitGroups) WaitFor(events []config.Event) {
	// To understand why this implementation has been chosen,
	// consider an action with tag "action2" adding more actions tagged
	// "action2.1", with an "action1" previously called that has been
	// configured to  Wait for "action2" as well as all "action2.1" created
	// by it.
	// At the time of calling WaitFor, no "action2.1" exists yet, and
	// in consequence also no WaitGroup for this name. Therefore we first
	// Wait for "action2" (or all other actions that already have a
	// corresponding WaitGroup) to give it a chance to spawn
	// the missing GeneratorWaitGroups and avoid a panic.

	for len(events) > 0 {
		var remainingEvents []config.Event

		e.mu.RLock()
		for _, eventKey := range events {
			foundAllWaitGroups := e.waitFor(eventKey)
			if !foundAllWaitGroups {
				remainingEvents = append(remainingEvents, eventKey)
			}
		}
		e.mu.RUnlock()

		if len(remainingEvents) == len(events) {
			panic(fmt.Sprintf("Wait group(s) for %v do not exist", events))
		}

		events = remainingEvents
	}
}

func (e *EventWaitGroups) waitFor(event config.Event) (success bool) {
	waitGroupSetsForTagsByTime := e.waitGroups.Containing(event.GetTags())
	if len(waitGroupSetsForTagsByTime) == 0 {
		_, ignoreMissmatch := event.(configinternal.At)
		return ignoreMissmatch
	}

	switch event := event.(type) {

	case configinternal.At:
		// Wait for events at time, ignore missing match
		for _, waitGroupsForTagsByTime := range waitGroupSetsForTagsByTime {
			wg, exists := waitGroupsForTagsByTime[event.Time.UnixMilli()]
			if !exists {
				return true
			}

			e.mu.RUnlock() // Unlock before waiting to avoid deadlocks
			wg.Wait()
			e.mu.RLock() // Reacquire the lock after waiting
		}

	case config.At:
		// Wait for events at time, don't ignore missing match
		for _, waitGroupsForTagsByTime := range waitGroupSetsForTagsByTime {
			wg, exists := waitGroupsForTagsByTime[event.Time.UnixMilli()]
			if !exists {
				return false
			}

			e.mu.RUnlock() // Unlock before waiting to avoid deadlocks
			wg.Wait()
			e.mu.RLock() // Reacquire the lock after waiting
		}

	case config.All:
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
