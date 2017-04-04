package backend

import (
	"errors"

	"github.com/sensu/sensu-go/backend/messaging"
	"github.com/sensu/sensu-go/backend/store"
)

// Eventd handles incoming sensu events and stores them in etcd.
type Eventd struct {
	Store      store.Store
	MessageBus messaging.MessageBus

	eventChan chan []byte
	errChan   chan error
}

// Start eventd.
func (e *Eventd) Start() error {
	if e.Store == nil {
		return errors.New("no store found")
	}

	if e.MessageBus == nil {
		return errors.New("no message bus found")
	}

	e.errChan = make(chan error, 1)

	ch := make(chan []byte, 100)
	err := e.MessageBus.Subscribe(messaging.TopicEvent, "eventd", ch)
	if err != nil {
		return err
	}

	return nil
}

// Stop eventd.
func (e *Eventd) Stop() error {
	return nil
}

// Status returns an error if eventd is unhealthy.
func (e *Eventd) Status() error {
	return nil
}

// Err returns a channel to listen for terminal errors on.
func (e *Eventd) Err() <-chan error {
	return e.errChan
}
