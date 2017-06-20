package schedulerd

import (
	"errors"

	"github.com/sensu/sensu-go/backend/messaging"
	"github.com/sensu/sensu-go/backend/store"
)

// Schedulerd handles scheduling check requests for each check's
// configured interval and publishing to the message bus.
type Schedulerd struct {
	Store      store.Store
	MessageBus messaging.MessageBus

	stateManager *StateManager
	scheduler    *Scheduler

	errChan chan error
}

// Start the Scheduler daemon.
func (s *Schedulerd) Start() error {
	if s.Store == nil {
		return errors.New("no store available")
	}

	if s.MessageBus == nil {
		return errors.New("no message bus found")
	}

	// State
	s.stateManager = NewStateManager(s.Store)

	// Scheduler
	s.scheduler = &Scheduler{
		StateManager: s.stateManager,
		MessageBus:   s.MessageBus,
	}

	// Sync
	s.errChan = make(chan error, 1)

	// Start
	s.scheduler.Start()
	s.stateManager.Start()

	return nil
}

// Stop the scheduler daemon.
func (s *Schedulerd) Stop() error {
	s.stateManager.Stop()
	s.scheduler.Stop()

	close(s.errChan)
	return nil
}

// Status returns the health of the scheduler daemon.
func (s *Schedulerd) Status() error {
	return nil
}

// Err returns a channel on which to listen for terminal errors.
func (s *Schedulerd) Err() <-chan error {
	return s.errChan
}
