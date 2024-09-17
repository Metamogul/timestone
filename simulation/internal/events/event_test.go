package events

import (
	"context"
	"testing"
	"time"

	"github.com/metamogul/timestone/v2"
	"github.com/stretchr/testify/require"
)

func Test_NewEvent(t *testing.T) {
	t.Parallel()

	type args struct {
		ctx        context.Context
		action     timestone.Action
		actionTime time.Time
		tags       []string
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
			name: "success, no tags provided",
			args: args{
				ctx:        context.Background(),
				action:     timestone.NewMockAction(t),
				actionTime: time.Time{},
				tags:       nil,
			},
			want: &Event{
				Context: context.Background(),
				Action:  timestone.NewMockAction(t),
				Time:    time.Time{},
				tags:    []string{DefaultTag},
			},
		},
		{
			name: "success, tags empty",
			args: args{
				ctx:        context.Background(),
				action:     timestone.NewMockAction(t),
				actionTime: time.Time{},
				tags:       []string{},
			},
			want: &Event{
				Context: context.Background(),
				Action:  timestone.NewMockAction(t),
				Time:    time.Time{},
				tags:    []string{DefaultTag},
			},
		},
		{
			name: "success",
			args: args{
				ctx:        context.Background(),
				action:     timestone.NewMockAction(t),
				actionTime: time.Time{},
				tags:       []string{"foo", "bar"},
			},
			want: &Event{
				Context: context.Background(),
				Action:  timestone.NewMockAction(t),
				Time:    time.Time{},
				tags:    []string{"foo", "bar"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if tt.requirePanic {
				require.Panics(t, func() {
					_ = NewEvent(tt.args.ctx, tt.args.action, tt.args.actionTime, tt.args.tags)
				})
				return
			}

			require.Equal(t, tt.want, NewEvent(tt.args.ctx, tt.args.action, tt.args.actionTime, tt.args.tags))
		})
	}
}

func Test_Event_Tags(t *testing.T) {
	t.Parallel()

	e := NewEvent(context.Background(), timestone.NewMockAction(t), time.Now(), []string{"test1", "test2"})
	require.Equal(t, []string{"test1", "test2"}, e.Tags())
}
