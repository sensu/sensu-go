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

// CheckScheduler TODO
type CheckScheduler struct {
	MessageBus messaging.MessageBus
	Store      store.Store
	Check      *types.Check

	wg       *sync.WaitGroup
	stopping chan struct{}
}

// Start scheduler, ...
func (s *CheckScheduler) Start() error {
	s.stopping = make(chan struct{})

	splayHash := calcExecutionSplay(s.Check.Name)

	s.wg.Add(1)
	// TODO(greg): Refactor this part to make the code more easily tested.
	go func() {
		nextExecution := calcNextExecution(splayHash, s.Check.Interval)
		timer := time.NewTimer(nextExecution)

		defer s.wg.Done()
		for {
			select {
			case <-timer.C:
				check, err := s.Store.GetCheckByName(s.Check.Name)
				if err != nil {
					logger.Info("error getting check from store: ", err.Error())
					// TODO(grep): what do we do when we cannot talk to the store?
					continue
				}

				if check == nil {
					// The check has been deleted, and there was no error talking to etcd.
					timer.Stop()
					close(s.stopping)
					return
				}

				// update our pointer to the check
				s.Check = check

				timer.Reset(time.Duration(time.Second * time.Duration(s.Check.Interval)))
				for _, sub := range s.Check.Subscriptions {
					evt := &types.Event{
						Timestamp: time.Now().Unix(),
						Check:     s.Check,
					}

					if err := s.MessageBus.Publish(sub, evt); err != nil {
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
	return nil
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
