package schedulerd

import (
	"errors"
	"sync"

	"github.com/sensu/sensu-go/backend/messaging"
	"github.com/sensu/sensu-go/backend/store"
	"github.com/sensu/sensu-go/types"
)

// Schedulerd handles scheduling check requests for each check's
// configured interval and publishing to the message bus.
type Schedulerd struct {
	Store      store.Store
	MessageBus messaging.MessageBus

	cache      *Cache
	schedulers *ScheduleManager

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
	s.cache = NewCheckCache()

	// Check Schedulers
	s.schedulers = newScheduleManager(s.MessageBus, s.cache)

	// Sync
	s.errChan = make(chan error, 1)

	// Sync
	s.syncResourceScheduler = &SyncResourceScheduler{
		Interval: 30,
		Syncers: []ResourceSyncer{
			&CheckSyncer{
				Store:    s.Store,
				OnChange: s.checksUpdatedHandler,
			},
			&AssetSyncer{
				Store:    s.Store,
				OnChange: s.assetsUpdatedHandler,
			},
		},
	}
	s.syncResourceScheduler.Start()

	return nil
}

// Stop the scheduler daemon.
func (s *Schedulerd) Stop() error {
	s.syncResourceScheduler.Stop()
	s.schedulers.Stop()

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

func (s *Schedulerd) checksUpdatedHandler(checks *types.Check) {
	// Update cache
	s.cache.SetChecks(checks)

	minInterval := s.syncResourceScheduler.Interval
	for _, check := range checks {
		// Ensure check scheduler has the check
		s.scheduler.Run(check)

		// Find min interval
		if check.Interval < minInterval {
			minInterval = check.Interval
		}
	}

	// Update sync interval
	s.syncResourceScheduler.Interval = minInterval
}

func (s *Schedulerd) assetsUpdatedHandler(assets *types.Asset) {
	cache.SetAssets(assets) // Update cache
}
