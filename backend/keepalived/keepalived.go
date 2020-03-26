package keepalived

import (
	"context"
	"fmt"
	"path"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/sensu/sensu-go/agent"
	corev2 "github.com/sensu/sensu-go/api/core/v2"
	"github.com/sensu/sensu-go/backend/liveness"
	"github.com/sensu/sensu-go/backend/messaging"
	"github.com/sensu/sensu-go/backend/ringv2"
	"github.com/sensu/sensu-go/backend/store"
	"github.com/sirupsen/logrus"
)

const (
	// DEPRECATED, use core/v2
	// KeepaliveCheckName is the name of the check that is created when a
	// keepalive timeout occurs.
	KeepaliveCheckName = "keepalive"

	// DEPRECATED, use core/v2
	// KeepaliveHandlerName is the name of the handler that is executed when
	// a keepalive timeout occurs.
	KeepaliveHandlerName = "keepalive"

	// DEPRECATED, use core/v2
	// RegistrationCheckName is the name of the check that is created when an
	// entity sends a keepalive and the entity does not yet exist in the store.
	RegistrationCheckName = "registration"

	// DEPRECATED, use core/v2
	// RegistrationHandlerName is the name of the handler that is executed when
	// a registration event is passed to pipelined.
	RegistrationHandlerName = "registration"
)

const deletedEventSentinel = -1

// Keepalived is responsible for monitoring keepalive events and recording
// keepalives for entities.
type Keepalived struct {
	bus                   messaging.MessageBus
	workerCount           int
	store                 store.Store
	eventStore            store.EventStore
	deregistrationHandler string
	mu                    *sync.Mutex
	wg                    *sync.WaitGroup
	keepaliveChan         chan interface{}
	subscription          messaging.Subscription
	errChan               chan error
	livenessFactory       liveness.Factory
	ringPool              *ringv2.Pool
	ctx                   context.Context
	cancel                context.CancelFunc
	storeTimeout          time.Duration
}

// Option is a functional option.
type Option func(*Keepalived) error

// Config configures Keepalived.
type Config struct {
	Store                 store.Store
	EventStore            store.EventStore
	Bus                   messaging.MessageBus
	LivenessFactory       liveness.Factory
	DeregistrationHandler string
	RingPool              *ringv2.Pool
	BufferSize            int
	WorkerCount           int
	StoreTimeout          time.Duration
}

// New creates a new Keepalived.
func New(c Config, opts ...Option) (*Keepalived, error) {
	if c.BufferSize == 0 {
		logger.Warn("BufferSize not set")
		c.BufferSize = 1
	}
	if c.WorkerCount == 0 {
		logger.Warn("WorkerCount not set")
		c.WorkerCount = 1
	}
	if c.StoreTimeout == 0 {
		logger.Warn("StoreTimeout not set")
		c.StoreTimeout = time.Minute
	}

	ctx, cancel := context.WithCancel(context.Background())

	k := &Keepalived{
		store:      c.Store,
		eventStore: c.EventStore,
		bus:        c.Bus,
		deregistrationHandler: c.DeregistrationHandler,
		livenessFactory:       c.LivenessFactory,
		keepaliveChan:         make(chan interface{}, c.BufferSize),
		workerCount:           c.WorkerCount,
		mu:                    &sync.Mutex{},
		errChan:               make(chan error, 1),
		ringPool:              c.RingPool,
		ctx:                   ctx,
		cancel:                cancel,
		storeTimeout:          c.StoreTimeout,
	}
	for _, o := range opts {
		if err := o(k); err != nil {
			return nil, err
		}
	}
	return k, nil
}

// Receiver returns the keepalive receiver channel.
func (k *Keepalived) Receiver() chan<- interface{} {
	return k.keepaliveChan
}

// Start starts the daemon, returning an error if preconditions for startup
// fail.
func (k *Keepalived) Start() error {
	sub, err := k.bus.Subscribe(messaging.TopicKeepalive, "keepalived", k)
	if err != nil {
		return err
	}

	k.subscription = sub
	if err := k.initFromStore(context.Background()); err != nil {
		_ = sub.Cancel()
		return err
	}

	k.startWorkers()

	return nil
}

// Stop stops the daemon, returning an error if one was encountered during
// shutdown.
func (k *Keepalived) Stop() error {
	k.cancel()
	err := k.subscription.Cancel()
	close(k.keepaliveChan)
	k.wg.Wait()
	close(k.errChan)
	return err
}

// Err returns a channel that the caller can use to listen for terminal errors
// indicating a premature shutdown of the Daemon.
func (k *Keepalived) Err() <-chan error {
	return k.errChan
}

// Name returns the daemon name
func (k *Keepalived) Name() string {
	return "keepalived"
}

