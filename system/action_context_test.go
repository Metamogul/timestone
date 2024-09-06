package system

import (
	"context"
	"testing"

	timing "github.com/metamogul/timestone"
	"github.com/stretchr/testify/require"
)

func TestNewActionContext(t *testing.T) {
	t.Parallel()

	clock := Clock{}
	ctx := context.Background()

	a := newActionContext(ctx, clock)
	require.NotNil(t, a)
	require.NotNil(t, a.Context)
	require.NotNil(t, a.clock)
}

func TestActionContext_Clock(t *testing.T) {
	t.Parallel()

	clock := Clock{}

	a := newActionContext(context.Background(), clock)
	gotClock := a.Clock()
	require.Equal(t, clock, gotClock)
}

func TestActionContext_DoneSchedulingNewEvents(t *testing.T) {
	t.Parallel()

	a := newActionContext(context.Background(), Clock{})
	a.DoneSchedulingNewActions()
}

func TestActionContext_Value_Clock(t *testing.T) {
	t.Parallel()

	clock := Clock{}

	a := newActionContext(context.Background(), clock)
	gotClock := a.Value(timing.ActionContextClockKey)
	require.Equal(t, clock, gotClock)
}

func TestActionContext_Value_Default(t *testing.T) {
	t.Parallel()

	a := newActionContext(context.Background(), Clock{})
	gotValue := a.Value("someNoneExistentKey")
	require.Nil(t, gotValue)
}
