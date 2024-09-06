package simulation

import (
	"slices"
)

type eventCombinator struct {
	configs            *eventConfigurations
	activeGenerators   []EventGenerator
	finishedGenerators []EventGenerator
}

func newEventCombinator(configs *eventConfigurations, inputs ...EventGenerator) *eventCombinator {
	combinator := &eventCombinator{
		configs:            configs,
		activeGenerators:   make([]EventGenerator, 0),
		finishedGenerators: make([]EventGenerator, 0),
	}

	for _, input := range inputs {
		if input.Finished() {
			combinator.finishedGenerators = append(combinator.finishedGenerators, input)
		} else {
			combinator.activeGenerators = append(combinator.activeGenerators, input)
		}
	}

	combinator.sortActiveGenerators()

	return combinator
}

func (e *eventCombinator) add(generator EventGenerator) {
	if generator.Finished() {
		e.finishedGenerators = append(e.finishedGenerators, generator)
		return
	}

	e.activeGenerators = append(e.activeGenerators, generator)

	e.sortActiveGenerators()
}

func (e *eventCombinator) Pop() *Event {
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

func (e *eventCombinator) Peek() Event {
	if e.Finished() {
		panic(ErrEventGeneratorFinished)
	}

	return e.activeGenerators[0].Peek()
}

func (e *eventCombinator) Finished() bool {
	return len(e.activeGenerators) == 0
}

func (e *eventCombinator) sortActiveGenerators() {
	slices.SortStableFunc(e.activeGenerators, func(a, b EventGenerator) int {
		eventA, eventB := a.Peek(), b.Peek()

		if timeComparison := eventA.Time.Compare(eventB.Time); timeComparison != 0 {
			return timeComparison
		}

		priorityComparison := e.configs.getPriority(&eventA) - e.configs.getPriority(&eventB)

		return priorityComparison
	})
}
