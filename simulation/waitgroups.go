package simulation

import (
	"fmt"
	"maps"
	"sync"
)

type waitGroups struct {
	waitGroups map[string]*waitGroup

	mu sync.RWMutex
}

func newWaitGroups() *waitGroups {
	return &waitGroups{
		waitGroups: make(map[string]*waitGroup),
	}
}

func (w *waitGroups) new(name string) {
	w.mu.Lock()
	defer w.mu.Unlock()

	if _, exists := w.waitGroups[name]; !exists {
		w.waitGroups[name] = &waitGroup{}
	}
}

func (w *waitGroups) add(name string, delta int) {
	w.mu.RLock()
	defer w.mu.RUnlock()

	if _, exists := w.waitGroups[name]; !exists {
		panic(fmt.Sprintf("wait group for \"%s\" does not exist", name))
	}

	w.waitGroups[name].add(delta)
}

func (w *waitGroups) done(name string) {
	w.mu.RLock()
	if _, exists := w.waitGroups[name]; !exists {
		w.mu.RUnlock()
		panic(fmt.Sprintf("wait group for \"%s\" does not exist", name))
	}
	w.mu.RUnlock()

	w.waitGroups[name].done()
}

func (w *waitGroups) waitFor(names ...string) {
	// To understand why this implementation has been chosen,
	// consider an action "action2" adding more actions "action2.1",
	// with an "action1" previously called that has been configured to
	// wait for "action2" as well as all "action2.1" created by it.
	// At the time of calling waitFor, no "action2.1" exists yet, and
	// in consequence also no WaitGroup for this name. Therefore we first
	// wait for "action2" (or all other actions that already have a
	// corresponding WaitGroup) to give it a chance to spawn
	// the missing WaitGroups and avoid a panic.

	for len(names) > 0 {
		var remainingNames []string

		w.mu.RLock()
		for _, name := range names {
			if wg, exists := w.waitGroups[name]; exists {
				w.mu.RUnlock() // Unlock before waiting to avoid deadlocks
				wg.wait()
				w.mu.RLock() // Reacquire the lock after waiting
			} else {
				remainingNames = append(remainingNames, name)
			}
		}
		w.mu.RUnlock()

		if len(remainingNames) == len(names) {
			panic(fmt.Sprintf("wait group(s) for \"%v\" do not exist", remainingNames))
		}

		names = remainingNames
	}
}

func (w *waitGroups) wait() {
	w.mu.RLock()
	var waitGroups = make(map[string]*waitGroup, len(w.waitGroups))
	maps.Copy(waitGroups, w.waitGroups)
	w.mu.RUnlock()

	for _, wg := range waitGroups {
		wg.wait()
	}
}
