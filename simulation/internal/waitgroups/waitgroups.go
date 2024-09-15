package waitgroups

import (
	"fmt"
	"github.com/metamogul/timestone/simulation/internal/tags"
	"sync"
)

type WaitGroups struct {
	waitGroups *tags.TaggedStore[*waitGroup]

	mu sync.RWMutex
}

func NewWaitGroups() *WaitGroups {
	return &WaitGroups{
		waitGroups: tags.NewTaggedStore[*waitGroup](),
	}
}

func (w *WaitGroups) Add(delta int, tags []string) {
	w.mu.Lock()
	defer w.mu.Unlock()

	matchingEntry := w.waitGroups.Matching(tags)
	if matchingEntry == nil {
		matchingEntry = &waitGroup{}
		w.waitGroups.Set(matchingEntry, tags)
	}

	matchingEntry.add(delta)
}

func (w *WaitGroups) Done(tags []string) {
	w.mu.RLock()
	defer w.mu.RUnlock()

	matchingEntries := w.waitGroups.ContainedIn(tags)
	if len(matchingEntries) == 0 {
		return
	}

	for _, matchingEntry := range matchingEntries {
		matchingEntry.done()
	}
}

func (w *WaitGroups) WaitFor(tags []string) {
	w.mu.RLock()
	defer w.mu.RUnlock()

	waitGroupForTags := w.waitGroups.Matching(tags)
	if waitGroupForTags == nil {
		panic(fmt.Sprintf("WaitGroup for %v does not exist", tags))
	}

	waitGroupForTags.wait()
}
