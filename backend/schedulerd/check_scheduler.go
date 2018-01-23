package schedulerd

import (
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/sensu/sensu-go/agent"
	"github.com/sensu/sensu-go/backend/messaging"
	"github.com/sensu/sensu-go/types"
	"github.com/sensu/sensu-go/types/dynamic"
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

	logger   *logrus.Entry
	stopping chan struct{}
}

// Start scheduler, ...
func (s *CheckScheduler) Start() error {
	s.stopping = make(chan struct{})
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
		executor := &CheckExecutor{Bus: s.MessageBus}

		// TODO(greg): Refactor this part to make the code more easily tested.
		timer.Start()
		defer timer.Stop()

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
	// Ensure the check is configured to publish check requests
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

// PublishProxyCheckRequest publishes a proxy check request for an entity. This
// method substitutes entity tokens in the check definition prior to publishling
// the check request. If there are unmatched entity tokens, it returns an error,
// and a check request is not published.
func (execPtr *CheckExecutor) PublishProxyCheckRequest(entity *types.Entity, check *types.CheckConfig) error {
	// Extract the extended attributes from the entity and combine them at the
	// top-level so they can be easily accessed using token substitution
	synthesizedEntity, err := dynamic.Synthesize(entity)
	if err != nil {
		return fmt.Errorf("could not synthesize the entity: %s", err)
	}

	// Substitute tokens within the check configuration with the synthesized
	// entity
	checkBytes, err := agent.TokenSubstitution(synthesizedEntity, check)
	if err != nil {
		return err
	}

	// Unmarshal the check configuration obtained after the token substitution
	// back into the check config struct
	err = json.Unmarshal(checkBytes, check)
	if err != nil {
		return fmt.Errorf("could not unmarshal the check: %s", err)
	}

	check.ProxyEntityID = entity.ID
	return execPtr.Execute(check)
}

// PublishProxyCheckRequests publishes proxy check requests for one or more
// entities. This method can optionally splay proxy check requests, evenly, over
// a period of time, determined by the check interval and a configurable splay
// coverage percentage. For example, splay proxy check requests over 60s * 90%,
// 54s, leaving 6s for the last proxy check execution before the the next round
// of proxy check requests for the same check. The PublishProxyCheckRequest
// method is used to publish the proxy check requests.
func (execPtr *CheckExecutor) PublishProxyCheckRequests(entities []*types.Entity, check *types.CheckConfig) error {
	var err error
	splay := float64(0)
	numEntities := float64(len(entities))
	if check.ProxyRequests.Splay {
		next := time.Duration(time.Second * time.Duration(check.Interval))
		if check.Cron != "" {
			if next, err = NextCronTime(check.Cron); err != nil {
				return err
			}
		}
		splayCoverage := float64(check.ProxyRequests.SplayCoverage)
		if splayCoverage == 0 {
			splayCoverage = types.DefaultSplayCoverage
		}
		splay = next.Seconds() * (splayCoverage / 100.0) / numEntities
	}

	for _, entity := range entities {
		time.Sleep(time.Duration(time.Millisecond * time.Duration(splay*1000)))
		if err := execPtr.PublishProxyCheckRequest(entity, check); err != nil {
			return err
		}
	}
	return nil
}
