package cmd

import (
	"context"
	"errors"
	"os"
	"os/signal"
	runtimedebug "runtime/debug"
	"sync"
	"syscall"

	"github.com/sensu/sensu-go/agent"
	"golang.org/x/sys/windows/svc"
)

var (
	AgentNewFunc = agent.NewAgentContext
)

func NewService(cfg *agent.Config) *Service {
	return &Service{cfg: cfg}
}

type Service struct {
	cfg *agent.Config
	wg  sync.WaitGroup
}

func (s *Service) start(ctx context.Context, cancel context.CancelFunc, changes chan<- svc.Status) chan error {
	s.wg.Add(1)
	result := make(chan error, 1)
	defer func() {
		if e := recover(); e != nil {
			changes <- svc.Status{State: svc.Stopped}
			stack := runtimedebug.Stack()
			result <- errors.New(string(stack))
		}
	}()
	changes <- svc.Status{State: svc.StartPending}
	accepts := svc.AcceptShutdown | svc.AcceptStop
	changes <- svc.Status{State: svc.Running, Accepts: accepts}

	sensuAgent, err := agent.NewAgentContext(ctx, s.cfg)
	if err != nil {
		result <- err
		return result
	}

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		defer cancel()
		logger.Info("signal received: ", <-sigs)
	}()

	go func() {
		defer s.wg.Done()
		if err := sensuAgent.Run(ctx); err != nil {
			result <- err
		}
	}()
	return result
}

func (s *Service) Execute(_ []string, r <-chan svc.ChangeRequest, changes chan<- svc.Status) (bool, uint32) {
	ctx, cancel := context.WithCancel(context.Background())
	logger.Info("sensu-agent service starting")
	errs := s.start(ctx, cancel, changes)
	for {
		select {
		case req := <-r:
			switch req.Cmd {
			case svc.Stop, svc.Shutdown:
				logger.Info("sensu-agent shutting down")
				changes <- svc.Status{State: svc.StopPending}
				cancel()
				s.wg.Wait()
				changes <- svc.Status{State: svc.Stopped}
				return false, 0
			}
		case err := <-errs:
			logger.WithError(err).Error("fatal error")
			return false, 1
		}
	}
	return false, 0
}
