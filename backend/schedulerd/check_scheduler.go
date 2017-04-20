package schedulerd

import (
	"crypto/md5"
	"encoding/binary"
	"encoding/json"
	"errors"
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

	splayHash, err := calcExecutionSplay(s.Check)
	if err != nil {
		return err
	}

	s.wg.Add(1)
	go func() {
		now := uint64(time.Now().UnixNano())
		checkInterval := time.Duration(s.Check.Interval) * time.Second
		nextExecution := time.Duration(splayHash-now) % checkInterval
		timer := time.NewTimer(time.Duration(nextExecution))

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
					evtBytes, err := json.Marshal(evt)
					if err != nil {
						logger.Info("error marshalling check in scheduler: ", err.Error())
						continue
					}

					if err := s.MessageBus.Publish(sub, evtBytes); err != nil {
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
func calcExecutionSplay(c *types.Check) (uint64, error) {
	sum := md5.Sum([]byte(c.Name))
	splayHash, n := binary.Uvarint(sum[0:7])
	if n < 0 {
		return 0, errors.New("check hashing failed")
	}

	return splayHash, nil
}
