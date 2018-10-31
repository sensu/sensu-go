package keepalived

import (
	"context"
	"sync"
	"time"

	"github.com/sensu/sensu-go/agent"
	"github.com/sensu/sensu-go/backend/messaging"
	"github.com/sensu/sensu-go/backend/monitor"
	"github.com/sensu/sensu-go/backend/store"
	"github.com/sensu/sensu-go/types"
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
	monitorFactory        monitor.Factory
}

// Option is a functional option.
type Option func(*Keepalived) error

// Config configures Keepalived.
type Config struct {
	Store                 store.Store
	Bus                   messaging.MessageBus
	MonitorFactory        monitor.Factory
	DeregistrationHandler string
}

// New creates a new Keepalived.
func New(c Config, opts ...Option) (*Keepalived, error) {
	k := &Keepalived{
		store:                 c.Store,
		bus:                   c.Bus,
		deregistrationHandler: c.DeregistrationHandler,
		monitorFactory:        c.MonitorFactory,
		keepaliveChan:         make(chan interface{}, 10),
		handlerCount:          DefaultHandlerCount,
		mu:                    &sync.Mutex{},
		errChan:               make(chan error, 1),
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
	if err := k.initFromStore(); err != nil {
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

func (k *Keepalived) initFromStore() error {
	// For which clients were we previously alerting?
	keepalives, err := k.store.GetFailingKeepalives(context.TODO())
	if err != nil {
		return err
	}

	for _, keepalive := range keepalives {
		entityCtx := context.WithValue(context.TODO(), types.NamespaceKey, keepalive.Namespace)
		event, err := k.store.GetEventByEntityCheck(entityCtx, keepalive.EntityID, "keepalive")
		if err != nil {
			return err
		}

		// if there's no event, the entity was deregistered/deleted.
		if event == nil {
			continue
		}

		// if another backend picked it up, it will be passing.
		if event.HasCheck() {
			if event.Check.Status == 0 {
				continue
			}
		}

		// Recreate the monitor with a time offset calculated from the keepalive
		// entry timestamp minus the current time.
		d := int64(keepalive.Time) - time.Now().Unix()

		if d < 0 {
			d = 0
		}

		supervisor := k.monitorFactory(k)
		err = supervisor.Monitor(context.TODO(), keepalive.EntityID, event, d)
		if err != nil {
			return err
		}
	}

	return nil
}

func (k *Keepalived) startWorkers() {
	k.wg = &sync.WaitGroup{}
	k.wg.Add(k.handlerCount)

	for i := 0; i < k.handlerCount; i++ {
		go k.processKeepalives()
	}
}

func (k *Keepalived) processKeepalives() {
	defer k.wg.Done()

	var (
		event *types.Event
		ok    bool
	)

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
		timeout := int64(types.DefaultKeepaliveTimeout)
		if event.Check != nil {
			timeout = int64(event.Check.Timeout)
		}

		supervisor := k.monitorFactory(k)
		if err := supervisor.Monitor(context.TODO(), entity.ID, event, timeout); err != nil {
			logger.WithError(err).Error("error monitoring entity")
		}

		if err := k.handleUpdate(event); err != nil {
			logger.WithError(err).Error("error updating event")
		}
	}
}

func (k *Keepalived) HandleError(err error) {
	logger.WithError(err).Error(err)
}

func (k *Keepalived) handleEntityRegistration(entity *types.Entity) error {
	if entity.Class != types.EntityAgentClass {
		return nil
	}

	ctx := types.SetContextFromResource(context.Background(), entity)
	fetchedEntity, err := k.store.GetEntityByID(ctx, entity.ID)

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

	return k.bus.Publish(messaging.TopicEventRaw, event)
}

// HandleFailure checks if the entity should be deregistered, and emits a
// keepalive event if the entity is still valid.
func (k *Keepalived) HandleFailure(e *types.Event) error {
	// Note, we don't want to use the e parameter here as we're
	// constructing a new event instead.
	entity := e.Entity
	ctx := types.SetContextFromResource(context.Background(), entity)

	deregisterer := &Deregistration{
		Store:      k.store,
		MessageBus: k.bus,
	}
	// if the entity is supposed to be deregistered, do so.
	if entity.Deregister {
		return deregisterer.Deregister(entity)
	}

	// this is a real keepalive event, emit it.
	event := createKeepaliveEvent(e)
	event.Check.Status = 1

	if err := k.bus.Publish(messaging.TopicEventRaw, event); err != nil {
		return err
	}

	logger.WithField("entity", entity.GetID()).Info("keepalive timed out, creating keepalive event for entity")
	timeout := time.Now().Unix() + int64(e.Check.Timeout)
	return k.store.UpdateFailingKeepalive(ctx, entity, timeout)
}
