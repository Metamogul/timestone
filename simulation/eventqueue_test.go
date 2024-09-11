package simulation

import (
	"context"
	"slices"
	"testing"
	"time"

	"github.com/metamogul/timestone"
	"github.com/stretchr/testify/require"
)

func Test_newEventQueue(t *testing.T) {
	t.Parallel()

	got := newEventQueue(newEventConfigurations())

	require.NotNil(t, got.configs)
	require.NotNil(t, got.activeGenerators)
	require.NotNil(t, got.finishedGenerators)
	require.NotNil(t, got.newGeneratorsWaitGroups)

	require.Len(t, got.activeGenerators, 0)
	require.Len(t, got.finishedGenerators, 0)
	require.Len(t, got.newGeneratorsWaitGroups.waitGroups, 0)

}

func Test_eventQueue_add(t *testing.T) {
	t.Parallel()

	now := time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC)

	tests := []struct {
		name                string
		generator           func() EventGenerator
		generatorIsFinished bool
	}{
		{
			name: "generator finished",
			generator: func() EventGenerator {
				mockEventGenerator := NewMockEventGenerator(t)
				mockEventGenerator.EXPECT().
					Finished().
					Return(true).
					Once()

				return mockEventGenerator
			},
			generatorIsFinished: true,
		},
		{
			name: "generator not finished",
			generator: func() EventGenerator {
				mockEventGenerator := NewMockEventGenerator(t)
				mockEventGenerator.EXPECT().
					Finished().
					Return(false).
					Once()
				mockEventGenerator.EXPECT().
					Peek().
					Return(
						*NewEvent(
							context.Background(),
							timestone.NewSimpleAction(func(context.Context) {}, "test"),
							now,
						),
					).
					Once()

				return mockEventGenerator
			},
			generatorIsFinished: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			e := newEventQueue(newEventConfigurations())

			e.add(tt.generator())

			if !tt.generatorIsFinished {
				require.Len(t, e.activeGenerators, 1)
				require.Len(t, e.finishedGenerators, 0)
				require.Len(t, e.newGeneratorsWaitGroups.waitGroups, 1)
				require.Equal(t, e.newGeneratorsWaitGroups.waitGroups["test"].count, 0)
			} else {
				require.Len(t, e.activeGenerators, 0)
				require.Len(t, e.finishedGenerators, 1)
				require.Len(t, e.newGeneratorsWaitGroups.waitGroups, 0)
			}

			sorted := slices.IsSortedFunc(e.activeGenerators, func(a, b EventGenerator) int {
				return a.Peek().Time.Compare(b.Peek().Time)
			})
			require.True(t, sorted)
		})
	}
}

func Test_eventQueue_pop(t *testing.T) {
	t.Parallel()

	type fields struct {
		activeGenerators   func() []EventGenerator
		finishedGenerators func() []EventGenerator
	}

	ctx := context.Background()

	newMockAction := func(t *testing.T, name string) *timestone.MockAction {
		mockAction := timestone.NewMockAction(t)
		mockAction.EXPECT().
			Name().
			Return(name).
			Maybe()
		return mockAction
	}

	tests := []struct {
		name              string
		fields            fields
		finishesGenerator bool
		want              *Event
		requirePanic      bool
	}{
		{
			name: "all generators finished",
			fields: fields{
				activeGenerators: func() []EventGenerator {
					return make([]EventGenerator, 0)
				},
				finishedGenerators: func() []EventGenerator {
					return make([]EventGenerator, 0)
				},
			},
			requirePanic: true,
		},
		{
			name: "success, generator not finished",
			fields: fields{
				activeGenerators: func() []EventGenerator {
					eventGenerator1 := newPeriodicEventGenerator(ctx, newMockAction(t, "test1"), time.Time{}, nil, time.Minute)
					eventGenerator2 := newPeriodicEventGenerator(ctx, newMockAction(t, "test2"), time.Time{}, nil, time.Second)
					return []EventGenerator{eventGenerator1, eventGenerator2}
				},
				finishedGenerators: func() []EventGenerator {
					return make([]EventGenerator, 0)
				},
			},
			finishesGenerator: false,
			want: &Event{
				Action:  newMockAction(t, "test2"),
				Time:    time.Time{}.Add(time.Second),
				Context: ctx,
			},
		},
		{
			name: "success, generator finished",
			fields: fields{
				activeGenerators: func() []EventGenerator {
					eventGenerator1 := newSingleEventGenerator(context.Background(), newMockAction(t, "test1"), time.Time{})
					eventGenerator2 := newPeriodicEventGenerator(context.Background(), newMockAction(t, "test2"), time.Time{}, nil, time.Second)
					return []EventGenerator{eventGenerator1, eventGenerator2}
				},
				finishedGenerators: func() []EventGenerator {
					return make([]EventGenerator, 0)
				},
			},
			finishesGenerator: true,
			want: &Event{
				Action:  newMockAction(t, "test1"),
				Time:    time.Time{},
				Context: ctx,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			e := &eventQueue{
				activeGenerators:   tt.fields.activeGenerators(),
				finishedGenerators: tt.fields.finishedGenerators(),
			}
			e.sortActiveGenerators()

			if tt.requirePanic {
				require.Panics(t, func() {
					_ = e.Pop()
				})
				return
			}

			require.Equal(t, tt.want.Name(), e.Pop().Name())

			if !tt.finishesGenerator {
				require.Len(t, e.activeGenerators, len(tt.fields.activeGenerators()))
				require.Len(t, e.finishedGenerators, len(tt.fields.finishedGenerators()))
			} else {
				require.Len(t, e.activeGenerators, len(tt.fields.activeGenerators())-1)
				require.Len(t, e.finishedGenerators, len(tt.fields.finishedGenerators())+1)
			}
		})
	}
}

