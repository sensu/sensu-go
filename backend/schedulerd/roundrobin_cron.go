package schedulerd

import (
	"context"
	"sync"

	corev2 "github.com/sensu/sensu-go/api/core/v2"
	corev3 "github.com/sensu/sensu-go/api/core/v3"
	"github.com/sensu/sensu-go/backend/messaging"
	"github.com/sensu/sensu-go/backend/ringv2"
	"github.com/sensu/sensu-go/backend/secrets"
	"github.com/sensu/sensu-go/backend/store"
	cachev2 "github.com/sensu/sensu-go/backend/store/cache/v2"
	"github.com/sirupsen/logrus"
)

// RoundRobinCronScheduler is like CronScheduler, but only schedules checks
// on a single entity at a time.
type RoundRobinCronScheduler struct {
	lastCronState string
	lastScheduler string
	check         *corev2.CheckConfig
	store         store.Store
	bus           messaging.MessageBus
	logger        *logrus.Entry
	ctx           context.Context
	cancel        context.CancelFunc
	interrupt     chan *corev2.CheckConfig
	ringPool      *ringv2.RingPool
	cancels       map[string]ringCancel
	executor      *CheckExecutor
	entityCache   *cachev2.Resource
	mu            sync.Mutex
	proxyEntities []*corev3.EntityConfig
}

// NewRoundRobinCronScheduler creates a new RoundRobinCronScheduler.
func NewRoundRobinCronScheduler(ctx context.Context, store store.Store, bus messaging.MessageBus, pool *ringv2.RingPool, check *corev2.CheckConfig, cache *cachev2.Resource, secretsProviderManager *secrets.ProviderManager) *RoundRobinCronScheduler {
	sched := &RoundRobinCronScheduler{
		store:         store,
		bus:           bus,
		check:         check,
		lastCronState: check.Cron,
		interrupt:     make(chan *corev2.CheckConfig),
		logger: logger.WithFields(logrus.Fields{
			"name":           check.Name,
			"namespace":      check.Namespace,
			"scheduler_type": RoundRobinCronType.String(),
		}),
		ringPool:    pool,
		cancels:     make(map[string]ringCancel),
		executor:    NewCheckExecutor(bus, check.Namespace, store, cache, secretsProviderManager),
		entityCache: cache,
	}
	sched.ctx, sched.cancel = context.WithCancel(ctx)
	sched.ctx = corev2.SetContextFromResource(sched.ctx, check)
	return sched
}

// Start starts the scheduler.
func (s *RoundRobinCronScheduler) Start() {
	rrCronCounter.WithLabelValues(s.check.Namespace).Inc()
	go s.start()
}

func (s *RoundRobinCronScheduler) handleEvent(executor *CheckExecutor, event ringv2.Event) {
	switch event.Type {
	case ringv2.EventError:
		s.logger.WithError(event.Err).Error("error scheduling check")

	case ringv2.EventAdd:
		s.logger.WithFields(
			logrus.Fields{"entity": getAgentEntity(event)}).Info(
			"adding entity to round-robin")

	case ringv2.EventRemove:
		s.logger.WithFields(
			logrus.Fields{"entity": getAgentEntity(event)}).Info(
			"removing entity from round-robin")

	case ringv2.EventTrigger:
		s.logger.Info("scheduling check")
		s.mu.Lock()
		s.schedule(executor, s.proxyEntities, event.Values)
		s.mu.Unlock()

	case ringv2.EventClosing:
		s.logger.Warn("shutting down scheduler")
	}
}

func (s *RoundRobinCronScheduler) start() {
	s.logger.Info("starting new round-robin cron scheduler")
	s.setLastState()
	s.updateRings()

	entityWatcher := s.entityCache.Watch(s.ctx)

	for {
		select {
		case <-s.ctx.Done():
			return
		case check := <-s.interrupt:
			s.check = check
			if s.toggleSchedule() {
				s.logger.Debug("cron schedule updated")
				s.updateRings()
			}
		case <-entityWatcher:
			if s.check.ProxyRequests != nil {
				// The set of proxy entities to consider may have changed
				s.logger.Debug("proxy entities updated")
				s.updateRings()
			}
		}
	}
}

