package keepalived

import (
	"context"
	"errors"
	"fmt"
	"path"
	"strings"
	"sync"
	"time"

	"github.com/sensu/sensu-go/agent"
	corev2 "github.com/sensu/sensu-go/api/core/v2"
	"github.com/sensu/sensu-go/backend/liveness"
	"github.com/sensu/sensu-go/backend/messaging"
	"github.com/sensu/sensu-go/backend/ringv2"
	"github.com/sensu/sensu-go/backend/store"
	"github.com/sensu/sensu-go/types"
	"github.com/sirupsen/logrus"
)

const (
	// DefaultHandlerCount is the default number of goroutines dedicated to
	// handling keepalive events.
	DefaultHandlerCount = 10

	// KeepaliveCheckName is the name of the check that is created when a
	// keepalive timeout occurs.
	KeepaliveCheckName = "keepalive"

	// KeepaliveHandlerName is the name of the handler that is executed when
	// a keepalive timeout occurs.
	KeepaliveHandlerName = "keepalive"

	// RegistrationCheckName is the name of the check that is created when an
	// entity sends a keepalive and the entity does not yet exist in the store.
	RegistrationCheckName = "registration"

	// RegistrationHandlerName is the name of the handler that is executed when
	// a registration event is passed to pipelined.
	RegistrationHandlerName = "registration"
)

// Keepalived is responsible for monitoring keepalive events and recording
// keepalives for entities.
type Keepalived struct {
	bus                   messaging.MessageBus
	handlerCount          int
	store                 store.Store
	deregistrationHandler string
	mu                    *sync.Mutex
	wg                    *sync.WaitGroup
	keepaliveChan         chan interface{}
	subscription          messaging.Subscription
	errChan               chan error
	livenessFactory       liveness.Factory
	ringPool              *ringv2.Pool
}

// Option is a functional option.
type Option func(*Keepalived) error

// Config configures Keepalived.
type Config struct {
	Store                 store.Store
	Bus                   messaging.MessageBus
	LivenessFactory       liveness.Factory
	DeregistrationHandler string
	RingPool              *ringv2.Pool
}

