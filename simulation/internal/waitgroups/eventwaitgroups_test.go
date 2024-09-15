package waitgroups

import (
	"fmt"
	"github.com/metamogul/timestone/internal"
	"github.com/metamogul/timestone/simulation/event"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

func Test_NewEventWaitGroups(t *testing.T) {
	t.Parallel()

	newEventWaitGroups := NewEventWaitGroups()
	require.NotNil(t, newEventWaitGroups.waitGroups)
	require.Empty(t, newEventWaitGroups.waitGroups.All())
}

func TestEventWaitGroups_New(t *testing.T) {
	t.Parallel()

	testcases := []struct {
		name     string
		time     time.Time
		tags     []string
		addCount int
	}{
		{
			name:     "wait groups for tags doesn't exist",
			time:     time.Time{},
			tags:     []string{"test"},
			addCount: -1,
		},
		{
			name:     "wait groups for tags exists",
			time:     time.Time{},
			tags:     []string{"testExists"},
			addCount: -1,
		},
		{
			name:     "wait group for tags exists and entry for time exists",
			time:     time.Time{}.Add(time.Second),
			tags:     []string{"testExists"},
			addCount: -2,
		},
	}

	for _, tt := range testcases {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			e := NewEventWaitGroups()
			_ = e.New(time.Time{}.Add(time.Second), []string{"testExists"})

			wg := e.New(tt.time, tt.tags)
			go func() { wg.Add(tt.addCount) }()
			wg.Wait()
		})
	}

}

func TestEventWaitGroups_WaitFor(t *testing.T) {
	t.Parallel()

	presentTags := []string{"test", "testGroup", "foo"}
	presentTime := time.Time{}

	e := NewEventWaitGroups()

	wg := e.New(presentTime, presentTags)
	go func() {
		wg2 := e.New(presentTime.Add(time.Second), presentTags)
		wg2.Done()
		wg.Done()
	}()

	e.WaitFor([]*event.Key{
		{
			Tags: presentTags,
			Time: internal.Ptr(presentTime.Add(time.Second)),
		},
		{
			Tags: presentTags,
			Time: &presentTime,
		},
	})

}

func TestEventWaitGroups_waitFor(t *testing.T) {
	t.Parallel()

	presentTags := []string{"test", "testGroup", "foo"}
	presentTime := time.Time{}

	testcases := []struct {
		name            string
		waitForEventKey *event.Key
		wantSuccess     bool
	}{
		{
			name:            "time wanted, no result for tags",
			waitForEventKey: &event.Key{Tags: []string{"baz"}, Time: &time.Time{}},
			wantSuccess:     false,
		},
		{
			name:            "not time wanted, no result for tags",
			waitForEventKey: &event.Key{Tags: []string{"baz"}, Time: &time.Time{}},
			wantSuccess:     false,
		},
		{
			name:            "time wanted, but not present",
			waitForEventKey: &event.Key{Tags: []string{"test"}, Time: internal.Ptr(time.Now())},
			wantSuccess:     false,
		},
		{
			name:            "time wanted, is present",
			waitForEventKey: &event.Key{Tags: []string{"test"}, Time: &time.Time{}},
			wantSuccess:     true,
		},
		{
			name:            "no time wanted",
			waitForEventKey: &event.Key{Tags: []string{"test"}},
			wantSuccess:     true,
		},
	}

	for _, tt := range testcases {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			e := NewEventWaitGroups()

			wg := e.New(presentTime, presentTags)
			go func() { wg.Done() }()

			e.mu.RLock()
			success := e.waitFor(tt.waitForEventKey)
			e.mu.RUnlock()

			require.Equal(t, tt.wantSuccess, success)
		})
	}
}

func TestEvenWaitGroups_Wait(t *testing.T) {
	t.Parallel()

	e := NewEventWaitGroups()

	for i := range 5 {
		testTag := fmt.Sprintf("test%d", i)
		wg1 := e.New(time.Time{}, []string{testTag})
		wg2 := e.New(time.Time{}.Add(time.Second), []string{testTag})
		go func() { wg1.Done(); wg2.Done() }()
	}

	e.Wait()
}
