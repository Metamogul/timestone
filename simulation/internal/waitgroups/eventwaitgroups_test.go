package waitgroups

import (
	"fmt"
	"github.com/metamogul/timestone/v2/simulation/config"
	configinternal "github.com/metamogul/timestone/v2/simulation/internal/config"
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

	e.WaitFor(
		[]config.Event{
			config.At{
				Time: presentTime.Add(time.Second),
				Tags: presentTags,
			},
			config.At{
				Time: presentTime,
				Tags: presentTags,
			},
		},
	)

}

func TestEventWaitGroups_waitFor(t *testing.T) {
	t.Parallel()

	testcases := []struct {
		name            string
		presentTags     []string
		presentTime     time.Time
		waitForEventKey config.Event
		wantSuccess     bool
	}{
		{
			name:            "match relative time, no result for tags",
			presentTags:     []string{"test", "testGroup", "foo"},
			presentTime:     time.Time{},
			waitForEventKey: configinternal.At{Time: time.Time{}, Tags: []string{"baz"}},
			wantSuccess:     true,
		},
		{
			name:            "match time, no result for tags",
			presentTags:     []string{"test", "testGroup", "foo"},
			presentTime:     time.Time{},
			waitForEventKey: config.At{Time: time.Time{}, Tags: []string{"baz"}},
			wantSuccess:     false,
		},
		{
			name:            "match all times, no result for tags",
			presentTags:     []string{"test", "testGroup", "foo"},
			presentTime:     time.Time{},
			waitForEventKey: config.All{Tags: []string{"baz"}},
			wantSuccess:     false,
		},
		{
			name:            "match relative time, has no match",
			waitForEventKey: configinternal.At{Time: time.Time{}.Add(-1), Tags: []string{"test"}},
			wantSuccess:     true,
		},
		{
			name:            "match relative time, has match",
			presentTags:     []string{"test", "testGroup", "foo"},
			presentTime:     time.Time{},
			waitForEventKey: configinternal.At{Time: time.Time{}, Tags: []string{"test"}},
			wantSuccess:     true,
		},
		{
			name:            "match time, has no match",
			presentTags:     []string{"test", "testGroup", "foo"},
			presentTime:     time.Time{},
			waitForEventKey: config.At{Time: time.Now(), Tags: []string{"test"}},
			wantSuccess:     false,
		},
		{
			name:            "match time, has match",
			presentTags:     []string{"test", "testGroup", "foo"},
			presentTime:     time.Time{},
			waitForEventKey: config.At{Time: time.Time{}, Tags: []string{"test"}},
			wantSuccess:     true,
		},
		{
			name:            "match all",
			presentTags:     []string{"test", "testGroup", "foo"},
			presentTime:     time.Time{},
			waitForEventKey: config.All{Tags: []string{"test"}},
			wantSuccess:     true,
		},
	}

	for _, tt := range testcases {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			e := NewEventWaitGroups()

			if tt.presentTags != nil {
				wg := e.New(tt.presentTime, tt.presentTags)
				go func() { wg.Done() }()
			}

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
