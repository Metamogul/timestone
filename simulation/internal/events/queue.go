package events

import (
	"github.com/metamogul/timestone/v2/simulation/config"
	"github.com/metamogul/timestone/v2/simulation/internal/waitgroups"
	"slices"
)

type Queue struct {
	configs            *Configs
	activeGenerators   []Generator
	finishedGenerators []Generator

	NewGeneratorsWaitGroups *waitgroups.GeneratorWaitGroups
}

func NewQueue(configs *Configs) *Queue {
	queue := &Queue{
		configs:                 configs,
		activeGenerators:        make([]Generator, 0),
		finishedGenerators:      make([]Generator, 0),
		NewGeneratorsWaitGroups: waitgroups.NewGeneratorWaitGroups(),
	}

	return queue
}

func (q *Queue) Add(generator Generator) {
	if generator.Finished() {
		q.finishedGenerators = append(q.finishedGenerators, generator)
		return
	}

	q.activeGenerators = append(q.activeGenerators, generator)

	generatorEventTags := generator.Peek().tags
	q.NewGeneratorsWaitGroups.Done(generatorEventTags)

	q.sortActiveGenerators()
}

func (q *Queue) ExpectGenerators(expectedGenerators []*config.Generator) {
	for _, expectation := range expectedGenerators {
		q.NewGeneratorsWaitGroups.Add(expectation.Count, expectation.Tags)
	}
}

func (q *Queue) WaitForExpectedGenerators(expectedGenerators []*config.Generator) {
	for _, expectedGenerator := range expectedGenerators {
		q.NewGeneratorsWaitGroups.WaitFor(expectedGenerator.Tags)
	}
}

func (q *Queue) Pop() *Event {
	if q.Finished() {
		panic(ErrGeneratorFinished)
	}

	nextEvent := q.activeGenerators[0].Pop()

	if q.activeGenerators[0].Finished() {
		q.finishedGenerators = append(q.finishedGenerators, q.activeGenerators[0])
		q.activeGenerators = q.activeGenerators[1:]
	}

	q.sortActiveGenerators()

	return nextEvent
}

func (q *Queue) Peek() Event {
	if q.Finished() {
		panic(ErrGeneratorFinished)
	}

	return q.activeGenerators[0].Peek()
}

func (q *Queue) Finished() bool {
	return len(q.activeGenerators) == 0
}

func (q *Queue) sortActiveGenerators() {
	slices.SortStableFunc(q.activeGenerators, func(a, b Generator) int {
		eventA, eventB := a.Peek(), b.Peek()

		if timeComparison := eventA.Time.Compare(eventB.Time); timeComparison != 0 {
			return timeComparison
		}

		priorityComparison := q.configs.Priority(&eventA) - q.configs.Priority(&eventB)

		return priorityComparison
	})
}
