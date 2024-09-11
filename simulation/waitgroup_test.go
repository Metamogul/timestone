package simulation

import (
	"testing"
)

func Test_waitGroup_add(t *testing.T) {
	t.Parallel()

	testcases := []struct {
		name         string
		initialCount int
		addDelta     int
		wantCount    int
	}{
		{
			name:         "add zero",
			initialCount: 0,
			addDelta:     0,
			wantCount:    0,
		},
		{
			name:         "add one",
			initialCount: 0,
			addDelta:     1,
			wantCount:    1,
		},
		{
			name:         "add multiple",
			initialCount: 0,
			addDelta:     10,
			wantCount:    10,
		},
		{
			name:         "add negative, count+delta < 0",
			initialCount: 1,
			addDelta:     -2,
			wantCount:    0,
		},
	}

	for _, tt := range testcases {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			w := waitGroup{}

			w.waitGroup.Add(tt.initialCount)
			w.count = tt.initialCount

			w.add(tt.addDelta)

			for range w.count {
				go func() {
					w.done()
				}()
			}
			w.wait()
		})
	}
}

func Test_waitGroup_done(t *testing.T) {
	t.Parallel()

	testcases := []struct {
		name         string
		initialCount int
		timesDone    int
		wantCount    int
	}{
		{
			name:         "count greater zero",
			initialCount: 1,
			timesDone:    1,
			wantCount:    0,
		},
		{
			name:         "count is zero",
			initialCount: 0,
			timesDone:    1,
			wantCount:    0,
		},
	}

	for _, tt := range testcases {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			w := waitGroup{}

			w.waitGroup.Add(tt.initialCount)
			w.count = tt.initialCount

			for range tt.timesDone {
				go func() { w.done() }()
			}
			w.wait()
		})
	}
}

func Test_waitGroup_wait(t *testing.T) {
	t.Parallel()

	w := waitGroup{}

	const delta = 5

	w.add(delta)
	for range delta {
		go func() { w.done() }()
	}

	w.wait()
}
