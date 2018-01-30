package leader

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/coreos/etcd/clientv3/concurrency"
	"github.com/google/uuid"
	"github.com/sensu/sensu-go/backend/store"
)

var super *supervisor

var keyBuilder = store.NewKeyBuilder(sensuLeaderKey)

type supervisor struct {
	session       *concurrency.Session
	election      *concurrency.Election
	isLeader      chan struct{}
	isFollower    chan struct{}
	work          chan *work
	cancel        context.CancelFunc
	nodeName      string
	logger        *logrus.Entry
	leaderName    atomic.Value
	workPerformed int64
	wg            sync.WaitGroup
	workInFlight  sync.WaitGroup
}

func newSupervisor(session *concurrency.Session) *supervisor {
	nodeName := fmt.Sprintf("sensu-backend-%s", uuid.New().String())
	s := &supervisor{
		session:    session,
		election:   concurrency.NewElection(session, keyBuilder.Build()),
		isLeader:   make(chan struct{}),
		isFollower: make(chan struct{}),
		work:       make(chan *work),
		nodeName:   nodeName,
		logger: logger.WithFields(logrus.Fields{
			"node_name": nodeName,
		}),
	}
	return s
}

// Start starts the supervisor.
func (s *supervisor) Start() {
	s.logger.Debug("campaigning for leadership")
	s.wg.Add(1)
	s.leaderName.Store("")
	ctx, cancel := context.WithCancel(context.Background())
	s.cancel = cancel

	go s.campaign(ctx)
	go s.observer(ctx)
	go s.worker(ctx)
	go s.periodicLogger(ctx)
}

func (s *supervisor) periodicLogger(ctx context.Context) {
	s.logger.Debug("starting up logger")
	defer s.logger.Debug("shutting down logger")
	ticker := time.NewTicker(getLogInterval())
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			logPeriodic(
				s.nodeName,
				s.leaderName.Load().(string),
				atomic.LoadInt64(&s.workPerformed))
		case <-ctx.Done():
			return
		}
	}
}

func (s *supervisor) worker(ctx context.Context) {
	s.logger.Debug("starting up worker")
	defer s.logger.Debug("shutting down worker")
	for {
		select {
		case <-ctx.Done():
			return
		case <-s.isLeader:
			s.processWork(ctx)
		}
	}
}

func (s *supervisor) processWork(ctx context.Context) {
	s.logger.Debug("starting up work processor")
	defer s.logger.Debug("shutting down work processor")
	workCtx, cancel := context.WithCancel(context.Background())
	defer cancel()
	for {
		select {
		case <-s.isFollower:
			s.logger.Warn("lost cluster leadership")
			// Cancel any work that's in-flight. The work is responsible for
			// terminating itself by waiting for its context to cancel.
			return
		case <-ctx.Done():
			return
		case work := <-s.work:
			s.workInFlight.Add(1)
			s.logger.Debugf("starting work (ID %d)", work.id)
			go func() {
				defer s.logger.Debugf("finished work (ID %d)", work.id)
				defer atomic.AddInt64(&s.workPerformed, 1)
				defer s.workInFlight.Done()
				work.result <- work.f(workCtx)
			}()
		}
	}
}

func (s *supervisor) campaign(ctx context.Context) {
	s.logger.Debug("starting up campaign")
	defer s.logger.Debug("finished campaign")
	if err := s.election.Campaign(ctx, s.nodeName); err != nil {
		if err != ctx.Err() {
			s.logger.WithError(err).Error("error running campaign")
		}
		return
	}
	// The Campaign method blocks until the node is elected
	s.logger.Info("gained cluster leadership")
	s.leaderName.Store(s.nodeName)
	s.wg.Done()
	// Ensure that all work from previous leadership round has completed first.
	s.workInFlight.Wait()
	s.isLeader <- struct{}{}
}

func (s *supervisor) observer(ctx context.Context) {
	s.logger.Debug("starting up election observer")
	defer s.logger.Debug("shutting down election observer")
	for {
		select {
		case <-ctx.Done():
			return
		default:
		}
		observer := s.election.Observe(ctx)
		for response := range observer {
			wasLeading := s.IsLeader()
			leaderName := string(response.Kvs[0].Value)
			s.leaderName.Store(leaderName)
			if leaderName != s.nodeName && wasLeading {
				s.isFollower <- struct{}{}
				s.wg.Add(1)
				s.campaign(ctx)
			}
		}
	}
}

// Stop stops the supervisor. It blocks until the election has been resigned,
// and all work has been completed or cancelled.
func (s *supervisor) Stop() error {
	s.logger.Info("resigning from leadership or election")
	s.cancel()
	s.workInFlight.Wait()
	return s.election.Resign(context.Background())
}

// Exec sends the work its given to the work channel.
func (s *supervisor) Exec(w *work) {
	s.work <- w
}

// WaitLeader blocks until the supervisor has attained leadership.
// This method is meant for testing purposes.
func (s *supervisor) WaitLeader() {
	s.wg.Wait()
}

// IsLeader returns true if the supervisor is the leader. This method is meant
// for testing purposes.
func (s *supervisor) IsLeader() bool {
	return s.leaderName.Load().(string) == s.nodeName
}
