package simulation

import (
	"context"
	"testing"
	"time"

	"github.com/metamogul/timestone"
	"github.com/stretchr/testify/require"
)

func Test_newSingleEventGenerator(t *testing.T) {
	t.Parallel()

	type args struct {
		action     timestone.Action
		actionTime time.Time
		ctx        context.Context
	}

	ctx := context.Background()

	tests := []struct {
		name         string
		args         args
		want         *singleEventGenerator
		requirePanic bool
	}{
		{
			name: "no Action",
			args: args{
				action:     nil,
				actionTime: time.Time{},
				ctx:        ctx,
			},
			requirePanic: true,
		},
		{
			name: "success",
			args: args{
				action:     timestone.NewMockAction(t),
				actionTime: time.Time{},
				ctx:        ctx,
			},
			want: &singleEventGenerator{
				Event: &Event{
					Action:  timestone.NewMockAction(t),
					Time:    time.Time{},
					Context: ctx,
				},
				ctx: ctx,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if tt.requirePanic {
				require.Panics(t, func() {
					_ = newSingleEventGenerator(tt.args.ctx, tt.args.action, tt.args.actionTime)
				})
				return
			}

			require.Equal(t, tt.want, newSingleEventGenerator(tt.args.ctx, tt.args.action, tt.args.actionTime))
		})
	}
}

func Test_singleEventStream_pop(t *testing.T) {
	t.Parallel()

	type fields struct {
		event *Event
		ctx   context.Context
	}

	ctx := context.Background()

	tests := []struct {
		name         string
		fields       fields
		want         *Event
		requirePanic bool
	}{
		{
			name: "already finished",
			fields: fields{
				event: nil,
				ctx:   ctx,
			},
			requirePanic: true,
		},
		{
			name: "success",
			fields: fields{
				event: NewEvent(ctx, timestone.NewMockAction(t), time.Time{}),
				ctx:   ctx,
			},
			want: NewEvent(ctx, timestone.NewMockAction(t), time.Time{}),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			s := &singleEventGenerator{
				Event: tt.fields.event,
				ctx:   tt.fields.ctx,
			}

			if tt.requirePanic {
				require.Panics(t, func() {
					_ = s.Pop()
				})
				return
			}

			require.Equal(t, tt.want, s.Pop())

			if tt.want != nil {
				require.True(t, s.Finished())
			} else {
				require.False(t, s.Finished())
			}
		})
	}
}

func Test_singleEventStream_peek(t *testing.T) {
	t.Parallel()

	type fields struct {
		event *Event
		ctx   context.Context
	}

	ctx := context.Background()

	tests := []struct {
		name         string
		fields       fields
		want         Event
		requirePanic bool
	}{
		{
			name: "already finished",
			fields: fields{
				event: nil,
				ctx:   ctx,
			},
			requirePanic: true,
		},
		{
			name: "success",
			fields: fields{
				event: NewEvent(ctx, timestone.NewMockAction(t), time.Time{}),
				ctx:   ctx,
			},
			want: *NewEvent(ctx, timestone.NewMockAction(t), time.Time{}),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			s := &singleEventGenerator{
				Event: tt.fields.event,
				ctx:   tt.fields.ctx,
			}

			if tt.requirePanic {
				require.Panics(t, func() {
					_ = s.Peek()
				})
				return
			}

			require.Equal(t, tt.want, s.Peek())
			require.False(t, s.Finished())
		})
	}
}

func Test_singleEventStream_finished(t *testing.T) {
	t.Parallel()

	type fields struct {
		event *Event
		ctx   context.Context
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	tests := []struct {
		name   string
		fields fields
		want   bool
	}{
		{
			name: "no event",
			fields: fields{
				event: nil,
				ctx:   context.Background(),
			},
			want: true,
		},
		{
			name: "context is done",
			fields: fields{
				event: NewEvent(ctx, timestone.NewMockAction(t), time.Time{}),
				ctx:   ctx,
			},
			want: true,
		},
		{
			name: "not finished",
			fields: fields{
				event: NewEvent(ctx, timestone.NewMockAction(t), time.Time{}),
				ctx:   context.Background(),
			},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			s := &singleEventGenerator{
				Event: tt.fields.event,
				ctx:   tt.fields.ctx,
			}

			require.Equal(t, tt.want, s.Finished())
		})
	}
}
