package events

import (
	"context"
	"github.com/metamogul/timestone/simulation/event"
	"slices"
	"testing"
	"time"

	"github.com/metamogul/timestone"
	"github.com/stretchr/testify/require"
)

func Test_NewQueue(t *testing.T) {
	t.Parallel()

	got := NewQueue(NewConfigs())

	require.NotNil(t, got.configs)
	require.NotNil(t, got.activeGenerators)
	require.NotNil(t, got.finishedGenerators)
	require.NotNil(t, got.NewGeneratorsWaitGroups)

	require.Len(t, got.activeGenerators, 0)
	require.Len(t, got.finishedGenerators, 0)

}

func TestQueue_Add(t *testing.T) {
	t.Parallel()

	now := time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC)

	tests := []struct {
		name                string
		generator           func() Generator
		generatorIsFinished bool
	}{
		{
			name: "generator finished",
			generator: func() Generator {
				mockEventGenerator := NewMockGenerator(t)
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
			generator: func() Generator {
				mockEventGenerator := NewMockGenerator(t)
				mockEventGenerator.EXPECT().
					Finished().
					Return(false).
					Once()
				mockEventGenerator.EXPECT().
					Peek().
					Return(
						*NewEvent(
							context.Background(),
							timestone.SimpleAction(func(context.Context) {}),
							now,
							[]string{"test"},
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

			e := NewQueue(NewConfigs())

			e.Add(tt.generator())

			if !tt.generatorIsFinished {
				require.Len(t, e.activeGenerators, 1)
				require.Len(t, e.finishedGenerators, 0)
				e.NewGeneratorsWaitGroups.Add(1, []string{"test"})
				go func() { e.NewGeneratorsWaitGroups.Done([]string{"test"}) }()
				e.NewGeneratorsWaitGroups.WaitFor([]string{"test"})
			} else {
				require.Len(t, e.activeGenerators, 0)
				require.Len(t, e.finishedGenerators, 1)
			}

			sorted := slices.IsSortedFunc(e.activeGenerators, func(a, b Generator) int {
				return a.Peek().Time.Compare(b.Peek().Time)
			})
			require.True(t, sorted)
		})
	}
}

func TestQueue_ExpectGenerators(t *testing.T) {
	t.Parallel()

	now := time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC)

	generatorMock := NewMockGenerator(t)
	generatorMock.EXPECT().
		Finished().
		Return(false).
		Once()
	generatorMock.EXPECT().
		Peek().
		Return(
			*NewEvent(
				context.Background(),
				timestone.SimpleAction(func(context.Context) {}),
				now,
				[]string{"test", "group", "foo"},
			),
		).
		Once()

	e := NewQueue(NewConfigs())

	generatorExpectations := []*event.GeneratorExpectation{{Tags: []string{"test"}, Count: 1}}

	e.ExpectGenerators(generatorExpectations)
	go func() { e.Add(generatorMock) }()
	e.WaitForExpectedGenerators(generatorExpectations)
}

func TestQueue_WaitForExpectedGenerators(t *testing.T) {
	t.Parallel()

	now := time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC)

	generatorMock := NewMockGenerator(t)
	generatorMock.EXPECT().
		Finished().
		Return(false).
		Once()
	generatorMock.EXPECT().
		Peek().
		Return(
			*NewEvent(
				context.Background(),
				timestone.SimpleAction(func(context.Context) {}),
				now,
				[]string{"test", "group", "foo"},
			),
		).
		Once()

	e := NewQueue(NewConfigs())

	generatorExpectations := []*event.GeneratorExpectation{{Tags: []string{"test"}, Count: 1}}

	e.ExpectGenerators(generatorExpectations)
	go func() { e.Add(generatorMock) }()
	e.WaitForExpectedGenerators(generatorExpectations)
}

