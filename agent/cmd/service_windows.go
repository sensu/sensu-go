package cmd

import (
	"context"
	"errors"
	"fmt"
	"os"
	"sync"

	runtimedebug "runtime/debug"

	"github.com/sensu/sensu-go/agent"
	"github.com/sirupsen/logrus"
	"golang.org/x/sys/windows/svc"
	"golang.org/x/sys/windows/svc/debug"
	"golang.org/x/sys/windows/svc/eventlog"
)

var (
	_            svc.Handler = &Service{}
	elog         debug.Log
	AgentNewFunc = agent.NewAgentContext
)

func NewService(args []string) *Service {
	return &Service{args: args}
}

type Service struct {
	args []string
	wg   sync.WaitGroup
	mu   sync.Mutex
}

func (s *Service) start(ctx context.Context, args []string, changes chan<- svc.Status) chan error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.wg.Wait()
	s.wg.Add(1)
	result := make(chan error, 1)
	go func() {
		defer func() {
			if e := recover(); e != nil {
				changes <- svc.Status{State: svc.Stopped}
				stack := runtimedebug.Stack()
				result <- errors.New(string(stack))
			}
		}()
		defer s.wg.Done()
		changes <- svc.Status{State: svc.StartPending}
		// Start service here
		binPath, err := exePath()
		if err != nil {
			panic(err)
		}
		configFile := args[0]
		logPath := args[1]
		logFile, err := os.OpenFile(logPath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0600)
		if err != nil {
			result <- fmt.Errorf("service quit: cant't open log file: %s", err)
		}
		defer logFile.Close()

		logrus.SetFormatter(&logrus.JSONFormatter{})
		logrus.SetOutput(logFile)
		logger := logrus.WithFields(logrus.Fields{
			"component": "cmd",
		})

		args = []string{binPath, "start", "-c", configFile}
		command := StartCommand(AgentNewFunc)
		accepts := svc.AcceptShutdown | svc.AcceptStop
		changes <- svc.Status{State: svc.Running, Accepts: accepts}

		if err := command.Execute(); err != nil {
			logger.WithError(err).Error("sensu-agent exited with error")
			result <- err
		}
	}()
	return result
}

func (s *Service) Execute(_ []string, r <-chan svc.ChangeRequest, changes chan<- svc.Status) (bool, uint32) {
	ctx, cancel := context.WithCancel(context.Background())
	errs := s.start(ctx, s.args, changes)
	elog, _ := eventlog.Open(serviceName)
	defer elog.Close()
	for {
		select {
		case req := <-r:
			switch req.Cmd {
			case svc.Stop, svc.Shutdown:
				elog.Info(1, "service shutting down")
				changes <- svc.Status{State: svc.StopPending}
				cancel()
				s.wg.Wait()
				changes <- svc.Status{State: svc.Stopped}
				return false, 0
			}
		case err := <-errs:
			elog.Error(1, fmt.Sprintf("restarting due to error (%v) %s", s.args, err))
			s.start(ctx, s.args, changes)
		}
	}
	return false, 0
}

func runService(args []string) error {
	elog, err := eventlog.Open(serviceName)
	if err != nil {
		return err
	}
	defer elog.Close()
	elog.Info(1, fmt.Sprintf("starting %s service (%v)", serviceName, args))
	if err := svc.Run(serviceName, NewService(args)); err != nil {
		return err
	}
	elog.Info(1, fmt.Sprintf("%s service terminated", serviceName))
	return nil
}
