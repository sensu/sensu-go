package schedulerd

import (
	"context"
	"strings"
	"sync"

	"github.com/sensu/sensu-go/backend/messaging"
	"github.com/sensu/sensu-go/backend/store"
	"github.com/sensu/sensu-go/types"
)

// CheckWatcher manages all the check schedulers
type CheckWatcher struct {
	items map[string]*CheckScheduler
	store store.Store
	bus   messaging.MessageBus
	mu    sync.Mutex
	ctx   context.Context
}

// NewCheckWatcher creates a new ScheduleManager.
func NewCheckWatcher(msgBus messaging.MessageBus, store store.Store, ctx context.Context) *CheckWatcher {
	watcher := &CheckWatcher{
		store: store,
		items: make(map[string]*CheckScheduler),
		bus:   msgBus,
		ctx:   ctx,
	}

	return watcher
}

// startScheduler starts a new scheduler for the given check. It assumes mu is locked.
func (c *CheckWatcher) startScheduler(check *types.CheckConfig) error {
	// Guard against updates while the daemon is shutting down
	if err := c.ctx.Err(); err != nil {
		return err
	}

	// Guard against creating a duplicate scheduler; schedulers are able to update
	// their internal state with any changes that occur to their associated check.
	key := concatUniqueKey(check.Name, check.Namespace)
	if existing := c.items[key]; existing != nil {
		return nil
	}

	// Create new scheduler & start it
	scheduler := NewCheckScheduler(c.store, c.bus, check, c.ctx)

	// Start scheduling check
	if err := scheduler.Start(); err != nil {
		return err
	}

	// Register new check scheduler
	c.items[key] = scheduler
	return nil
}

// Start starts the CheckWatcher.
func (c *CheckWatcher) Start() error {
	// for each check
	checkConfigs, err := c.store.GetCheckConfigs(c.ctx)
	if err != nil {
		return err
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	for _, cfg := range checkConfigs {
		if err := c.startScheduler(cfg); err != nil {
			return err
		}
	}

	go c.startWatcher()

	return nil
}

func (c *CheckWatcher) startWatcher() {
	watchChan := c.store.GetCheckConfigWatcher(c.ctx)
	for {
		select {
		case watchEvent, ok := <-watchChan:
			if !ok {
				// The watchChan has closed. Restart the watcher.
				watchChan = c.store.GetCheckConfigWatcher(c.ctx)
				continue
			}
			c.handleWatchEvent(watchEvent)
		case <-c.ctx.Done():
			c.mu.Lock()
			defer c.mu.Unlock()
			for _, scheduler := range c.items {
				if err := scheduler.Stop(); err != nil {
					logger.Debug(err)
				}
			}
			return
		}
	}
}

func (c *CheckWatcher) handleWatchEvent(watchEvent store.WatchEventCheckConfig) {
	check := watchEvent.CheckConfig
	key := concatUniqueKey(check.Name, check.Namespace)

	c.mu.Lock()
	defer c.mu.Unlock()

	switch watchEvent.Action {
	case store.WatchCreate:
		// we need to spin up a new CheckScheduler for the newly created check
		if err := c.startScheduler(check); err != nil {
			logger.WithError(err).Error("unable to start check scheduler")
		}
	case store.WatchUpdate:
		// Interrupt the check scheduler, causing the check to execute and the timer to be reset.
		sched, ok := c.items[key]
		if ok {
			sched.Interrupt()
		}
	case store.WatchDelete:
		// Call stop on the scheduler.
		sched, ok := c.items[key]
		if ok {
			sched.Stop()
			delete(c.items, key)
		}
	}
}

func concatUniqueKey(args ...string) string {
	return strings.Join(args, "-")
}
