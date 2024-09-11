package simulation

import (
	"slices"
)

type eventQueue struct {
	configs            *eventConfigurations
	activeGenerators   []EventGenerator
	finishedGenerators []EventGenerator

	newGeneratorsWaitGroups *waitGroups
}

func newEventQueue(configs *eventConfigurations) *eventQueue {
	queue := &eventQueue{
		configs:                 configs,
		activeGenerators:        make([]EventGenerator, 0),
		finishedGenerators:      make([]EventGenerator, 0),
		newGeneratorsWaitGroups: newWaitGroups(),
	}

	return queue
}

func (e *eventQueue) add(generator EventGenerator) {
	if generator.Finished() {
		e.finishedGenerators = append(e.finishedGenerators, generator)
		return
	}

	e.activeGenerators = append(e.activeGenerators, generator)

	generatorEventName := generator.Peek().Name()
	e.newGeneratorsWaitGroups.new(generatorEventName)
	e.newGeneratorsWaitGroups.done(generatorEventName)

	e.sortActiveGenerators()
}

func (e *eventQueue) Pop() *Event {
	if e.Finished() {
		panic(ErrEventGeneratorFinished)
	}

	nextEvent := e.activeGenerators[0].Pop()

	if e.activeGenerators[0].Finished() {
		e.finishedGenerators = append(e.finishedGenerators, e.activeGenerators[0])
		e.activeGenerators = e.activeGenerators[1:]
	}

	e.sortActiveGenerators()

	return nextEvent
}

func (e *eventQueue) Peek() Event {
	if e.Finished() {
		panic(ErrEventGeneratorFinished)
	}

	return e.activeGenerators[0].Peek()
}

func (e *eventQueue) Finished() bool {
	return len(e.activeGenerators) == 0
}

func (e *eventQueue) sortActiveGenerators() {
	slices.SortStableFunc(e.activeGenerators, func(a, b EventGenerator) int {
		eventA, eventB := a.Peek(), b.Peek()

		if timeComparison := eventA.Time.Compare(eventB.Time); timeComparison != 0 {
			return timeComparison
		}

		priorityComparison := e.configs.getPriority(&eventA) - e.configs.getPriority(&eventB)

		return priorityComparison
	})
}
