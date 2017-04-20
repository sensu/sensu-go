package schedulerd

import (
	"context"
	"encoding/json"
	"errors"
	"sync"
	"time"

	"github.com/coreos/etcd/clientv3"
	"github.com/sensu/sensu-go/backend/messaging"
	"github.com/sensu/sensu-go/backend/store"
	"github.com/sensu/sensu-go/types"
)

// Schedulerd handles scheduling check requests for each check's
// configured interval and publishing to the message bus.
type Schedulerd struct {
	Client     *clientv3.Client
	Store      store.Store
	MessageBus messaging.MessageBus

	schedulers      map[string]*CheckScheduler
	schedulersMutex *sync.Mutex
	watcher         clientv3.Watcher
	wg              *sync.WaitGroup
	errChan         chan error
	stopping        chan struct{}
}

// Start the Scheduler daemon.
func (s *Schedulerd) Start() error {
	if s.Client == nil {
		return errors.New("no etcd client available")
	}

	if s.Store == nil {
		return errors.New("no store available")
	}

	if s.MessageBus == nil {
		return errors.New("no message bus found")
	}

	s.schedulers = map[string]*CheckScheduler{}
	s.schedulersMutex = &sync.Mutex{}

	s.errChan = make(chan error, 1)

	// The reconciler and the watchers have to be a little coordinated. We start
	// the watcher first, so that we don't miss any checks that are created
	// during our initial reconciliation phase.
	s.wg = &sync.WaitGroup{}
	s.startWatcher()

	s.stopping = make(chan struct{})
	s.reconcile()
	s.startReconciler()
	return nil
}

// Stop the scheduler daemon.
func (s *Schedulerd) Stop() error {
	close(s.stopping)
	s.watcher.Close()
	// let the event queue drain so that we don't panic inside the loop.
	// TODO(greg): get ride of this dependency.
	s.wg.Wait()
	close(s.errChan)
	return nil
}

// Status returns the health of the scheduler daemon.
func (s *Schedulerd) Status() error {
	return nil
}

// Err returns a channel on which to listen for terminal errors.
func (s *Schedulerd) Err() <-chan error {
	return s.errChan
}

func (s *Schedulerd) startReconciler() {
	s.wg.Add(1)

	go func() {
		ticker := time.NewTicker(30 * time.Second)
		for {
			select {
			case <-s.stopping:
				ticker.Stop()
				s.wg.Done()
				return
			case <-ticker.C:
				s.reconcile()
			}
		}
	}()
}

// All the watcher has to do is make sure that we have schedulers for any checks
// that are created. Once the scheduler is in place, it will self manage.
func (s *Schedulerd) startWatcher() {
	s.wg.Add(1)

	go func() {
		defer s.wg.Done()
		for {
			select {
			case <-s.stopping:
				return
			default:
				// TODO(grep): this should probably come from our own factory. have a
				// WatchFactory interface that takes a *clientv3.Client and returns a
				// clientv3.Watcher (interface). Then we can have the etcd factory and
				// the testing factory so we can do unit testing.
				s.watcher = clientv3.NewWatcher(s.Client)
			}
			for resp := range s.watcher.Watch(
				context.TODO(),
				"/sensu.io/checks",
				clientv3.WithPrefix(),
				clientv3.WithFilterDelete(),
				clientv3.WithFilterPut(),
				clientv3.WithCreatedNotify(),
			) {
				for _, ev := range resp.Events {
					s.schedulersMutex.Lock()
					check := &types.Check{}
					err := json.Unmarshal(ev.Kv.Value, check)
					if err != nil {
						logger.Errorf("error unmarshalling check \"%s\": %s", string(ev.Kv.Value), err.Error())
						s.schedulersMutex.Unlock()
						continue
					}
					scheduler := s.newSchedulerFromCheck(check)
					s.schedulers[check.Name] = scheduler
					err = scheduler.Start()
					if err != nil {
						logger.Error("error starting scheduler for check: ", check.Name)
						s.schedulersMutex.Unlock()
					}
					s.schedulersMutex.Unlock()
				}
			}
			// TODO(greg): exponential backoff
			time.Sleep(1 * time.Second)
		}
	}()
}

func (s *Schedulerd) reconcile() error {
	checks, err := s.Store.GetChecks()
	if err != nil {
		return err
	}

	for _, check := range checks {
		s.schedulersMutex.Lock()
		if _, ok := s.schedulers[check.Name]; !ok {
			scheduler := s.newSchedulerFromCheck(check)
			err = scheduler.Start()
			if err != nil {
				s.schedulersMutex.Unlock()
				return err
			}
			s.schedulers[check.Name] = scheduler
		}
		s.schedulersMutex.Unlock()
	}
	return nil
}

func (s *Schedulerd) newSchedulerFromCheck(check *types.Check) *CheckScheduler {
	scheduler := &CheckScheduler{
		MessageBus: s.MessageBus,
		Store:      s.Store,
		Check:      check,
		wg:         s.wg,
	}
	return scheduler
}
