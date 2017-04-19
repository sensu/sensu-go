package backend

import (
	"encoding/json"
	"errors"
	"log"
	"sync"

	"github.com/sensu/sensu-go/backend/messaging"
	"github.com/sensu/sensu-go/backend/store"
	"github.com/sensu/sensu-go/types"
)

// Eventd handles incoming sensu events and stores them in etcd.
type Eventd struct {
	Store        store.Store
	MessageBus   messaging.MessageBus
	HandlerCount int

	eventChan    chan []byte
	errChan      chan error
	shutdownChan chan struct{}
	wg           *sync.WaitGroup
}

// Start eventd.
func (e *Eventd) Start() error {
	if e.Store == nil {
		return errors.New("no store found")
	}

	if e.MessageBus == nil {
		return errors.New("no message bus found")
	}

	if e.HandlerCount == 0 {
		e.HandlerCount = 10
	}

	e.errChan = make(chan error, 1)
	e.shutdownChan = make(chan struct{}, 1)

	e.wg = &sync.WaitGroup{}
	e.wg.Add(e.HandlerCount)
	e.startHandlers()

	ch := make(chan []byte, 100)
	err := e.MessageBus.Subscribe(messaging.TopicEventRaw, "eventd", ch)
	if err != nil {
		return err
	}

	return nil
}

func (e *Eventd) startHandlers() {
	for i := 0; i < e.HandlerCount; i++ {
		go func() {
			var event *types.Event
			var err error

			for {
				select {
				case <-e.shutdownChan:
					e.wg.Done()
					return

				case msg := <-e.eventChan:
					event = &types.Event{}
					err = json.Unmarshal(msg, event)
					if err != nil {
						log.Printf("eventd - error handling event: %s\n", err.Error())
						continue
					}

					prevEvent, err := e.Store.GetEventByEntityCheck(event.Entity.ID, event.Check.Name)
					if err != nil {
						log.Printf("eventd - error handling event: %s\n", err.Error())
						continue
					}

					if prevEvent == nil {
						err = e.Store.UpdateEvent(event)
						if err != nil {
							log.Printf("eventd - error handling event: %s\n", err.Error())
						}
						continue
					}

					if prevEvent.Check == nil {
						log.Printf("eventd - error handling event: invalid previous event")
						continue
					}

					event.Check.MergeWith(prevEvent.Check)

					err = e.Store.UpdateEvent(event)
					if err != nil {
						log.Printf("eventd - error handling event: %s\n", err.Error())
					}
				}
			}
		}()
	}
}

// Stop eventd.
func (e *Eventd) Stop() error {
	close(e.shutdownChan)
	e.wg.Wait()
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
