package waitgroups

import (
	"fmt"
	"github.com/metamogul/timestone/v2/simulation/internal/data"
	"sync"
)

type GeneratorWaitGroups struct {
	waitGroups *data.TaggedStore[*waitGroup]

	mu sync.RWMutex
}

func NewGeneratorWaitGroups() *GeneratorWaitGroups {
	return &GeneratorWaitGroups{
		waitGroups: data.NewTaggedStore[*waitGroup](),
	}
}

func (w *GeneratorWaitGroups) Add(delta int, tags []string) {
	w.mu.Lock()
	defer w.mu.Unlock()

	matchingEntry := w.waitGroups.Matching(tags)
	if matchingEntry == nil {
		matchingEntry = &waitGroup{}
		w.waitGroups.Set(matchingEntry, tags)
	}

	matchingEntry.add(delta)
}

func (w *GeneratorWaitGroups) Done(tags []string) {
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

func (w *GeneratorWaitGroups) WaitFor(tags []string) {
	w.mu.RLock()
	defer w.mu.RUnlock()

	waitGroupForTags := w.waitGroups.Matching(tags)
	if waitGroupForTags == nil {
		panic(fmt.Sprintf("WaitGroup for %v does not exist", tags))
	}

	waitGroupForTags.wait()
}
