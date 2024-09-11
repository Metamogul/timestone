package simulation

import (
	"context"
	"testing"
	"time"

	"github.com/metamogul/timestone"
	"github.com/stretchr/testify/require"
)

func Test_newPeriodicEventGenerator(t *testing.T) {
	t.Parallel()

	type args struct {
		action   timestone.Action
		from     time.Time
		to       *time.Time
		interval time.Duration
		ctx      context.Context
	}

	ctx := context.Background()

	tests := []struct {
		name         string
		args         args
		want         *periodicEventGenerator
		requirePanic bool
	}{
		{
			name: "no Action",
			args: args{
				action:   nil,
				from:     time.Time{},
				to:       ptr(time.Time{}.Add(time.Second)),
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
				to:       ptr(time.Time{}),
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
				to:       ptr(time.Time{}.Add(time.Second)),
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
				to:       ptr(time.Time{}.Add(time.Second)),
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
				to:       ptr(time.Time{}.Add(time.Second)),
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
				to:       ptr(time.Time{}.Add(2 * time.Second)),
				interval: time.Second,
				ctx:      ctx,
			},
			want: &periodicEventGenerator{
				action:   timestone.NewMockAction(t),
				from:     time.Time{},
				to:       ptr(time.Time{}.Add(2 * time.Second)),
				interval: time.Second,
				nextEvent: &Event{
					Action:  timestone.NewMockAction(t),
					Time:    time.Time{}.Add(time.Second),
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
					_ = newPeriodicEventGenerator(tt.args.ctx, tt.args.action, tt.args.from, tt.args.to, tt.args.interval)
				})
				return
			}

			require.Equal(t, tt.want, newPeriodicEventGenerator(tt.args.ctx, tt.args.action, tt.args.from, tt.args.to, tt.args.interval))
		})
	}
}

func Test_periodicEventGenerator_Pop(t *testing.T) {
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
				to:           ptr(time.Time{}.Add(time.Minute)),
				interval:     10 * time.Second,
				currentEvent: NewEvent(ctx, timestone.NewMockAction(t), time.Time{}.Add(55*time.Second)),
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
				currentEvent: NewEvent(ctx, timestone.NewMockAction(t), time.Time{}.Add(time.Second)),
				ctx:          context.Background(),
			},
			want: NewEvent(ctx, timestone.NewMockAction(t), time.Time{}.Add(time.Second)),
		},
		{
			name: "success, not finished 2",
			fields: fields{
				action:       timestone.NewMockAction(t),
				from:         time.Time{},
				to:           ptr(time.Time{}.Add(time.Minute)),
				interval:     10 * time.Second,
				currentEvent: NewEvent(ctx, timestone.NewMockAction(t), time.Time{}.Add(40*time.Second)),
				ctx:          context.Background(),
			},
			want: NewEvent(ctx, timestone.NewMockAction(t), time.Time{}.Add(40*time.Second)),
		},
		{
			name: "success, finished",
			fields: fields{
				action:       timestone.NewMockAction(t),
				from:         time.Time{},
				to:           ptr(time.Time{}.Add(time.Minute)),
				interval:     10 * time.Second,
				currentEvent: NewEvent(ctx, timestone.NewMockAction(t), time.Time{}.Add(50*time.Second)),
				ctx:          context.Background(),
			},
			want:            NewEvent(ctx, timestone.NewMockAction(t), time.Time{}.Add(50*time.Second)),
			requireFinished: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			p := &periodicEventGenerator{
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

func Test_periodicEventGenerator_Peek(t *testing.T) {
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
				to:           ptr(time.Time{}.Add(time.Minute)),
				interval:     10 * time.Second,
				currentEvent: NewEvent(ctx, timestone.NewMockAction(t), time.Time{}.Add(55*time.Second)),
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
				currentEvent: NewEvent(ctx, timestone.NewMockAction(t), time.Time{}.Add(time.Second)),
				ctx:          context.Background(),
			},
			want: *NewEvent(ctx, timestone.NewMockAction(t), time.Time{}.Add(time.Second)),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			p := &periodicEventGenerator{
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

func Test_periodicEventGenerator_Finished(t *testing.T) {
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
				to:           ptr(time.Time{}.Add(time.Minute)),
				interval:     10 * time.Second,
				currentEvent: NewEvent(ctx, timestone.NewMockAction(t), time.Time{}.Add(45*time.Second)),
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
				currentEvent: NewEvent(ctx, timestone.NewMockAction(t), time.Time{}),
				ctx:          context.Background(),
			},
			want: false,
		},
		{
			name: "to is set, finished",
			fields: fields{
				action:       timestone.NewMockAction(t),
				from:         time.Time{},
				to:           ptr(time.Time{}.Add(time.Minute)),
				interval:     10 * time.Second,
				currentEvent: NewEvent(ctx, timestone.NewMockAction(t), time.Time{}.Add(55*time.Second)),
				ctx:          context.Background(),
			},
			want: true,
		},
		{
			name: "to is set, not finished yet",
			fields: fields{
				action:       timestone.NewMockAction(t),
				from:         time.Time{},
				to:           ptr(time.Time{}.Add(time.Minute)),
				interval:     10 * time.Second,
				currentEvent: NewEvent(ctx, timestone.NewMockAction(t), time.Time{}.Add(45*time.Second)),
				ctx:          context.Background(),
			},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			p := &periodicEventGenerator{
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
