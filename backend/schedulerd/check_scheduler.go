package schedulerd

import (
	"crypto/md5"
	"encoding/binary"
	"sync"
	"time"

	"github.com/sensu/sensu-go/backend/messaging"
	"github.com/sensu/sensu-go/backend/store"
	"github.com/sensu/sensu-go/types"
)

// A CheckScheduler schedules checks to be executed on a timer
type CheckScheduler struct {
	CheckName string
	CheckOrg  string

	Cache      *CheckCache
	MessageBus messaging.MessageBus
	WaitGroup  *sync.WaitGroup

	stopping chan struct{}
}

// Start scheduler, ...
func (s *CheckScheduler) Start(initialInterval int64) error {
	s.stopping = make(chan struct{})
	s.WaitGroup.Add(1)

	timer := NewCheckTimer(s.CheckName, initialInterval)

	// TODO(greg): Refactor this part to make the code more easily tested.
	go func() {
		timer.Start()
		defer timer.Stop()
		defer s.WaitGroup.Done()

		for {
			select {
			case <-timer.C:
				request := s.buildRequest()
				if request == nil {
					// The check has been deleted, and there was no error talking to etcd.
					close(s.stopping)
					return
				}

				timer.SetInterval(request.Config.Interval)
				timer.Next()

				for _, sub := range request.Config.Subscriptions {
					topic := messaging.SubscriptionTopic(s.CheckOrg, sub)
					logger.Debugf("Sending check request for %s on topic %s", s.CheckName, topic)
					if err := s.MessageBus.Publish(topic, request); err != nil {
						logger.Info("error publishing check request: ", err.Error())
					}
				}
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

func (s *CheckScheduler) buildRequest() (*types.CheckRequest, error) {
	cache := s.Cache
	request := &types.CheckRequest{}

	// Get Check
	check := cache.GetCheck(s.CheckName, s.CheckOrg)
	request.Check = check

	if check == nil {
		return nil
	}

	// Guard against iterating over assets if there are no assets associated with
	// the check in the first place.
	if len(check.RuntimeAssets) == 0 {
		return request
	}

	// Get Assets & filter out those that are irrelevant
	allAssets := cache.GetAssetsInOrg(s.CheckOrg)
	for _, asset := range allAssets {
		if assetIsRelevant(asset, check) {
			request.Assets = append(request.Assets, asset)
		}
	}

	return request
}

func assetIsRelevant(asset *types.Asset, check *types.CheckConfig) bool {
	for _, assetName := range check.RuntimeAssets {
		return strings.HasPrefix(asset.Name, assetName)
	}
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
func (timerPtr *CheckTimer) C() <-chan Time {
	return timerPtr.timer.C
}

// SetInterval updates the interval in which timers are set
func (timerPtr *CheckTimer) SetInterval(i int) {
	timerPrt.interval = time.Duration(time.Second * time.Duration(i))
}

// Start sets up a new timer
func (timerPtr *CheckTimer) Start() {
	// Calculate the first execution time using splay & interval
	now := time.Now().UnixNano() / int64(time.Millisecond)
	offset := (splay - uint64(now)) % uint64(timerPtr.interval)

	// Initialize new timer w/ initial exec time
	timerPtr.timer = time.NewTimer(time.Duration(offset) * time.Millisecond)
}

// Next reset's timer using interval
func (timerPtr *CheckTimer) Next() {
	timerPtr.timer.Reset(timer.interval)
}

// Stop ends the timer
func (timerPtr *CheckTimer) Stop() {
	timerPtr.timer.Stop()
}
