package eventd

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/sensu/sensu-go/backend/messaging"
	"github.com/sensu/sensu-go/backend/monitor"
	"github.com/sensu/sensu-go/backend/store"
	"github.com/sensu/sensu-go/types"
)

const (
	// ComponentName identifies Eventd as the component/daemon implemented in this
	// package.
	ComponentName = "eventd"
)

var (
	logger = logrus.WithFields(logrus.Fields{
		"component": ComponentName,
	})
)

// Eventd handles incoming sensu events and stores them in etcd.
type Eventd struct {
	Store          store.Store
	MessageBus     messaging.MessageBus
	HandlerCount   int
	MonitorFactory monitor.FactoryFunc

	eventChan    chan interface{}
	errChan      chan error
	monitors     map[string]monitor.Interface
	mu           *sync.Mutex
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

	if e.MonitorFactory == nil {
		e.MonitorFactory = func(entity *types.Entity, event *types.Event, t time.Duration, updateHandler monitor.UpdateHandler, failureHandler monitor.FailureHandler) monitor.Interface {
			return monitor.New(entity, event, t, updateHandler, failureHandler)
		}
	}

	e.errChan = make(chan error, 1)
	e.shutdownChan = make(chan struct{}, 1)

	ch := make(chan interface{}, 100)
	e.eventChan = ch

	err := e.MessageBus.Subscribe(messaging.TopicEventRaw, ComponentName, ch)
	if err != nil {
		return err
	}

	e.wg = &sync.WaitGroup{}
	e.wg.Add(e.HandlerCount)
	e.startHandlers()

	e.mu = &sync.Mutex{}
	e.monitors = map[string]monitor.Interface{}
	return nil
}

func (e *Eventd) startHandlers() {
	for i := 0; i < e.HandlerCount; i++ {
		go func() {
			defer e.wg.Done()

			for {
				select {
				case <-e.shutdownChan:
					// drain the event channel.
					for msg := range e.eventChan {
						if err := e.handleMessage(msg); err != nil {
							logger.Errorf("eventd - error handling event: %s", err.Error())
						}
					}
					return

				case msg, ok := <-e.eventChan:
					// The message bus will close channels when it's shut down which means
					// we will end up reading from a closed channel. If it's closed,
					// return from this goroutine and emit a fatal error. It is then
					// the responsility of eventd's parent to shutdown eventd.
					//
					// NOTE: Should that be the case? If eventd is signalling that it has,
					// effectively, shutdown, why would something else be responsible for
					// shutting it down.
					if !ok {
						// This only buffers a single error. We can't block on
						// sending these or shutdown will block indefinitely.
						select {
						case e.errChan <- errors.New("event channel closed"):
						default:
						}
						return
					}

					if err := e.handleMessage(msg); err != nil {
						logger.Errorf("eventd - error handling event: %s", err.Error())
					}
				}
			}
		}()
	}
}

func (e *Eventd) handleMessage(msg interface{}) error {
	var (
		mon monitor.Interface
		ok  bool
	)

	event, ok := msg.(*types.Event)
	if !ok {
		return errors.New("received non-Event on event channel")
	}

	// Validate the received event
	if err := event.Validate(); err != nil {
		return err
	}

	ctx := context.WithValue(context.Background(), types.OrganizationKey, event.Entity.Organization)
	ctx = context.WithValue(ctx, types.EnvironmentKey, event.Entity.Environment)

	prevEvent, err := e.Store.GetEventByEntityCheck(
		ctx, event.Entity.ID, event.Check.Name,
	)
	if err != nil {
		return err
	}

	// Maintain check history.
	if prevEvent != nil {
		if prevEvent.Check == nil {
			return errors.New("invalid previous event")
		}

		event.Check.MergeWith(prevEvent.Check)
	}

	// Calculate percent state change for this check's history
	event.Check.TotalStateChange = totalStateChange(event)

	// Determine the check's state
	state(event)

	// Add any silenced subscriptions to the event
	err = getSilenced(ctx, event, e.Store)
	if err != nil {
		return err
	}

	// Handle expire on resolve silenced entries
	err = handleExpireOnResolveEntries(ctx, event, e.Store)
	if err != nil {
		return err
	}

	err = e.Store.UpdateEvent(ctx, event)
	if err != nil {
		return err
	}

	entity := event.Entity

	if event.Check.Ttl > 0 && !event.Check.RoundRobin {
		// create a monitor for the event's entity if it doesn't exist in the
		// monitor map
		// only monitor if there is a check TTL and the check is not a
		// round robin check.
		e.mu.Lock()
		mon, ok = e.monitors[entity.ID]
		if !ok || mon.IsStopped() {
			timeout := time.Duration(event.Check.Ttl) * time.Second
			mon = e.MonitorFactory(entity, event, timeout, e, e)
			e.monitors[entity.ID] = mon
		}
		e.mu.Unlock()
		return mon.HandleUpdate(event)
	}

	err = e.Store.UpdateEvent(ctx, event)
	if err != nil {
		return err
	}

	return e.MessageBus.Publish(messaging.TopicEvent, event)
}

// HandleUpdate updates the event in the store and publishes it to TopicEvent.
func (e *Eventd) HandleUpdate(event *types.Event) error {
	ctx := context.WithValue(context.Background(), types.OrganizationKey, event.Entity.Organization)
	ctx = context.WithValue(ctx, types.EnvironmentKey, event.Entity.Environment)

	err := e.Store.UpdateEvent(ctx, event)
	if err != nil {
		return err
	}

	return e.MessageBus.Publish(messaging.TopicEvent, event)
}

// HandleFailure creates a check event with a warn status and publishes it to
// TopicEvent.
func (e *Eventd) HandleFailure(entity *types.Entity, event *types.Event) error {
	ctx := context.WithValue(context.Background(), types.OrganizationKey, entity.Organization)
	ctx = context.WithValue(ctx, types.EnvironmentKey, entity.Environment)

	failedCheckEvent, err := e.createFailedCheckEvent(ctx, event)
	if err != nil {
		return err
	}
	err = e.Store.UpdateEvent(ctx, failedCheckEvent)
	if err != nil {
		return err
	}

	return e.MessageBus.Publish(messaging.TopicEvent, failedCheckEvent)
}

func (e *Eventd) createFailedCheckEvent(ctx context.Context, event *types.Event) (*types.Event, error) {
	lastCheckResult, err := e.Store.GetEventByEntityCheck(
		ctx, event.Entity.ID, event.Check.Name,
	)
	if err != nil {
		return nil, err
	}

	output := fmt.Sprintf("Last check execution was %d seconds ago", time.Now().Unix()-lastCheckResult.Check.Executed)

	lastCheckResult.Check.Output = output
	lastCheckResult.Check.Status = 1
	lastCheckResult.Timestamp = time.Now().Unix()

	return lastCheckResult, nil
}

// Stop eventd.
func (e *Eventd) Stop() error {
	logger.Info("shutting down eventd")
	err := e.MessageBus.Unsubscribe(messaging.TopicEventRaw, ComponentName)
	close(e.eventChan)
	close(e.shutdownChan)
	e.wg.Wait()
	return err
}

// Status returns an error if eventd is unhealthy.
func (e *Eventd) Status() error {
	return nil
}

// Err returns a channel to listen for terminal errors on.
func (e *Eventd) Err() <-chan error {
	return e.errChan

}