func (k *Keepalived) initFromStore(ctx context.Context) error {
	// For which clients were we previously alerting?
	tctx, cancel := context.WithTimeout(ctx, k.storeTimeout)
	defer cancel()
	keepalives, err := k.store.GetFailingKeepalives(tctx)
	if err != nil {
		return err
	}

	switches := k.livenessFactory(k.Name(), k.dead, k.alive, logger)

	for _, keepalive := range keepalives {
		entityCtx := context.WithValue(ctx, corev2.NamespaceKey, keepalive.Namespace)
		tctx, cancel := context.WithTimeout(entityCtx, k.storeTimeout)
		defer cancel()
		event, err := k.eventStore.GetEventByEntityCheck(tctx, keepalive.Name, "keepalive")
		if err != nil {
			return err
		}

		id := path.Join(keepalive.Namespace, keepalive.Name)

		// if there's no event, the entity was deregistered/deleted.
		if event == nil {
			tctx, cancel := context.WithTimeout(entityCtx, k.storeTimeout)
			defer cancel()
			if err := switches.Bury(tctx, id); err != nil {
				return err
			}
			continue
		}

		if !event.HasCheck() {
			logger.WithFields(logrus.Fields{"event": event}).Error("keepalive event malformed")
			continue
		}

		// if another backend picked it up, it will be passing.
		if event.Check.Status == 0 {
			continue
		}

		ttl := int64(event.Check.Timeout)
		tctx, cancel = context.WithTimeout(entityCtx, k.storeTimeout)
		defer cancel()
		if err := switches.Dead(tctx, id, ttl); err != nil {
			return fmt.Errorf("error initializing keepalive %q: %s", id, err)
		}
	}

	return nil
}

func (k *Keepalived) startWorkers() {
	k.wg = &sync.WaitGroup{}
	k.wg.Add(k.workerCount)

	for i := 0; i < k.workerCount; i++ {
		go k.processKeepalives(context.Background())
	}
}

func (k *Keepalived) processKeepalives(ctx context.Context) {
	defer k.wg.Done()

	var (
		event *corev2.Event
		ok    bool
	)

	switches := k.livenessFactory(k.Name(), k.alive, k.dead, logger)

	for msg := range k.keepaliveChan {
		event, ok = msg.(*corev2.Event)
		if !ok {
			logger.Error("keepalived received non-Event on keepalive channel")
			continue
		}

		entity := event.Entity
		if entity == nil {
			logger.Error("keepalive channel received keepalive with nil event")
			continue
		}

		if err := entity.Validate(); err != nil {
			logger.WithError(err).Error("invalid keepalive event")
			continue
		}

		if event.Timestamp == deletedEventSentinel {
			// The keepalive event was deleted, so we should bury its associated switch
			id := path.Join(entity.Namespace, entity.Name)
			tctx, cancel := context.WithTimeout(ctx, k.storeTimeout)
			err := switches.Bury(tctx, id)
			cancel()
			if err != nil {
				if _, ok := err.(*store.ErrInternal); ok {
					// Fatal error
					select {
					case k.errChan <- err:
					case <-k.ctx.Done():
					}
					return
				}
				logger.WithError(err).Error("error deleting keepalive")
			}
			continue
		}

		if err := k.handleEntityRegistration(entity); err != nil {
			logger.WithError(err).Error("error handling entity registration")
			if _, ok := err.(*store.ErrInternal); ok {
				// Fatal error
				select {
				case k.errChan <- err:
				case <-k.ctx.Done():
				}
				return
			}
		}

		// Retrieve the keepalive timeout or use a default value in case an older
		// agent version was used, since entity.KeepaliveTimeout no longer exist
		ttl := int64(corev2.DefaultKeepaliveTimeout)
		if event.Check != nil {
			ttl = int64(event.Check.Timeout)
		}

		key := path.Join(entity.Namespace, entity.Name)

		tctx, cancel := context.WithTimeout(ctx, k.storeTimeout)
		err := switches.Alive(tctx, key, ttl)
		cancel()
		if err != nil {
			logger.WithError(err).Errorf("error on switch %q", key)
			if _, ok := err.(*store.ErrInternal); ok {
				// Fatal error
				select {
				case k.errChan <- err:
				case <-k.ctx.Done():
				}
				return
			}
			continue
		}

		if err := k.handleUpdate(event); err != nil {
			logger.WithError(err).Error("error updating event")
			if _, ok := err.(*store.ErrInternal); ok {
				// Fatal error
				select {
				case k.errChan <- err:
				case <-k.ctx.Done():
				}
				return
			}
		}
	}
}

// HandleError logs an error
func (k *Keepalived) HandleError(err error) {
	logger.WithError(err).Error(err)
}

