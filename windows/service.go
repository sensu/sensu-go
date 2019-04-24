//+build windows

package windows

import (
	"fmt"

	"github.com/spf13/cobra"
	"golang.org/x/sys/windows/svc"
	"golang.org/x/sys/windows/svc/debug"
	"golang.org/x/sys/windows/svc/eventlog"
)

var (
	_    svc.Handler = &Service{}
	elog debug.Log
)

func NewService(command *cobra.Command) *Service {
	return &Service{
		command: command,
	}
}

type Service struct {
	command *cobra.Command
}

func (s *Service) Execute(args []string, r <-chan svc.ChangeRequest, changes chan<- svc.Status) (bool, uint32) {
	changes <- svc.Status{State: svc.StartPending}
	// Start service here
	changes <- svc.Status{State: svc.Running, Accepts: svc.AcceptShutdown}
	// Block until shutdown
	// Service has been given stop signal
	changes <- svc.Status{State: svc.StopPending}
	// Service has stopped
	changes <- svc.Status{State: svc.Stopped}
	return false, 0
}

func runService(name string, isDebug bool) {
	var err error
	if isDebug {
		elog = debug.New(name)
	} else {
		elog, err = eventlog.Open(name)
		if err != nil {
			return
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
		elog.Error(1, fmt.Sprintf("%s service failed: %v", name, err))
		return
	}
	elog.Info(1, fmt.Sprintf("%s service stopped", name))
}
