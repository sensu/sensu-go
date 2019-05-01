package cmd

import (
	"context"
	"errors"
	"fmt"
	"os"
	"sync"

	runtimedebug "runtime/debug"

	"github.com/sirupsen/logrus"
	"golang.org/x/sys/windows/svc"
	"golang.org/x/sys/windows/svc/debug"
	"golang.org/x/sys/windows/svc/eventlog"
)

var (
	_    svc.Handler = &Service{}
	elog debug.Log
)

func NewService() *Service {
	return &Service{}
}

type Service struct {
	wg sync.WaitGroup
	mu sync.Mutex
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
		configFile := args[len(args)-2]
		logPath := args[len(args)-1]
		logFile, err := os.OpenFile(logPath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0600)
		if err != nil {
			result <- fmt.Errorf("service quit: cant't open log file: %s", err)
		}
		defer logFile.Close()

		logrus.SetFormatter(&logrus.JSONFormatter{})
		logger := logrus.New()
		logger.SetOutput(logFile)
		entry := logger.WithFields(logrus.Fields{
			"component": "cmd",
		})

		args = []string{args[0], "start", "-c", configFile}
		command := newStartCommand(ctx, args, entry)
		accepts := svc.AcceptShutdown | svc.AcceptStop
		changes <- svc.Status{State: svc.Running, Accepts: accepts}

		if err := command.Execute(); err != nil {
			logger.WithError(err).Error("sensu-agent exited with error")
			result <- err
		}
	}()
	return result
}

func (s *Service) Execute(args []string, r <-chan svc.ChangeRequest, changes chan<- svc.Status) (bool, uint32) {
	ctx, cancel := context.WithCancel(context.Background())
	errs := s.start(ctx, args, changes)
	elog, _ := eventlog.Open(serviceName)
	defer elog.Close()
	for {
		select {
		case req := <-r:
			switch req.Cmd {
			case svc.Stop, svc.Pause:
				changes <- svc.Status{State: svc.StopPending}
				cancel()
				s.wg.Wait()
				changes <- svc.Status{State: svc.Stopped}
			case svc.Shutdown:
				cancel()
				s.wg.Wait()
				return false, 0
			case svc.Continue:
				s.wg.Wait()
				ctx, cancel = context.WithCancel(context.Background())
				errs = s.start(ctx, args, changes)
			}
		case err := <-errs:
			elog.Error(1, fmt.Sprintf("restarting due to error: %s", err))
			s.start(ctx, args, changes)
		}
	}
	return false, 0
}

func runService() error {
	elog, err := eventlog.Open(serviceName)
	if err != nil {
		return err
	}
	defer elog.Close()
	elog.Info(1, fmt.Sprintf("starting %s service", serviceName))
	if err := svc.Run(serviceName, NewService()); err != nil {
		return err
	}
	elog.Info(1, fmt.Sprintf("%s service stopped", serviceName))
	return nil
}