func (k *Keepalived) handleEntityRegistration(entity *corev2.Entity) error {
	if entity.EntityClass != corev2.EntityAgentClass {
		return nil
	}

	ctx := corev2.SetContextFromResource(k.ctx, entity)
	tctx, cancel := context.WithTimeout(ctx, k.storeTimeout)
	defer cancel()
	fetchedEntity, err := k.store.GetEntityByName(tctx, entity.Name)

	if err != nil {
		// Warning: do not wrap this error
		return err
	}

	if fetchedEntity == nil {
		event := createRegistrationEvent(entity)
		err = k.bus.Publish(messaging.TopicEvent, event)
	}

	return err
}

func createKeepaliveEvent(rawEvent *corev2.Event) *corev2.Event {
	check := rawEvent.Check
	if check == nil {
		check = &corev2.Check{
			Interval: agent.DefaultKeepaliveInterval,
			Timeout:  corev2.DefaultKeepaliveTimeout,
		}
	}

	// Use the entity keepalive handlers if defined, otherwise fallback to the
	// default keepalive handler
	handlers := []string{corev2.KeepaliveHandlerName}
	if len(rawEvent.Entity.KeepaliveHandlers) > 0 {
		handlers = rawEvent.Entity.KeepaliveHandlers
	}

	keepaliveCheck := &corev2.Check{
		ObjectMeta: corev2.ObjectMeta{
			Name:      corev2.KeepaliveCheckName,
			Namespace: rawEvent.Entity.Namespace,
		},
		Interval: check.Interval,
		Timeout:  check.Timeout,
		Ttl:      check.Ttl,
		Handlers: handlers,
		Executed: time.Now().Unix(),
		Issued:   time.Now().Unix(),
	}
	keepaliveEvent := &corev2.Event{
		ObjectMeta: rawEvent.ObjectMeta,
		Timestamp:  time.Now().Unix(),
		Entity:     rawEvent.Entity,
		Check:      keepaliveCheck,
		ID:         rawEvent.ID,
	}

	if len(keepaliveEvent.ID) == 0 {
		uid, _ := uuid.NewRandom()
		keepaliveEvent.ID = uid[:]
	}

	return keepaliveEvent
}

func createRegistrationEvent(entity *corev2.Entity) *corev2.Event {
	registrationCheck := &corev2.Check{
		ObjectMeta: corev2.ObjectMeta{
			Name:      corev2.RegistrationCheckName,
			Namespace: entity.Namespace,
		},
		Interval: 1,
		Handlers: []string{corev2.RegistrationHandlerName},
		Status:   1,
	}
	registrationEvent := &corev2.Event{
		ObjectMeta: corev2.ObjectMeta{
			Namespace: entity.Namespace,
		},
		Timestamp: time.Now().Unix(),
		Entity:    entity,
		Check:     registrationCheck,
	}

	return registrationEvent
}

func (k *Keepalived) alive(key string, prev liveness.State, leader bool) bool {
	lager := logger.WithFields(logrus.Fields{
		"status":          liveness.Alive.String(),
		"previous_status": prev.String(),
		"is_leader":       fmt.Sprintf("%v", leader),
	})

	namespace, name, err := parseKey(key)
	if err != nil {
		lager.Error(err)
		return false
	}

	lager = lager.WithFields(logrus.Fields{"entity": name, "namespace": namespace})
	lager.Debug("entity is alive")
	return false
}

