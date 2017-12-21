package schedulerd

import (
	"strings"
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

	StateManager *StateManager
	MessageBus   messaging.MessageBus
	WaitGroup    *sync.WaitGroup

	logger   *logrus.Entry
	stopping chan struct{}
}

// Start scheduler, ...
func (s *CheckScheduler) Start() error {
	s.stopping = make(chan struct{})
	s.WaitGroup.Add(1)

	s.logger = logger.WithFields(logrus.Fields{"name": s.CheckName, "org": s.CheckOrg, "env": s.CheckEnv})
	s.logger.Infof("starting new scheduler")

	timer := NewIntervalTimer(s.CheckName, uint(s.CheckInterval))
	executor := &CheckExecutor{Bus: s.MessageBus}

	// TODO(greg): Refactor this part to make the code more easily tested.
	go func() {
		timer.Start()
		defer timer.Stop()
		defer s.WaitGroup.Done()

		for {
			select {
			case <-s.stopping:
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

				if subdue := check.GetSubdue(); subdue != nil {
					isSubdued, err := sensutime.InWindows(time.Now(), *subdue)
					if err == nil && isSubdued {
						// Check is subdued at this time
						continue
					}
					if err != nil {
						s.logger.WithError(err).Print("check scheduler: subdued time window")
					}
				}

				// Reset timer
				timer.SetInterval(uint(check.Interval))
				timer.Next()

				// Point executor to lastest copy of the scheduler state
				executor.State = state

				// Publish check request
				if err := executor.Execute(check); err != nil {
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
	close(s.stopping)

	return nil
}

// CheckExecutor builds request & publishes
type CheckExecutor struct {
	Bus   messaging.MessageBus
	State *SchedulerState
}

// Execute queues reqest on message bus
func (execPtr *CheckExecutor) Execute(check *types.CheckConfig) error {
	// Ensure the check if configured to publish check requests
	if !check.Publish {
		return nil
	}

	var err error
	request := execPtr.BuildRequest(check)

	for _, sub := range check.Subscriptions {
		topic := messaging.SubscriptionTopic(check.Organization, check.Environment, sub)
		logger.Debugf("sending check request for %s on topic %s", check.Name, topic)

		if pubErr := execPtr.Bus.Publish(topic, request); err != nil {
			logger.Info("error publishing check request: ", err.Error())
			err = pubErr
		}
	}

	return err
}

// BuildRequest given check config fetches associated assets and builds request
func (execPtr *CheckExecutor) BuildRequest(check *types.CheckConfig) *types.CheckRequest {
	request := &types.CheckRequest{}
	request.Config = check

	// Guard against iterating over assets if there are no assets associated with
	// the check in the first place.
	if len(check.RuntimeAssets) != 0 {
		// Explode assets; get assets & filter out those that are irrelevant
		allAssets := execPtr.State.GetAssetsInOrg(check.Organization)
		for _, asset := range allAssets {
			if assetIsRelevant(asset, check) {
				request.Assets = append(request.Assets, *asset)
			}
		}
	}

	// Guard against iterating over hooks if there are no hooks associated with
	// the check in the first place.
	if len(check.CheckHooks) != 0 {
		// Explode hooks; get hooks & filter out those that are irrelevant
		allHooks := execPtr.State.GetHooksInOrg(check.Organization)
		for _, hook := range allHooks {
			if hookIsRelevant(hook, check) {
				request.Hooks = append(request.Hooks, *hook)
			}
		}
	}

	return request
}

func assetIsRelevant(asset *types.Asset, check *types.CheckConfig) bool {
	for _, assetName := range check.RuntimeAssets {
		if strings.HasPrefix(asset.Name, assetName) {
			return true
		}
	}

	return false
}

func hookIsRelevant(hook *types.HookConfig, check *types.CheckConfig) bool {
	for _, checkHook := range check.CheckHooks {
		for _, hookName := range checkHook.Hooks {
			if hookName == hook.Name {
				return true
			}
		}
	}

	return false
}
