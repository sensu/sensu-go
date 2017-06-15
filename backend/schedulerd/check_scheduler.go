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

	splayHash := calcExecutionSplay(s.CheckName)

	// TODO(greg): Refactor this part to make the code more easily tested.
	go func() {
		nextExecution := calcNextExecution(splayHash, initialInterval)
		timer := time.NewTimer(nextExecution)

		defer s.WaitGroup.Done()
		for {
			select {
			case <-timer.C:
				request := s.buildRequest()
				if request == nil {
					// The check has been deleted, and there was no error talking to etcd.
					timer.Stop()
					close(s.stopping)
					return
				}

				timer.Reset(time.Duration(time.Second * time.Duration(checkConfig.Interval)))
				for _, sub := range request.Config.Subscriptions {
					topic := messaging.SubscriptionTopic(s.CheckOrg, sub)
					logger.Debugf("Sending check request for %s on topic %s", s.CheckName, topic)
					if err := s.MessageBus.Publish(topic, request); err != nil {
						logger.Info("error publishing check request: ", err.Error())
					}
				}
			case <-s.stopping:
				timer.Stop()
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

// Calculate a check execution splay to ensure
// execution is consistent between process restarts.
func calcExecutionSplay(checkName string) uint64 {
	sum := md5.Sum([]byte(checkName))

	return binary.LittleEndian.Uint64(sum[:])
}

// Calculate the next execution time for a given time and a check interval
// (in seconds) as an int.
func calcNextExecution(splay uint64, intervalSeconds int) time.Duration {
	// current_time = (Time.now.to_f * 1000).to_i
	now := time.Now().UnixNano() / int64(time.Millisecond)
	offset := (splay - uint64(now)) % uint64(intervalSeconds*1000)
	return time.Duration(offset) * time.Millisecond
}
