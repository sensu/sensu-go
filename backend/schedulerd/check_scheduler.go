package schedulerd

import (
	"crypto/md5"
	"encoding/binary"
	"strings"
	"sync"
	"time"

	"github.com/sensu/sensu-go/backend/messaging"
	"github.com/sensu/sensu-go/types"
)

// A CheckScheduler schedules checks to be executed on a timer
type CheckScheduler struct {
	CheckName string
	CheckOrg  string

	StateManager *StateManager
	MessageBus   messaging.MessageBus
	WaitGroup    *sync.WaitGroup

	stopping chan struct{}
}

// Start scheduler, ...
func (s *CheckScheduler) Start(initialInterval int) error {
	s.stopping = make(chan struct{})
	s.WaitGroup.Add(1)

	timer := NewCheckTimer(s.CheckName, initialInterval)
	executor := &CheckExecutor{Bus: s.MessageBus}

	// TODO(greg): Refactor this part to make the code more easily tested.
	go func() {
		timer.Start()
		defer timer.Stop()
		defer s.WaitGroup.Done()

		for {
			select {
			case <-timer.C():
				// Fetch check from scheduler's state
				state := s.StateManager.State()
				check := state.GetCheck(s.CheckName, s.CheckOrg)

				// The check has been deleted
				if check == nil {
					close(s.stopping)
					return
				}

				// Reset timer
				timer.SetInterval(check.Interval)
				timer.Next()

				// Point executor to lastest copy of the scheduler state
				executor.State = &state

				// Publish check request
				executor.Execute(check)
			case <-s.stopping:
				return
			}
		}
	}()

	return nil
}

// Stop stops the CheckScheduler
func (s *CheckScheduler) Stop() error {
	close(s.stopping)
	s.WaitGroup.Wait()
	return nil
}

// CheckExecutor builds request & publishes
type CheckExecutor struct {
	Bus   messaging.MessageBus
	State *SchedulerState
}

// Execute queues reqest on message bus
func (execPtr *CheckExecutor) Execute(check *types.CheckConfig) error {
	var err error
	request := execPtr.BuildRequest(check)

	for _, sub := range check.Subscriptions {
		topic := messaging.SubscriptionTopic(check.Organization, sub)
		logger.Debugf("Sending check request for %s on topic %s", check.Name, topic)

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
	if len(check.RuntimeAssets) == 0 {
		return request
	}

	// Explode assets; get assets & filter out those that are irrelevant
	allAssets := execPtr.State.GetAssetsInOrg(check.Organization)
	for _, asset := range allAssets {
		if assetIsRelevant(asset, check) {
			request.ExpandedAssets = append(request.ExpandedAssets, *asset)
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

// A CheckTimer handles starting a stopping timers for a given check
type CheckTimer struct {
	interval time.Duration
	splay    uint64
	timer    *time.Timer
}

// NewCheckTimer establishes new check timer given a name & an initial interval
func NewCheckTimer(name string, interval int) *CheckTimer {
	// Calculate a check execution splay to ensure
	// execution is consistent between process restarts.
	sum := md5.Sum([]byte(name))
	splay := binary.LittleEndian.Uint64(sum[:])

	timer := &CheckTimer{splay: splay}
	timer.SetInterval(interval)
	return timer
}

// C channel emits events when timer's duration has reaached 0
func (timerPtr *CheckTimer) C() <-chan time.Time {
	return timerPtr.timer.C
}

// SetInterval updates the interval in which timers are set
func (timerPtr *CheckTimer) SetInterval(i int) {
	timerPtr.interval = time.Duration(time.Second * time.Duration(i))
}

// Start sets up a new timer
func (timerPtr *CheckTimer) Start() {
	initOffset := timerPtr.calcInitialOffset()
	timerPtr.timer = time.NewTimer(initOffset)
}

// Next reset's timer using interval
func (timerPtr *CheckTimer) Next() {
	timerPtr.timer.Reset(timerPtr.interval)
}

// Stop ends the timer
func (timerPtr *CheckTimer) Stop() bool {
	return timerPtr.timer.Stop()
}

// Calculate the first execution time using splay & interval
func (timerPtr *CheckTimer) calcInitialOffset() time.Duration {
	now := uint64(time.Now().UnixNano())
	offset := (timerPtr.splay - now) % uint64(timerPtr.interval)
	return time.Duration(offset) / time.Nanosecond
}
