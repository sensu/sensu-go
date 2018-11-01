package schedulerd

import (
	"context"

	"github.com/sensu/sensu-go/backend/messaging"
	"github.com/sensu/sensu-go/backend/store"
	"github.com/sensu/sensu-go/types"
	"github.com/sirupsen/logrus"
)

type schedulerMode int

const (
	intervalMode schedulerMode = iota
	cronMode
)

// A CheckScheduler schedules checks to be executed on a timer
type CheckScheduler struct {
	checkName      string
	checkNamespace string
	checkInterval  uint32
	checkCron      string
	lastCronState  string
	store          store.Store
	bus            messaging.MessageBus
	logger         *logrus.Entry
	ctx            context.Context
	cancel         context.CancelFunc
	interrupt      chan struct{}
}

func NewCheckScheduler(store store.Store, bus messaging.MessageBus, check *types.CheckConfig, ctx context.Context) *CheckScheduler {
	sched := &CheckScheduler{
		store:          store,
		bus:            bus,
		checkName:      check.Name,
		checkNamespace: check.Namespace,
		checkInterval:  check.Interval,
		checkCron:      check.Cron,
		lastCronState:  check.Cron,
		interrupt:      make(chan struct{}),
		logger: logger.WithFields(logrus.Fields{
			"name":      check.Name,
			"namespace": check.Namespace,
		}),
	}
	sched.ctx, sched.cancel = context.WithCancel(ctx)
	sched.ctx = types.SetContextFromResource(sched.ctx, check)
	return sched
}

func (s *CheckScheduler) toggleCron(check *types.CheckConfig) (stateChanged bool) {
	lastCron := s.lastCronState
	cron := check.Cron
	if (lastCron != "" && cron == "") || (lastCron == "" && cron != "") {
		// Update the CheckScheduler with current check state and last cron state
		s.lastCronState = cron
		s.checkCron = cron
		s.checkInterval = check.Interval
		return true
	}
	return false
}

func (s *CheckScheduler) schedule(timer CheckTimer, executor *CheckExecutor) (restart bool) {
	check, err := s.store.GetCheckConfigByName(s.ctx, s.checkName)
	if err != nil {
		s.logger.WithError(err).Error("unable to retrieve check in check scheduler")
		return false
	}

	// The check has been deleted
	if check == nil {
		s.logger.Info("check is no longer in state")
		return false
	}

	// Indicates a state change from cron to interval or interval to cron
	if s.toggleCron(check) {
		s.logger.Info("check schedule type has changed")
		return true
	}

	// Update the CheckScheduler with the last cron state
	s.lastCronState = check.Cron

	if check.IsSubdued() {
		s.logger.Debug("check is subdued")
		// Reset the timer so the check is scheduled again for the next
		// interval, since it might no longer be subdued
		timer.SetDuration(check.Cron, uint(check.Interval))
		timer.Next()
		return false
	}

	s.logger.Debug("check is not subdued")

	// Reset timer
	timer.SetDuration(check.Cron, uint(check.Interval))
	timer.Next()

	if err := executor.processCheck(s.ctx, check); err != nil {
		logger.Error(err)
	}

	return false
}

// Start starts the CheckScheduler. It always returns nil error.
func (s *CheckScheduler) Start() error {
	go s.start()
	return nil
}

func (s *CheckScheduler) mode() schedulerMode {
	// cron scheduling mode
	if s.checkCron != "" {
		return cronMode
	}
	return intervalMode
}

func (s *CheckScheduler) start() {
	var timer CheckTimer

	switch s.mode() {
	case cronMode:
		s.logger.Info("starting new cron scheduler")
		timer = NewCronTimer(s.checkName, s.checkCron)
	default:
		s.logger.Info("starting new interval scheduler")
		timer = NewIntervalTimer(s.checkName, uint(s.checkInterval))
	}

	executor := NewCheckExecutor(
		s.bus, newRoundRobinScheduler(s.ctx, s.bus), s.checkNamespace, s.store)

	timer.Start()

	for {
		select {
		case <-s.ctx.Done():
			timer.Stop()
			return
		case <-s.interrupt:
		case <-timer.C():
		}
		restart := s.schedule(timer, executor)
		if restart {
			timer.Stop()
			defer s.Start()
			return
		}
	}
}

// Interrupt causes the scheduler to immediately fire and ignore the timer.
func (s *CheckScheduler) Interrupt() {
	s.interrupt <- struct{}{}
}

// Stop stops the CheckScheduler
func (s *CheckScheduler) Stop() error {
	logger.Info("stopping scheduler")
	s.cancel()

	return nil
}
