package schedulerd

import (
	"context"
	"errors"

	"github.com/sensu/sensu-go/backend/messaging"
	"github.com/sensu/sensu-go/backend/queue"
	"github.com/sensu/sensu-go/types"
)

// Store specifies the storage requirements for Schedulerd.
type Store interface {
	StateManagerStore
	types.RingGetter
	queue.Get
}

// Schedulerd handles scheduling check requests for each check's
// configured interval and publishing to the message bus.
type Schedulerd struct {
	Store      Store
	MessageBus messaging.MessageBus

	stateManager         *StateManager
	schedulerManager     *ScheduleManager
	adhocRequestExecutor *AdhocRequestExecutor

	errChan chan error
}

// Start the Scheduler daemon.
func (s *Schedulerd) Start() error {
	ctx := context.TODO()

	if s.Store == nil {
		return errors.New("no store available")
	}

	if s.MessageBus == nil {
		return errors.New("no message bus found")
	}

	// State
	s.stateManager = NewStateManager(s.Store)

	// Check Schedulers
	s.schedulerManager = NewScheduleManager(s.MessageBus, s.stateManager, s.Store)

	// Adhoc Request Executor
	s.adhocRequestExecutor = NewAdhocRequestExecutor(ctx, s.Store, s.MessageBus)

	// Sync
	s.errChan = make(chan error, 1)

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