func (s *RoundRobinCronScheduler) handleEvents(executor *CheckExecutor, ch <-chan ringv2.Event) {
	for event := range ch {
		s.handleEvent(executor, event)
	}
}

// this function technically can leak its cancel, but the design makes it
// difficult to fix, and there are no known issues with it.
//nolint:govet
func (s *RoundRobinCronScheduler) updateRings() {
	agentEntitiesRequest := 1
	if s.check.ProxyRequests != nil {
		entities := s.entityCache.Get(s.check.Namespace)
		s.mu.Lock()
		s.proxyEntities = matchEntities(entities, s.check.ProxyRequests)
		agentEntitiesRequest = len(s.proxyEntities)
		s.mu.Unlock()
		if agentEntitiesRequest == 0 {
			s.logger.Error("check not published, no matching entities for proxy request")
			return
		}
	}
	newCancels := make(map[string]ringCancel)
	for _, sub := range s.check.Subscriptions {
		key := ringv2.Path(s.check.Namespace, sub)
		watcher, ok := s.cancels[key]
		if ok {
			if watcher.AgentEntitiesRequest == agentEntitiesRequest {
				// don't need to recreate the watcher
				newCancels[key] = watcher
				continue
			}
			watcher.Cancel()
		}

		// Create a new watcher
		ctx, cancel := context.WithCancel(s.ctx)
		sub := ringv2.Subscription{
			Name:             s.check.Name,
			Items:            agentEntitiesRequest,
			IntervalSchedule: int(s.check.Interval),
			CronSchedule:     s.check.Cron,
		}
		if err := sub.Validate(); err != nil {
			logger.WithField("check", s.check.Name).WithError(err).Error("error scheduling round-robin check")
			continue
		}
		wc := s.ringPool.Get(key).Subscribe(ctx, sub)
		val := ringCancel{Cancel: cancel, AgentEntitiesRequest: agentEntitiesRequest}
		go s.handleEvents(s.executor, wc)
		newCancels[key] = val
	}
	// clean up any remaining watchers that are no longer valid
	for key, watcher := range s.cancels {
		if _, ok := newCancels[key]; !ok {
			watcher.Cancel()
		}
	}
	s.cancels = newCancels
}

func (s *RoundRobinCronScheduler) schedule(executor *CheckExecutor, proxyEntities []*corev3.EntityConfig, agentEntities []string) {
	if s.check.IsSubdued() {
		s.logger.Debug("check is subdued")
		return
	}

	s.logger.Debug("check is not subdued")

	if err := processRoundRobinCheck(s.ctx, executor, s.check, proxyEntities, agentEntities); err != nil {
		logger.WithError(err).Error("error executing check")
	}
}

// Indicates a state change in the schedule, and if a timer needs to be reset.
func (s *RoundRobinCronScheduler) toggleSchedule() (stateChanged bool) {
	defer s.setLastState()

	if s.lastCronState != s.check.Cron {
		s.logger.Debug("cron schedule has changed")
		return true
	}
	if s.lastScheduler != s.check.Scheduler {
		s.logger.Debug("cron schedule has changed")
		return true
	}
	s.logger.Debug("cron schedule has not changed")
	return false
}

// Update the CronScheduler with the last schedule states
func (s *RoundRobinCronScheduler) setLastState() {
	s.lastCronState = s.check.Cron
	s.lastScheduler = s.check.Scheduler
}

// Interrupt refreshes the scheduler with a revised check config.
func (s *RoundRobinCronScheduler) Interrupt(check *corev2.CheckConfig) {
	s.interrupt <- check
}

// Stop stops the scheduler
func (s *RoundRobinCronScheduler) Stop() error {
	rrCronCounter.WithLabelValues(s.check.Namespace).Dec()
	s.logger.Info("stopping scheduler")
	s.cancel()
	return nil
}

// Type returns RoundRobinCronType
func (s *RoundRobinCronScheduler) Type() SchedulerType {
	return RoundRobinCronType
}
