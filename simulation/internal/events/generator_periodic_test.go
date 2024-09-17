package events

import (
	"context"
	"github.com/metamogul/timestone/v2/internal"
	"testing"
	"time"

	"github.com/metamogul/timestone/v2"
	"github.com/stretchr/testify/require"
)

func Test_NewPeriodicGenerator(t *testing.T) {
	t.Parallel()

	type args struct {
		action   timestone.Action
		from     time.Time
		to       *time.Time
		interval time.Duration
		ctx      context.Context
		tags     []string
	}

	ctx := context.Background()

	tests := []struct {
		name         string
		args         args
		want         *PeriodicGenerator
		requirePanic bool
	}{
		{
			name: "no Action",
			args: args{
				action:   nil,
				from:     time.Time{},
				to:       internal.Ptr(time.Time{}.Add(time.Second)),
				interval: time.Second,
				ctx:      ctx,
			},
			requirePanic: true,
		},
		{
			name: "to before from",
			args: args{
				action:   timestone.NewMockAction(t),
				from:     time.Time{}.Add(time.Second),
				to:       internal.Ptr(time.Time{}),
				interval: time.Second,
				ctx:      ctx,
			},
			requirePanic: true,
		},
		{
			name: "to equals from",
			args: args{
				action:   timestone.NewMockAction(t),
				from:     time.Time{}.Add(time.Second),
				to:       internal.Ptr(time.Time{}.Add(time.Second)),
				interval: time.Second,
				ctx:      ctx,
			},
			requirePanic: true,
		},
		{
			name: "interval is zero",
			args: args{
				action:   timestone.NewMockAction(t),
				from:     time.Time{},
				to:       internal.Ptr(time.Time{}.Add(time.Second)),
				interval: 0,
				ctx:      ctx,
			},
			requirePanic: true,
		},
		{
			name: "interval is too long",
			args: args{
				action:   timestone.NewMockAction(t),
				from:     time.Time{},
				to:       internal.Ptr(time.Time{}.Add(time.Second)),
				interval: time.Second * 2,
				ctx:      ctx,
			},
			requirePanic: true,
		},
		{
			name: "success",
			args: args{
				action:   timestone.NewMockAction(t),
				from:     time.Time{},
				to:       internal.Ptr(time.Time{}.Add(2 * time.Second)),
				interval: time.Second,
				ctx:      ctx,
				tags:     []string{"test"},
			},
			want: &PeriodicGenerator{
				action:   timestone.NewMockAction(t),
				from:     time.Time{},
				to:       internal.Ptr(time.Time{}.Add(2 * time.Second)),
				interval: time.Second,
				tags:     []string{"test"},
				nextEvent: &Event{
					Action:  timestone.NewMockAction(t),
					Time:    time.Time{}.Add(time.Second),
					Context: ctx,
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
					_ = NewPeriodicGenerator(tt.args.ctx, tt.args.action, tt.args.from, tt.args.to, tt.args.interval, tt.args.tags)
				})
				return
			}

			newGenerator := NewPeriodicGenerator(tt.args.ctx, tt.args.action, tt.args.from, tt.args.to, tt.args.interval, tt.args.tags)
			require.Equal(t, tt.want, newGenerator)
		})
	}
}

func Test_PeriodicGenerator_Pop(t *testing.T) {
	t.Parallel()

	type fields struct {
		action       timestone.Action
		from         time.Time
		to           *time.Time
		interval     time.Duration
		currentEvent *Event
		ctx          context.Context
	}

	ctx := context.Background()

	tests := []struct {
		name            string
		fields          fields
		want            *Event
		requirePanic    bool
		requireFinished bool
	}{
		{
			name: "already finished",
			fields: fields{
				action:       timestone.NewMockAction(t),
				from:         time.Time{},
				to:           internal.Ptr(time.Time{}.Add(time.Minute)),
				interval:     10 * time.Second,
				currentEvent: NewEvent(ctx, timestone.NewMockAction(t), time.Time{}.Add(55*time.Second), []string{}),
				ctx:          context.Background(),
			},
			requirePanic: true,
		},
		{
			name: "success, not finished 1",
			fields: fields{
				action:       timestone.NewMockAction(t),
				from:         time.Time{},
				to:           nil,
				interval:     time.Second,
				currentEvent: NewEvent(ctx, timestone.NewMockAction(t), time.Time{}.Add(time.Second), []string{}),
				ctx:          context.Background(),
			},
			want: NewEvent(ctx, timestone.NewMockAction(t), time.Time{}.Add(time.Second), []string{}),
		},
		{
			name: "success, not finished 2",
			fields: fields{
				action:       timestone.NewMockAction(t),
				from:         time.Time{},
				to:           internal.Ptr(time.Time{}.Add(time.Minute)),
				interval:     10 * time.Second,
				currentEvent: NewEvent(ctx, timestone.NewMockAction(t), time.Time{}.Add(40*time.Second), []string{}),
				ctx:          context.Background(),
			},
			want: NewEvent(ctx, timestone.NewMockAction(t), time.Time{}.Add(40*time.Second), []string{}),
		},
		{
			name: "success, finished",
			fields: fields{
				action:       timestone.NewMockAction(t),
				from:         time.Time{},
				to:           internal.Ptr(time.Time{}.Add(time.Minute)),
				interval:     10 * time.Second,
				currentEvent: NewEvent(ctx, timestone.NewMockAction(t), time.Time{}.Add(50*time.Second), []string{}),
				ctx:          context.Background(),
			},
			want:            NewEvent(ctx, timestone.NewMockAction(t), time.Time{}.Add(50*time.Second), []string{}),
			requireFinished: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			p := &PeriodicGenerator{
				action:    tt.fields.action,
				from:      tt.fields.from,
				to:        tt.fields.to,
				interval:  tt.fields.interval,
				nextEvent: tt.fields.currentEvent,
				ctx:       tt.fields.ctx,
			}

			if tt.requirePanic {
				require.Panics(t, func() {
					_ = p.Pop()
				})
				return
			}

			require.Equal(t, tt.want, p.Pop())

			if tt.requireFinished {
				require.True(t, p.Finished())
			} else {
				require.False(t, p.Finished())
			}
		})
	}
}

