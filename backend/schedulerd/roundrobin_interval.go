package schedulerd

import (
	"context"
	"reflect"
	"sync"

	corev2 "github.com/sensu/core/v2"
	corev3 "github.com/sensu/core/v3"
	"github.com/sensu/sensu-go/backend/messaging"
	"github.com/sensu/sensu-go/backend/ringv2"
	"github.com/sensu/sensu-go/backend/secrets"
	storev2 "github.com/sensu/sensu-go/backend/store/v2"
	"github.com/sirupsen/logrus"
)

type ringCancel struct {
	AgentEntitiesRequest int
	cancel               context.CancelFunc
	wg                   *sync.WaitGroup
}

func (r ringCancel) Cancel() {
	r.cancel()
	r.wg.Wait()
}

// RoundRobinIntervalScheduler schedules checks with a determined interval on a
// single entity at a time
type RoundRobinIntervalScheduler struct {
	lastIntervalState      uint32
	lastSubscriptionsState []string
	lastScheduler          string
	check                  *corev2.CheckConfig
	store                  storev2.Interface
	bus                    messaging.MessageBus
	logger                 *logrus.Entry
	ctx                    context.Context
	cancel                 context.CancelFunc
	interrupt              chan *corev2.CheckConfig
	ringPool               *ringv2.RingPool
	executor               *CheckExecutor
	cancels                map[string]ringCancel
	entityCache            EntityCache
	mu                     sync.Mutex
	proxyEntities          []*corev3.EntityConfig
	stopWg                 sync.WaitGroup
}

// NewRoundRobinIntervalScheduler initializes a RoundRobinIntervalScheduler
func NewRoundRobinIntervalScheduler(ctx context.Context, store storev2.Interface, bus messaging.MessageBus, pool *ringv2.RingPool, check *corev2.CheckConfig, cache EntityCache, secretsProviderManager *secrets.ProviderManager) *RoundRobinIntervalScheduler {
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
			"interval":       check.Interval,
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

// this function technically can leak its cancel, but the design makes it
// difficult to fix, and there are no known issues with it.
//nolint:govet
func (s *RoundRobinIntervalScheduler) updateRings() {
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
	// clean up old watchers synchronously
	for key, watcher := range s.cancels {
		s.logger.WithField("ring", key).Debug("cancelling old ring watcher")
		watcher.Cancel()
	}
	newCancels := make(map[string]ringCancel)
	for _, sub := range s.check.Subscriptions {
		key := ringv2.Path(s.check.Namespace, sub)

		s.logger.WithField("ring", key).Debug("creating new ring watcher")
		// Create a new watcher
		ctx, cancel := context.WithCancel(s.ctx)
		ring := s.ringPool.Get(key)
		sub := ringv2.Subscription{
			Name:             s.check.Name,
			Items:            agentEntitiesRequest,
			IntervalSchedule: int(s.check.Interval),
			CronSchedule:     s.check.Cron,
		}
		if err := sub.Validate(); err != nil {
			s.logger.WithError(err).Error("error scheduling round-robin check")
			continue
		}
		wc := ring.Subscribe(ctx, sub)
		wg := new(sync.WaitGroup)
		wg.Add(1)
		val := ringCancel{cancel: cancel, AgentEntitiesRequest: agentEntitiesRequest, wg: wg}
		go s.handleEvents(s.executor, wc, wg)
		newCancels[key] = val
	}
	s.cancels = newCancels
}

// Start starts the round robin interval scheduler.
func (s *RoundRobinIntervalScheduler) Start() {
	rrIntervalCounter.WithLabelValues(s.check.Namespace).Inc()
	s.stopWg.Add(1)
	go s.start()
}

func (s *RoundRobinIntervalScheduler) handleEvents(executor *CheckExecutor, ch <-chan ringv2.Event, wg *sync.WaitGroup) {
	defer wg.Done()
	for event := range ch {
		s.handleEvent(executor, event)
	}
}

func getAgentEntity(event ringv2.Event) string {
	var entity string
	if len(event.Values) > 0 {
		entity = event.Values[0]
	}
	return entity
}

func (s *RoundRobinIntervalScheduler) handleEvent(executor *CheckExecutor, event ringv2.Event) {
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
		if len(event.Values) == 0 {
			s.logger.Error("round robin check scheduled, but no available entities")
			return
		}
		// The ring has produced a trigger for the entity, and a check should
		// be executed.
		s.logger.WithFields(logrus.Fields{"agents": event.Values}).Info("executing round robin check on agents")
		s.mu.Lock()
		s.schedule(executor, s.proxyEntities, event.Values)
		s.mu.Unlock()

	case ringv2.EventClosing:
		s.logger.Warn("shutting down scheduler")
	}
}

func (s *RoundRobinIntervalScheduler) start() {
	defer s.stopWg.Done()
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

func (s *RoundRobinIntervalScheduler) schedule(executor *CheckExecutor, proxyEntities []*corev3.EntityConfig, agentEntities []string) {
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
	if s.lastScheduler != s.check.Scheduler {
		s.logger.WithField("previous_scheduler", s.lastScheduler).WithField("new scheduler", s.check.Scheduler).Info("scheduler backend has changed")
		return true
	}
	s.logger.Debug("interval schedule has not changed")
	return false
}

// Update the IntervalScheduler with the last schedule states
func (s *RoundRobinIntervalScheduler) setLastState() {
	s.lastIntervalState = s.check.Interval
	s.lastSubscriptionsState = s.check.Subscriptions
	s.lastScheduler = s.check.Scheduler
}

// Interrupt refreshes the scheduler with a revised check config.
func (s *RoundRobinIntervalScheduler) Interrupt(check *corev2.CheckConfig) {
	s.interrupt <- check
}

// Stop stops the scheduler
func (s *RoundRobinIntervalScheduler) Stop() error {
	s.logger.Info("stopping scheduler")
	s.cancel()
	s.stopWg.Wait()
	rrIntervalCounter.WithLabelValues(s.check.Namespace).Dec()
	return nil
}

// Type returns the type of the round robin interval scheduler.
func (s *RoundRobinIntervalScheduler) Type() SchedulerType {
	return RoundRobinIntervalType
}
