package schedulerd

import (
	"context"
	"strings"
	"sync"

	corev2 "github.com/sensu/core/v2"
	"github.com/sensu/sensu-go/backend/messaging"
	"github.com/sensu/sensu-go/backend/ringv2"
	"github.com/sensu/sensu-go/backend/secrets"
	"github.com/sensu/sensu-go/backend/store"
	cachev2 "github.com/sensu/sensu-go/backend/store/cache/v2"
)

// CheckWatcher manages all the check schedulers
type CheckWatcher struct {
	items                  map[string]Scheduler
	store                  store.Store
	bus                    messaging.MessageBus
	mu                     sync.Mutex
	ctx                    context.Context
	ringPool               *ringv2.RingPool
	entityCache            *cachev2.Resource
	secretsProviderManager *secrets.ProviderManager
}

// NewCheckWatcher creates a new ScheduleManager.
func NewCheckWatcher(ctx context.Context, msgBus messaging.MessageBus, store store.Store, pool *ringv2.RingPool, cache *cachev2.Resource, secretsProviderManager *secrets.ProviderManager) *CheckWatcher {
	watcher := &CheckWatcher{
		store:                  store,
		items:                  make(map[string]Scheduler),
		bus:                    msgBus,
		ctx:                    ctx,
		ringPool:               pool,
		entityCache:            cache,
		secretsProviderManager: secretsProviderManager,
	}

	return watcher
}

// startScheduler starts a new scheduler for the given check. It assumes mu is locked.
func (c *CheckWatcher) startScheduler(check *corev2.CheckConfig) error {
	// Guard against updates while the daemon is shutting down
	if err := c.ctx.Err(); err != nil {
		return err
	}

	// Guard against creating a duplicate scheduler; schedulers are able to update
	// their internal state with any changes that occur to their associated check.
	key := concatUniqueKey(check.Name, check.Namespace)
	if existing := c.items[key]; existing != nil {
		if existing.Type() == GetSchedulerType(check) {
			logger.Error("scheduler already exists")
			return nil
		}
		if err := existing.Stop(); err != nil {
			return err
		}
	}

	var scheduler Scheduler

	switch GetSchedulerType(check) {
	case IntervalType:
		scheduler = NewIntervalScheduler(c.ctx, c.store, c.bus, check, c.entityCache, c.secretsProviderManager)
	case CronType:
		scheduler = NewCronScheduler(c.ctx, c.store, c.bus, check, c.entityCache, c.secretsProviderManager)
	case RoundRobinIntervalType:
		scheduler = NewRoundRobinIntervalScheduler(c.ctx, c.store, c.bus, c.ringPool, check, c.entityCache, c.secretsProviderManager)
	case RoundRobinCronType:
		scheduler = NewRoundRobinCronScheduler(c.ctx, c.store, c.bus, c.ringPool, check, c.entityCache, c.secretsProviderManager)
	default:
		logger.Error("bad scheduler type, falling back to interval scheduler")
		scheduler = NewIntervalScheduler(c.ctx, c.store, c.bus, check, c.entityCache, c.secretsProviderManager)
	}

	// Start scheduling check
	scheduler.Start()

	// Register new check scheduler
	c.items[key] = scheduler
	return nil
}

// Start starts the CheckWatcher.
func (c *CheckWatcher) Start() error {
	// for each check
	checkConfigs, err := c.store.GetCheckConfigs(c.ctx, &store.SelectionPredicate{})
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
			if ok {
				c.handleWatchEvent(watchEvent)
			}
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

	if check == nil {
		logger.Error("nil check config received from check config watcher")
		return
	}

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
		logger.Info("check configs updated")
		sched, ok := c.items[key]
		if !ok {
			logger.Info("starting new scheduler")
			if err := c.startScheduler(check); err != nil {
				logger.WithError(err).Error("unable to start check scheduler")
			}
			return
		}
		if sched.Type() == GetSchedulerType(check) {
			logger.Info("restarting scheduler")
			sched.Interrupt(check)
		} else {
			logger.Info("stopping existing scheduler, starting new scheduler")
			if err := sched.Stop(); err != nil {
				logger.WithError(err).Error("error stopping check scheduler")
			}
			delete(c.items, key)
			if err := c.startScheduler(check); err != nil {
				logger.WithError(err).Error("unable to start check scheduler")
			}
		}
	case store.WatchDelete:
		// Call stop on the scheduler.
		sched, ok := c.items[key]
		if ok {
			if err := sched.Stop(); err != nil {
				logger.WithError(err).Error("error stopping check scheduler")
			}
			delete(c.items, key)
		}
	}
}

func concatUniqueKey(args ...string) string {
	return strings.Join(args, "-")
}