func Test_PeriodicGenerator_Peek(t *testing.T) {
	t.Parallel()

	type fields struct {
		action       timestone.Action
		from         time.Time
		to           *time.Time
		interval     time.Duration
		currentEvent *Event
		ctx          context.Context
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
				action:       timestone.NewMockAction(t),
				from:         time.Time{},
				to:           internal.Ptr(time.Time{}.Add(time.Minute)),
				interval:     10 * time.Second,
				currentEvent: NewEvent(ctx, timestone.NewMockAction(t), time.Time{}.Add(55*time.Second), []string{}),
				ctx:          context.Background(),
			},
			requirePanic: true,
		},
		{
			name: "success",
			fields: fields{
				action:       timestone.NewMockAction(t),
				from:         time.Time{},
				to:           nil,
				interval:     time.Second,
				currentEvent: NewEvent(ctx, timestone.NewMockAction(t), time.Time{}.Add(time.Second), []string{}),
				ctx:          context.Background(),
			},
			want: *NewEvent(ctx, timestone.NewMockAction(t), time.Time{}.Add(time.Second), []string{}),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			p := &PeriodicGenerator{
				action:    tt.fields.action,
				from:      tt.fields.from,
				to:        tt.fields.to,
				interval:  tt.fields.interval,
				nextEvent: tt.fields.currentEvent,
				ctx:       tt.fields.ctx,
			}

			if tt.requirePanic {
				require.Panics(t, func() {
					_ = p.Peek()
				})
				return
			}

			require.Equal(t, tt.want, p.Peek())

			require.False(t, p.Finished())
		})
	}
}

func Test_PeriodicGenerator_Finished(t *testing.T) {
	t.Parallel()

	type fields struct {
		action       timestone.Action
		from         time.Time
		to           *time.Time
		interval     time.Duration
		currentEvent *Event
		ctx          context.Context
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	tests := []struct {
		name   string
		fields fields
		want   bool
	}{
		{
			name: "context is done",
			fields: fields{
				action:       timestone.NewMockAction(t),
				from:         time.Time{},
				to:           internal.Ptr(time.Time{}.Add(time.Minute)),
				interval:     10 * time.Second,
				currentEvent: NewEvent(ctx, timestone.NewMockAction(t), time.Time{}.Add(45*time.Second), []string{}),
				ctx:          ctx,
			},
			want: true,
		},
		{
			name: "to is nil",
			fields: fields{
				action:       timestone.NewMockAction(t),
				from:         time.Time{},
				to:           nil,
				interval:     0,
				currentEvent: NewEvent(ctx, timestone.NewMockAction(t), time.Time{}, []string{}),
				ctx:          context.Background(),
			},
			want: false,
		},
		{
			name: "to is set, finished",
			fields: fields{
				action:       timestone.NewMockAction(t),
				from:         time.Time{},
				to:           internal.Ptr(time.Time{}.Add(time.Minute)),
				interval:     10 * time.Second,
				currentEvent: NewEvent(ctx, timestone.NewMockAction(t), time.Time{}.Add(55*time.Second), []string{}),
				ctx:          context.Background(),
			},
			want: true,
		},
		{
			name: "to is set, not finished yet",
			fields: fields{
				action:       timestone.NewMockAction(t),
				from:         time.Time{},
				to:           internal.Ptr(time.Time{}.Add(time.Minute)),
				interval:     10 * time.Second,
				currentEvent: NewEvent(ctx, timestone.NewMockAction(t), time.Time{}.Add(45*time.Second), []string{}),
				ctx:          context.Background(),
			},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			p := &PeriodicGenerator{
				action:    tt.fields.action,
				from:      tt.fields.from,
				to:        tt.fields.to,
				interval:  tt.fields.interval,
				nextEvent: tt.fields.currentEvent,
				ctx:       tt.fields.ctx,
			}

			require.Equal(t, tt.want, p.Finished())
		})
	}
}