func Test_eventQueue_peek(t *testing.T) {
	t.Parallel()

	type fields struct {
		activeGenerators   func() []EventGenerator
		finishedGenerators func() []EventGenerator
	}

	ctx := context.Background()

	tests := []struct {
		name         string
		fields       fields
		want         Event
		requirePanic bool
	}{
		{
			name: "all generators finished",
			fields: fields{
				activeGenerators: func() []EventGenerator {
					return make([]EventGenerator, 0)
				},
				finishedGenerators: func() []EventGenerator {
					return make([]EventGenerator, 0)
				},
			},
			requirePanic: true,
		},
		{
			name: "success",
			fields: fields{
				activeGenerators: func() []EventGenerator {
					eventGenerator1 := newPeriodicEventGenerator(ctx, timestone.NewMockAction(t), time.Time{}, nil, time.Minute)
					eventGenerator2 := newPeriodicEventGenerator(ctx, timestone.NewMockAction(t), time.Time{}, nil, time.Second)
					return []EventGenerator{eventGenerator1, eventGenerator2}
				},
				finishedGenerators: func() []EventGenerator {
					return make([]EventGenerator, 0)
				},
			},
			want: Event{
				Action:  timestone.NewMockAction(t),
				Time:    time.Time{}.Add(time.Second),
				Context: ctx,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			e := &eventQueue{
				activeGenerators:   tt.fields.activeGenerators(),
				finishedGenerators: tt.fields.finishedGenerators(),
			}
			e.sortActiveGenerators()

			if tt.requirePanic {
				require.Panics(t, func() {
					_ = e.Peek()
				})
				return
			}

			require.Equal(t, tt.want, e.Peek())
			require.Len(t, e.activeGenerators, len(tt.fields.activeGenerators()))
			require.Len(t, e.finishedGenerators, len(tt.fields.finishedGenerators()))

		})
	}
}

func Test_eventQueue_finished(t *testing.T) {
	t.Parallel()

	type fields struct {
		activeGenerators   []EventGenerator
		finishedGenerators []EventGenerator
	}

	tests := []struct {
		name   string
		fields fields
		want   bool
	}{
		{
			name: "not finished",
			fields: fields{
				activeGenerators:   []EventGenerator{NewMockEventGenerator(t)},
				finishedGenerators: make([]EventGenerator, 0),
			},
			want: false,
		},
		{
			name: "finished",
			fields: fields{
				activeGenerators:   make([]EventGenerator, 0),
				finishedGenerators: make([]EventGenerator, 0),
			},
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			e := &eventQueue{
				activeGenerators:   tt.fields.activeGenerators,
				finishedGenerators: tt.fields.finishedGenerators,
			}

			require.Equal(t, tt.want, e.Finished())
		})
	}
}

func Test_eventQueue_sortActiveGeneratos(t *testing.T) {
	t.Parallel()

	eventGenerator1 := newPeriodicEventGenerator(context.Background(), timestone.NewMockAction(t), time.Time{}, nil, time.Minute)
	eventGenerator2 := newPeriodicEventGenerator(context.Background(), timestone.NewMockAction(t), time.Time{}, nil, time.Second)
	eventGenerator3 := newPeriodicEventGenerator(context.Background(), timestone.NewMockAction(t), time.Time{}, nil, time.Hour)

	activeGenerators := []EventGenerator{eventGenerator1, eventGenerator2, eventGenerator3}

	e := &eventQueue{
		activeGenerators:   activeGenerators,
		finishedGenerators: make([]EventGenerator, 0),
	}
	e.sortActiveGenerators()

	sorted := slices.IsSortedFunc(e.activeGenerators, func(a, b EventGenerator) int {
		return a.Peek().Time.Compare(b.Peek().Time)
	})
	require.True(t, sorted)
}
