package eventd

import (
	"context"
	"errors"
	"fmt"
	"path"
	"strings"
	"sync"
	"time"

	"github.com/coreos/etcd/clientv3"
	"github.com/prometheus/client_golang/prometheus"
	corev2 "github.com/sensu/sensu-go/api/core/v2"
	"github.com/sensu/sensu-go/backend/keepalived"
	"github.com/sensu/sensu-go/backend/liveness"
	"github.com/sensu/sensu-go/backend/messaging"
	"github.com/sensu/sensu-go/backend/store"
	"github.com/sensu/sensu-go/backend/store/cache"
	"github.com/sirupsen/logrus"
)

const (
	// ComponentName identifies Eventd as the component/daemon implemented in this
	// package.
	ComponentName = "eventd"

	// EventsProcessedCounterVec is the name of the prometheus counter vec used to count events processed.
	EventsProcessedCounterVec = "sensu_go_events_processed"

	// EventsProcessedLabelName is the name of the label which stores prometheus values.
	EventsProcessedLabelName = "status"

	// EventsProcessedLabelSuccess is the name of the label used to count events processed successfully.
	EventsProcessedLabelSuccess = "success"
)

var (
	logger = logrus.WithFields(logrus.Fields{
		"component": ComponentName,
	})

	// EventsProcessed counts the number of sensu go events processed.
	EventsProcessed = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: EventsProcessedCounterVec,
			Help: "The total number of processed events",
		},
		[]string{EventsProcessedLabelName},
	)
)

const deletedEventSentinel = -1

// Eventd handles incoming sensu events and stores them in etcd.
type Eventd struct {
	ctx             context.Context
	cancel          context.CancelFunc
	store           store.Store
	eventStore      store.EventStore
	bus             messaging.MessageBus
	workerCount     int
	livenessFactory liveness.Factory
	eventChan       chan interface{}
	subscription    messaging.Subscription
	errChan         chan error
	mu              *sync.Mutex
	shutdownChan    chan struct{}
	wg              *sync.WaitGroup
	Logger          Logger
	silencedCache   *cache.Resource
}

// Option is a functional option.
type Option func(*Eventd) error

// Config configures Eventd
type Config struct {
	Store           store.Store
	EventStore      store.EventStore
	Bus             messaging.MessageBus
	LivenessFactory liveness.Factory
	Client          *clientv3.Client
	BufferSize      int
	WorkerCount     int
}

// New creates a new Eventd.
func New(ctx context.Context, c Config, opts ...Option) (*Eventd, error) {
	if c.BufferSize == 0 {
		c.BufferSize = 1
	}
	if c.WorkerCount == 0 {
		c.WorkerCount = 1
	}

	e := &Eventd{
		store:           c.Store,
		eventStore:      c.EventStore,
		bus:             c.Bus,
		workerCount:     c.WorkerCount,
		livenessFactory: c.LivenessFactory,
		errChan:         make(chan error, 1),
		shutdownChan:    make(chan struct{}, 1),
		eventChan:       make(chan interface{}, c.BufferSize),
		wg:              &sync.WaitGroup{},
		mu:              &sync.Mutex{},
		Logger:          &RawLogger{},
	}

	e.ctx, e.cancel = context.WithCancel(ctx)
	cache, err := cache.New(e.ctx, c.Client, &corev2.Silenced{}, false)
	if err != nil {
		return nil, err
	}
	e.silencedCache = cache

	for _, o := range opts {
		if err := o(e); err != nil {
			return nil, err
		}
	}

	// Initialize the most likely labels
	EventsProcessed.WithLabelValues(EventsProcessedLabelSuccess)
	_ = prometheus.Register(EventsProcessed)

	return e, nil
}

// Receiver returns the event receiver channel.
func (e *Eventd) Receiver() chan<- interface{} {
	return e.eventChan
}

// Start eventd.
func (e *Eventd) Start() error {
	e.wg.Add(e.workerCount)
	sub, err := e.bus.Subscribe(messaging.TopicEventRaw, "eventd", e)
	e.subscription = sub
	if err != nil {
		return err
	}
	e.startHandlers()

	return nil
}

func (e *Eventd) startHandlers() {
	for i := 0; i < e.workerCount; i++ {
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
		e.Logger.Println(event)
		return e.bus.Publish(messaging.TopicEvent, event)
	}

	ctx := context.WithValue(context.Background(), corev2.NamespaceKey, event.Entity.Namespace)

	// Add any silenced subscriptions to the event
	getSilenced(ctx, event, e.silencedCache)

	// Merge the new event with the stored event if a match is found
	event, prevEvent, err := e.eventStore.UpdateEvent(ctx, event)
	if err != nil {
		return err
	}

	e.Logger.Println(event)

	switches := e.livenessFactory("eventd", e.dead, e.alive, logger)
	switchKey := eventKey(event)

	if event.Check.Ttl > 0 {
		// Reset the switch
		timeout := int64(event.Check.Ttl)
		if err := switches.Alive(context.TODO(), switchKey, timeout); err != nil {
			return err
		}
	} else if (prevEvent != nil && prevEvent.Check.Ttl > 0) || event.Check.Ttl == deletedEventSentinel {
		// The check TTL has been disabled, there is no longer a need to track it
		if err := switches.Bury(context.TODO(), switchKey); err != nil {
			// It's better to publish the event even if this fails, so
			// don't return the error here.
			logger.WithError(err).Error("error burying switch")
		}
	}

	EventsProcessed.WithLabelValues(EventsProcessedLabelSuccess).Inc()

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
		lager.Error("round robin check ttl not supported")
		return true
	}

	ctx := store.NamespaceContext(context.Background(), namespace)

	// The entity has been deleted, and so there is no reason to track check
	// TTL for it anymore.
	if ent, err := e.store.GetEntityByName(ctx, entity); err == nil && ent == nil {
		return true
	} else if err != nil {
		lager.WithError(err).Error("check ttl: error retrieving entity")
		return false
	}

	event, err := e.eventStore.GetEventByEntityCheck(ctx, entity, check)
	if err != nil {
		lager.WithError(err).Error("check ttl: error retrieving event")
		return false
	}

	if event == nil {
		// The user deleted the check event but not the entity
		return true
	}

	if leader {
		if err := e.handleFailure(event); err != nil {
			lager.WithError(err).Error("can't handle check TTL failure")
		}
	}

	return false
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

// handleFailure creates a check event with a warn status and publishes it to
// TopicEvent.
func (e *Eventd) handleFailure(event *corev2.Event) error {
	// don't update the event with ttl output for keepalives,
	// there is a different mechanism for that
	if event.Check.Name == keepalived.KeepaliveCheckName {
		return nil
	}

	entity := event.Entity
	ctx := context.WithValue(context.Background(), corev2.NamespaceKey, entity.Namespace)

	failedCheckEvent, err := e.createFailedCheckEvent(ctx, event)
	if err != nil {
		return err
	}
	updatedEvent, _, err := e.eventStore.UpdateEvent(ctx, failedCheckEvent)
	if err != nil {
		return err
	}

	return e.bus.Publish(messaging.TopicEvent, updatedEvent)
}

func (e *Eventd) createFailedCheckEvent(ctx context.Context, event *corev2.Event) (*corev2.Event, error) {
	if !event.HasCheck() {
		return nil, errors.New("event does not contain a check")
	}

	event, err := e.eventStore.GetEventByEntityCheck(
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
	e.cancel()
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