// this is a callback - it should not write to k.errChan
func (k *Keepalived) dead(key string, prev liveness.State, leader bool) bool {
	// Parse the key to determine the namespace and entity name. The error will be
	// verified later
	namespace, name, err := parseKey(key)

	lager := logger.WithFields(logrus.Fields{
		"status":          liveness.Dead.String(),
		"previous_status": prev.String(),
		"is_leader":       fmt.Sprintf("%v", leader),
		"entity":          name,
		"namespace":       namespace,
	})

	if !leader {
		// If this client isn't the one that flipped the keepalive switch,
		// don't do anything further.
		lager.Debug("not the leader of this keepalive switch, stopping here")
		return false
	}

	lager.Warn("keepalive timed out")

	// Now verify if we encountered an error while parsing the key
	if err != nil {
		// We couldn't parse the key, which probably means the key didn't contain a
		// namespace. Log the error and then try to bury the key so it stops sending
		// events to the watcher.
		lager.Error(err)
		return true
	}

	ctx := store.NamespaceContext(k.ctx, namespace)

	entity, err := k.store.GetEntityByName(ctx, name)
	if err != nil {
		lager.WithError(err).Error("error while reading entity")
		return false
	}

	if entity == nil {
		// The entity has been deleted, there is no longer a need to
		// track keepalives for it.
		lager.Debug("nil entity")
		return true
	}

	if entity.Deregister {
		deregisterer := &Deregistration{
			EntityStore:  k.store,
			EventStore:   k.eventStore,
			MessageBus:   k.bus,
			StoreTimeout: k.storeTimeout,
		}
		if err := deregisterer.Deregister(entity); err != nil {
			lager.WithError(err).Error("error deregistering entity")
		}
		lager.Debug("deregistering entity")
		return true
	}

	currentEvent, err := k.eventStore.GetEventByEntityCheck(ctx, name, "keepalive")
	if err != nil {
		lager.WithError(err).Error("error while reading event")
		return false
	}
	if currentEvent == nil {
		// The keepalive was deleted, so bury the switch
		lager.Debug("nil event")
		return true
	}

	// this is a real keepalive event, emit it.
	event := createKeepaliveEvent(currentEvent)
	timeSinceLastSeen := time.Now().Unix() - entity.LastSeen
	warningTimeout := int64(event.Check.Timeout)
	criticalTimeout := event.Check.Ttl
	var timeout int64
	if warningTimeout != 0 && timeSinceLastSeen >= warningTimeout {
		// warning keepalive
		timeout = warningTimeout
		event.Check.Status = 1
	}
	if criticalTimeout != 0 && timeSinceLastSeen >= criticalTimeout {
		// critical keepalive
		timeout = criticalTimeout
		event.Check.Status = 2
	}
	event.Check.Output = fmt.Sprintf("No keepalive sent from %s for %v seconds (>= %v)", entity.Name, timeSinceLastSeen, timeout)

	if err := k.bus.Publish(messaging.TopicEventRaw, event); err != nil {
		lager.WithError(err).Error("error publishing event")
		return false
	}

	expiration := time.Now().Unix() + int64(event.Check.Timeout)

	if err := k.store.UpdateFailingKeepalive(ctx, entity, expiration); err != nil {
		lager.WithError(err).Error("error updating keepalive")
		return false
	}

	if entity.EntityClass != corev2.EntityAgentClass {
		return false
	}

	for _, sub := range entity.Subscriptions {
		ring := k.ringPool.Get(ringv2.Path(namespace, sub))
		if err := ring.Remove(ctx, name); err != nil {
			lager := lager.WithFields(logrus.Fields{"subscription": sub})
			lager.WithError(err).Error("error removing entity from ring")
		}
	}

	return false
}

func parseKey(key string) (namespace, name string, err error) {
	parts := strings.Split(key, "/")
	if len(parts) != 2 {
		return "", "", fmt.Errorf("bad key: '%s'", key)
	}
	return parts[0], parts[1], nil
}

// handleUpdate sets the entity's last seen time and publishes an OK check event
// to the message bus.
func (k *Keepalived) handleUpdate(e *corev2.Event) error {
	entity := e.Entity

	ctx := corev2.SetContextFromResource(context.Background(), entity)
	if err := k.store.DeleteFailingKeepalive(ctx, e.Entity); err != nil {
		// Warning: do not wrap this error
		return err
	}

	entity.LastSeen = e.Timestamp

	if err := k.store.UpdateEntity(ctx, entity); err != nil {
		logger.WithError(err).Error("error updating entity in store")
		// Warning: do not wrap this error
		return err
	}
	event := createKeepaliveEvent(e)
	event.Check.Status = 0
	event.Check.Output = fmt.Sprintf("Keepalive last sent from %s at %s", entity.Name, time.Unix(entity.LastSeen, 0).String())

	if entity.EntityClass == corev2.EntityAgentClass {
		// Refresh the rings that the entity is involved in
		for _, sub := range entity.Subscriptions {
			if strings.HasPrefix(sub, "entity:") {
				// Entity subscriptions don't get rings
				continue
			}
			ring := k.ringPool.Get(ringv2.Path(entity.Namespace, sub))
			if e.Check.Timeout == 0 {
				// This should never happen but guards against a crash
				e.Check.Timeout = corev2.DefaultKeepaliveTimeout
			}
			tctx, cancel := context.WithTimeout(ctx, k.storeTimeout)
			defer cancel()
			if err := ring.Add(tctx, entity.Name, int64(e.Check.Timeout)); err != nil {
				lager := logger.WithFields(logrus.Fields{
					"entity":       entity.Name,
					"namespace":    entity.Namespace,
					"subscription": sub,
				})
				lager.WithError(err).Error("error adding entity to ring")
				if _, ok := err.(*store.ErrInternal); ok {
					return err
				}
			}
		}
	}

	return k.bus.Publish(messaging.TopicEventRaw, event)
}
