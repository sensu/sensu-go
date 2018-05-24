package schedulerd

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"sync/atomic"

	"github.com/sensu/sensu-go/backend/messaging"
	"github.com/sensu/sensu-go/backend/store"
	"github.com/sensu/sensu-go/types"
)

// CheckSchedulerManager manages all the check schedulers
type CheckSchedulerManager struct {
	items map[string]*CheckScheduler
	store store.Store
	mutex *sync.Mutex

	stopped *atomic.Value
	wg      *sync.WaitGroup

	newSchedulerFn func(check *types.CheckConfig) *CheckScheduler
}

// NewCheckSchedulerManager creates a new ScheduleManager.
func NewCheckSchedulerManager(msgBus messaging.MessageBus, store store.Store) *CheckSchedulerManager {
	wg := &sync.WaitGroup{}
	stopped := &atomic.Value{}

	newSchedulerFn := func(check *types.CheckConfig) *CheckScheduler {
		return &CheckScheduler{
			checkName:     check.Name,
			checkEnv:      check.Environment,
			checkOrg:      check.Organization,
			checkInterval: check.Interval,
			checkCron:     check.Cron,
			lastCronState: check.Cron,
			store:         store,
			bus:           msgBus,
			wg:            wg,
		}
	}

	manager := &CheckSchedulerManager{
		newSchedulerFn: newSchedulerFn,
		store:          store,
		items:          map[string]*CheckScheduler{},
		mutex:          &sync.Mutex{},
		stopped:        stopped,
		wg:             wg,
	}

	return manager
}

// Run starts a new scheduler for the given check
func (mngrPtr *CheckSchedulerManager) run(check *types.CheckConfig) error {
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
	if err := scheduler.Start(); err != nil {
		return err
	}

	// Register new check scheduler
	mngrPtr.items[key] = scheduler
	return nil
}

// Start ...
func (mngrPtr *CheckSchedulerManager) Start(ctx context.Context) error {
	logger.Info("starting scheduler manager")
	mngrPtr.stopped.Store(false)

	// for each check
	checkConfigs, err := mngrPtr.store.GetCheckConfigs(ctx)
	if err != nil {
		return err
	}

	for _, cfg := range checkConfigs {
		if err := mngrPtr.run(cfg); err != nil {
			return err
		}
	}

	go mngrPtr.startWatcher(ctx)

	return nil
}

func (mngrPtr *CheckSchedulerManager) startWatcher(ctx context.Context) {
	watchChan := mngrPtr.store.GetCheckConfigWatcher(context.Background())
	for {
		select {
		case watchEvent := <-watchChan:
			mngrPtr.handleWatchEvent(watchEvent)
		case <-ctx.Done():
			for _, scheduler := range mngrPtr.items {
				if err := scheduler.Stop(); err != nil {
					logger.Debug(err)
				}
			}
			return
		}
	}
}

func (mngrPtr *CheckSchedulerManager) handleWatchEvent(watchEvent store.WatchEventCheckConfig) {
	check := watchEvent.CheckConfig
	fmt.Println(check)
	key := concatUniqueKey(check.Name, check.Organization, check.Environment)
	fmt.Println("first: ", key)

	switch watchEvent.Action {
	case store.WatchCreate:
		// we need to spin up a new CheckScheduler for the newly created check
		if err := mngrPtr.run(check); err != nil {
			logger.WithError(err).Error("unable to start check scheduler")
		}

	case store.WatchUpdate:
		// Interrupt the check scheduler, causing the check to execute and the timer to be reset.
		mngrPtr.items[key].Interrupt()

	case store.WatchDelete:
		// Call stop on the scheduler.
		fmt.Println(mngrPtr.items)
		fmt.Println(key)
		mngrPtr.items[key].Stop()
	}
}

// Stop closes all the schedulers
func (mngrPtr *CheckSchedulerManager) Stop() {
	mngrPtr.stopped.Store(true)
	mngrPtr.wg.Wait()
}

func concatUniqueKey(args ...string) string {
	return strings.Join(args, "-")
}
