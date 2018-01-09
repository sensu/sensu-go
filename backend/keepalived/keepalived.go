package keepalived

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"time"

	"github.com/sensu/sensu-go/backend/messaging"
	"github.com/sensu/sensu-go/backend/monitor"
	"github.com/sensu/sensu-go/backend/store"
	"github.com/sensu/sensu-go/types"
)

const (
	// DefaultHandlerCount is the default number of goroutines dedicated to
	// handling keepalive events.
	DefaultHandlerCount = 10

	// DefaultKeepaliveTimeout is the amount of time we consider a Keepalive
	// valid for.
	DefaultKeepaliveTimeout = 120 // seconds

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

// MonitorFactoryFunc takes an entity and returns a Monitor. Keepalived can
// take a MonitorFactoryFunc that stubs/mocks a Deregisterer and/or an
// EventCreator to make it easier to test.
type MonitorFactoryFunc func(e *types.Entity) *Monitor

// Keepalived is responsible for monitoring keepalive events and recording
// keepalives for entities.
type Keepalived struct {
	MessageBus            messaging.MessageBus
	HandlerCount          int
	Store                 store.Store
	DeregistrationHandler string
	MonitorFactory        MonitorFactoryFunc

	mu            *sync.Mutex
	monitors      map[string]*Monitor
	wg            *sync.WaitGroup
	keepaliveChan chan interface{}
	errChan       chan error
}

// Start starts the daemon, returning an error if preconditions for startup
// fail.
func (k *Keepalived) Start() error {
	if k.MessageBus == nil {
		return errors.New("no message bus found")
	}

	if k.Store == nil {
		return errors.New("no keepalive store found")
	}

	if k.MonitorFactory == nil {
		k.MonitorFactory = func(e *types.Entity) *KeepaliveMonitor {
			return &KeepaliveMonitor{
				Entity: e,
				Deregisterer: &Deregistration{
					Store:      k.Store,
					MessageBus: k.MessageBus,
				},
				// this should be the monitor struct with methods
				EventCreator: &Monitor{},

				EventCreator: &MessageBusEventCreator{
					MessageBus: k.MessageBus,
				},
				MessageBus: k.MessageBus,
				Store:      k.Store,
			}
		}
	}

	k.keepaliveChan = make(chan interface{}, 10)
	err := k.MessageBus.Subscribe(messaging.TopicKeepalive, "keepalived", k.keepaliveChan)
	if err != nil {
		return err
	}

	if k.HandlerCount == 0 {
		k.HandlerCount = DefaultHandlerCount
	}

	k.mu = &sync.Mutex{}
	k.monitors = map[string]*KeepaliveMonitor{}

	if err := k.initFromStore(); err != nil {
		return err
	}

	k.startWorkers()

	k.startMonitorSweeper()

	k.errChan = make(chan error, 1)
	return nil
}

// Stop stops the daemon, returning an error if one was encountered during
// shutdown.
func (k *Keepalived) Stop() error {
	close(k.keepaliveChan)
	k.wg.Wait()
	for _, monitor := range k.monitors {
		go monitor.Stop()
	}
	err := k.MessageBus.Unsubscribe(messaging.TopicKeepalive, "keepalived")
	close(k.errChan)
	return err
}

// Status returns nil if the Daemon is healthy, otherwise it returns an error.
func (k *Keepalived) Status() error {
	return nil
}

// Err returns a channel that the caller can use to listen for terminal errors
// indicating a premature shutdown of the Daemon.
func (k *Keepalived) Err() <-chan error {
	return k.errChan
}

func (k *Keepalived) initFromStore() error {
	// For which clients were we previously alerting?
	keepalives, err := k.Store.GetFailingKeepalives(context.TODO())
	if err != nil {
		return err
	}

	for _, keepalive := range keepalives {
		entityCtx := context.WithValue(context.TODO(), types.OrganizationKey, keepalive.Organization)
		entityCtx = context.WithValue(entityCtx, types.EnvironmentKey, keepalive.Environment)
		event, err := k.Store.GetEventByEntityCheck(entityCtx, keepalive.EntityID, "keepalive")
		if err != nil {
			return err
		}

		// if there's no event, the entity was deregistered/deleted.
		if event == nil {
			continue
		}

		// if another backend picked it up, it will be passing.
		if event.Check.Status == 0 {
			continue
		}

		// Recreate the monitor and reset its timer to alert when it's going to
		// timeout.
		monitor := k.MonitorFactory(event.Entity)
		monitor.Reset(keepalive.Time)
		k.monitors[keepalive.EntityID] = monitor
	}

	return nil
}

func (k *Keepalived) startWorkers() {
	k.wg = &sync.WaitGroup{}
	k.wg.Add(k.HandlerCount)

	for i := 0; i < k.HandlerCount; i++ {
		go k.processKeepalives()
	}
}

func (k *Keepalived) processKeepalives() {
	defer k.wg.Done()

	var (
		monitor *KeepaliveMonitor
		event   *types.Event
		ok      bool
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

		k.mu.Lock()
		monitor, ok = k.monitors[entity.ID]
		// create if it doesn't exist
		if !ok || monitor.IsStopped() {
			monitor = k.MonitorFactory(entity)
			monitor.Start()
			k.monitors[entity.ID] = monitor
		}
		k.mu.Unlock()

		if err := monitor.Update(event); err != nil {
			logger.WithError(err).Error("error monitoring entity")
		}
	}
}

func (k *Keepalived) handleEntityRegistration(entity *types.Entity) error {
	if entity.Class != types.EntityAgentClass {
		return nil
	}

	ctx := types.SetContextFromResource(context.Background(), entity)
	fetchedEntity, err := k.Store.GetEntityByID(ctx, entity.ID)

	if err != nil {
		return err
	}

	if fetchedEntity == nil {
		event := createRegistrationEvent(entity)
		err = k.MessageBus.Publish(messaging.TopicEvent, event)
	}

	return err
}

// startMonitorSweeper spins off into oblivion if Keepalived is stopped until
// the monitors map is empty, and then the goroutine stops.
func (k *Keepalived) startMonitorSweeper() {
	go func() {
		timer := time.NewTimer(10 * time.Minute)
		for {
			<-timer.C
			for key, monitor := range k.monitors {
				if monitor.IsStopped() {
					k.mu.Lock()
					delete(k.monitors, key)
					k.mu.Unlock()
				}
			}
		}
	}()
}

func createKeepaliveEvent(entity *types.Entity) *types.Event {
	keepaliveCheck := &types.Check{
		Config: &types.CheckConfig{
			Name:         KeepaliveCheckName,
			Interval:     entity.KeepaliveTimeout,
			Handlers:     []string{KeepaliveHandlerName},
			Environment:  entity.Environment,
			Organization: entity.Organization,
		},
		Status: 1,
	}
	keepaliveEvent := &types.Event{
		Timestamp: time.Now().Unix(),
		Entity:    entity,
		Check:     keepaliveCheck,
	}

	return keepaliveEvent
}

func createRegistrationEvent(entity *types.Entity) *types.Event {
	registrationCheck := &types.Check{
		Config: &types.CheckConfig{
			Name:         RegistrationCheckName,
			Interval:     entity.KeepaliveTimeout,
			Handlers:     []string{RegistrationHandlerName},
			Environment:  entity.Environment,
			Organization: entity.Organization,
		},
		Status: 1,
	}
	registrationEvent := &types.Event{
		Timestamp: time.Now().Unix(),
		Entity:    entity,
		Check:     registrationCheck,
	}

	return registrationEvent
}

type keepaliveUpdateHandler struct{}

func (handler *keepaliveUpdateHandler) HandleUpdate(e *types.Event) error {
	entity := event.Entity

	if atomic.CompareAndSwapInt32(&monitorPtr.failing, 1, 0) {
		if err := monitorPtr.Store.DeleteFailingKeepalive(context.Background(), entity); err != nil {
			logger.Debug(err)
		}
	}

	entity.LastSeen = event.Timestamp
	ctx := types.SetContextFromResource(context.Background(), entity)

	if err := monitorPtr.Store.UpdateEntity(ctx, entity); err != nil {
		logger.WithError(err).Error("error updating entity in store")
	}
}

func (handler *keepaliveUpdateHandler) HandleFailure(t time.T) error {
	// timed out keepalive
	ctx := types.SetContextFromResource(context.Background(), monitorPtr.Entity)
	// if the entity is supposed to be deregistered, do so.
	if monitorPtr.Entity.Deregister {
		if err = monitorPtr.Deregisterer.Deregister(monitorPtr.Entity); err != nil {
			logger.WithError(err).Error("error deregistering entity")
		}
		monitorPtr.Stop()
		return nil
	}

	// this is a real keepalive event, emit it.
	if err = monitorPtr.EventCreator.Warn(monitorPtr.Entity); err != nil {
		logger.WithError(err).Error("error sending keepalive event")
	}
	timeout = time.Now().Unix() + int64(monitorPtr.Entity.KeepaliveTimeout)
	if err = monitorPtr.Store.UpdateFailingKeepalive(ctx, monitorPtr.Entity, timeout); err != nil {
		logger.WithError(err).Error("error updating failing keepalive in store")
	}
}
