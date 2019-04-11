package schedulerd

import (
	"context"

	corev2 "github.com/sensu/sensu-go/api/core/v2"
	"github.com/sensu/sensu-go/backend/messaging"
	"github.com/sensu/sensu-go/backend/ringv2"
	"github.com/sensu/sensu-go/backend/store"
	"github.com/sensu/sensu-go/types"
	"github.com/sirupsen/logrus"
)

// RoundRobinCronScheduler is like CronScheduler, but only schedules checks
// on a single entity at a time.
type RoundRobinCronScheduler struct {
	lastCronState string
	check         *types.CheckConfig
	store         store.Store
	bus           messaging.MessageBus
	logger        *logrus.Entry
	ctx           context.Context
	cancel        context.CancelFunc
	interrupt     chan *corev2.CheckConfig
	ringPool      *ringv2.Pool
	cancels       map[string]ringCancel
	executor      *CheckExecutor
	entityCache   *EntityCache
}

// NewRoundRobinCronScheduler creates a new RoundRobinCronScheduler.
func NewRoundRobinCronScheduler(ctx context.Context, store store.Store, bus messaging.MessageBus, pool *ringv2.Pool, check *corev2.CheckConfig, cache *EntityCache) *RoundRobinCronScheduler {
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
		executor:    NewCheckExecutor(bus, check.Namespace, store, cache),
		entityCache: cache,
	}
	sched.ctx, sched.cancel = context.WithCancel(ctx)
	sched.ctx = corev2.SetContextFromResource(sched.ctx, check)
	return sched
}

// Start starts the scheduler.
func (s *RoundRobinCronScheduler) Start() {
	go s.start()
}

func (s *RoundRobinCronScheduler) handleEvent(executor *CheckExecutor, event ringv2.Event, proxyEntities []*corev2.Entity) {
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
		s.schedule(executor, proxyEntities, event.Values)

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
				s.logger.Info("cron schedule updated")
				s.updateRings()
			}
		case <-entityWatcher:
			if s.check.ProxyRequests != nil {
				// The set of proxy entities to consider may have changed
				s.logger.Info("proxy entities updated")
				s.updateRings()
			}
		}
	}
}

func (s *RoundRobinCronScheduler) handleEvents(executor *CheckExecutor, ch <-chan ringv2.Event, proxyEntities []*corev2.Entity) {
	for event := range ch {
		s.handleEvent(executor, event, proxyEntities)
	}
}

func (s *RoundRobinCronScheduler) updateRings() {
	newCancels := make(map[string]ringCancel)
	agentEntitiesRequest := 1
	var proxyEntities []*corev2.Entity
	if s.check.ProxyRequests != nil {
		entities := s.entityCache.GetEntities(s.check.Namespace)
		proxyEntities = matchEntities(entities, s.check.ProxyRequests)
		agentEntitiesRequest = len(proxyEntities)
		if agentEntitiesRequest == 0 {
			s.logger.Error("check not published, no matching entities for proxy request")
			return
		}
	}
	// Cancel any ring watchers that should no longer exist
	for _, watcher := range s.cancels {
		watcher.Cancel()
	}
	for _, sub := range s.check.Subscriptions {
		key := ringv2.Path(s.check.Namespace, sub)

		// Create a new watcher
		ctx, cancel := context.WithCancel(s.ctx)
		wc := s.ringPool.Get(key).Watch(ctx, s.check.Name, agentEntitiesRequest, int(s.check.Interval), s.check.Cron)
		val := ringCancel{Cancel: cancel, AgentEntitiesRequest: agentEntitiesRequest}
		go s.handleEvents(s.executor, wc, proxyEntities)
		newCancels[key] = val
	}
	s.cancels = newCancels
}

func (s *RoundRobinCronScheduler) schedule(executor *CheckExecutor, proxyEntities []*corev2.Entity, agentEntities []string) {
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
		s.logger.Info("cron schedule has changed")
		return true
	}
	s.logger.Info("cron schedule has not changed")
	return false
}

// Update the CronScheduler with the last schedule states
func (s *RoundRobinCronScheduler) setLastState() {
	s.lastCronState = s.check.Cron
}

// Interrupt refreshes the scheduler with a revised check config.
func (s *RoundRobinCronScheduler) Interrupt(check *corev2.CheckConfig) {
	s.interrupt <- check
}

// Stop stops the scheduler
func (s *RoundRobinCronScheduler) Stop() error {
	s.logger.Info("stopping scheduler")
	s.cancel()
	return nil
}

// Type returns RoundRobinCronType
func (s *RoundRobinCronScheduler) Type() SchedulerType {
	return RoundRobinCronType
}
