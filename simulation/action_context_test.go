package simulation

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/metamogul/timestone"
	"github.com/stretchr/testify/require"
)

func TestNewActionContext(t *testing.T) {
	t.Parallel()

	now := time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC)

	newActionContext := newActionContext(context.Background(), newClock(now), &sync.WaitGroup{})
	require.NotNil(t, newActionContext)
	require.NotNil(t, newActionContext.Context)
	require.NotNil(t, newActionContext.clock)
	require.NotNil(t, newActionContext.eventLoopBlocker)
}

func TestActionContext_Clock(t *testing.T) {
	t.Parallel()

	now := time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC)
	clock := newClock(now)

	a := newActionContext(context.Background(), clock, &sync.WaitGroup{})
	gotClock := a.Clock()
	require.Equal(t, clock, gotClock)
}

func TestActionContext_DoneSchedulingNewActions_blockerIsNil(t *testing.T) {
	t.Parallel()

	now := time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC)

	a := newActionContext(context.Background(), newClock(now), nil)
	a.DoneSchedulingNewActions()
}

func TestActionContext_DoneSchedulingNewEvents_blockerNotNil(t *testing.T) {
	t.Parallel()

	now := time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC)
	eventLoopBlocker := &sync.WaitGroup{}

	a := newActionContext(context.Background(), newClock(now), eventLoopBlocker)

	eventLoopBlocker.Add(1)
	go func() {
		defer a.DoneSchedulingNewActions()
	}()
	eventLoopBlocker.Wait()
}

func TestActionContext_Value_Clock(t *testing.T) {
	t.Parallel()

	now := time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC)
	clock := newClock(now)

	a := newActionContext(context.Background(), clock, &sync.WaitGroup{})
	gotClock := a.Value(timestone.ActionContextClockKey)
	require.Equal(t, clock, gotClock)
}

func TestActionContext_Value_EventLoopBlocker(t *testing.T) {
	t.Parallel()

	now := time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC)
	eventLoopBlocker := &sync.WaitGroup{}

	a := newActionContext(context.Background(), newClock(now), eventLoopBlocker)
	gotEventLoopBlocker := a.Value(ActionContextEventLoopBlockerKey)
	require.Equal(t, eventLoopBlocker, gotEventLoopBlocker)
}

func TestActionContext_Value_Default(t *testing.T) {
	t.Parallel()

	now := time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC)

	a := newActionContext(context.Background(), newClock(now), &sync.WaitGroup{})
	gotValue := a.Value("someNoneExistentKey")
	require.Nil(t, gotValue)
}
