package waitgroups

import "sync"

type waitGroup struct {
	waitGroup sync.WaitGroup
	count     int
	mu        sync.Mutex
}

func (w *waitGroup) add(delta int) {
	if delta == 0 {
		return
	}

	w.mu.Lock()
	defer w.mu.Unlock()

	if w.count+delta < 0 {
		delta = -w.count
	}

	w.count += delta
	w.waitGroup.Add(delta)
}

func (w *waitGroup) done() {
	w.mu.Lock()
	defer w.mu.Unlock()

	if w.count == 0 {
		return
	}

	w.count--
	w.waitGroup.Done()
}

func (w *waitGroup) wait() {
	w.waitGroup.Wait()
}
