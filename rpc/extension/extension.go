package extension

import (
	"errors"

	"github.com/sensu/sensu-go/types"
)

// ErrNotImplemented is returned by extension methods that haven't been
// mapped.
var ErrNotImplemented = errors.New("method not implemented")

type HandlerFunc func(*types.Event, []byte) error
type MutatorFunc func(*types.Event) ([]byte, error)
type FilterFunc func(*types.Event) (bool, error)

var _ Interface = &Extension{}

func (h HandlerFunc) HandleEvent(e *types.Event, m []byte) error {
	return h(e, m)
}

func (m MutatorFunc) MutateEvent(e *types.Event) ([]byte, error) {
	return m(e)
}

func (f FilterFunc) FilterEvent(e *types.Event) (bool, error) {
	return f(e)
}

// Interface is the definition of an extension.
type Interface interface {
	// FilterEvent filters an event.
	FilterEvent(*types.Event) (bool, error)

	// MutateEvent mutates an event.
	MutateEvent(*types.Event) ([]byte, error)

	// HandleEvent handles an event. It is passed both the original event
	// and the mutated event, if the event was mutated.
	HandleEvent(*types.Event, []byte) error
}

// Extension is a convenience type for implementing Interface.
type Extension struct {
	HandlerFunc
	MutatorFunc
	FilterFunc
}

// New creates a new Extension.
func New() *Extension {
	return &Extension{
		HandlerFunc: func(*types.Event, []byte) error { return ErrNotImplemented },
		MutatorFunc: func(*types.Event) ([]byte, error) { return nil, ErrNotImplemented },
		FilterFunc:  func(*types.Event) (bool, error) { return false, ErrNotImplemented },
	}
}

// Handle registers fn as the func to call on HandleEvent.
func (e *Extension) Handle(fn HandlerFunc) *Extension {
	e.HandlerFunc = fn
	return e
}

// Mutate registers fn as the func to call on MutateEvent.
func (e *Extension) Mutate(fn MutatorFunc) *Extension {
	e.MutatorFunc = fn
	return e
}

// Filter registers fn as the func to call on FilterEvent.
func (e *Extension) Filter(fn FilterFunc) *Extension {
	e.FilterFunc = fn
	return e
}
