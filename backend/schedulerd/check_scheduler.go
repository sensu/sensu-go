package schedulerd

import (
	"context"
	"sync"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/sensu/sensu-go/backend/messaging"
	"github.com/sensu/sensu-go/types"
	sensutime "github.com/sensu/sensu-go/util/time"
)

// A CheckScheduler schedules checks to be executed on a timer
type CheckScheduler struct {
	CheckName     string
	CheckEnv      string
	CheckOrg      string
	CheckInterval uint32
	CheckCron     string
	LastCronState string

	StateManager *StateManager
	MessageBus   messaging.MessageBus
	WaitGroup    *sync.WaitGroup

	logger *logrus.Entry

	ringGetter types.RingGetter
	ctx        context.Context
	cancel     context.CancelFunc
}

// Start starts the CheckScheduler. It always returns nil error.
func (s *CheckScheduler) Start() error {
	s.ctx, s.cancel = context.WithCancel(context.Background())
	s.WaitGroup.Add(1)
	defer s.WaitGroup.Done()

	s.logger = logger.WithFields(logrus.Fields{"name": s.CheckName, "org": s.CheckOrg, "env": s.CheckEnv})

	go func() {
	toggle:
		var timer CheckTimer
		if s.CheckCron != "" {
			s.logger.Infof("starting new cron scheduler")
			timer = NewCronTimer(s.CheckName, s.CheckCron)
		}
		if timer == nil || s.CheckCron == "" {
			s.logger.Infof("starting new interval scheduler")
			timer = NewIntervalTimer(s.CheckName, uint(s.CheckInterval))
		}

		executor := NewCheckExecutor(s.MessageBus, newRoundRobinScheduler(s.ctx, s.MessageBus, s.ringGetter), s.CheckOrg, s.CheckEnv)

		// TODO(greg): Refactor this part to make the code more easily tested.
		timer.Start()
		defer timer.Stop()

		for {
			select {
			case <-s.ctx.Done():
				return
			case <-timer.C():
				// Fetch check from scheduler's state
				state := s.StateManager.State()
				check := state.GetCheck(s.CheckName, s.CheckOrg, s.CheckEnv)

				// The check has been deleted
				if check == nil {
					s.logger.Info("check is no longer in state")
					return
				}

				// Indicates a state change from cron to interval or interval to cron
				if (s.LastCronState != "" && check.Cron == "") ||
					(s.LastCronState == "" && check.Cron != "") {
					s.logger.Info("check schedule type has changed")
					// Update the CheckScheduler with current check state and last cron state
					s.LastCronState = check.Cron
					s.CheckCron = check.Cron
					s.CheckInterval = check.Interval
					timer.Stop()
					goto toggle
				}

				// Update the CheckScheduler with the last cron state
				s.LastCronState = check.Cron

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

				// Point executor to lastest copy of the scheduler state
				executor.setState(state)

				if err := executor.processCheck(s.ctx, check); err != nil {
					logger.Error(err)
				}
			}
		}
	}()

	return nil
}

// Stop stops the CheckScheduler
func (s *CheckScheduler) Stop() error {
	s.logger.Infof("stopping scheduler")
	s.cancel()

	return nil
}
