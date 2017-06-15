package schedulerd

import (
	"sync"
	"sync/atomic"

	"github.com/coreos/etcd/store"
	"github.com/sensu/sensu-go/backend/messaging"
	"github.com/sensu/sensu-go/types"
)

// ScheduleManager manages all the check schedulers
type ScheduleManager struct {
	items map[string]*CheckScheduler
	mutex *sync.Mutex

	stopped *atomic.Value
	wg      *sync.WaitGroup

	createScheduler func(check *types.CheckConfig) *CheckScheduler
}

func newScheduleManager(msgBus messaging.MessageBus, cache *CheckCache) *ScheduleManager {
	wg := &sync.WaitGroup{}
	stopped := &atomic.Value{}
	stopped.Store(false)

	createScheduler := func(check *types.CheckConfig) *CheckScheduler {
		return &CheckScheduler{
			CheckName:  check.Name,
			CheckOrg:   check.Organization,
			MessageBus: msgBus,
			WaitGroup:  wg,
			Cache:      cache,
		}
	}

	collection := &ScheduleManager{
		createScheduler: createScheduler,
		items:           map[string]*CheckScheduler{},
		mutex:           &sync.Mutex{},
		stopped:         stopped,
		wg:              wg,
	}

	return collection
}

// Run starts a new scheduler for the given check
func (collectionPtr *ScheduleManager) Run(check *types.CheckConfig) error {
	// Guard against updates while the daemon is shutting down
	if collectionPtr.stopped.Load().(bool) {
		return nil
	}

	// Avoid competing updates
	collectionPtr.mutex.Lock()
	defer collectionPtr.mutex.Unlock()

	// Guard against creating a duplicate scheduler; schedulers are able to update
	// their internal state with any changes that occur to their associated check.
	key := concatUniqueKey(check.Name, check.Organization)
	if existing := collectionPtr.items[key]; existing != nil {
		return nil
	}

	// Create new scheduler & start it
	scheduler := collectionPtr.createScheduler(check)
	if err := scheduler.Start(check.Interval); err != nil {
		return err
	}

	// Register new check scheduler
	collectionPtr.items[check.Name] = scheduler
	return nil
}

// Stop closes all the schedulers
func (collectionPtr *ScheduleManager) Stop() {
	// Await any pending updates before shutting down
	collectionPtr.mutex.Lock()
	collectionPtr.stopped.Store(true)

	for n, scheduler := range collectionPtr.items {
		delete(collectionPtr.items, n)
		scheduler.Stop()
	}

	collectionPtr.wg.Wait()
}
