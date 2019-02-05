package eventd

import (
	"context"
	"errors"
	"fmt"
	"path"
	"strings"
	"sync"
	"time"

	corev2 "github.com/sensu/sensu-go/api/core/v2"
	"github.com/sensu/sensu-go/backend/liveness"
	"github.com/sensu/sensu-go/backend/messaging"
	"github.com/sensu/sensu-go/backend/store"
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
	store           store.Store
	bus             messaging.MessageBus
	handlerCount    int
	livenessFactory liveness.Factory
	eventChan       chan interface{}
	subscription    messaging.Subscription
	errChan         chan error
	mu              *sync.Mutex
	shutdownChan    chan struct{}
	wg              *sync.WaitGroup
}

// Option is a functional option.
type Option func(*Eventd) error

// Config configures Eventd
type Config struct {
	Store           store.Store
	Bus             messaging.MessageBus
	LivenessFactory liveness.Factory
}

// New creates a new Eventd.
func New(c Config, opts ...Option) (*Eventd, error) {
	e := &Eventd{
		store:           c.Store,
		bus:             c.Bus,
		handlerCount:    10,
		livenessFactory: c.LivenessFactory,
		errChan:         make(chan error, 1),
		shutdownChan:    make(chan struct{}, 1),
		eventChan:       make(chan interface{}, 100),
		wg:              &sync.WaitGroup{},
		mu:              &sync.Mutex{},
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

// eventKey creates a key to identify the event for liveness monitoring
func eventKey(event *corev2.Event) string {
	// Typically we want the entity name to be the thing we monitor, but if
	// it's a round robin check, and there is no proxy entity, then use
	// the check name instead.
	if event.Check.RoundRobin && event.Entity.EntityClass != corev2.EntityProxyClass {
		return path.Join(event.Check.Namespace, event.Check.Name)
	}
	return path.Join(event.Entity.Namespace, event.Check.Name, event.Entity.Name)
}

func (e *Eventd) handleMessage(msg interface{}) error {
	event, ok := msg.(*corev2.Event)
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

	ctx := context.WithValue(context.Background(), corev2.NamespaceKey, event.Entity.Namespace)

	prevEvent, err := e.store.GetEventByEntityCheck(
		ctx, event.Entity.Name, event.Check.Name,
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

	switches := e.livenessFactory("eventd", e.dead, e.alive, logger)
	switchKey := eventKey(event)

	if event.Check.Ttl > 0 {
		// Reset the switch
		timeout := int64(event.Check.Ttl)
		if err := switches.Alive(context.TODO(), switchKey, timeout); err != nil {
			return err
		}
		return e.handleUpdate(event)
	} else {
		// The check TTL has been disabled, there is no longer a need to track it
		if err := switches.Bury(context.TODO(), switchKey); err != nil {
			// It's better to publish the event even if this fails, so
			// don't return the error here.
			logger.WithError(err).Error("error burying switch")
		}
	}

	return e.bus.Publish(messaging.TopicEvent, event)
}

func (e *Eventd) alive(key string, prev liveness.State, leader bool) (bury bool) {
	lager := logger.WithFields(logrus.Fields{
		"status":          liveness.Alive.String(),
		"previous_status": prev.String()})

	namespace, check, entity, err := parseKey(key)
	if err != nil {
		lager.Error(err)
		return false
	}

	lager = lager.WithFields(logrus.Fields{
		"check":     check,
		"entity":    entity,
		"namespace": namespace})

	lager.Info("check TTL reset")

	return false
}

func (e *Eventd) dead(key string, prev liveness.State, leader bool) (bury bool) {
	lager := logger.WithFields(logrus.Fields{
		"status":          liveness.Dead.String(),
		"previous_status": prev.String()})

	namespace, check, entity, err := parseKey(key)
	if err != nil {
		lager.Error(err)
		return false
	}

	lager = lager.WithFields(logrus.Fields{
		"check":     check,
		"entity":    entity,
		"namespace": namespace})

	lager.Warn("check TTL expired")

	// NOTE: To support check TTL for round robin scheduling, load all events
	// here, filter by check, and update all events involved in the round robin
	if entity == "" {
		lager.Error("round robin check TTL not supported")
		return false
	}

	ctx := store.NamespaceContext(context.Background(), namespace)

	// The entity has been deleted, and so there is no reason to track check
	// TTL for it anymore.
	if ent, err := e.store.GetEntityByName(ctx, entity); err == nil && ent == nil {
		return true
	}

	event, err := e.store.GetEventByEntityCheck(ctx, entity, check)
	if err != nil {
		lager.WithError(err).Error("can't handle check TTL failure")
		return false
	}

	if event == nil {
		// The user deleted the check event but not the entity
		lager.Error("event is nil")
		return false
	}

	if leader {
		if err := e.handleFailure(event); err != nil {
			lager.WithError(err).Error("can't handle check TTL failure")
		}
	}

	return false
}

func updateOccurrences(event *corev2.Event) {
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

func parseKey(key string) (namespace, check, entity string, err error) {
	parts := strings.Split(key, "/")
	if len(parts) == 2 {
		return parts[0], parts[1], "", nil
	}
	if len(parts) == 3 {
		return parts[0], parts[1], parts[2], nil
	}
	return "", "", "", errors.New("bad key")
}

// handleUpdate updates the event in the store and publishes it to TopicEvent.
func (e *Eventd) handleUpdate(event *corev2.Event) error {
	ctx := context.WithValue(context.Background(), corev2.NamespaceKey, event.Entity.Namespace)

	err := e.store.UpdateEvent(ctx, event)
	if err != nil {
		return err
	}

	return e.bus.Publish(messaging.TopicEvent, event)
}

// handleFailure creates a check event with a warn status and publishes it to
// TopicEvent.
func (e *Eventd) handleFailure(event *corev2.Event) error {
	entity := event.Entity
	ctx := context.WithValue(context.Background(), corev2.NamespaceKey, entity.Namespace)

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

func (e *Eventd) createFailedCheckEvent(ctx context.Context, event *corev2.Event) (*corev2.Event, error) {
	if !event.HasCheck() {
		return nil, errors.New("event does not contain a check")
	}

	event, err := e.store.GetEventByEntityCheck(
		ctx, event.Entity.Name, event.Check.Name,
	)
	if err != nil {
		return nil, err
	}

	check := corev2.NewCheck(corev2.NewCheckConfigFromFace(event.Check))
	output := fmt.Sprintf("Last check execution was %d seconds ago", time.Now().Unix()-event.Check.Executed)

	check.Output = output
	check.Status = 1
	check.State = corev2.EventFailingState
	check.Executed = time.Now().Unix()

	check.MergeWith(event.Check)

	event.Timestamp = time.Now().Unix()
	event.Check = check

	return event, nil
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
