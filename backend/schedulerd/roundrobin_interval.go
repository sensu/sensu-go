package schedulerd

import (
	"context"
	"reflect"

	corev2 "github.com/sensu/sensu-go/api/core/v2"
	"github.com/sensu/sensu-go/backend/messaging"
	"github.com/sensu/sensu-go/backend/ringv2"
	"github.com/sensu/sensu-go/backend/store"
	"github.com/sensu/sensu-go/backend/store/cache"
	"github.com/sensu/sensu-go/types"
	"github.com/sirupsen/logrus"
)

type ringCancel struct {
	AgentEntitiesRequest int
	Cancel               context.CancelFunc
}

// RoundRobinIntervalScheduler schedules checks with a determined interval on a
// single entity at a time
type RoundRobinIntervalScheduler struct {
	lastIntervalState      uint32
	lastSubscriptionsState []string
	check                  *types.CheckConfig
	store                  store.Store
	bus                    messaging.MessageBus
	logger                 *logrus.Entry
	ctx                    context.Context
	cancel                 context.CancelFunc
	interrupt              chan *corev2.CheckConfig
	ringPool               *ringv2.Pool
	executor               *CheckExecutor
	cancels                map[string]ringCancel
	entityCache            *cache.Resource
}

// NewRoundRobinIntervalScheduler initializes a RoundRobinIntervalScheduler
func NewRoundRobinIntervalScheduler(ctx context.Context, store store.Store, bus messaging.MessageBus, pool *ringv2.Pool, check *corev2.CheckConfig, cache *cache.Resource) *RoundRobinIntervalScheduler {
	sched := &RoundRobinIntervalScheduler{
		store:             store,
		bus:               bus,
		check:             check,
		lastIntervalState: check.Interval,
		interrupt:         make(chan *corev2.CheckConfig),
		logger: logger.WithFields(logrus.Fields{
			"name":           check.Name,
			"namespace":      check.Namespace,
			"scheduler_type": RoundRobinIntervalType.String(),
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

func (s *RoundRobinIntervalScheduler) updateRings() {
	newCancels := make(map[string]ringCancel)
	agentEntitiesRequest := 1
	var proxyEntities []*corev2.Entity
	if s.check.ProxyRequests != nil {
		entities := s.entityCache.Get(s.check.Namespace)
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
		ring := s.ringPool.Get(key)
		wc := ring.Watch(ctx, s.check.Name, agentEntitiesRequest, int(s.check.Interval), s.check.Cron)
		val := ringCancel{Cancel: cancel, AgentEntitiesRequest: agentEntitiesRequest}
		go s.handleEvents(s.executor, wc, proxyEntities)
		newCancels[key] = val
	}
	s.cancels = newCancels
}

func (s *RoundRobinIntervalScheduler) Start() {
	go s.start()
}

func (s *RoundRobinIntervalScheduler) handleEvents(executor *CheckExecutor, ch <-chan ringv2.Event, proxyEntities []*corev2.Entity) {
	for event := range ch {
		s.handleEvent(executor, event, proxyEntities)
	}
}

func getAgentEntity(event ringv2.Event) string {
	var entity string
	if len(event.Values) > 0 {
		entity = event.Values[0]
	}
	return entity
}

func (s *RoundRobinIntervalScheduler) handleEvent(executor *CheckExecutor, event ringv2.Event, proxyEntities []*corev2.Entity) {
	switch event.Type {
	case ringv2.EventError:
		s.logger.WithError(event.Err).Error("error scheduling check")

	case ringv2.EventAdd:
		s.logger.WithFields(logrus.Fields{
			"agent_entity": getAgentEntity(event),
		}).Info("adding entity to round-robin")

	case ringv2.EventRemove:
		s.logger.WithFields(logrus.Fields{
			"agent_entity": getAgentEntity(event),
		}).Info("removing entity from round-robin")

	case ringv2.EventTrigger:
		// The ring has produced a trigger for the entity, and a check should
		// be executed.
		s.logger.WithFields(logrus.Fields{"agents": event.Values}).Info("executing round robin check on agents")
		s.schedule(executor, proxyEntities, event.Values)

	case ringv2.EventClosing:
		s.logger.Info("shutting down scheduler")
	}
}

func (s *RoundRobinIntervalScheduler) start() {
	s.logger.Info("starting new round-robin interval scheduler")
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
				s.updateRings()
			}
		case <-entityWatcher:
			if s.check.ProxyRequests != nil {
				// The set of proxy entities to consider may have changed
				s.updateRings()
			}
		}
	}
}

func (s *RoundRobinIntervalScheduler) schedule(executor *CheckExecutor, proxyEntities []*corev2.Entity, agentEntities []string) {
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
func (s *RoundRobinIntervalScheduler) toggleSchedule() (stateChanged bool) {
	defer s.setLastState()

	if s.lastIntervalState != s.check.Interval {
		s.logger.Debug("schedule has changed")
		return true
	}
	if !reflect.DeepEqual(s.lastSubscriptionsState, s.check.Subscriptions) {
		s.logger.Debug("subscriptions have changed")
		return true
	}
	s.logger.Debug("check schedule has not changed")
	return false
}

// Update the IntervalScheduler with the last schedule states
func (s *RoundRobinIntervalScheduler) setLastState() {
	s.lastIntervalState = s.check.Interval
	s.lastSubscriptionsState = s.check.Subscriptions
}

// Interrupt refreshes the scheduler with a revised check config.
func (s *RoundRobinIntervalScheduler) Interrupt(check *corev2.CheckConfig) {
	s.interrupt <- check
}

// Stop stops the scheduler
func (s *RoundRobinIntervalScheduler) Stop() error {
	s.logger.Info("stopping scheduler")
	s.cancel()
	return nil
}

func (s *RoundRobinIntervalScheduler) Type() SchedulerType {
	return RoundRobinIntervalType
}
