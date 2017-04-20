package eventd

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

	ch := make(chan []byte, 100)
	err := e.MessageBus.Subscribe(messaging.TopicEventRaw, "eventd", ch)
	if err != nil {
		return err
	}
	e.eventChan = ch

	e.wg = &sync.WaitGroup{}
	e.wg.Add(e.HandlerCount)
	e.startHandlers()

	return nil
}

func (e *Eventd) startHandlers() {
	for i := 0; i < e.HandlerCount; i++ {
		go func() {
			var event *types.Event
			var err error
			defer e.wg.Done()

			for {
				select {
				case <-e.shutdownChan:
					return

				case msg, ok := <-e.eventChan:
					// The message bus will close channels when it's shut down which means
					// we will end up reading from a closed channel. If it's closed,
					// return from this goroutine. Shutdown eventd.
					if !ok {
						e.errChan <- errors.New("event channel closed")
						return
					}
					log.Printf("eventd - handling event: %s\n", string(msg))
					event = &types.Event{}
					err = json.Unmarshal(msg, event)
					if err != nil {
						log.Printf("eventd - error handling event: %s\n", err.Error())
						continue
					}

					if event.Check == nil || event.Entity == nil {
						log.Println("eventd - error handling event: event invalid")
						continue
					}

					if err := event.Check.Validate(); err != nil {
						log.Printf("eventd - error handling event: %s\n", err.Error())
						continue
					}

					if err := event.Entity.Validate(); err != nil {
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
	log.Println("shutting down eventd")
	close(e.shutdownChan)
	e.wg.Wait()
	close(e.eventChan)
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
