package keepalived

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/sensu/sensu-go/backend/messaging"
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
)

// MonitorFactoryFunc takes an entity and returns a Monitor. Keepalived can
// take a MonitorFactoryFunc that stubs/mocks a Deregisterer and/or an
// EventCreator to make it easier to test.
type MonitorFactoryFunc func(e *types.Entity) *KeepaliveMonitor

// Keepalived is responsible for monitoring keepalive events and recording
// keepalives for entities.
type Keepalived struct {
	MessageBus            messaging.MessageBus
	HandlerCount          int
	Store                 store.Store
	DeregistrationHandler string
	MonitorFactory        MonitorFactoryFunc

	mu            *sync.Mutex
	monitors      map[string]*KeepaliveMonitor
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
				EventCreator: &MessageBusEventCreator{
					MessageBus: k.MessageBus,
				},
				Store: k.Store,
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
	k.MessageBus.Unsubscribe(messaging.TopicKeepalive, "keepalived")
	close(k.errChan)
	return nil
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

func (k *Keepalived) startWorkers() {
	// concurrent access to entityChannels map is problematic
	// because of race conditions :(
	k.HandlerCount = 1

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
		if err := entity.Validate(); err != nil {
			logger.WithError(err).Error("invalid keepalive event")
			continue
		}
		entity.LastSeen = event.Timestamp

		ctx := context.WithValue(context.Background(), types.OrganizationKey, event.Entity.Organization)
		if err := k.Store.UpdateEntity(ctx, event.Entity); err != nil {
			logger.WithError(err).Error("error updating entity in store")
		}

		// TODO(greg): This is a good candidate for a concurrent map
		// when it's released.
		k.mu.Lock()
		monitor, ok = k.monitors[entity.ID]
		if !ok {
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
