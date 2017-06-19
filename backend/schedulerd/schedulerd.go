package schedulerd

import (
	"errors"

	"github.com/sensu/sensu-go/backend/messaging"
	"github.com/sensu/sensu-go/backend/store"
	"github.com/sensu/sensu-go/types"
)

// Schedulerd handles scheduling check requests for each check's
// configured interval and publishing to the message bus.
type Schedulerd struct {
	Store      store.Store
	MessageBus messaging.MessageBus

	stateManager          *StateManager
	schedulerManager      *ScheduleManager
	syncResourceScheduler *SyncResourceScheduler

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

	// Cache
	s.stateManager = NewStateManager()

	// Check Schedulers
	s.schedulerManager = newScheduleManager(s.MessageBus, s.stateManager)

	// Sync
	s.errChan = make(chan error, 1)

	// Sync
	s.syncResourceScheduler = &SyncResourceScheduler{
		Interval: 30,
		Syncers: []ResourceSync{
			&SyncronizeChecks{
				Store:    s.Store,
				OnUpdate: s.checksUpdatedHandler,
			},
			&SyncronizeAssets{
				Store:    s.Store,
				OnUpdate: s.assetsUpdatedHandler,
			},
		},
	}
	s.syncResourceScheduler.Start()

	return nil
}

// Stop the scheduler daemon.
func (s *Schedulerd) Stop() error {
	s.syncResourceScheduler.Stop()
	s.schedulerManager.Stop()

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

func (s *Schedulerd) checksUpdatedHandler(checks []*types.CheckConfig) {
	// Update state
	s.stateManager.Update(func(state *SchedulerState) {
		state.SetChecks(checks)
	})

	minInterval := s.syncResourceScheduler.Interval
	for _, check := range checks {
		// Ensure check scheduler has the check
		s.schedulerManager.Run(check)

		// Find min interval
		if check.Interval < minInterval {
			minInterval = check.Interval
		}
	}

	// Update sync interval
	s.syncResourceScheduler.Interval = minInterval
}

func (s *Schedulerd) assetsUpdatedHandler(assets []*types.Asset) {
	// Update state
	s.stateManager.Update(func(state *SchedulerState) {
		state.SetAssets(assets)
	})
}
