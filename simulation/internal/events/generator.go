package events

import (
	"errors"
)

var ErrGeneratorFinished = errors.New("event generator is finished")

type Generator interface {
	Pop() *Event
	Peek() Event

	Finished() bool
}