func TestQueue_Pop(t *testing.T) {
	t.Parallel()

	type fields struct {
		activeGenerators   func() []Generator
		finishedGenerators func() []Generator
	}

	ctx := context.Background()

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
				activeGenerators: func() []Generator {
					return make([]Generator, 0)
				},
				finishedGenerators: func() []Generator {
					return make([]Generator, 0)
				},
			},
			requirePanic: true,
		},
		{
			name: "success, generator not finished",
			fields: fields{
				activeGenerators: func() []Generator {
					eventGenerator1 := NewPeriodicGenerator(ctx, timestone.NewMockAction(t), time.Time{}, nil, time.Minute, []string{"test1"})
					eventGenerator2 := NewPeriodicGenerator(ctx, timestone.NewMockAction(t), time.Time{}, nil, time.Second, []string{"test2"})
					return []Generator{eventGenerator1, eventGenerator2}
				},
				finishedGenerators: func() []Generator {
					return make([]Generator, 0)
				},
			},
			finishesGenerator: false,
			want: &Event{
				Action:  timestone.NewMockAction(t),
				Time:    time.Time{}.Add(time.Second),
				Context: ctx,
				tags:    []string{"test2"},
			},
		},
		{
			name: "success, generator finished",
			fields: fields{
				activeGenerators: func() []Generator {
					eventGenerator1 := NewOnceGenerator(context.Background(), timestone.NewMockAction(t), time.Time{}, []string{"test1"})
					eventGenerator2 := NewPeriodicGenerator(context.Background(), timestone.NewMockAction(t), time.Time{}, nil, time.Second, []string{"test2"})
					return []Generator{eventGenerator1, eventGenerator2}
				},
				finishedGenerators: func() []Generator {
					return make([]Generator, 0)
				},
			},
			finishesGenerator: true,
			want: &Event{
				Action:  timestone.NewMockAction(t),
				Time:    time.Time{},
				Context: ctx,
				tags:    []string{"test1"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			e := &Queue{
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

			require.Equal(t, tt.want.tags, e.Pop().tags)

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

func TestQueue_Peek(t *testing.T) {
	t.Parallel()

	type fields struct {
		activeGenerators   func() []Generator
		finishedGenerators func() []Generator
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
				activeGenerators: func() []Generator {
					return make([]Generator, 0)
				},
				finishedGenerators: func() []Generator {
					return make([]Generator, 0)
				},
			},
			requirePanic: true,
		},
		{
			name: "success",
			fields: fields{
				activeGenerators: func() []Generator {
					eventGenerator1 := NewPeriodicGenerator(ctx, timestone.NewMockAction(t), time.Time{}, nil, time.Minute, []string{"test1"})
					eventGenerator2 := NewPeriodicGenerator(ctx, timestone.NewMockAction(t), time.Time{}, nil, time.Second, []string{"test2"})
					return []Generator{eventGenerator1, eventGenerator2}
				},
				finishedGenerators: func() []Generator {
					return make([]Generator, 0)
				},
			},
			want: Event{
				Action:  timestone.NewMockAction(t),
				Time:    time.Time{}.Add(time.Second),
				Context: ctx,
				tags:    []string{"test2"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			e := &Queue{
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

func TestQueue_Finished(t *testing.T) {
	t.Parallel()

	type fields struct {
		activeGenerators   []Generator
		finishedGenerators []Generator
	}

	tests := []struct {
		name   string
		fields fields
		want   bool
	}{
		{
			name: "not finished",
			fields: fields{
				activeGenerators:   []Generator{NewMockGenerator(t)},
				finishedGenerators: make([]Generator, 0),
			},
			want: false,
		},
		{
			name: "finished",
			fields: fields{
				activeGenerators:   make([]Generator, 0),
				finishedGenerators: make([]Generator, 0),
			},
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			e := &Queue{
				activeGenerators:   tt.fields.activeGenerators,
				finishedGenerators: tt.fields.finishedGenerators,
			}

			require.Equal(t, tt.want, e.Finished())
		})
	}
}

func TestQueue_sortActiveGenerators(t *testing.T) {
	t.Parallel()

	eventGenerator1 := NewPeriodicGenerator(context.Background(), timestone.NewMockAction(t), time.Time{}, nil, time.Minute, []string{})
	eventGenerator2 := NewPeriodicGenerator(context.Background(), timestone.NewMockAction(t), time.Time{}, nil, time.Second, []string{})
	eventGenerator3 := NewPeriodicGenerator(context.Background(), timestone.NewMockAction(t), time.Time{}, nil, time.Hour, []string{})

	activeGenerators := []Generator{eventGenerator1, eventGenerator2, eventGenerator3}

	e := &Queue{
		activeGenerators:   activeGenerators,
		finishedGenerators: make([]Generator, 0),
	}
	e.sortActiveGenerators()

	sorted := slices.IsSortedFunc(e.activeGenerators, func(a, b Generator) int {
		return a.Peek().Time.Compare(b.Peek().Time)
	})
	require.True(t, sorted)
}
