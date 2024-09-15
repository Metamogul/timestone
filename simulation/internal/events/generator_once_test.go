package events

import (
	"context"
	"testing"
	"time"

	"github.com/metamogul/timestone"
	"github.com/stretchr/testify/require"
)

func Test_NewOnceGenerator(t *testing.T) {
	t.Parallel()

	type args struct {
		ctx        context.Context
		action     timestone.Action
		actionTime time.Time
		tags       []string
	}

	ctx := context.Background()

	tests := []struct {
		name         string
		args         args
		want         *OnceGenerator
		requirePanic bool
	}{
		{
			name: "no Action",
			args: args{
				ctx:        ctx,
				action:     nil,
				actionTime: time.Time{},
				tags:       []string{"test"},
			},
			requirePanic: true,
		},
		{
			name: "success",
			args: args{
				ctx:        ctx,
				action:     timestone.NewMockAction(t),
				actionTime: time.Time{},
				tags:       []string{"test"},
			},
			want: &OnceGenerator{
				event: &Event{
					Context: ctx,
					Action:  timestone.NewMockAction(t),
					Time:    time.Time{},
					tags:    []string{"test"},
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
					_ = NewOnceGenerator(tt.args.ctx, tt.args.action, tt.args.actionTime, tt.args.tags)
				})
				return
			}

			require.Equal(t, tt.want, NewOnceGenerator(tt.args.ctx, tt.args.action, tt.args.actionTime, tt.args.tags))
		})
	}
}

func Test_OnceGenerator_Pop(t *testing.T) {
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
				event: NewEvent(ctx, timestone.NewMockAction(t), time.Time{}, []string{"test"}),
				ctx:   ctx,
			},
			want: NewEvent(ctx, timestone.NewMockAction(t), time.Time{}, []string{"test"}),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			o := &OnceGenerator{
				event: tt.fields.event,
				ctx:   tt.fields.ctx,
			}

			if tt.requirePanic {
				require.Panics(t, func() {
					_ = o.Pop()
				})
				return
			}

			require.Equal(t, tt.want, o.Pop())

			if tt.want != nil {
				require.True(t, o.Finished())
			} else {
				require.False(t, o.Finished())
			}
		})
	}
}

func Test_OnceGenerator_Peek(t *testing.T) {
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
				event: NewEvent(ctx, timestone.NewMockAction(t), time.Time{}, []string{"test"}),
				ctx:   ctx,
			},
			want: *NewEvent(ctx, timestone.NewMockAction(t), time.Time{}, []string{"test"}),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			o := &OnceGenerator{
				event: tt.fields.event,
				ctx:   tt.fields.ctx,
			}

			if tt.requirePanic {
				require.Panics(t, func() {
					_ = o.Peek()
				})
				return
			}

			require.Equal(t, tt.want, o.Peek())
			require.False(t, o.Finished())
		})
	}
}

func Test_OnceGenerator_Finished(t *testing.T) {
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
				event: NewEvent(ctx, timestone.NewMockAction(t), time.Time{}, []string{"test"}),
				ctx:   ctx,
			},
			want: true,
		},
		{
			name: "not finished",
			fields: fields{
				event: NewEvent(ctx, timestone.NewMockAction(t), time.Time{}, []string{"test"}),
				ctx:   context.Background(),
			},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			o := &OnceGenerator{
				event: tt.fields.event,
				ctx:   tt.fields.ctx,
			}

			require.Equal(t, tt.want, o.Finished())
		})
	}
}
