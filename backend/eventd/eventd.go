package eventd

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/sensu/sensu-go/backend/messaging"
	"github.com/sensu/sensu-go/backend/monitor"
	"github.com/sensu/sensu-go/backend/store"
	"github.com/sensu/sensu-go/types"
	"github.com/sirupsen/logrus"
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
	store          store.Store
	bus            messaging.MessageBus
	handlerCount   int
	monitorFactory monitor.Factory
	eventChan      chan interface{}
	subscription   messaging.Subscription
	errChan        chan error
	mu             *sync.Mutex
	shutdownChan   chan struct{}
	wg             *sync.WaitGroup
}

// Option is a functional option.
type Option func(*Eventd) error

// Config configures Eventd
type Config struct {
	Store          store.Store
	Bus            messaging.MessageBus
	MonitorFactory monitor.Factory
}

// New creates a new Eventd.
func New(c Config, opts ...Option) (*Eventd, error) {
	e := &Eventd{
		store:          c.Store,
		bus:            c.Bus,
		handlerCount:   10,
		monitorFactory: c.MonitorFactory,
		errChan:        make(chan error, 1),
		shutdownChan:   make(chan struct{}, 1),
		eventChan:      make(chan interface{}, 100),
		wg:             &sync.WaitGroup{},
		mu:             &sync.Mutex{},
	}
	for _, o := range opts {
		if err := o(e); err != nil {
			return nil, err
		}
	}
	return e, nil
}

// Receiver returns the event receiver channel.
func (e *Eventd) Receiver() chan<- interface{} {
	return e.eventChan
}

// Start eventd.
func (e *Eventd) Start() error {
	e.wg.Add(e.handlerCount)
	sub, err := e.bus.Subscribe(messaging.TopicEventRaw, "eventd", e)
	e.subscription = sub
	if err != nil {
		return err
	}
	e.startHandlers()

	return nil
}

func (e *Eventd) startHandlers() {
	for i := 0; i < e.handlerCount; i++ {
		go func() {
			defer e.wg.Done()

			for {
				select {
				case <-e.shutdownChan:
					// drain the event channel.
					for msg := range e.eventChan {
						if err := e.handleMessage(msg); err != nil {
							logger.WithError(err).Error("eventd - error handling event")
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
						logger.WithError(err).Error("eventd - error handling event")
					}
				}
			}
		}()
	}
}

func (e *Eventd) HandleError(err error) {
	logger.WithError(err).Error("error monitoring event")
}

func (e *Eventd) handleMessage(msg interface{}) error {
	event, ok := msg.(*types.Event)
	if !ok {
		return errors.New("received non-Event on event channel")
	}

	// Validate the received event
	if err := event.Validate(); err != nil {
		return err
	}

	// If the event does not contain a check (rather, it contains metrics)
	// publish the event without writing to the store
	if !event.HasCheck() {
		return e.bus.Publish(messaging.TopicEvent, event)
	}

	ctx := context.WithValue(context.Background(), types.NamespaceKey, event.Entity.Namespace)

	prevEvent, err := e.store.GetEventByEntityCheck(
		ctx, event.Entity.ID, event.Check.Name,
	)
	if err != nil {
		return err
	}

	// Maintain check history.
	if prevEvent != nil {
		if !prevEvent.HasCheck() {
			return errors.New("invalid previous event")
		}

		event.Check.MergeWith(prevEvent.Check)
	}

	updateOccurrences(event)

	// Calculate percent state change for this check's history
	event.Check.TotalStateChange = totalStateChange(event)

	// Determine the check's state
	state(event)

	// Add any silenced subscriptions to the event
	err = getSilenced(ctx, event, e.store)
	if err != nil {
		return err
	}

	// Handle expire on resolve silenced entries
	if err = handleExpireOnResolveEntries(ctx, event, e.store); err != nil {
		return err
	}

	if err := e.store.UpdateEvent(ctx, event); err != nil {
		return err
	}

	if event.Check.Ttl > 0 {
		// Reset the TTL monitor
		var monitorKey string

		// Typically we want the entity ID to be the thing we monitor, but if
		// it's a round robin check and there is no proxy entity, then just use
		// the check name instead.
		if event.Check.RoundRobin && event.Entity.Class != types.EntityProxyClass {
			monitorKey = event.Check.Name
		} else {
			monitorKey = event.Entity.ID
		}

		timeout := int64(event.Check.Ttl)
		supervisor := e.monitorFactory(e)
		err := supervisor.Monitor(context.TODO(), monitorKey, event, timeout)
		if err == nil {
			// HandleUpdate also publishes the event
			err = e.HandleUpdate(event)
		}
		return err
	}

	return e.bus.Publish(messaging.TopicEvent, event)
}

func updateOccurrences(event *types.Event) {
	if !event.HasCheck() {
		return
	}
	if len(event.Check.History) > 1 && (event.IsIncident() || isFlapping(event)) {
		historyLen := len(event.Check.History)
		if event.Check.History[historyLen-1].Status == event.Check.History[historyLen-2].Status {
			event.Check.Occurrences++
		} else {
			event.Check.Occurrences = 1
		}
	} else {
		event.Check.Occurrences = 1
	}

	if event.Check.Occurrences > event.Check.OccurrencesWatermark {
		event.Check.OccurrencesWatermark = event.Check.Occurrences
	}
}

// HandleUpdate updates the event in the store and publishes it to TopicEvent.
func (e *Eventd) HandleUpdate(event *types.Event) error {
	ctx := context.WithValue(context.Background(), types.NamespaceKey, event.Entity.Namespace)

	err := e.store.UpdateEvent(ctx, event)
	if err != nil {
		return err
	}

	return e.bus.Publish(messaging.TopicEvent, event)
}

// HandleFailure creates a check event with a warn status and publishes it to
// TopicEvent.
func (e *Eventd) HandleFailure(event *types.Event) error {
	entity := event.Entity
	ctx := context.WithValue(context.Background(), types.NamespaceKey, entity.Namespace)

	failedCheckEvent, err := e.createFailedCheckEvent(ctx, event)
	if err != nil {
		return err
	}
	err = e.store.UpdateEvent(ctx, failedCheckEvent)
	if err != nil {
		return err
	}

	return e.bus.Publish(messaging.TopicEvent, failedCheckEvent)
}

func (e *Eventd) createFailedCheckEvent(ctx context.Context, event *types.Event) (*types.Event, error) {
	if !event.HasCheck() {
		return nil, errors.New("event does not contain a check")
	}

	lastCheckResult, err := e.store.GetEventByEntityCheck(
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
	if err := e.subscription.Cancel(); err != nil {
		logger.WithError(err).Error("unable to unsubscribe from message bus")
	}
	close(e.eventChan)
	close(e.shutdownChan)
	e.wg.Wait()
	return nil
}

// Err returns a channel to listen for terminal errors on.
func (e *Eventd) Err() <-chan error {
	return e.errChan
}

// Name returns the daemon name
func (e *Eventd) Name() string {
	return "eventd"
}
