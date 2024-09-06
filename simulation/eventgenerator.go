package simulation

import (
	"errors"
)

var ErrEventGeneratorFinished = errors.New("event generator is finished")

type EventGenerator interface {
	Pop() *Event
	Peek() Event

	Finished() bool
}
