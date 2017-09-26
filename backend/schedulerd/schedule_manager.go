package schedulerd

import (
	"sync"
	"sync/atomic"

	"github.com/sensu/sensu-go/backend/messaging"
	"github.com/sensu/sensu-go/types"
)

// ScheduleManager manages all the check schedulers
type ScheduleManager struct {
	items map[string]*CheckScheduler
	mutex *sync.Mutex

	stopped *atomic.Value
	wg      *sync.WaitGroup

	newSchedulerFn func(check *types.CheckConfig) *CheckScheduler
}

// NewScheduleManager ...
func NewScheduleManager(msgBus messaging.MessageBus, stateMngr *StateManager) *ScheduleManager {
	wg := &sync.WaitGroup{}
	stopped := &atomic.Value{}

	newSchedulerFn := func(check *types.CheckConfig) *CheckScheduler {
		return &CheckScheduler{
			CheckName:    check.Name,
			CheckEnv:     check.Environment,
			CheckOrg:     check.Organization,
			MessageBus:   msgBus,
			WaitGroup:    wg,
			StateManager: stateMngr,
		}
	}

	manager := &ScheduleManager{
		newSchedulerFn: newSchedulerFn,

		items:   map[string]*CheckScheduler{},
		mutex:   &sync.Mutex{},
		stopped: stopped,
		wg:      wg,
	}

	// Listen to state updates and add schedulers if necassarily
	stateMngr.OnChecksChange = func(state *SchedulerState) {
		for _, check := range state.checks {
			manager.Run(check)
		}
	}

	return manager
}

// Run starts a new scheduler for the given check
func (mngrPtr *ScheduleManager) Run(check *types.CheckConfig) error {
	// Guard against updates while the daemon is shutting down
	if mngrPtr.stopped.Load().(bool) {
		return nil
	}

	// Avoid competing updates
	mngrPtr.mutex.Lock()
	defer mngrPtr.mutex.Unlock()

	// Guard against creating a duplicate scheduler; schedulers are able to update
	// their internal state with any changes that occur to their associated check.
	key := concatUniqueKey(check.Name, check.Organization, check.Environment)
	if existing := mngrPtr.items[key]; existing != nil {
		return nil
	}

	// Create new scheduler & start it
	scheduler := mngrPtr.newSchedulerFn(check)

	// Start scheduling check
	if err := scheduler.Start(check.Interval); err != nil {
		return err
	}

	// Register new check scheduler
	mngrPtr.items[key] = scheduler
	return nil
}

// Start ...
func (mngrPtr *ScheduleManager) Start() {
	logger.Info("starting scheduler manager")
	mngrPtr.stopped.Store(false)
}

// Stop closes all the schedulers
func (mngrPtr *ScheduleManager) Stop() {
	// Await any pending updates before shutting down
	mngrPtr.mutex.Lock()
	mngrPtr.stopped.Store(true)

	for n, scheduler := range mngrPtr.items {
		delete(mngrPtr.items, n)
		scheduler.Stop()
	}

	mngrPtr.wg.Wait()
}
