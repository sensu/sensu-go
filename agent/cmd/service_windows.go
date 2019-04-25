package cmd

import (
	"context"
	"fmt"
	"log"

	"golang.org/x/sys/windows/svc"
	"golang.org/x/sys/windows/svc/debug"
	"golang.org/x/sys/windows/svc/eventlog"
)

var (
	_    svc.Handler = &Service{}
	elog debug.Log
)

func NewService() *Service {
	ctx, cancel := context.WithCancel(context.Background())
	return &Service{
		ctx:    ctx,
		cancel: cancel,
	}
}

type Service struct {
	ctx    context.Context
	cancel context.CancelFunc
}

func (s *Service) Execute(args []string, r <-chan svc.ChangeRequest, changes chan<- svc.Status) (bool, uint32) {
	go func() {
		defer func() {
			changes <- svc.Status{State: svc.Stopped}
		}()
		defer s.cancel()
		for req := range r {
			switch req.Cmd {
			case svc.Stop, svc.Shutdown:
				changes <- svc.Status{State: svc.StopPending}
				return
			default:
				// TODO log this to the appropriate place? or?
				elog.Error(1, fmt.Sprintf("got change request: %v", req))
			}
		}
	}()
	changes <- svc.Status{State: svc.StartPending}
	// Start service here
	command := newStartCommand(args, s.ctx)
	changes <- svc.Status{State: svc.Running, Accepts: svc.AcceptShutdown}

	if err := command.Execute(); err != nil {
		// TODO figure out how best to handle this
		log.Println(err)
		return false, 1
	}
	// Block until shutdown
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
