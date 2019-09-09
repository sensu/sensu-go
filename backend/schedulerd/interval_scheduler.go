package schedulerd

import (
	"context"

	corev2 "github.com/sensu/sensu-go/api/core/v2"
	"github.com/sensu/sensu-go/backend/messaging"
	"github.com/sensu/sensu-go/backend/store"
	"github.com/sensu/sensu-go/backend/store/cache"
	"github.com/sensu/sensu-go/types"
	"github.com/sirupsen/logrus"
)

// IntervalScheduler schedules checks to be executed on a timer
type IntervalScheduler struct {
	lastIntervalState uint32
	check             *types.CheckConfig
	store             store.Store
	bus               messaging.MessageBus
	logger            *logrus.Entry
	ctx               context.Context
	cancel            context.CancelFunc
	interrupt         chan *corev2.CheckConfig
	entityCache       *cache.Resource
}

// NewIntervalScheduler initializes an IntervalScheduler
func NewIntervalScheduler(ctx context.Context, store store.Store, bus messaging.MessageBus, check *types.CheckConfig, cache *cache.Resource) *IntervalScheduler {
	sched := &IntervalScheduler{
		store:             store,
		bus:               bus,
		check:             check,
		lastIntervalState: check.Interval,
		interrupt:         make(chan *corev2.CheckConfig),
		logger: logger.WithFields(logrus.Fields{
			"name":           check.Name,
			"namespace":      check.Namespace,
			"scheduler_type": IntervalType.String(),
		}),
		entityCache: cache,
	}
	sched.ctx, sched.cancel = context.WithCancel(ctx)
	sched.ctx = types.SetContextFromResource(sched.ctx, check)
	return sched
}

func (s *IntervalScheduler) schedule(timer CheckTimer, executor *CheckExecutor) {
	s.resetTimer(timer)

	if s.check.IsSubdued() {
		s.logger.Debug("check is subdued")
		return
	}

	s.logger.Debug("check is not subdued")

	if err := executor.processCheck(s.ctx, s.check); err != nil {
		logger.WithError(err).Error("error executing check")
	}
}

// Start starts the IntervalScheduler.
func (s *IntervalScheduler) Start() {
	go s.start()
}

func (s *IntervalScheduler) start() {
	s.logger.Info("starting new interval scheduler")
	timer := NewIntervalTimer(s.check.Name, uint(s.check.Interval))
	executor := NewCheckExecutor(s.bus, s.check.Namespace, s.store, s.entityCache)

	timer.Start()

	for {
		select {
		case <-s.ctx.Done():
			timer.Stop()
			return
		case check := <-s.interrupt:
			// if a schedule change is detected, restart the timer
			s.check = check
			if s.toggleSchedule() {
				timer.Stop()
				defer s.Start()
				return
			}
			continue
		case <-timer.C():
		}
		s.schedule(timer, executor)
	}
}

// Interrupt refreshes the scheduler with a revised check config.
func (s *IntervalScheduler) Interrupt(check *corev2.CheckConfig) {
	s.interrupt <- check
}

// Stop stops the IntervalScheduler
func (s *IntervalScheduler) Stop() error {
	s.logger.Info("stopping scheduler")
	s.cancel()

	return nil
}

// Indicates a state change in the schedule, and if a timer needs to be reset.
func (s *IntervalScheduler) toggleSchedule() (stateChanged bool) {
	defer s.setLastState()

	if s.lastIntervalState != s.check.Interval {
		s.logger.Info("interval schedule has changed")
		return true
	}
	s.logger.Info("check schedule has not changed")
	return false
}

// Update the IntervalScheduler with the last schedule states
func (s *IntervalScheduler) setLastState() {
	s.lastIntervalState = s.check.Interval
}

// Reset timer
func (s *IntervalScheduler) resetTimer(timer CheckTimer) {
	timer.SetDuration("", uint(s.check.Interval))
	timer.Next()
}

// Type returns the type of the interval scheduler.
func (s *IntervalScheduler) Type() SchedulerType {
	return IntervalType
}
