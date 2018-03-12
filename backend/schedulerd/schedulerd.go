package schedulerd

import (
	"context"

	"github.com/sensu/sensu-go/backend/messaging"
	"github.com/sensu/sensu-go/backend/store"
	"github.com/sensu/sensu-go/types"
)

// Schedulerd handles scheduling check requests for each check's
// configured interval and publishing to the message bus.
type Schedulerd struct {
	store                store.Store
	ringGetter           types.RingGetter
	queueGetter          types.QueueGetter
	bus                  messaging.MessageBus
	stateManager         *StateManager
	schedulerManager     *ScheduleManager
	adhocRequestExecutor *AdhocRequestExecutor
	errChan              chan error
}

// Option is a functional option.
type Option func(*Schedulerd) error

// Config configures Schedulerd.
type Config struct {
	Store       store.Store
	RingGetter  types.RingGetter
	QueueGetter types.QueueGetter
	Bus         messaging.MessageBus
}

// New creates a new Schedulerd.
func New(c Config, opts ...Option) (*Schedulerd, error) {
	s := &Schedulerd{
		store:        c.Store,
		ringGetter:   c.RingGetter,
		queueGetter:  c.QueueGetter,
		bus:          c.Bus,
		errChan:      make(chan error, 1),
		stateManager: NewStateManager(c.Store),
	}
	s.schedulerManager = NewScheduleManager(s.bus, s.stateManager, s.ringGetter)
	for _, o := range opts {
		if err := o(s); err != nil {
			return nil, err
		}
	}
	return s, nil
}

// Start the Scheduler daemon.
func (s *Schedulerd) Start() error {
	ctx := context.TODO()

	// Adhoc Request Executor
	s.adhocRequestExecutor = NewAdhocRequestExecutor(ctx, s.store, s.queueGetter.GetQueue(adhocQueueName), s.bus)

	// Start
	s.schedulerManager.Start()
	s.stateManager.Start(ctx)

	return nil
}

// Stop the scheduler daemon.
func (s *Schedulerd) Stop() error {
	err := s.stateManager.Stop()
	s.schedulerManager.Stop()
	s.adhocRequestExecutor.Stop()

	close(s.errChan)
	return err
}

// Status returns the health of the scheduler daemon.
func (s *Schedulerd) Status() error {
	return nil
}

// Err returns a channel on which to listen for terminal errors.
func (s *Schedulerd) Err() <-chan error {
	return s.errChan
}
