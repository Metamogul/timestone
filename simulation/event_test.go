package simulation

import (
	"context"
	"testing"
	"time"

	"github.com/metamogul/timestone"
	"github.com/stretchr/testify/require"
)

func Test_newEvent(t *testing.T) {
	t.Parallel()

	type args struct {
		action     timestone.Action
		actionTime time.Time
		ctx        context.Context
	}

	tests := []struct {
		name         string
		args         args
		want         *Event
		requirePanic bool
	}{
		{
			name: "no Action",
			args: args{
				action:     nil,
				actionTime: time.Time{},
			},
			requirePanic: true,
		},
		{
			name: "success",
			args: args{
				action:     timestone.NewMockAction(t),
				actionTime: time.Time{},
				ctx:        context.Background(),
			},
			want: &Event{
				Action:  timestone.NewMockAction(t),
				Time:    time.Time{},
				Context: context.Background(),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if tt.requirePanic {
				require.Panics(t, func() {
					_ = NewEvent(tt.args.ctx, tt.args.action, tt.args.actionTime)
				})
				return
			}

			require.Equal(t, tt.want, NewEvent(tt.args.ctx, tt.args.action, tt.args.actionTime))
		})
	}
}

func Test_event_perform(t *testing.T) {
	t.Parallel()

	actionContextArg := context.WithValue(context.Background(), timestone.ActionContextClockKey, newClock(time.Now()))

	e := &Event{
		Action: func() timestone.Action {
			mockedAction := timestone.NewMockAction(t)
			mockedAction.EXPECT().
				Perform(actionContextArg).
				Once()

			return mockedAction
		}(),
		Time: time.Time{},
	}

	e.Perform(actionContextArg)
}
