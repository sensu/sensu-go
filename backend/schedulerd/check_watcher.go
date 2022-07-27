package schedulerd

import (
	"context"
	"strings"
	"sync"

	corev2 "github.com/sensu/sensu-go/api/core/v2"
	"github.com/sensu/sensu-go/backend/messaging"
	"github.com/sensu/sensu-go/backend/ringv2"
	"github.com/sensu/sensu-go/backend/secrets"
	"github.com/sensu/sensu-go/backend/store"
	cachev2 "github.com/sensu/sensu-go/backend/store/cache/v2"
	storev2 "github.com/sensu/sensu-go/backend/store/v2"
)

// CheckWatcher manages all the check schedulers
type CheckWatcher struct {
	items                  map[string]Scheduler
	store                  storev2.Interface
	bus                    messaging.MessageBus
	mu                     sync.Mutex
	ctx                    context.Context
	ringPool               *ringv2.RingPool
	entityCache            *cachev2.Resource
	secretsProviderManager *secrets.ProviderManager
}

// NewCheckWatcher creates a new ScheduleManager.
func NewCheckWatcher(ctx context.Context, msgBus messaging.MessageBus, store storev2.Interface, pool *ringv2.RingPool, cache *cachev2.Resource, secretsProviderManager *secrets.ProviderManager) *CheckWatcher {
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
	checkConfigs := []*corev2.CheckConfig{}
	req := storev2.NewResourceRequestFromResource(&corev2.CheckConfig{})
	list, err := c.store.List(context.TODO(), req, &store.SelectionPredicate{})
	if err != nil {
		return err
	}
	if err := list.UnwrapInto(&checkConfigs); err != nil {
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
	watchChan := c.store.Watch(c.ctx, storev2.NewResourceRequestFromResource(&corev2.CheckConfig{}))
	for {
		select {
		case watchEvents, ok := <-watchChan:
			if ok {
				for _, watchEvent := range watchEvents {
					c.handleWatchEvent(watchEvent)
				}
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

func (c *CheckWatcher) handleWatchEvent(watchEvent storev2.WatchEvent) {
	// TODO(ccressent): is there a better way to check types?
	if watchEvent.Key.Type != "CheckConfig" {
		logger.Error("check watcher received wrong type")
		return
	}

	check := &corev2.CheckConfig{}
	if err := watchEvent.Value.UnwrapInto(check); err != nil {
		logger.WithError(err).Error("could not unwrap watch event value")
		return
	}

	key := concatUniqueKey(check.Name, check.Namespace)

	c.mu.Lock()
	defer c.mu.Unlock()

	switch watchEvent.Type {
	case storev2.WatchCreate:
		// we need to spin up a new CheckScheduler for the newly created check
		if err := c.startScheduler(check); err != nil {
			logger.WithError(err).Error("unable to start check scheduler")
		}
	case storev2.WatchUpdate:
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
	case storev2.WatchDelete:
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
