package schedulerd

import (
	"context"

	"github.com/sensu/sensu-go/backend/messaging"
	"github.com/sensu/sensu-go/backend/store"
	"github.com/sensu/sensu-go/backend/store/cache"
	"github.com/sensu/sensu-go/backend/secrets"
	"github.com/sirupsen/logrus"

	corev2 "github.com/sensu/sensu-go/api/core/v2"
)

// CronScheduler schedules checks to be executed on a cron schedule.
type CronScheduler struct {
	lastCronState          string
	check                  *corev2.CheckConfig
	store                  store.Store
	bus                    messaging.MessageBus
	logger                 *logrus.Entry
	ctx                    context.Context
	cancel                 context.CancelFunc
	interrupt              chan *corev2.CheckConfig
	entityCache            *cache.Resource
	secretsProviderManager *secrets.ProviderManager
}

// NewCronScheduler initializes a CronScheduler
func NewCronScheduler(ctx context.Context, store store.Store, bus messaging.MessageBus, check *corev2.CheckConfig, cache *cache.Resource, secretsProviderManager *secrets.ProviderManager) *CronScheduler {
	sched := &CronScheduler{
		store:         store,
		bus:           bus,
		check:         check,
		lastCronState: check.Cron,
		interrupt:     make(chan *corev2.CheckConfig),
		logger: logger.WithFields(logrus.Fields{
			"name":           check.Name,
			"namespace":      check.Namespace,
			"scheduler_type": CronType.String(),
		}),
		entityCache:            cache,
		secretsProviderManager: secretsProviderManager,
	}
	sched.ctx, sched.cancel = context.WithCancel(ctx)
	sched.ctx = corev2.SetContextFromResource(sched.ctx, check)
	return sched
}

func (s *CronScheduler) schedule(timer *CronTimer, executor *CheckExecutor) {
	defer s.resetTimer(timer)

	if s.check.IsSubdued() {
		s.logger.Debug("check is subdued")
		return
	}

	s.logger.Debug("check is not subdued")

	if err := executor.processCheck(s.ctx, s.check); err != nil {
		logger.Error(err)
	}
}

// Start starts the cron scheduler.
func (s *CronScheduler) Start() {
	cronCounter.WithLabelValues(s.check.Namespace).Inc()
	go s.start()
}

func (s *CronScheduler) start() {
	s.logger.Info("starting new cron scheduler")
	timer := NewCronTimer(s.check.Name, s.check.Cron)
	executor := NewCheckExecutor(s.bus, s.check.Namespace, s.store, s.entityCache, s.secretsProviderManager)
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
func (s *CronScheduler) Interrupt(check *corev2.CheckConfig) {
	s.interrupt <- check
}

// Stop stops the cron scheduler.
func (s *CronScheduler) Stop() error {
	cronCounter.WithLabelValues(s.check.Namespace).Dec()
	logger.Info("stopping cron scheduler")
	s.cancel()

	return nil
}

func (s *CronScheduler) toggleSchedule() (stateChanged bool) {
	defer s.setLastState()

	if s.lastCronState != s.check.Cron {
		s.logger.Info("cron schedule has changed")
		return true
	}

	s.logger.Debug("schedule unchanged")
	return false
}

func (s *CronScheduler) setLastState() {
	s.lastCronState = s.check.Cron
}

func (s *CronScheduler) resetTimer(timer *CronTimer) {
	timer.SetDuration(s.check.Cron, 0)
	timer.Next()
}

// Type returns the type of the cron scheduler.
func (s *CronScheduler) Type() SchedulerType {
	return CronType
}