// New creates a new Keepalived.
func New(c Config, opts ...Option) (*Keepalived, error) {
	k := &Keepalived{
		store: c.Store,
		bus:   c.Bus,
		deregistrationHandler: c.DeregistrationHandler,
		livenessFactory:       c.LivenessFactory,
		keepaliveChan:         make(chan interface{}, 10),
		handlerCount:          DefaultHandlerCount,
		mu:                    &sync.Mutex{},
		errChan:               make(chan error, 1),
		ringPool:              c.RingPool,
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
	keepalives, err := k.store.GetFailingKeepalives(context.TODO())
	if err != nil {
		return err
	}

	switches := k.livenessFactory(k.Name(), k.dead, k.alive, logger)

	for _, keepalive := range keepalives {
		entityCtx := context.WithValue(context.TODO(), types.NamespaceKey, keepalive.Namespace)
		event, err := k.store.GetEventByEntityCheck(entityCtx, keepalive.Name, "keepalive")
		if err != nil {
			return err
		}

		// if there's no event, the entity was deregistered/deleted.
		if event == nil {
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

		if err := switches.Dead(ctx, keepalive.Name, ttl); err != nil {
			return fmt.Errorf("error initializing keepalive %q: %s", keepalive.Name, err)
		}
	}

	return nil
}

func (k *Keepalived) startWorkers() {
	k.wg = &sync.WaitGroup{}
	k.wg.Add(k.handlerCount)

	for i := 0; i < k.handlerCount; i++ {
		go k.processKeepalives(context.Background())
	}
}

func (k *Keepalived) processKeepalives(ctx context.Context) {
	defer k.wg.Done()

	var (
		event *types.Event
		ok    bool
	)

	switches := k.livenessFactory(k.Name(), k.alive, k.dead, logger)

	for msg := range k.keepaliveChan {
		event, ok = msg.(*types.Event)
		if !ok {
			logger.Error("keepalived received non-Event on keepalive channel")
			continue
		}

		entity := event.Entity
		if entity == nil {
			logger.Error("received keepalive with nil entity")
			continue
		}

		if err := entity.Validate(); err != nil {
			logger.WithError(err).Error("invalid keepalive event")
			continue
		}

		if err := k.handleEntityRegistration(entity); err != nil {
			logger.WithError(err).Error("error handling entity registration")
		}

		// Retrieve the keepalive timeout or use a default value in case an older
		// agent version was used, since entity.KeepaliveTimeout no longer exist
		ttl := int64(types.DefaultKeepaliveTimeout)
		if event.Check != nil {
			ttl = int64(event.Check.Timeout)
		}

		key := path.Join(entity.Namespace, entity.Name)

		if err := switches.Alive(ctx, key, ttl); err != nil {
			logger.WithError(err).Errorf("error on switch %q", key)
			continue
		}

		if err := k.handleUpdate(event); err != nil {
			logger.WithError(err).Error("error updating event")
		}
	}
}

// HandleError logs an error
func (k *Keepalived) HandleError(err error) {
	logger.WithError(err).Error(err)
}

func (k *Keepalived) handleEntityRegistration(entity *types.Entity) error {
	if entity.EntityClass != types.EntityAgentClass {
		return nil
	}

	ctx := types.SetContextFromResource(context.Background(), entity)
	fetchedEntity, err := k.store.GetEntityByName(ctx, entity.Name)

	if err != nil {
		return err
	}

	if fetchedEntity == nil {
		event := createRegistrationEvent(entity)
		err = k.bus.Publish(messaging.TopicEvent, event)
	}

	return err
}

func createKeepaliveEvent(rawEvent *types.Event) *types.Event {
	check := rawEvent.Check
	if check == nil {
		check = &types.Check{
			Interval: agent.DefaultKeepaliveInterval,
			Timeout:  types.DefaultKeepaliveTimeout,
		}
	}
	keepaliveCheck := &types.Check{
		ObjectMeta: types.ObjectMeta{
			Name:      KeepaliveCheckName,
			Namespace: rawEvent.Entity.Namespace,
		},
		Interval: check.Interval,
		Timeout:  check.Timeout,
		Handlers: []string{KeepaliveHandlerName},
		Executed: time.Now().Unix(),
		Issued:   time.Now().Unix(),
	}
	keepaliveEvent := &types.Event{
		Timestamp: time.Now().Unix(),
		Entity:    rawEvent.Entity,
		Check:     keepaliveCheck,
	}

	return keepaliveEvent
}

func createRegistrationEvent(entity *types.Entity) *types.Event {
	registrationCheck := &types.Check{
		ObjectMeta: types.ObjectMeta{
			Name:      RegistrationCheckName,
			Namespace: entity.Namespace,
		},
		Interval: 1,
		Handlers: []string{RegistrationHandlerName},
		Status:   1,
	}
	registrationEvent := &types.Event{
		Timestamp: time.Now().Unix(),
		Entity:    entity,
		Check:     registrationCheck,
	}

	return registrationEvent
}

func (k *Keepalived) alive(key string, prev liveness.State, leader bool) bool {
	lager := logger.WithFields(logrus.Fields{
		"status":          liveness.Alive.String(),
		"previous_status": prev.String()})

	namespace, name, err := parseKey(key)
	if err != nil {
		lager.Error(err)
		return false
	}

	lager = lager.WithFields(logrus.Fields{"entity": name, "namespace": namespace})
	lager.Debug("entity is alive")
	return false
}

func (k *Keepalived) dead(key string, prev liveness.State, leader bool) bool {
	lager := logger.WithFields(logrus.Fields{
		"status":          liveness.Dead.String(),
		"previous_status": prev.String()})

	namespace, name, err := parseKey(key)
	if err != nil {
		lager.Error(err)
		return false
	}

	lager = lager.WithFields(logrus.Fields{"entity": name, "namespace": namespace})
	lager.Warn("keepalive timed out")

	if !leader {
		// If this client isn't the one that flipped the keepalive switch,
		// don't do anything further.
		return false
	}

	ctx := store.NamespaceContext(context.Background(), namespace)

	entity, err := k.store.GetEntityByName(ctx, name)
	if err != nil {
		lager.WithError(err).Error("error while reading entity")
		return false
	}

	if entity == nil {
		// The entity has been deleted, there is no longer a need to
		// track keepalives for it.
		return true
	}

	deregisterer := &Deregistration{
		Store:      k.store,
		MessageBus: k.bus,
	}

	if entity.Deregister {
		if err := deregisterer.Deregister(entity); err != nil {
			lager.WithError(err).Error("error deregistering entity")
		}
		return true
	}

	currentEvent, err := k.store.GetEventByEntityCheck(ctx, name, "keepalive")
	if err != nil {
		lager.WithError(err).Error("error while reading event")
		return false
	}
	if currentEvent == nil {
		lager.Error("keepalive event not found")
		return false
	}

	// this is a real keepalive event, emit it.
	event := createKeepaliveEvent(currentEvent)
	event.Check.Status = 1
	event.Check.Output = fmt.Sprintf("No keepalive sent from %s for %v seconds (>= %v)", entity.Name, time.Now().Unix()-entity.LastSeen, event.Check.Timeout)

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
		return "", "", errors.New("bad key")
	}
	return parts[0], parts[1], nil
}

// handleUpdate sets the entity's last seen time and publishes an OK check event
// to the message bus.
func (k *Keepalived) handleUpdate(e *types.Event) error {
	entity := e.Entity

	ctx := types.SetContextFromResource(context.Background(), entity)
	if err := k.store.DeleteFailingKeepalive(ctx, e.Entity); err != nil {
		return err
	}

	entity.LastSeen = e.Timestamp

	if err := k.store.UpdateEntity(ctx, entity); err != nil {
		logger.WithError(err).Error("error updating entity in store")
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
				e.Check.Timeout = agent.DefaultKeepaliveTimeout
			}
			if err := ring.Add(ctx, entity.Name, int64(e.Check.Timeout)); err != nil {
				lager := logger.WithFields(logrus.Fields{
					"entity":       entity.Name,
					"namespace":    entity.Namespace,
					"subscription": sub,
				})
				lager.WithError(err).Error("error adding entity to ring")
			}
		}
	}

	return k.bus.Publish(messaging.TopicEventRaw, event)
}
