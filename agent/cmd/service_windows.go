package cmd

import (
	"context"
	"fmt"
	"log"
	"sync"

	runtimedebug "runtime/debug"

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
				runtimedebug.PrintStack()
				result <- fmt.Errorf("%v", e)
			}
		}()
		defer s.wg.Done()
		changes <- svc.Status{State: svc.StartPending}
		// Start service here
		args = []string{args[0], "start", "-c", args[len(args)-1]}
		command := newStartCommand(ctx, args)
		accepts := svc.AcceptShutdown | svc.AcceptStop
		changes <- svc.Status{State: svc.Running, Accepts: accepts}

		if err := command.Execute(); err != nil {
			result <- err
		}
	}()
	return result
}

func (s *Service) Execute(args []string, r <-chan svc.ChangeRequest, changes chan<- svc.Status) (bool, uint32) {
	ctx, cancel := context.WithCancel(context.Background())
	errs := s.start(ctx, args, changes)
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
			log.Printf("restarting due to error: %s", err)
			s.start(ctx, args, changes)
		}
	}
	return false, 0
}

func runService(name string, isDebug bool) error {
	var err error
	if isDebug {
		elog = debug.New(name)
	} else {
		elog, err = eventlog.Open(name)
		if err != nil {
			return err
		}
	}
	defer elog.Close()
	elog.Info(1, fmt.Sprintf("starting %s service", name))
	run := svc.Run
	if isDebug {
		run = debug.Run
	}
	err = run(name, NewService())
	if err != nil {
		return err
	}
	elog.Info(1, fmt.Sprintf("%s service stopped", name))
	return nil
}
