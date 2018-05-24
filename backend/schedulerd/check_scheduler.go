package schedulerd

import (
	"context"
	"sync"
	"time"

	"github.com/sensu/sensu-go/backend/messaging"
	"github.com/sensu/sensu-go/backend/store"
	"github.com/sensu/sensu-go/types"
	sensutime "github.com/sensu/sensu-go/util/time"
	"github.com/sirupsen/logrus"
)

// A CheckScheduler schedules checks to be executed on a timer
type CheckScheduler struct {
	checkName     string
	checkEnv      string
	checkOrg      string
	checkInterval uint32
	checkCron     string
	lastCronState string
	store         store.Store
	bus           messaging.MessageBus
	wg            *sync.WaitGroup
	logger        *logrus.Entry
	ctx           context.Context
	cancel        context.CancelFunc
	interrupt     chan struct{}
}

// Start starts the CheckScheduler. It always returns nil error.
func (s *CheckScheduler) Start() error {
	check := &types.CheckConfig{
		Organization: s.checkOrg,
		Environment:  s.checkEnv,
		Name:         s.checkName,
	}
	ctx := types.SetContextFromResource(context.Background(), check)
	s.ctx, s.cancel = context.WithCancel(ctx)

	s.wg.Add(1)
	defer s.wg.Done()

	s.logger = logger.WithFields(logrus.Fields{"name": s.checkName, "org": s.checkOrg, "env": s.checkEnv})

	go func() {
	toggle:
		var timer CheckTimer
		if s.checkCron != "" {
			s.logger.Info("starting new cron scheduler")
			timer = NewCronTimer(s.checkName, s.checkCron)
		}
		if timer == nil || s.checkCron == "" {
			s.logger.Info("starting new interval scheduler")
			timer = NewIntervalTimer(s.checkName, uint(s.checkInterval))
		}

		executor := NewCheckExecutor(s.bus, newRoundRobinScheduler(s.ctx, s.bus), s.checkOrg, s.checkEnv, s.store)

		// TODO(greg): Refactor this part to make the code more easily tested.
		timer.Start()
		defer timer.Stop()

		for {
			select {
			case <-s.ctx.Done():
				return
			case <-s.interrupt:
			case <-timer.C():
			}

			check, err := s.store.GetCheckConfigByName(s.ctx, s.checkName)
			if err != nil {
				s.logger.WithError(err).Error("unable to retrieve check in check scheduler")
				continue
			}

			// The check has been deleted
			if check == nil {
				s.logger.Info("check is no longer in state")
				return
			}

			// Indicates a state change from cron to interval or interval to cron
			if (s.lastCronState != "" && check.Cron == "") ||
				(s.lastCronState == "" && check.Cron != "") {
				s.logger.Info("check schedule type has changed")
				// Update the CheckScheduler with current check state and last cron state
				s.lastCronState = check.Cron
				s.checkCron = check.Cron
				s.checkInterval = check.Interval
				timer.Stop()
				goto toggle
			}

			// Update the CheckScheduler with the last cron state
			s.lastCronState = check.Cron

			if subdue := check.GetSubdue(); subdue != nil {
				isSubdued, err := sensutime.InWindows(time.Now(), *subdue)
				if err == nil && isSubdued {
					// Check is subdued at this time
					s.logger.Debug("check is not scheduled to be executed")
				}
				if err != nil {
					s.logger.WithError(err).Print("unexpected error with time windows")
				}

				if err != nil || isSubdued {
					// Reset the timer so the check is scheduled again for the next
					// interval, since it might no longer be subdued
					timer.SetDuration(check.Cron, uint(check.Interval))
					timer.Next()
					continue
				}
			}

			// Reset timer
			timer.SetDuration(check.Cron, uint(check.Interval))
			timer.Next()

			if err := executor.processCheck(s.ctx, check); err != nil {
				logger.Error(err)
			}
		}
	}()

	return nil
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
