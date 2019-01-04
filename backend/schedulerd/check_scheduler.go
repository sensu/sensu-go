package schedulerd

import (
	"context"
	"fmt"

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
	lastIntervalState uint32
	lastCronState     string
	check             *types.CheckConfig
	store             store.Store
	bus               messaging.MessageBus
	logger            *logrus.Entry
	ctx               context.Context
	cancel            context.CancelFunc
	interrupt         chan struct{}
}

func NewCheckScheduler(store store.Store, bus messaging.MessageBus, check *types.CheckConfig, ctx context.Context) *CheckScheduler {
	sched := &CheckScheduler{
		store:             store,
		bus:               bus,
		check:             check,
		lastIntervalState: check.Interval,
		lastCronState:     check.Cron,
		interrupt:         make(chan struct{}),
		logger: logger.WithFields(logrus.Fields{
			"name":      check.Name,
			"namespace": check.Namespace,
		}),
	}
	sched.ctx, sched.cancel = context.WithCancel(ctx)
	sched.ctx = types.SetContextFromResource(sched.ctx, check)
	return sched
}

func (s *CheckScheduler) schedule(timer CheckTimer, executor *CheckExecutor) {
	if err := s.refreshCheckFromStore(); err != nil {
		s.logger.WithError(err).Error("unable to retrieve check in check scheduler")
		return
	}

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

// Start starts the CheckScheduler. It always returns nil error.
func (s *CheckScheduler) Start() error {
	go s.start()
	return nil
}

func (s *CheckScheduler) mode() schedulerMode {
	// cron scheduling mode
	if s.check.Cron != "" {
		return cronMode
	}
	return intervalMode
}

func (s *CheckScheduler) start() {
	var timer CheckTimer

	switch s.mode() {
	case cronMode:
		s.logger.Info("starting new cron scheduler")
		timer = NewCronTimer(s.check.Name, s.check.Cron)
	default:
		s.logger.Info("starting new interval scheduler")
		timer = NewIntervalTimer(s.check.Name, uint(s.check.Interval))
	}

	executor := NewCheckExecutor(
		s.bus, newRoundRobinScheduler(s.ctx, s.bus), s.check.Namespace, s.store)

	timer.Start()

	for {
		select {
		case <-s.ctx.Done():
			timer.Stop()
			return
		case <-s.interrupt:
			if err := s.refreshCheckFromStore(); err != nil {
				s.logger.WithError(err).Error("unable to retrieve check in check scheduler")
				return
			}
			// if a schedule change is detected, restart the timer
			if s.toggleSchedule() {
				timer.Stop()
				defer func() {
					if err := s.Start(); err != nil {
						s.logger.Error(err)
					}
				}()
			}
			continue
		case <-timer.C():
		}
		s.schedule(timer, executor)
	}
}

// Interrupt causes the scheduler to immediately fetch the check to determine if a timer reset is required.
// It is called when the check watcher detects an update to the check.
func (s *CheckScheduler) Interrupt() {
	s.interrupt <- struct{}{}
}

// Stop stops the CheckScheduler
func (s *CheckScheduler) Stop() error {
	logger.Info("stopping scheduler")
	s.cancel()

	return nil
}

func (s *CheckScheduler) refreshCheckFromStore() error {
	check, err := s.store.GetCheckConfigByName(s.ctx, s.check.Name)
	if err != nil {
		return err
	}
	if check == nil {
		return fmt.Errorf("check %s is no longer in state", s.check.Name)
	}
	s.check = check
	return nil
}

// Indicates a state change in the schedule, and if a timer needs to be reset.
func (s *CheckScheduler) toggleSchedule() (stateChanged bool) {
	defer s.setLastState()

	if (s.lastIntervalState != s.check.Interval) || (s.lastCronState != s.check.Cron) {
		s.logger.Info("check schedule has changed")
		return true
	}
	s.logger.Info("check schedule has not changed")
	return false
}

// Update the CheckScheduler with the last schedule states
func (s *CheckScheduler) setLastState() {
	s.lastCronState = s.check.Cron
	s.lastIntervalState = s.check.Interval
}

// Reset timer
func (s *CheckScheduler) resetTimer(timer CheckTimer) {
	timer.SetDuration(s.check.Cron, uint(s.check.Interval))
	timer.Next()
}
